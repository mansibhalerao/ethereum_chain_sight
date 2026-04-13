# ChainSight Go Backend

This backend proxies Ethereum JSON-RPC through your own API (using a hosted RPC URL).

## Endpoints

- `GET /api/health`
- `GET /api/blocks/latest`
- `GET /api/blocks/{number}`
- `GET /api/blocks/{number}/transactions?limit=5`
- `GET /api/wallets/{address}/balance`

### DB-backed endpoints (require Postgres configured)

- `GET /api/addresses/{address}/summary`
- `GET /api/leaderboards/top-senders?limit=10`
- `GET /api/leaderboards/top-receivers?limit=10`
- `GET /api/leaderboards/most-active?hours=24&limit=10`
- `GET /api/analytics/metrics?granularity=minute&hours=24&limit=200`

## Run locally

1. From this folder:
   - `cd /Users/.../chainsight/server`
2. Copy env.example to local env file, then edit .env for your env:
   - `cp .env.example .env`
3. build:
   `go build ./...`
3. Run API:
   - `go run ./cmd/chainsight-api`

Default server: `http://localhost:8080`

## Database (Postgres)

use `POSTGRES_DSN` from `.env.example` to connect with db.

## Indexer behavior and throttling

The indexer runs continuously in intervals

Useful env vars:

- `INDEXER_ENABLED` (default: `true`)
- `INDEXER_INITIAL_LOOKBACK` (default: `20`) - only used on first bootstrap when DB has no indexed blocks.
- `INDEXER_MAX_BLOCKS_PER_TICK` (default: `1`) - max blocks processed each interval.
- `INDEXER_INTERVAL_SECONDS` (default: `8`) - how often an indexing tick runs.

Conservative rate-limit-safe profile:

- `INDEXER_MAX_BLOCKS_PER_TICK=1`
- `INDEXER_INTERVAL_SECONDS=8`


