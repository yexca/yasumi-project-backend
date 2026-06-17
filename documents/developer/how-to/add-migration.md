# Add A Migration

## What This Document Solves

This document shows how to add a database migration that stays aligned with repository code, sync rules, and verification expectations.

## When To Read It

- Before editing `internal/migrations/sql`.
- Before changing repository records or database constraints.

## Current Rule

The backend uses embedded forward migrations only. Schema changes must stay incremental, reviewable, and tied to a real backend use case.

## Change Path

1. Decide the exact schema rule that needs to change.
2. Add a new ordered `.up.sql` file under `internal/migrations/sql`.
3. Update repository records and SQL in `internal/repository`.
4. Update service or auth logic if server behavior depends on the new field or constraint.
5. Add or update repository integration tests.
6. Update reference docs if the durable database rule changed.

## Rules For Writing The Migration

- Keep statements focused on the current behavior change.
- Prefer explicit constraints and indexes over relying on application code alone.
- Preserve user-scoped integrity rules.
- Preserve append-only rules where they already exist.
- Do not mix unrelated cleanup with a behavior change migration.

## Current Hotspots

Migration changes often affect:

- synced data constraints in `items`, `areas`, `recurring_task_templates`, `operation_history`, or `user_settings`
- account and session flows in `users`, `user_credentials`, or `user_sessions`
- repository integration tests that assert indexes, uniqueness, and foreign-key rules

## How To Verify

- Apply migrations to an empty database.
- Run repository integration tests.
- Run affected service or auth tests.
- If a synced table changed, verify the sync upload path still accepts valid writes and rejects invalid ones.

## Related Documents

- `../concepts/data-and-migrations.md`
- `extend-sync-upload.md`
- `change-configuration.md`
- `../reference/database-schema-and-migration-rules.md`
