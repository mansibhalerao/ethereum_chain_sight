# ChainSight

ChainSight is a full-stack Ethereum analytics playground built with a React frontend and a Go backend. The goal is to make it easy to explore blocks, transactions, and wallet activity with a clean UI and a typed, well-structured API.

## Tech Stack

- Frontend: React, React Router, TailwindCSS, Create React App
- Backend: Go 1.22, GORM, Postgres
- Infra: Docker, Docker Compose, Nginx

## Project Structure

```text
.
├── build/                  # Compiled frontend assets (Create React App)
├── public/                 # Static assets for the React app
├── src/                    # React application source
│   ├── api/                # API clients (e.g. Ethereum JSON-RPC)
│   ├── components/         # Reusable UI components
│   ├── pages/              # Top-level pages (routes)
│   ├── services/           # Frontend-domain service layer
│   └── ...                 # App entry, styles, tests
├── server/                 # Go backend API
│   ├── cmd/chainsight-api/ # API entrypoint (main.go)
│   └── internal/           # Config, HTTP API, indexer, storage
└── docs/                   # High-level documentation (architecture, etc.)
```

See [docs/architecture.md](docs/architecture.md) for a deeper dive into how the pieces fit together.
See [docs/Screenshots.md](docs/architecture.md) for page views

## Running the Project

### Option A: Run everything with Docker (recommended)

From the project root:

1. Build and start all services:
	- `docker compose up --build -d` 
2. Open the app in your browser:
	- Frontend: `http://localhost:3000`
	- Backend API (if you want to inspect it directly): `http://localhost:8080`

What this does:

- Starts Postgres in a `db` container.
- Starts the Go API in a `backend` container (port 8080).
- Builds the React app and serves it via nginx in a `frontend` container (port 3000).

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

In Docker Compose, these values are already set for the backend service.

### Option B: Run services locally without Docker

#### 1. Start the Go backend

From [server](server):

1. Copy env vars and configure them for your environment:
	- `cp .env.example .env`
2. Start Postgres (Docker):
	- `docker compose -f docker-compose.postgres.yml up -d`
3. Run the API server:
	- `go run ./cmd/chainsight-api`

The API will be available at `http://localhost:8080`.

#### 2. Start the React frontend

From the project root:

1. Install dependencies:
	- `npm install`
2. Start the dev server:
	- `npm start`

The frontend runs on `http://localhost:3000` and is configured (via `src/.env`) to talk to the Go API at `http://localhost:8080`.

## Frontend Overview

- Routing is handled by React Router.
- Pages live under [src/pages](src/pages) (e.g. `BlockPage`, `WalletTest`).
- Reusable display components live under [src/components](src/components) (e.g. `BlockInfo`, `TransactionInfo`).
- Domain logic for talking to the backend and Ethereum lives under [src/services](src/services) and [src/api](src/api).

To run tests:

- `npm test`

To build production assets:

- `npm run build`

## Backend Overview

Backend documentation, endpoints, and examples are in [server/README.md](server/README.md).

At a glance:

- `internal/httpapi` – HTTP router and handlers
- `internal/indexer` – blockchain indexing and aggregation logic
- `internal/store` – models and repositories using GORM and Postgres

