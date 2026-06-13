# Yasumi Backend

Phase 04 authentication and direct API boundary for the Yasumi Go backend.

## What Exists

- Go executables at `cmd/yasumi-api` and `cmd/yasumi-migrate`.
- Internal packages for app wiring, typed config, HTTP routing, structured logging, migrations, and PostgreSQL repository access.
- Infrastructure endpoints: `GET /healthz` and `GET /readyz`.
- Direct API endpoints: `GET /v1/session` and `POST /v1/sync/token`.
- Request ID, request timeout, bearer authentication boundary, stable API error shape, and sync token issuance.
- Embedded PostgreSQL migrations for `items`, `recurring_task_templates`, `areas`, `operation_history`, and `user_settings`.
- Repository transaction support and explicit user-scoped query methods.
- Local environment files under `env/`.
- Dockerfile for the API runtime and migration command.

No MVP business CRUD routes are implemented in Phase 04.

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
http://localhost:7650/healthz
http://localhost:7650/readyz
```

Authenticated local development calls use the placeholder bearer token from `env/local.env.example`:

```powershell
curl -H "Authorization: Bearer local-dev-session-token" http://localhost:7650/v1/session
curl -X POST -H "Authorization: Bearer local-dev-session-token" -H "Content-Type: application/json" -d "{\"device_id\":\"device-01\",\"client_version\":\"0.1.0\"}" http://localhost:7650/v1/sync/token
```

## Docker Local Environment

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml up --build
```

The default compose stack starts PostgreSQL, applies migrations, and starts the API.

To apply migrations directly:

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml run --rm migrate
```

PowerSync is configured with the official `journeyapps/powersync-service` image and can be started when sync work begins:

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml --profile sync up --build
```

PowerSync service wiring exists for later sync phases. `/readyz` reports the configured sync service as unavailable unless that dependency is reachable.

## Configuration

Configuration is read once at startup from environment variables. See `env/local.env.example` for the supported keys.

Invalid configuration fails fast with a startup error.

## Legacy

The previous Gin/Gorm prototype is preserved under `legacy/old-gin-gorm` for reference only. Active code uses the `cmd/` and `internal/` layout from the backend guide.
