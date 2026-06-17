# Data And Migrations

## What This Document Solves

This document explains how persistent backend data is organized and what rules protect schema changes from drifting away from current code behavior.

## When To Read It

- Before adding or changing SQL migrations.
- Before changing repository records or query methods.
- Before changing synced-table constraints or indexes.

## Current Persistence Structure

The backend uses explicit PostgreSQL access in `internal/repository` and embedded SQL migrations in `internal/migrations`.

Key runtime pieces:

- `internal/repository`
  - transaction shell, account SQL, synced-data SQL, row records
- `internal/migrations`
  - embedded migration runner
- `internal/migrations/sql`
  - ordered `.up.sql` files

## Current Table Groups

Account and session tables:

- `users`
- `user_credentials`
- `user_sessions`

Synced business tables:

- `areas`
- `recurring_task_templates`
- `items`
- `operation_history`
- `user_settings`

The schema is intentionally user-scoped. Cross-user references are rejected by constraints and repository access patterns.

## Why The Repository Stays Explicit

The current codebase benefits from explicit SQL because it keeps:

- ownership rules reviewable
- sync write behavior debuggable
- index needs visible
- migration-to-query alignment straightforward

Do not replace explicit repository SQL with a broad ORM layer as part of normal feature work.

## Migration Rules

- Add only forward `.up.sql` migrations under `internal/migrations/sql`.
- Keep migrations incremental and tied to an actual use case.
- Add only fields, constraints, or indexes required by current behavior.
- Update repository records and SQL in the same change when persisted shape changes.
- Update documentation when operators or developers need to know the new rule.

## Index And Constraint Intent

The current schema uses indexes and constraints to protect:

- per-user lookup paths
- visibility and status queries
- recurring instance uniqueness
- idempotency uniqueness
- cross-user foreign-key safety
- append-only history behavior

Repository integration tests already enforce several of these guarantees. Treat those tests as part of the schema contract.

## Change Path

1. Decide whether the rule belongs in SQL, domain validation, or both.
2. Add the migration file.
3. Update repository records and repository methods.
4. Update service logic if server-managed state or acceptance flow changes.
5. Add or update integration tests.
6. Update developer reference docs when the durable rule changes.

## How To Verify

- Run migration application against an empty database.
- Run repository integration tests for constraints, indexes, and scope rules.
- Verify the changed queries still use the intended user-scoped access path.
- Verify sync upload behavior if the schema change affects synced tables.

## Related Documents

- `architecture.md`
- `sync-model.md`
- `../how-to/add-migration.md`
- `../reference/database-schema-and-migration-rules.md`
- `../reference/test-matrix.md`
