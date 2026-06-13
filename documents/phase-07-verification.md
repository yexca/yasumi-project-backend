# Phase 07 Verification

Date: 2026-06-13

## Scope

This verification reviews the current backend against `backend_coding_guide/08-phase-07-hardening-release.md`.

## Result

Phase 07 verification passed on 2026-06-13.

The backend now has release-hardening coverage for fixture compatibility, offline conflict behavior, PowerSync stream authorization rules, account behavior, production-safe logging, metrics, dependency readiness, database indexes, calendar guardrails, and deployment documentation.

## Verified Working

- `/metrics` exposes request, error, auth failure, validation rejection, sync upload result, and dependency readiness metrics.
- Request logging records operational labels only and does not log request bodies, authorization headers, passwords, raw tokens, item titles, or notes.
- `/readyz` checks PostgreSQL and PowerSync and updates dependency readiness gauges.
- Backend constants cover the shared contract values in `shared_constants/`.
- Backend item and deadline validation matches the shared data fixture.
- Fixture compatibility tests read `dev_documents` through `YASUMI_CONTRACTS_DIR` in the Docker toolchain. If an external contract directory is explicitly configured but missing, the tests fail clearly.
- PowerSync streams for `areas`, `recurring_task_templates`, `items`, `operation_history`, and `user_settings` are scoped by `auth.user_id()`.
- Offline restore preserves completed status and clears delete/archive metadata only through a restored operation.
- Delete-versus-edit conflict keeps the tombstone until explicit restore.
- Duplicate postponed activation returns the existing accepted result.
- Expected MVP query indexes are present after migrations.
- Active backend endpoints, migrations, and environment configuration do not implement calendar integration.
- Deployment operations and release checklist are documented.

## Commands Run

Docker-based Go toolchain:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

PostgreSQL-backed integration tests:

```powershell
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/auth -count=1 -v
```

Deployment build:

```powershell
docker build -t yasumi-backend-phase7-verification .
```

## Residual Limitations

- Login and registration rate limiting remains deferred.
- `/metrics` is public in the local service and should be restricted by deployment network or ingress controls.
- Email verification remains deferred for MVP.
- Calendar integration remains intentionally absent.

## Sign-off

Phase 07 is accepted for the current backend state. The backend is ready for MVP integration after deployment-specific secrets, network restrictions, and PowerSync production configuration are applied.
