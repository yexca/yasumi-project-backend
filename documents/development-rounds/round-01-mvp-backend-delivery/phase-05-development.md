# Phase 05 Development Notes

Date: 2026-06-13

## Scope

Phase 05 implements the backend acceptance boundary for local-first sync uploads.

This phase keeps MVP business data writes on the sync path. It does not add first-class REST CRUD endpoints for items, areas, recurring templates, operation history, or user settings.

## Implemented

- Authenticated sync upload adapter endpoint at `POST /v1/sync/upload`.
- Sync upload service under `internal/service` for ownership checks, domain row validation, semantic item transition validation, idempotency detection, server timestamps, and revision assignment.
- `areas` and `recurring_task_templates` upload acceptance through the same authenticated sync boundary.
- User-owned item upsert repository method with explicit PostgreSQL columns and row-level user scoping.
- Append-only `operation_history` acceptance with duplicate non-null idempotency keys treated as retries.
- `user_settings` upload acceptance through the same authenticated sync boundary.
- Stable HTTP error mapping for sync upload validation failures.
- PowerSync Sync Streams configuration scoped by `auth.user_id()` for MVP synced tables.
- PowerSync-compatible HS256 JWT sync tokens with `sub`, `aud`, and local development `kid` claims.
- Service tests for missing `user_id` normalization, cross-user rejection, semantic intent rejection, semantic updates with companion operation history, area and recurring template writes, settings revision assignment, restore intent validation, postponed activation, and duplicate idempotency keys.
- Repository integration coverage for item upsert and server-accepted metadata.
- HTTP boundary tests for authenticated sync upload and stable error shape.

## Sync Upload Contract

The current adapter endpoint accepts the conceptual envelope from `interface_design/03-sync-write-interface.md`:

```json
{
  "client_batch_id": "batch-01",
  "device_id": "device-01",
  "base_revision": 42,
  "mutations": [
    {
      "table": "items",
      "op": "upsert",
      "row": {},
      "client_observed": {
        "revision": 41,
        "status": "active",
        "deleted_at": null,
        "archived_at": null,
        "hidden_reason": null
      }
    }
  ]
}
```

Rows without `user_id` are normalized to the authenticated user. Rows with a different `user_id` are rejected as `forbidden`.

PowerSync stream authorization uses the issued JWT subject as the authenticated user identity. The local development PowerSync config trusts the matching `local-dev-sync-key` HS256 key.

## Semantic Write Rules

For item updates, changes to `status`, `deleted_at`, `archived_at`, or `hidden_reason` require companion `operation_history` or observed baseline metadata. Status transitions are validated against the server-visible previous row state.

Physical delete operations are rejected as `unsupported_operation`.

## Out of Scope

- Full recurrence next-instance generation.
- Field-level conflict review.
- Calendar integration.
- UI recovery copy.

## Verification

Run:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
```

Repository integration tests with PostgreSQL:

```powershell
docker compose -f .\env\docker-compose.yml up -d postgres
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository -count=1 -v
```
