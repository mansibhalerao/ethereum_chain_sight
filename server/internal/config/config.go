package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	EthRPCURL               string
	Env                     string
	DBEnabled               bool
	PostgresDSN             string
	AutoMigrate             bool
	IndexerEnabled          bool
	IndexerInitialLookback  int
	IndexerMaxBlocksPerTick int
	IndexerIntervalSeconds  int
}

func Load() Config {
	loadDotEnvOnce()

	port := getEnv("PORT", "8080")
	rpcURL := getEnv("ETH_RPC_URL", "https://eth.llamarpc.com")
	env := getEnv("APP_ENV", "development")
	DBEnabled := getEnvBool("DB_ENABLED", false)
	postgresDSN := getEnv("POSTGRES_DSN", "")
	autoMigrate := getEnvBool("DB_AUTO_MIGRATE", true)

	// easy to toggle Indexer (case-insensitive)
	indexerEnabled := getEnvBool("INDEXER_ENABLED", false)

	indexerInitialLookback := getEnvInt("INDEXER_INITIAL_LOOKBACK", 20)
	indexerMaxBlocksPerTick := getEnvInt("INDEXER_MAX_BLOCKS_PER_TICK", 2)
	indexerIntervalSeconds := getEnvInt("INDEXER_INTERVAL_SECONDS", 8)

	return Config{
		Port:                    port,
		EthRPCURL:               rpcURL,
		Env:                     env,
		DBEnabled:               DBEnabled,
		PostgresDSN:             postgresDSN,
		AutoMigrate:             autoMigrate,
		IndexerEnabled:          indexerEnabled,
		IndexerInitialLookback:  indexerInitialLookback,
		IndexerMaxBlocksPerTick: indexerMaxBlocksPerTick,
		IndexerIntervalSeconds:  indexerIntervalSeconds,
	}
}

var dotEnvOnce sync.Once

func loadDotEnvOnce() {
	dotEnvOnce.Do(func() {
		// Load from a few common locations. We only attempt files that exist.
		// This does not override already-set environment variables.
		candidates := []string{".env", "server/.env", "../.env"}
		paths := make([]string, 0, len(candidates))
		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				paths = append(paths, p)
			}
		}
		if len(paths) == 0 {
			return
		}

		if err := godotenv.Load(paths...); err != nil {
			// File existed but couldn't be read/parsed.
			log.Printf("failed to load .env file(s): %v", err)
		}
	})
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	raw := getEnv(key, "")
	if raw == "" {
		return fallback
	}

	value := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return fallback
		}
		value = value*10 + int(ch-'0')
	}

	if value <= 0 {
		return fallback
	}

	return value
}

func getEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(getEnv(key, ""))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}
