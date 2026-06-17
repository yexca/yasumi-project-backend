# Run And Verify Locally

## What This Document Solves

This document shows how to verify backend changes locally using the current repository workflows.

## When To Read It

- Before submitting a backend change.
- When you need to prove a route, migration, or sync rule still works.

## Current Verification Paths

### Fast code-level checks

Use the local Go toolchain when available:

```powershell
go test ./...
go fmt ./...
go vet ./...
```

Use the Docker-based toolchain when Go is not installed locally:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

### Repository integration checks

```powershell
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
```

### Full local stack

```powershell
Copy-Item .env.example .env
docker compose -f .\docker-compose.example.yml up --build
```

Useful endpoints:

- `http://localhost:7659/healthz`
- `http://localhost:7659/readyz`
- `http://localhost:7659/metrics`

## What To Verify By Change Type

- Route change:
  - router behavior, auth behavior, stable error shape
- Sync upload change:
  - accepted and rejected mutations, user scoping, duplicate handling
- Migration change:
  - migration apply, repository integration constraints, affected SQL paths
- Configuration change:
  - config parsing, startup validation, example env files
- Registration or auth change:
  - register, login, refresh, logout, session, onboarding seed behavior

## Suggested Local Flow

1. Run focused tests for the changed area.
2. Run broader package or repository tests if the blast radius is wider.
3. Bring up the full stack when the change affects sync, readiness, or startup wiring.
4. Update documentation before considering the change complete.

## Related Documents

- `contribution-workflow.md`
- `change-configuration.md`
- `../reference/test-matrix.md`
- `../../README.md`
