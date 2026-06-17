# Phase 03 Development Notes

Date: 2026-06-13

## Scope

Phase 03 implements the PostgreSQL schema, embedded migration runner, and repository baseline for Yasumi synced user data.

This phase does not expose public CRUD endpoints and does not implement full sync upload acceptance.

## Implemented

- Embedded SQL migration runner in `internal/migrations`.
- Initial PostgreSQL schema for `items`, `recurring_task_templates`, `areas`, `operation_history`, and `user_settings`.
- Minimum database constraints for enum-like values, non-empty titles, deadline shapes, recurrence positivity, recurring instance uniqueness, operation idempotency uniqueness, and settings ranges.
- User-scoped indexes for the MVP synced tables.
- Append-only enforcement for `operation_history` through database triggers.
- PostgreSQL repository package with connection setup, transaction support, user-scoped item locking, operation history insert/lookup, item server metadata update, and user settings upsert.
- `yasumi-migrate` command for deployment and local Docker use.
- Docker image now includes both `yasumi-api` and `yasumi-migrate`.
- Compose now runs migrations before the API service.

## Migration Tooling

Migrations are plain SQL files embedded into the Go binary and applied through `cmd/yasumi-migrate`.

Rollback stance:

- Local and test databases may drop the migrated schema when resetting data.
- Production rollback should restore from backup because synced user data may already exist.
- Future schema changes should add forward migrations and document compatibility notes when they affect PowerSync/local storage shape.

## Verification

Ran:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
```

Integration coverage verifies:

- Migrations apply to an empty schema.
- Invalid item type and empty titles are rejected.
- Valid and invalid deadline shapes behave as expected.
- Duplicate generated recurring instances are rejected.
- Duplicate non-null operation idempotency keys are rejected.
- Invalid recurrence intervals are rejected.
- `operation_history` rejects update and delete attempts.
- Repository item lookup is scoped by authenticated `user_id`.
- Operation idempotency lookup and settings upsert work through repository transactions.
