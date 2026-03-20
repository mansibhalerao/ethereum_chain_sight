package eth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Client struct {
	rpcURL string
	http   *http.Client
}

func NewClient(rpcURL string) *Client {
	return &Client{
		rpcURL: rpcURL,
		http: &http.Client{
			// No client-level timeout. Long-running indexer RPCs should not fail
			// due to an arbitrary fixed deadline.
			Timeout: 0,
		},
	}
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCBlock struct {
	Number       string   `json:"number"`
	Hash         string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	Timestamp    string   `json:"timestamp"`
	Transactions []string `json:"transactions"`
}

type RPCBlockWithTransactions struct {
	Number       string           `json:"number"`
	Hash         string           `json:"hash"`
	ParentHash   string           `json:"parentHash"`
	Timestamp    string           `json:"timestamp"`
	Transactions []RPCTransaction `json:"transactions"`
}

type RPCTransaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	BlockNumber string `json:"blockNumber"`
}

type Block struct {
	Number       uint64   `json:"number"`
	Hash         string   `json:"hash"`
	ParentHash   string   `json:"parentHash"`
	Timestamp    uint64   `json:"timestamp"`
	Transactions []string `json:"transactions"`
}

type Transaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Gas         string `json:"gas"`
	GasPrice    string `json:"gasPrice"`
	BlockNumber uint64 `json:"blockNumber"`
}

func (c *Client) GetLatestBlock(ctx context.Context) (*Block, error) {
	return c.getBlockByTag(ctx, "latest")
}

func (c *Client) GetBlockByNumber(ctx context.Context, number uint64) (*Block, error) {
	tag := "0x" + strconv.FormatUint(number, 16)
	return c.getBlockByTag(ctx, tag)
}

func (c *Client) GetBlockTransactions(ctx context.Context, number uint64, limit int) ([]Transaction, error) {
	_, txs, err := c.GetBlockWithTransactions(ctx, number)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(txs) > limit {
		return txs[:limit], nil
	}
	return txs, nil
}

func (c *Client) GetBlockWithTransactions(ctx context.Context, number uint64) (*Block, []Transaction, error) {
	tag := "0x" + strconv.FormatUint(number, 16)
	return c.getBlockByTagWithTransactions(ctx, tag)
}

func (c *Client) GetWalletBalance(ctx context.Context, address string) (string, error) {
	var result string
	if err := c.callRPC(ctx, "eth_getBalance", []interface{}{address, "latest"}, &result); err != nil {
		return "", err
	}
	return result, nil
}

func (c *Client) GetTransactionByHash(ctx context.Context, hash string) (*Transaction, error) {
	var rpcTx RPCTransaction
	if err := c.callRPC(ctx, "eth_getTransactionByHash", []interface{}{hash}, &rpcTx); err != nil {
		return nil, err
	}

	blockNumber, err := hexToUint64Safe(rpcTx.BlockNumber)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		Hash:        rpcTx.Hash,
		From:        rpcTx.From,
		To:          rpcTx.To,
		Value:       rpcTx.Value,
		Gas:         rpcTx.Gas,
		GasPrice:    rpcTx.GasPrice,
		BlockNumber: blockNumber,
	}, nil
}

func (c *Client) getBlockByTag(ctx context.Context, tag string) (*Block, error) {
	var rpcBlock RPCBlock
	if err := c.callRPC(ctx, "eth_getBlockByNumber", []interface{}{tag, false}, &rpcBlock); err != nil {
		return nil, err
	}

	number, err := hexToUint64Safe(rpcBlock.Number)
	if err != nil {
		return nil, err
	}
	timestamp, err := hexToUint64Safe(rpcBlock.Timestamp)
	if err != nil {
		return nil, err
	}

	return &Block{
		Number:       number,
		Hash:         rpcBlock.Hash,
		ParentHash:   rpcBlock.ParentHash,
		Timestamp:    timestamp,
		Transactions: rpcBlock.Transactions,
	}, nil
}

func (c *Client) getBlockByTagWithTransactions(ctx context.Context, tag string) (*Block, []Transaction, error) {
	var rpcBlock RPCBlockWithTransactions
	if err := c.callRPC(ctx, "eth_getBlockByNumber", []interface{}{tag, true}, &rpcBlock); err != nil {
		return nil, nil, err
	}

	number, err := hexToUint64Safe(rpcBlock.Number)
	if err != nil {
		return nil, nil, err
	}
	timestamp, err := hexToUint64Safe(rpcBlock.Timestamp)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([]string, 0, len(rpcBlock.Transactions))
	txs := make([]Transaction, 0, len(rpcBlock.Transactions))
	for _, rpcTx := range rpcBlock.Transactions {
		hashes = append(hashes, rpcTx.Hash)

		blockNumber := number
		if strings.TrimSpace(rpcTx.BlockNumber) != "" {
			if parsed, parseErr := hexToUint64Safe(rpcTx.BlockNumber); parseErr == nil {
				blockNumber = parsed
			}
		}

		txs = append(txs, Transaction{
			Hash:        rpcTx.Hash,
			From:        rpcTx.From,
			To:          rpcTx.To,
			Value:       rpcTx.Value,
			Gas:         rpcTx.Gas,
			GasPrice:    rpcTx.GasPrice,
			BlockNumber: blockNumber,
		})
	}

	block := &Block{
		Number:       number,
		Hash:         rpcBlock.Hash,
		ParentHash:   rpcBlock.ParentHash,
		Timestamp:    timestamp,
		Transactions: hashes,
	}

	return block, txs, nil
}

func (c *Client) callRPC(ctx context.Context, method string, params interface{}, out interface{}) error {
	payload, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	})
	if err != nil {
		return fmt.Errorf("marshal rpc request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send rpc request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read rpc response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("rpc http error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return fmt.Errorf("decode rpc response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error code=%d message=%s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if string(rpcResp.Result) == "null" {
		return fmt.Errorf("rpc method %s returned null", method)
	}

	if err := json.Unmarshal(rpcResp.Result, out); err != nil {
		return fmt.Errorf("decode rpc result: %w", err)
	}

	return nil
}

func hexToUint64Safe(hexValue string) (uint64, error) {
	hexValue = strings.TrimSpace(hexValue)
	if hexValue == "" || hexValue == "0x" {
		return 0, nil
	}
	if !strings.HasPrefix(hexValue, "0x") {
		return 0, fmt.Errorf("invalid hex value: %s", hexValue)
	}

	parsed, err := strconv.ParseUint(strings.TrimPrefix(hexValue, "0x"), 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse hex value %s: %w", hexValue, err)
	}
	return parsed, nil
}
