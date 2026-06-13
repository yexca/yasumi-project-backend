# Local Environment

This directory contains reproducible local development environment files.

## Go Toolchain

The backend targets Go `1.23`, matching `go.mod`, the CI workflow, and the Dockerfile build image.

Use Docker when Go is not installed on the host:

```powershell
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine go test ./...
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine gofmt -w cmd internal
docker run --rm -v "${PWD}:/src" -w /src golang:1.23-alpine go vet ./...
```

The production-style runtime is built from the repository root `Dockerfile`.

## Local Services

`docker-compose.yml` defines PostgreSQL for the backend and an optional PowerSync profile for later sync phases.

`local.env.example` contains local-only placeholder values. Do not store real credentials in this repository.

