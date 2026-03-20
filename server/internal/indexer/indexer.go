package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/mansi/chainsight/server/internal/eth"
	"github.com/mansi/chainsight/server/internal/store"
)

type Config struct {
	InitialLookback  uint64
	MaxBlocksPerTick int
}

type Indexer struct {
	ethClient *eth.Client
	repo      *store.Repository
	config    Config

	lastIndexed uint64
}

func New(ethClient *eth.Client, repo *store.Repository, config Config) *Indexer {
	if config.InitialLookback == 0 {
		config.InitialLookback = 20
	}
	if config.MaxBlocksPerTick <= 0 {
		config.MaxBlocksPerTick = 2
	}

	return &Indexer{
		ethClient: ethClient,
		repo:      repo,
		config:    config,
	}
}

func (i *Indexer) RunOnce(ctx context.Context) error {
	if i.repo == nil {
		log.Printf("[indexer] repository is nil; indexer disabled")
		return nil
	}

	if err := i.initializeCursor(ctx); err != nil {
		return err
	}

	if err := i.indexOnce(ctx); err != nil {
		return err
	}

	return nil
}

func (i *Indexer) Run(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 8 * time.Second
	}

	if err := i.RunOnce(ctx); err != nil {
		log.Printf("[indexer] initial sync failed: %v", err)
	} else {
		log.Printf("[indexer] initial sync completed")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[indexer] stopped: %v", ctx.Err())
			return
		case <-ticker.C:
			if err := i.RunOnce(ctx); err != nil {
				log.Printf("[indexer] tick sync failed: %v", err)
			}
		}
	}
}

func (i *Indexer) initializeCursor(ctx context.Context) error {
	latestIndexed, err := i.repo.GetLatestIndexedBlockNumber()
	if err != nil {
		return err
	}

	if latestIndexed > 0 {
		i.lastIndexed = latestIndexed
		log.Printf("[indexer] resume from indexed block: %d", i.lastIndexed)
		return nil
	}

	var latest *eth.Block
	if err := i.retryRPC(ctx, "eth_getBlockByNumber(latest)", func() error {
		var rpcErr error
		latest, rpcErr = i.ethClient.GetLatestBlock(ctx)
		return rpcErr
	}); err != nil {
		return err
	}

	if latest.Number > i.config.InitialLookback {
		i.lastIndexed = latest.Number - i.config.InitialLookback
	} else {
		i.lastIndexed = 0
	}

	log.Printf("[indexer] bootstrap cursor set to block: %d", i.lastIndexed)
	return nil
}

func (i *Indexer) indexOnce(ctx context.Context) error {
	var latest *eth.Block
	if err := i.retryRPC(ctx, "eth_getBlockByNumber(latest)", func() error {
		var rpcErr error
		latest, rpcErr = i.ethClient.GetLatestBlock(ctx)
		return rpcErr
	}); err != nil {
		return err
	}

	start := i.lastIndexed + 1
	if start == 0 {
		start = latest.Number
	}
	if start > latest.Number {
		return nil
	}

	end := latest.Number
	maxEnd := start + uint64(i.config.MaxBlocksPerTick) - 1
	if end > maxEnd {
		end = maxEnd
	}

	for number := start; number <= end; number++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := i.indexBlock(ctx, number); err != nil {
			return err
		}
		i.lastIndexed = number

		// Small pause between blocks to avoid bursty request spikes.
		time.Sleep(250 * time.Millisecond)
	}

	if end < latest.Number {
		log.Printf("[indexer] progress: indexed through %d (chain tip %d)", end, latest.Number)
	}

	return nil
}

func (i *Indexer) indexBlock(ctx context.Context, number uint64) error {
	var (
		block *eth.Block
		txs   []eth.Transaction
	)
	if err := i.retryRPC(ctx, fmt.Sprintf("eth_getBlockByNumber(%d, true)", number), func() error {
		var rpcErr error
		block, txs, rpcErr = i.ethClient.GetBlockWithTransactions(ctx, number)
		return rpcErr
	}); err != nil {
		return err
	}

	blockModel := store.Block{
		Number:       block.Number,
		Hash:         strings.ToLower(block.Hash),
		ParentHash:   strings.ToLower(block.ParentHash),
		TimestampUTC: time.Unix(int64(block.Timestamp), 0).UTC(),
		TxCount:      len(block.Transactions),
	}
	if err := i.repo.UpsertBlock(blockModel); err != nil {
		return err
	}

	storeTxs := make([]store.Transaction, 0, len(txs))
	for _, tx := range txs {
		var toAddr *string
		if strings.TrimSpace(tx.To) != "" {
			to := strings.ToLower(strings.TrimSpace(tx.To))
			toAddr = &to
		}

		storeTxs = append(storeTxs, store.Transaction{
			Hash:        strings.ToLower(strings.TrimSpace(tx.Hash)),
			BlockNumber: number,
			FromAddress: strings.ToLower(strings.TrimSpace(tx.From)),
			ToAddress:   toAddr,
			ValueWei:    tx.Value,
		})
	}

	if err := i.repo.UpsertTransactions(storeTxs); err != nil {
		return err
	}

	if err := i.repo.UpsertAddressSummariesFromBlock(number); err != nil {
		return err
	}

	metric, err := i.buildMetricForBlock(block, txs)
	if err != nil {
		return err
	}

	if err := i.repo.UpsertNetworkMetric(metric); err != nil {
		return err
	}

	log.Printf("[indexer] indexed block %d (%d tx)", number, len(txs))
	return nil
}

func (i *Indexer) buildMetricForBlock(block *eth.Block, txs []eth.Transaction) (store.NetworkMetric, error) {
	bucketStart := time.Unix(int64(block.Timestamp), 0).UTC().Truncate(time.Minute)

	totalGasPrice := big.NewInt(0)
	totalFees := big.NewInt(0)
	validGasPriceCount := int64(0)

	for _, tx := range txs {
		gasPrice, ok := hexToBig(tx.GasPrice)
		if !ok {
			continue
		}
		gasLimit, gasOK := hexToBig(tx.Gas)

		totalGasPrice.Add(totalGasPrice, gasPrice)
		validGasPriceCount++

		if gasOK {
			feeEstimate := new(big.Int).Mul(gasPrice, gasLimit)
			totalFees.Add(totalFees, feeEstimate)
		}
	}

	avgGasPrice := "0"
	if validGasPriceCount > 0 {
		avg := new(big.Int).Div(totalGasPrice, big.NewInt(validGasPriceCount))
		avgGasPrice = avg.String()
	}

	avgBlockTimeSec := 0.0
	if block.Number > 0 {
		prev, err := i.repo.GetBlockByNumber(block.Number - 1)
		if err == nil && prev != nil {
			delta := time.Unix(int64(block.Timestamp), 0).UTC().Sub(prev.TimestampUTC).Seconds()
			if delta > 0 {
				avgBlockTimeSec = delta
			}
		}
	}

	return store.NetworkMetric{
		Granularity:     "minute",
		BucketStart:     bucketStart,
		AvgGasPriceWei:  avgGasPrice,
		TotalFeesWei:    totalFees.String(),
		TxCount:         int64(len(txs)),
		AvgBlockTimeSec: avgBlockTimeSec,
		BlockCount:      1,
	}, nil
}

func hexToBig(hexValue string) (*big.Int, bool) {
	v := strings.TrimSpace(hexValue)
	if v == "" || v == "0x" {
		return big.NewInt(0), true
	}
	if !strings.HasPrefix(v, "0x") {
		return nil, false
	}

	n := new(big.Int)
	_, ok := n.SetString(strings.TrimPrefix(v, "0x"), 16)
	if !ok {
		return nil, false
	}
	return n, true
}

func (i *Indexer) retryRPC(ctx context.Context, label string, op func() error) error {
	backoff := time.Second
	const maxAttempts = 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := op()
		if err == nil {
			return nil
		}

		if !isRetryableRPCError(err) || attempt == maxAttempts {
			return err
		}

		log.Printf("[indexer] %s failed (attempt %d/%d): %v; retrying in %s", label, attempt, maxAttempts, err, backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		if backoff < 15*time.Second {
			backoff *= 2
			if backoff > 15*time.Second {
				backoff = 15 * time.Second
			}
		}
	}

	return nil
}

func isRetryableRPCError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "status=429") ||
		strings.Contains(msg, "status=408") ||
		strings.Contains(msg, "status=500") ||
		strings.Contains(msg, "status=502") ||
		strings.Contains(msg, "status=503") ||
		strings.Contains(msg, "status=504") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "tempor") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "refused")
}
