# ChainSight

ChainSight is a full-stack Ethereum analytics playground built with a React frontend and a Go backend. The goal is to make it easy to explore blocks, transactions, and wallet activity with a clean UI and a typed, well-structured API.

- **Frontend:** React app (served by Nginx in Docker)
- **Backend:** Go API
- **Optional data layer:** Postgres + background indexer

## Tech Stack

- Frontend: React, React Router, TailwindCSS, Create React App
- Backend: Go 1.22, GORM
- Database: Postgres 16 (optional by mode)
- Infra: Docker, Docker Compose, Nginx

## Project Structure

```text
.
├── public/                    # Static assets for the React app
├── src/                       # React application source
│   ├── api/                   # API clients (e.g. Ethereum JSON-RPC)
│   ├── components/            # Reusable UI components
│   ├── pages/                 # Top-level pages (routes)
│   ├── services/              # Frontend-domain service layer
│   └── ...
├── server/
│   ├── cmd/chainsight-api/    # API entrypoint (main.go)
│   └── internal/              # Config, HTTP API, indexer, storage
│       ├── config/
│       ├── eth/
│       ├── httpapi/           # HTTP router and handlers
│       ├── indexer/           # used in full mode, blockchain indexing and aggregation logic
│       └── store/             # used in full mode, models and repositories using GORM and Postgres
├── docs/
├── docker-compose.yml
├── Dockerfile.frontend
└── server/Dockerfile
```

See [docs/architecture.md](docs/architecture.md) for a deeper dive into how the pieces fit together.
See [docs/Screenshots.md](docs/architecture.md) for page views

## Running the Project, Runtime Modes

ChainSight now supports **two explicit modes** controlled by env flags:

Endpoint behavior by mode
- Live RPC endpoints always on.
- Analytics endpoints require DB (and fresh data requires indexer).

### 1) Core mode (RPC only)

- Uses hosted Ethereum RPC
- No DB, no indexer
- Live block flow works

Required flags:
- `DB_ENABLED=false`
- `INDEXER_ENABLED=false`

Run:
```bash
cd /Users/mansi/projects/chainsight
docker compose up --build
```

---

### 2) Full mode (RPC + DB + indexer)

- Uses RPC + Postgres
- Runs background indexer
- Enables DB-backed analytics endpoints/data

Required flags:
- `DB_ENABLED=true`
- `INDEXER_ENABLED=true`

Run:
```bash
cd /Users/mansi/projects/chainsight
DB_ENABLED=true INDEXER_ENABLED=true docker compose --profile full up --build
```

---
## Environment Variable Flow (important)

Priority order at runtime:

1. Container/runtime environment variables (highest priority)
2. `.env` loaded by backend config (`godotenv`)
3. Defaults in `server/internal/config/config.go`

Notes:

- `.env.example` is a template only (documentation), not runtime.
- In Docker runs, `docker-compose.yml` is the main toggle point.

---

## Key Backend Flags

- `DB_ENABLED` → enables/disables Postgres usage in backend startup
- `DB_AUTO_MIGRATE` → runs schema migration on startup when DB is enabled


## Indexer Runtime Behavior (New)

The backend indexer now runs continuously at fixed intervals (instead of a one-time startup run).

You can tune it with environment variables:

- `INDEXER_ENABLED` (default: `true`)
- `INDEXER_INITIAL_LOOKBACK` (default: `20`)
- `INDEXER_MAX_BLOCKS_PER_TICK` (default: `2`, recommended `1` for strict rate limits)
- `INDEXER_INTERVAL_SECONDS` (default: `8`)

Recommended conservative profile (rate-limit friendly):

- `INDEXER_MAX_BLOCKS_PER_TICK=1`
- `INDEXER_INTERVAL_SECONDS=8`

## Docker Compose Notes

- Single root `docker-compose.yml` is used.
- `db` service is under `profiles: ["full"]`

---

## Local (non-Docker) run

Backend:
```bash
cd /Users/mansi/projects/chainsight/server
cp .env.example .env
go run ./cmd/chainsight-api
```

Frontend:
```bash
cd /Users/mansi/projects/chainsight
npm install
npm start
```

Frontend URL: `http://localhost:3000`  
Backend URL: `http://localhost:8080`

---
