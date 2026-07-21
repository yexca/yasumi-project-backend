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

The default local runtime image name is `yexca/yasumi-backend:0.1.1`.

## Local Services

`docker-compose.yml` defines PostgreSQL, MongoDB for PowerSync, PowerSync, a one-shot `migrate` service, and the API.

The API image includes both:

- `/usr/local/bin/yasumi-api`
- `/usr/local/bin/yasumi-migrate`

`local.env.example` contains local-only placeholder values for PostgreSQL and token signing. Do not store real credentials in this repository.

When publishing the runtime image, inject runtime secrets through environment variables or your deployment secret manager instead of baking them into the image.

Create a local account before calling authenticated routes:

```powershell
$auth = Invoke-RestMethod -Uri http://localhost:7659/v1/auth/register -Method Post -ContentType "application/json" -Body "{\"username\":\"local_user\",\"email\":\"local@example.com\",\"password\":\"password123\"}"
curl -H "Authorization: Bearer $($auth.session.access_token)" http://localhost:7659/v1/session
```
