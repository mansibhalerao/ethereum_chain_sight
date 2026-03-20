package store

import (
	"errors"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertBlock(block Block) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "number"}},
		DoUpdates: clause.AssignmentColumns([]string{"hash", "parent_hash", "timestamp_utc", "tx_count", "updated_at"}),
	}).Create(&block).Error
}

func (r *Repository) UpsertTransactions(transactions []Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	for i := range transactions {
		transactions[i].FromAddress = strings.ToLower(strings.TrimSpace(transactions[i].FromAddress))
		if transactions[i].ToAddress != nil {
			addr := strings.ToLower(strings.TrimSpace(*transactions[i].ToAddress))
			transactions[i].ToAddress = &addr
		}

		normalized, err := normalizeWeiString(transactions[i].ValueWei)
		if err != nil {
			return err
		}
		transactions[i].ValueWei = normalized
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hash"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "from_address", "to_address", "value_wei", "updated_at"}),
	}).Create(&transactions).Error
}

func (r *Repository) SaveWalletBalance(address, balanceWei string, blockNumber uint64) error {
	snapshot := WalletBalanceSnapshot{
		Address:     address,
		BalanceWei:  balanceWei,
		BlockNumber: blockNumber,
		CheckedAt:   time.Now().UTC(),
	}

	return r.db.Create(&snapshot).Error
}

func (r *Repository) GetBlockByNumber(number uint64) (*Block, error) {
	var block Block
	err := r.db.Where("number = ?", number).First(&block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (r *Repository) GetTransactionsByBlock(number uint64, limit int) ([]Transaction, error) {
	var txs []Transaction
	query := r.db.Where("block_number = ?", number).Order("id asc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&txs).Error; err != nil {
		return nil, err
	}

	return txs, nil
}

func (r *Repository) GetLatestIndexedBlockNumber() (uint64, error) {
	var latest uint64
	err := r.db.Model(&Block{}).
		Select("COALESCE(MAX(number), 0)").
		Scan(&latest).Error
	if err != nil {
		return 0, err
	}
	return latest, nil
}

type AddressLeaderboardRow struct {
	Address  string `json:"address"`
	ValueWei string `json:"valueWei"`
	TxCount  int64  `json:"txCount"`
}

type AddressActivityRow struct {
	Address string `json:"address"`
	TxCount int64  `json:"txCount"`
}

func (r *Repository) UpsertAddressSummariesFromBlock(blockNumber uint64) error {
	// Sender side
	if err := r.db.Exec(`
INSERT INTO address_summaries (address, tx_count, total_sent_wei, total_received_wei, first_seen_block, last_seen_block, created_at, updated_at)
SELECT lower(t.from_address), COUNT(*), COALESCE(SUM(t.value_wei),0), 0, ?, ?, NOW(), NOW()
FROM transactions t
WHERE t.block_number = ?
GROUP BY lower(t.from_address)
ON CONFLICT (address) DO UPDATE SET
	tx_count = address_summaries.tx_count + EXCLUDED.tx_count,
	total_sent_wei = address_summaries.total_sent_wei + EXCLUDED.total_sent_wei,
	first_seen_block = LEAST(address_summaries.first_seen_block, EXCLUDED.first_seen_block),
	last_seen_block = GREATEST(address_summaries.last_seen_block, EXCLUDED.last_seen_block),
	updated_at = NOW();
	`, blockNumber, blockNumber, blockNumber).Error; err != nil {
		return err
	}

	// Receiver side
	return r.db.Exec(`
INSERT INTO address_summaries (address, tx_count, total_sent_wei, total_received_wei, first_seen_block, last_seen_block, created_at, updated_at)
SELECT lower(t.to_address), COUNT(*), 0, COALESCE(SUM(t.value_wei),0), ?, ?, NOW(), NOW()
FROM transactions t
WHERE t.block_number = ? AND t.to_address IS NOT NULL
GROUP BY lower(t.to_address)
ON CONFLICT (address) DO UPDATE SET
	tx_count = address_summaries.tx_count + EXCLUDED.tx_count,
	total_received_wei = address_summaries.total_received_wei + EXCLUDED.total_received_wei,
	first_seen_block = LEAST(address_summaries.first_seen_block, EXCLUDED.first_seen_block),
	last_seen_block = GREATEST(address_summaries.last_seen_block, EXCLUDED.last_seen_block),
	updated_at = NOW();
	`, blockNumber, blockNumber, blockNumber).Error
}

func (r *Repository) GetAddressSummary(address string) (*AddressSummary, error) {
	var summary AddressSummary
	err := r.db.Where("address = ?", strings.ToLower(strings.TrimSpace(address))).First(&summary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *Repository) GetTopSenders(limit int) ([]AddressLeaderboardRow, error) {
	if limit <= 0 {
		limit = 10
	}

	var rows []AddressLeaderboardRow
	err := r.db.Raw(`
		SELECT address, total_sent_wei::text AS value_wei, tx_count
		FROM address_summaries
		ORDER BY total_sent_wei DESC
		LIMIT ?
	`, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) GetTopReceivers(limit int) ([]AddressLeaderboardRow, error) {
	if limit <= 0 {
		limit = 10
	}

	var rows []AddressLeaderboardRow
	err := r.db.Raw(`
		SELECT address, total_received_wei::text AS value_wei, tx_count
		FROM address_summaries
		ORDER BY total_received_wei DESC
		LIMIT ?
	`, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) GetMostActiveAddressesSince(since time.Time, limit int) ([]AddressActivityRow, error) {
	if limit <= 0 {
		limit = 10
	}

	var rows []AddressActivityRow
	err := r.db.Raw(`
		WITH addresses AS (
			SELECT lower(t.from_address) AS address, b.timestamp_utc AS ts
			FROM transactions t
			JOIN blocks b ON b.number = t.block_number
			UNION ALL
			SELECT lower(t.to_address) AS address, b.timestamp_utc AS ts
			FROM transactions t
			JOIN blocks b ON b.number = t.block_number
			WHERE t.to_address IS NOT NULL
		)
		SELECT address, COUNT(*) AS tx_count
		FROM addresses
		WHERE ts >= ?
		GROUP BY address
		ORDER BY tx_count DESC
		LIMIT ?
	`, since, limit).Scan(&rows).Error
	return rows, err
}

func (r *Repository) UpsertNetworkMetric(metric NetworkMetric) error {
	metric.Granularity = strings.ToLower(strings.TrimSpace(metric.Granularity))
	if metric.Granularity == "" {
		metric.Granularity = "minute"
	}

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "granularity"}, {Name: "bucket_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_fees_wei": gorm.Expr("network_metrics.total_fees_wei + EXCLUDED.total_fees_wei"),
			"tx_count":       gorm.Expr("network_metrics.tx_count + EXCLUDED.tx_count"),
			"block_count":    gorm.Expr("network_metrics.block_count + EXCLUDED.block_count"),
			"avg_gas_price_wei": gorm.Expr(`
				CASE
					WHEN (network_metrics.tx_count + EXCLUDED.tx_count) = 0 THEN 0
					ELSE (
						(network_metrics.avg_gas_price_wei * network_metrics.tx_count) +
						(EXCLUDED.avg_gas_price_wei * EXCLUDED.tx_count)
					) / (network_metrics.tx_count + EXCLUDED.tx_count)
				END`),
			"avg_block_time_sec": gorm.Expr(`
				CASE
					WHEN (network_metrics.block_count + EXCLUDED.block_count) = 0 THEN 0
					ELSE (
						(network_metrics.avg_block_time_sec * network_metrics.block_count) +
						(EXCLUDED.avg_block_time_sec * EXCLUDED.block_count)
					) / (network_metrics.block_count + EXCLUDED.block_count)
				END`),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}).Create(&metric).Error
}

func (r *Repository) GetNetworkMetrics(granularity string, since time.Time, limit int) ([]NetworkMetric, error) {
	if limit <= 0 {
		limit = 200
	}
	if granularity == "" {
		granularity = "minute"
	}

	var rows []NetworkMetric
	err := r.db.Where("granularity = ? AND bucket_start >= ?", strings.ToLower(granularity), since).
		Order("bucket_start asc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func normalizeWeiString(input string) (string, error) {
	v := strings.TrimSpace(input)
	if v == "" {
		return "0", nil
	}
	if strings.HasPrefix(v, "0x") {
		n := new(big.Int)
		_, ok := n.SetString(strings.TrimPrefix(v, "0x"), 16)
		if !ok {
			return "", errors.New("invalid hex wei value")
		}
		return n.String(), nil
	}
	// Validate decimal numeric string
	n := new(big.Int)
	_, ok := n.SetString(v, 10)
	if !ok {
		return "", errors.New("invalid decimal wei value")
	}
	return n.String(), nil
}
