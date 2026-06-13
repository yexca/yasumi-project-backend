# Local Environment

This directory contains reproducible local development environment files.

## Go Toolchain

The backend targets Go `1.23`, matching `go.mod`, the CI workflow, and the Dockerfile build image.

Use Docker when Go is not installed on the host. The preferred project-local toolchain entrypoint is the `go-toolchain` service:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

Repository integration tests need PostgreSQL:

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

The production-style runtime is built from the repository root `Dockerfile`.

## Local Services

`docker-compose.yml` defines PostgreSQL, a one-shot `migrate` service, the API, and an optional PowerSync profile for later sync phases.

The API image includes both:

- `/usr/local/bin/yasumi-api`
- `/usr/local/bin/yasumi-migrate`

`local.env.example` contains local-only placeholder values for PostgreSQL, development bearer authentication, and sync token signing. Do not store real credentials in this repository.

The default development bearer token is only for local testing:

```powershell
curl -H "Authorization: Bearer local-dev-session-token" http://localhost:7650/v1/session
```
