# Yasumi Backend

Phase 03 database and repository baseline for the Yasumi Go backend.

## What Exists

- Go executables at `cmd/yasumi-api` and `cmd/yasumi-migrate`.
- Internal packages for app wiring, typed config, HTTP routing, structured logging, migrations, and PostgreSQL repository access.
- Infrastructure endpoints: `GET /healthz`, `GET /readyz`, and `GET /v1`.
- Embedded PostgreSQL migrations for `items`, `recurring_task_templates`, `areas`, `operation_history`, and `user_settings`.
- Repository transaction support and explicit user-scoped query methods.
- Local environment files under `env/`.
- Dockerfile for the API runtime and migration command.

No MVP business CRUD routes are implemented in Phase 03.

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
http://localhost:8080/healthz
http://localhost:8080/readyz
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

PowerSync service wiring exists for later sync phases; Phase 03 only owns the database schema and repository baseline.

## Configuration

Configuration is read once at startup from environment variables. See `env/local.env.example` for the supported keys.

Invalid configuration fails fast with a startup error.

## Legacy

The previous Gin/Gorm prototype is preserved under `legacy/old-gin-gorm` for reference only. Active code uses the `cmd/` and `internal/` layout from the backend guide.
