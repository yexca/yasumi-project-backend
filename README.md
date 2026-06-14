# Yasumi Backend

Yasumi Go backend with first-party accounts and the MVP sync validation boundary.

## What Exists

- Go executables at `cmd/yasumi-api` and `cmd/yasumi-migrate`.
- Internal packages for app wiring, typed config, HTTP routing, structured logging, migrations, and PostgreSQL repository access.
- Infrastructure endpoints: `GET /healthz` and `GET /readyz`.
- Direct API endpoints: `POST /v1/auth/register`, `POST /v1/auth/login`, `POST /v1/auth/logout`, `POST /v1/auth/refresh`, `GET /v1/session`, `POST /v1/profile`, `POST /v1/profile/password`, `GET /v1/weather`, and `POST /v1/sync/token`.
- Sync upload adapter endpoint: `POST /v1/sync/upload`.
- Request ID, request timeout, account-backed bearer authentication boundary, stable API error shape, production-safe request logs, `/metrics`, and PowerSync-compatible JWT sync token issuance.
- Embedded PostgreSQL migrations for account tables, `items`, `recurring_task_templates`, `areas`, `operation_history`, and `user_settings`.
- Repository transaction support and explicit user-scoped query/write methods.
- Service-layer sync upload validation for ownership, item shapes, semantic item transitions, idempotency, server timestamps, and revisions.
- Local environment examples at `.env.example` and `env/local.env.example`.
- Dockerfile for the API runtime and migration command.
- Root `docker-compose.example.yml` for building and running the local stack from the repository root.

No MVP business CRUD routes are implemented. Synced business writes enter through the sync upload boundary.

## Local Commands

If Go is installed locally:

```powershell
go test ./...
go fmt ./...
go vet ./...
go run ./cmd/yasumi-api
go run ./cmd/yasumi-migrate
```

If Go is not installed locally, use the project-local Docker toolchain:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

To run repository integration tests against the compose PostgreSQL service:

```powershell
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
```

One-off Docker commands are also valid:

```powershell
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine go test ./...
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine gofmt -w cmd internal
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine go vet ./...
```

Then open:

```text
http://localhost:7659/healthz
http://localhost:7659/readyz
http://localhost:7659/metrics
```

The local frontend defaults to `http://127.0.0.1:7659` for direct API calls. If you serve the frontend from a different origin, add that origin to `YASUMI_HTTP_ALLOWED_ORIGINS`.

Create or log in to a local account, then use the returned `access_token` for authenticated calls:

```powershell
$auth = Invoke-RestMethod -Uri http://localhost:7659/v1/auth/register -Method Post -ContentType "application/json" -Body "{\"username\":\"local_user\",\"email\":\"local@example.com\",\"password\":\"password123\"}"
$token = $auth.session.access_token
curl -H "Authorization: Bearer $token" http://localhost:7659/v1/session
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"display_name\":\"Quiet Planner\"}" http://localhost:7659/v1/profile
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"current_password\":\"password123\",\"new_password\":\"password456\"}" http://localhost:7659/v1/profile/password
curl -H "Authorization: Bearer $token" "http://localhost:7659/v1/weather?city=Tokyo"
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"device_id\":\"device-01\",\"client_version\":\"0.1.0\"}" http://localhost:7659/v1/sync/token
curl -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d "{\"client_batch_id\":\"batch-01\",\"device_id\":\"device-01\",\"mutations\":[]}" http://localhost:7659/v1/sync/upload
```

## Docker Local Environment

The recommended root-level example is:

```powershell
Copy-Item .env.example .env
docker compose -f .\docker-compose.example.yml up --build
```

Docker Compose automatically reads the root `.env` file when commands are run from this repository root. The `.env` file is ignored by Git and Docker build contexts; keep local secrets there and commit only `.env.example`.

The root compose stack starts PostgreSQL, MongoDB for PowerSync, PowerSync, applies migrations, and starts the API. It exposes:

```text
http://localhost:7659/healthz
http://localhost:7659/readyz
http://localhost:7659/metrics
```

To apply migrations directly:

```powershell
docker compose -f .\docker-compose.example.yml run --rm migrate
```

The same command brings up the required PowerSync infrastructure for local sync validation:

```powershell
docker compose -f .\docker-compose.example.yml up --build
```

The same Compose file can also run the production-style frontend container when this repository is next to `../yasumi-project-frontend`:

```powershell
docker compose -f .\docker-compose.example.yml --profile frontend up --build
```

The frontend is served from `http://localhost:5173` and receives backend and PowerSync URLs through runtime environment values.

`/readyz` reports the configured sync service as unavailable until PowerSync is reachable. Use `/healthz` if you are intentionally validating only process liveness.

The older project-local development compose file remains available:

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml up --build
```

PowerSync service wiring includes user-scoped Sync Streams for MVP synced tables and a local HS256 development key matching `YASUMI_SYNC_TOKEN_SECRET`. The default local stack now treats PowerSync as required infrastructure, so `/readyz` should become healthy only after that dependency is reachable.

The Docker toolchain mounts `../dev_documents` read-only as `/dev_documents` so fixture and shared-constant compatibility tests run with the same contract documents used during development.

## Configuration

Configuration is read once at startup from environment variables. See `.env.example` for the root Docker Compose defaults and `env/local.env.example` for the older `env/docker-compose.yml` workflow.

The application binary does not load `.env` files by itself. Docker Compose reads the root `.env` file for interpolation, then passes the configured values into the containers.

Invalid configuration fails fast with a startup error.

## Release Readiness

Deployment and operational checks are documented in `documents/deployment-operations.md`.

The backend MVP release checklist is in `documents/mvp-release-checklist.md`.

## Legacy

The previous Gin/Gorm prototype is preserved under `legacy/old-gin-gorm` for reference only. Active code uses the `cmd/` and `internal/` layout from the backend guide.
