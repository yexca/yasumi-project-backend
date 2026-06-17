# Phase 05 Verification

Date: 2026-06-13

## Scope

This verification reviews the current Phase 05 implementation against `backend_coding_guide/06-phase-05-sync-validation.md` and its referenced sync contracts.

Phase 05 is expected to provide a safe backend acceptance boundary for local-first sync uploads, including ownership enforcement, domain validation, semantic write validation, idempotency, server metadata assignment, and fixture-backed acceptance coverage for MVP sync behavior.

## Result

Phase 05 verification passed on 2026-06-13 after the post-Phase-06 remediation work.

The current backend accepts MVP synced table uploads through the local-first sync boundary, rejects cross-user writes, validates row shape through domain rules, validates semantic item changes against companion operation intent or observed server state, handles duplicate operation idempotency keys as retries, assigns server timestamps and revisions from accepted server state, and keeps accepted row changes plus operation history in one transaction.

## Verified Working

- `POST /v1/sync/upload` is implemented, authenticated, and returns the stable accepted DTO for valid requests.
- Upload acceptance covers `areas`, `recurring_task_templates`, `items`, `operation_history`, and `user_settings`.
- Missing row `user_id` is normalized to the authenticated account user, and mismatched row `user_id` is rejected with `forbidden`.
- Item, area, recurring template, and settings writes use shared domain validation.
- Item, area, recurring template, and settings writes assign `updated_at`, `server_updated_at`, and server-owned `revision`.
- Semantic item updates require companion operation intent or matching `client_observed` server state.
- Invalid observed revisions and mismatched restore/delete/archive operation intent are rejected with stable domain errors.
- `operation_history` remains append-only, and duplicate non-null idempotency keys return the existing accepted result.
- Due postponed activation is checked for due date eligibility and idempotency key shape.
- User settings revision is incremented from the accepted server row, not trusted from the client payload.
- PowerSync Sync Streams are scoped by `auth.user_id()` for MVP synced tables.
- Sync JWT issuance uses the authenticated internal `users.id` as `sub` and stream scope.

## Evidence

- `internal/service/sync_upload.go` routes all MVP synced tables and wraps upload acceptance in `Repository.InTx`.
- `internal/service/sync_upload_test.go` covers cross-user rejection, semantic intent requirements, observed revision mismatch, area and recurring template acceptance, settings revision assignment, restore intent validation, postponed activation, and duplicate idempotency behavior.
- `internal/repository/repository.go` scopes user-data reads and upserts by `user_id`.
- `env/powersync/sync-config.yaml` defines user-scoped Sync Streams for synced MVP tables.
- `internal/synctoken/issuer_test.go` validates sync token claims and subject scoping.

## Commands Run

Docker-based Go toolchain:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go vet ./...
```

Deployment build:

```powershell
docker build -t yasumi-backend-phase56-verify .
```

Runtime probes against the compose API:

```powershell
docker compose -f .\env\docker-compose.yml up -d api
```

Observed runtime results:

- `/healthz` returned `{"status":"ok"}`.
- Account registration returned an access token and refresh token.
- `/v1/session` accepted the account access token and returned the registered account user.
- `/v1/sync/token` returned a sync token with `stream_scope.user_id` equal to the registered internal `users.id`.
- `/v1/sync/upload` accepted an empty mutation batch and returned `accepted: []`.
- `/readyz` returned `not_ready` with `database: ok` and `sync: unavailable` because PowerSync was not running in this verification pass.

## Residual Limitations

- PowerSync service itself was not started in this pass; stream rules were reviewed from configuration and sync-token scope was verified through the API.
- Coverage is fixture-informed and targets the documented Phase 05 scenarios, but the backend tests do not dynamically load every JSON fixture under `dev_documents/contracts/fixtures`.

## Sign-off

Phase 05 is accepted for the current backend state and can remain a precondition for Phase 07 hardening.
