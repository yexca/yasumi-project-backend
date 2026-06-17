# Sync Model

## What This Document Solves

This document explains how synced backend writes enter the service today and which boundaries must stay stable when sync behavior changes.

## When To Read It

- Before changing `internal/service`.
- Before changing synced table rules in `internal/domain` or `internal/repository`.
- Before changing onboarding seed data or sync upload acceptance behavior.

## Current Model

The backend does not expose direct CRUD routes for MVP synced business data.

Current synced writes enter through:

- `POST /v1/sync/upload`

Current sync token issuance enters through:

- `POST /v1/sync/token`

Authenticated account flows live separately in `internal/auth` and still shape synced data indirectly because registration creates default synced rows.

## Current Synced Tables

The sync-facing data model currently includes:

- `areas`
- `recurring_task_templates`
- `items`
- `operation_history`
- `user_settings`

The route layer accepts upload payloads, but the service and repository layers decide whether a row is accepted.

## Ownership And Authority

Current ownership rules are strict:

- The authenticated backend user scope is authoritative.
- Client-supplied cross-user writes are normalized or rejected.
- Repository methods read and write synced rows under explicit user scope.

This rule appears across:

- `internal/httpapi`
- `internal/service`
- `internal/repository`
- `internal/auth`

Do not relax ownership rules in handlers to make a client mutation pass.

## Registration Seeds Real Synced Data

Account registration in `internal/auth` creates:

- `users`
- `user_credentials`
- `user_settings`
- onboarding `items`
- `user_sessions`

The current onboarding seed rules are durable because they already shape real account behavior:

- seed rows are created in the registration transaction
- rows use the real `user_id`
- device fields use `account-registration`
- revision starts at `1`
- dates derive from the registration time in Tokyo timezone

The historical decision record remains in `../../original/round-02-real-sync-and-structure/onboarding-seed-data.md`.

## Accepted Sync Write Shape

`internal/service` currently treats sync upload as:

1. Validate upload envelope and device context.
2. Detect duplicate semantic operations through `operation_history`.
3. Decode row payloads by table.
4. Normalize ownership and server-managed fields.
5. Validate structural and semantic constraints.
6. Upsert or append data inside one transaction.

Current acceptance is table-specific:

- `areas`
- `recurring_task_templates`
- `items`
- `operation_history`
- `user_settings`

## Important Stability Rules

- Keep `POST /v1/sync/token` and `POST /v1/sync/upload` stable unless the contract change is intentional and documented.
- Keep business logic in `internal/service` and `internal/domain`, not in `internal/httpapi`.
- Keep SQL ownership and persistence details in `internal/repository`.
- Keep `operation_history` append-only.
- Keep idempotency behavior explicit when semantic actions can repeat.

## Change Path

When sync behavior changes:

1. Update domain validation or semantic rules in `internal/domain` if needed.
2. Update aggregate-specific acceptance in `internal/service`.
3. Update repository records and SQL in `internal/repository`.
4. Update migrations only when the persisted schema must change.
5. Update tests and this guide if the durable rule changed.

## How To Verify

- Run router, service, and repository tests relevant to the changed table.
- Verify accepted writes remain scoped to the authenticated user.
- Verify duplicate semantic operations converge through current idempotency rules.
- Verify registration still creates valid synced defaults when the change affects onboarding data.

## Related Documents

- `architecture.md`
- `data-and-migrations.md`
- `../how-to/extend-sync-upload.md`
- `../reference/api-surface.md`
- `../reference/database-schema-and-migration-rules.md`
