# Extend Sync Upload

## What This Document Solves

This document shows how to extend `POST /v1/sync/upload` for a synced table or a new acceptance rule without turning the sync path into an unreviewable blob.

## When To Read It

- Before editing `internal/service/sync_upload*.go`.
- Before adding repository support for a synced table.
- Before changing semantic acceptance rules or idempotency behavior.

## Current Rule

The sync upload path is intentionally layered:

- `internal/httpapi`
  - request decoding and HTTP error mapping
- `internal/service`
  - mutation dispatch, normalization, semantic checks, orchestration
- `internal/domain`
  - validation and semantic rule helpers
- `internal/repository`
  - scoped reads, upserts, append-only writes

Keep aggregate-specific logic in the split `sync_upload_*.go` files instead of rebuilding a large single-file flow.

## Change Path

1. Confirm the table belongs in the current synced model.
2. Add or update row records and scoped repository methods in `internal/repository`.
3. Add aggregate-specific decode, normalize, validate, and accept logic in `internal/service`.
4. Use `internal/domain` for reusable structural or semantic rules.
5. Keep server-managed fields explicit:
   - ownership
   - revision
   - timestamps
   - device ids
   - idempotency keys when applicable
6. Update route, service, and repository tests.
7. Update durable documentation if the long-term rule changed.

## Current Table Coverage

The service currently accepts mutations for:

- `areas`
- `recurring_task_templates`
- `items`
- `operation_history`
- `user_settings`

If you add another synced table later, document why it belongs in this contract.

## Structure Rules

- Split by concern before introducing a new package.
- Do not move sync business rules into handlers.
- Do not hide ownership checks inside generic helpers that obscure the table rule.
- Keep append-only `operation_history` semantics explicit.
- Keep duplicate semantic action handling reviewable.

## How To Verify

- Add or update service tests for accepted, rejected, and duplicate cases.
- Run router tests for error shape stability when HTTP-visible behavior changes.
- Run repository integration tests if a new SQL path or schema rule is introduced.
- Verify the authenticated user remains authoritative over the accepted rows.

## Related Documents

- `../concepts/sync-model.md`
- `../concepts/architecture.md`
- `add-migration.md`
- `../reference/database-schema-and-migration-rules.md`
- `../reference/test-matrix.md`
