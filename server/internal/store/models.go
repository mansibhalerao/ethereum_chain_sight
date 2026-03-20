package store

import "time"

// Block stores fields currently used in the UI block cards/pages.
type Block struct {
	ID           uint      `gorm:"primaryKey"`
	Number       uint64    `gorm:"not null;uniqueIndex;index"`
	Hash         string    `gorm:"size:66;not null;uniqueIndex"`
	ParentHash   string    `gorm:"size:66;not null;index"`
	TimestampUTC time.Time `gorm:"not null;index"`
	TxCount      int       `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Transactions []Transaction `gorm:"foreignKey:BlockNumber;references:Number"`
}

// Transaction stores fields shown in transaction cards.
type Transaction struct {
	ID uint `gorm:"primaryKey"`

	Hash        string `gorm:"size:66;not null;uniqueIndex"`
	BlockNumber uint64 `gorm:"not null;index"`

	FromAddress string  `gorm:"column:from_address;size:42;not null;index"`
	ToAddress   *string `gorm:"column:to_address;size:42;index"`

	// Store wei as numeric string (Postgres NUMERIC) for leaderboard aggregation.
	ValueWei string `gorm:"type:numeric(78,0);not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// WalletBalanceSnapshot supports the wallet checker screen and history.
type WalletBalanceSnapshot struct {
	ID uint `gorm:"primaryKey"`

	Address     string    `gorm:"size:42;not null;index:idx_wallet_block,priority:1;index"`
	BalanceWei  string    `gorm:"type:text;not null"`
	BlockNumber uint64    `gorm:"not null;index:idx_wallet_block,priority:2;index"`
	CheckedAt   time.Time `gorm:"not null;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// AddressSummary powers address quick stats and leaderboards.
type AddressSummary struct {
	ID uint `gorm:"primaryKey"`

	Address          string `gorm:"size:42;not null;uniqueIndex"`
	TxCount          int64  `gorm:"not null;default:0;index"`
	TotalSentWei     string `gorm:"column:total_sent_wei;type:numeric(78,0);not null;default:0"`
	TotalReceivedWei string `gorm:"column:total_received_wei;type:numeric(78,0);not null;default:0"`
	FirstSeenBlock   uint64 `gorm:"not null;default:0;index"`
	LastSeenBlock    uint64 `gorm:"not null;default:0;index"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NetworkMetric supports dashboard charts for minute/block time series.
// Granularity examples: "minute", "block".
type NetworkMetric struct {
	ID uint `gorm:"primaryKey"`

	Granularity     string    `gorm:"size:20;not null;uniqueIndex:idx_metric_bucket,priority:1"`
	BucketStart     time.Time `gorm:"not null;uniqueIndex:idx_metric_bucket,priority:2"`
	AvgGasPriceWei  string    `gorm:"type:numeric(78,0);not null;default:0"`
	TotalFeesWei    string    `gorm:"type:numeric(78,0);not null;default:0"`
	TxCount         int64     `gorm:"not null;default:0"`
	AvgBlockTimeSec float64   `gorm:"not null;default:0"`
	BlockCount      int64     `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
