# Database Schema And Migration Rules

## What This Document Solves

This document records the durable database rules that backend changes should preserve.

## When To Read It

- Before changing synced-table shape or SQL constraints.
- Before editing repository SQL or adding a migration.

## Current Table Groups

Account and session tables:

- `users`
- `user_credentials`
- `user_sessions`

Synced tables:

- `areas`
- `recurring_task_templates`
- `items`
- `operation_history`
- `user_settings`

## Durable Rules

- Persisted data is user-scoped.
- Repository access should be explicit about user scope.
- Cross-user references must stay invalid.
- `operation_history` remains append-only.
- Semantic duplicate handling relies on stable idempotency rules.
- Synced rows keep server-managed metadata such as revision, timestamps, and device ids.

## Migration Rules

- Add new ordered `.up.sql` files only.
- Keep each migration tied to a real backend change.
- Update repository records and SQL in the same change.
- Update tests when a constraint, index, or scope rule changes.
- Document lasting schema rules in `developer/`, not only in round documents.

## Repository Alignment

Schema changes commonly require updates in:

- `internal/repository/records.go`
- account-specific repository files
- sync-specific repository files
- service code that sets server-managed fields

## Verification Rules

- Apply migrations from an empty database.
- Run repository integration tests.
- Re-check sync upload or auth behavior when the changed tables participate in those flows.

## Related Documents

- `../concepts/data-and-migrations.md`
- `../how-to/add-migration.md`
- `test-matrix.md`
