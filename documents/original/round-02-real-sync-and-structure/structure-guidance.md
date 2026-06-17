# Structure Guidance

This note integrates the earlier backend structure optimization guidance into the active Round 02 plan.

The backend structure is healthy enough to evolve incrementally. The goal is not to redesign the project; it is to reduce pressure in the files most likely to receive real-sync changes.

## Integrated Assessment

The earlier structure review found that the backend already has useful boundaries:

- Entrypoints, assembly, protocol, business, and persistence layers are separated.
- Runtime code stays close to the Go standard library, which keeps the service lightweight.
- Sync write rules are concentrated in `service` and `domain`, which is the right direction.
- SQL remains explicit in the repository layer, which helps debugging and operations.

The same review identified two main pressure points for Round 02:

- `internal/service/sync_upload.go` combines decoding, normalization, validation, semantic transitions, idempotency, and write orchestration.
- `internal/repository/repository.go` combines account, session, synced data, operation history, and settings SQL.

The recommended response is still local file splitting, not a package rewrite.

## Keep

- The current handler, service, domain, repository, and migration layers.
- Explicit SQL in the repository package.
- Go standard-library-first implementation style.
- The current sync upload contract and domain error shape.

## Priority 1: Split Sync Upload Logic

`internal/service/sync_upload.go` should be split when sync-facing changes become harder to review in place.

Suggested file ownership:

- `sync_upload_service.go`
  - Upload entrypoint, transaction flow, mutation dispatch.
- `sync_upload_items.go`
  - `items` acceptance and semantic item update handling.
- `sync_upload_areas.go`
  - `areas` acceptance.
- `sync_upload_recurring_templates.go`
  - `recurring_task_templates` acceptance.
- `sync_upload_operations.go`
  - `operation_history` append behavior and idempotency checks.
- `sync_upload_settings.go`
  - `user_settings` acceptance.
- `sync_upload_validation.go`
  - Shared validation, ownership normalization, and error mapping helpers.

Keep all of these in the existing `service` package unless a stronger reason appears.

## Priority 2: Split Repository SQL

`internal/repository/repository.go` should be split by concern, not by introducing a new abstraction layer.

Suggested file ownership:

- `repository.go`
  - `Repository`, `Tx`, `InTx`, and shared errors.
- `account_repository.go`
  - Users, credentials, sessions, and account registration SQL.
- `sync_items_repository.go`
  - `items` and `operation_history` SQL.
- `sync_meta_repository.go`
  - `areas`, `recurring_task_templates`, and `user_settings` SQL.

Keep the public package usage stable.

## Priority 3: Keep Assembly Thin

`internal/app/app.go` can stay as-is until dependency construction grows enough to obscure the server setup.

If needed later, add a lightweight `internal/app/providers.go` for dependency construction. Do not introduce a DI framework.

## Deferred Decoupling

The service layer currently uses `repository.*Record` types as practical data carriers. That is acceptable for Round 02 while the sync contract is being stabilized.

Only introduce smaller use-case inputs, repository interfaces, or conversion helpers when a concrete change needs them. Do not add a broad DTO layer that only mirrors table records.

## Delivery Order

1. Keep document entry points clear and aligned with the deployment repository.
2. Split `internal/service/sync_upload.go` when touching sync upload behavior.
3. Split `internal/repository/repository.go` when touching account, session, or synced-data SQL.
4. Add provider helpers or smaller use-case structures only if complexity appears in the changed area.

## Avoid During Round 02

- Moving sync logic into HTTP handlers.
- Replacing explicit SQL with an ORM.
- Splitting `internal` into many small packages for style alone.
- Creating broad DTO layers that only mirror repository records.
- Combining large structure splits with product behavior changes.

## Practical Rule

When a change touches real sync behavior, first preserve behavior with tests, then split only the concern you are already modifying. Structure work should make the next sync change safer, not become its own migration project.
