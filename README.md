# Yasumi Backend

Phase 01 foundation for the Yasumi Go backend.

## What Exists

- Go executable at `cmd/yasumi-api`.
- Internal packages for app wiring, typed config, HTTP routing, and structured logging.
- Infrastructure endpoints: `GET /healthz`, `GET /readyz`, and `GET /v1`.
- Local environment files under `env/`.
- Dockerfile for the API runtime.

No MVP business CRUD routes are implemented in Phase 01.

## Local Commands

If Go is installed locally:

```powershell
go test ./...
go fmt ./...
go vet ./...
go run ./cmd/yasumi-api
```

If Go is not installed locally, use Docker:

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

The default compose stack starts PostgreSQL and the API.

PowerSync is configured with the official `journeyapps/powersync-service` image and can be started when sync work begins:

```powershell
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml --profile sync up --build
```

The API does not connect to PostgreSQL or PowerSync in Phase 01; these settings exist so later phases have reproducible local dependency names and ports.

## Configuration

Configuration is read once at startup from environment variables. See `env/local.env.example` for the supported keys.

Invalid configuration fails fast with a startup error.

## Legacy

The previous Gin/Gorm prototype is preserved under `legacy/old-gin-gorm` for reference only. Active Phase 01 code uses the `cmd/` and `internal/` layout from the backend guide.
