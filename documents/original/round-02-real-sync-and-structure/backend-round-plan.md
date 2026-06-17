# Backend Round Plan

## Context

The deployment-level Round 02 guidance identifies two backend pressure points:

- `internal/service/sync_upload.go` concentrates sync mutation decoding, normalization, validation, semantic transition handling, idempotency, and repository error mapping.
- `internal/repository/repository.go` concentrates account, session, synced data, operation history, and settings SQL.

The backend is already close to the desired sync contract. The round should protect that contract while the frontend moves from fixture-first state to real PowerSync reads and writes.

## Workstream 1: Preserve Sync Contract Compatibility

Keep `/v1/sync/token` and `/v1/sync/upload` stable while frontend writes move to real synced local tables.

Expected backend behavior:

- Authenticated user scope remains authoritative.
- Uploaded `user_id` values are normalized or rejected according to existing ownership rules.
- `items`, `areas`, `recurring_task_templates`, `operation_history`, and `user_settings` writes keep their current validation behavior.
- Semantic item changes continue to require matching action intent through `operation_history` or observed-state metadata where the service currently requires it.

Avoid widening the contract simply to make an incomplete frontend mutation pass.

## Workstream 2: Registration Onboarding Seed Data

New registered accounts should receive a small set of real synced rows so a fresh account can demonstrate the app model without frontend fixtures.

Rules:

- Seed rows are created in the same transaction as account registration.
- Seed rows go into synced tables, currently `items`.
- Seed rows must pass the same schema constraints as uploaded rows.
- Seed content should teach the data model, not simulate a user's private life.
- Seed data must not depend on frontend-only fixture IDs.

See `onboarding-seed-data.md`.

## Workstream 3: Backend Structure Split

Use the previous optimization guidance as the baseline, but keep this round incremental.

Recommended delivery order:

1. Split `internal/service/sync_upload.go` by aggregate and shared validation concerns.
2. Split `internal/repository/repository.go` by transaction shell, account/session SQL, and synced-data SQL.
3. Add focused tests only where extraction makes coverage less obvious.

Do not combine broad file splitting with unrelated behavior changes.

## Workstream 4: Verification Support

Backend verification should support the deployment-level end-to-end playbook:

- New account registration creates default synced settings and onboarding items.
- Sync token requests work for the registered account.
- Real non-empty sync upload mutations are accepted or rejected with stable domain errors.
- Accepted writes appear in PostgreSQL and remain streamable by PowerSync.

Repository-local tests should stay focused. Full multi-device behavior belongs to the deployment stack verification.
