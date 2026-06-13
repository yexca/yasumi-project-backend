# Phase 04 Verification

Date: 2026-06-13

## Scope

This verification reviews the Phase 04 authentication and direct API boundary against `backend_coding_guide/05-phase-04-auth-direct-api.md`.

Phase 04 remains limited to session context, sync token issuance, health, readiness, middleware, and stable HTTP error responses. MVP synced business data CRUD routes remain out of scope.

## Result

Phase 04 verification passed.

Verified behaviors:

- `GET /healthz` is public and returns liveness.
- `GET /readyz` reports dependency status without sensitive details.
- Protected routes return the stable `unauthorized` error shape without a bearer token.
- `GET /v1/session` returns the authenticated local development user context.
- `POST /v1/sync/token` requires authentication and `device_id`.
- Sync token scope is derived from the authenticated user, ignoring client-supplied `user_id`.
- API validation errors use the stable `{ code, message, fields, retryable }` DTO shape.
- First-class MVP business CRUD routes for items, areas, recurring templates, operation history, and user settings are absent.
- Database integration checks still pass after Phase 04 changes.
- The production-style Docker image builds both `yasumi-api` and `yasumi-migrate`.

## Commands Run

Docker-based Go toolchain:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

Repository integration tests with PostgreSQL:

```powershell
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
```

Docker deployment build:

```powershell
docker build -t yasumi-backend-phase4-verify .
docker compose -f .\env\docker-compose.yml config
```

Runtime HTTP probes were executed against a temporary API container mapped to host port `18080` because host port `8080` was already allocated by an unrelated local service:

```powershell
curl http://localhost:18080/healthz
curl -H "Authorization: Bearer local-dev-session-token" http://localhost:18080/v1/session
curl -X POST -H "Authorization: Bearer local-dev-session-token" -H "Content-Type: application/json" -d "{\"device_id\":\"device-01\",\"client_version\":\"0.1.0\",\"user_id\":\"00000000-0000-4000-8000-999999999999\"}" http://localhost:18080/v1/sync/token
curl http://localhost:18080/readyz
```

Observed runtime results:

- `/healthz` returned `{"status":"ok"}`.
- `/v1/session` returned authenticated user `00000000-0000-4000-8000-000000000001`.
- `/v1/sync/token` returned `stream_scope.user_id` as the authenticated user, not the client-supplied user.
- `/readyz` returned `503` with `database: ok` and `sync: unavailable` when PowerSync was not running.

After this verification pass, the backend default HTTP port was changed from `8080` to `7650` to avoid the local `8080` conflict. Runtime configuration, Docker exposure, compose port mapping, and local documentation were updated together.

The updated compose API service was also started on `http://localhost:7650`; `/healthz`, `/v1/session`, `/v1/sync/token`, and `/readyz` behaved as expected on the new default port.

## Notes

Host-level Go tooling was not available on this machine, so verification used the repository-local Docker toolchain under `env/`.

`go test -race ./...` was attempted through the Alpine toolchain image but did not run because race detection requires cgo. The standard Phase 04 checks (`gofmt`, `go test`, `go vet`) passed through Docker.

No additional root-level files needed to be moved to `legacy/`. Active project files are under `cmd/`, `internal/`, `env/`, and `documents/`; the previous Gin/Gorm implementation is already preserved under `legacy/old-gin-gorm`.
