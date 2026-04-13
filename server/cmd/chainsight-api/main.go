package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mansi/chainsight/server/internal/config"
	"github.com/mansi/chainsight/server/internal/eth"
	"github.com/mansi/chainsight/server/internal/httpapi"

	"github.com/mansi/chainsight/server/internal/indexer"
	"github.com/mansi/chainsight/server/internal/store"
)

func main() {
	cfg := config.Load()
	var repo *store.Repository

	// DB mode is explicit
	if cfg.DBEnabled {
		if cfg.PostgresDSN == "" {
			log.Fatal("DB_ENABLED=true but POSTGRES_DSN is empty")
		}

		db, err := store.OpenPostgres(cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("failed to connect postgres: %v", err)
		}

		if cfg.AutoMigrate {
			if err := store.AutoMigrate(db); err != nil {
				log.Fatalf("failed to auto-migrate schema: %v", err)
			}
			log.Printf("Postgres schema migrated successfully")
		}

		repo = store.NewRepository(db)
		log.Printf("DB mode: enabled")
	} else {
		log.Printf("DB mode: disabled (core mode)")
	}

	ethClient := eth.NewClient(cfg.EthRPCURL)
	// pass DB feature flag into HTTP layer
	server := httpapi.NewServer(ethClient, repo, cfg.DBEnabled)

	// Indexer mode is explicit and requires DB
	if cfg.IndexerEnabled {
		if !cfg.DBEnabled || repo == nil {
			log.Fatal("INDEXER_ENABLED=true requires DB_ENABLED=true")
		}

		idx := indexer.New(ethClient, repo, indexer.Config{
			InitialLookback:  uint64(cfg.IndexerInitialLookback),
			MaxBlocksPerTick: cfg.IndexerMaxBlocksPerTick,
		})

		interval := time.Duration(cfg.IndexerIntervalSeconds) * time.Second
		go idx.Run(context.Background(), interval)
		log.Printf("Indexer started (interval=%s, lookback=%d, maxBlocksPerTick=%d)", interval, cfg.IndexerInitialLookback, cfg.IndexerMaxBlocksPerTick)
	}

	addr := ":" + cfg.Port
	log.Printf("ChainSight API listening on %s (env=%s)", addr, cfg.Env)
	log.Printf("Using hosted RPC: %s", cfg.EthRPCURL)

	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
