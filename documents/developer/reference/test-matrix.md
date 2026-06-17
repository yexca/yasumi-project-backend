# Test Matrix

## What This Document Solves

This document maps common backend changes to the verification layers that should move with them.

## When To Read It

- Before finishing a backend change.
- When deciding whether focused tests are enough or wider verification is needed.

## Current Test Layers

- `internal/httpapi/router_test.go`
  - route behavior, auth requirement, error shape, metrics, log safety
- `internal/config/config_test.go`
  - config parsing and validation
- `internal/repository/repository_integration_test.go`
  - schema constraints, indexes, user scoping, repository behavior
- `internal/service/sync_upload_test.go`
  - sync upload acceptance and rejection behavior
- `internal/auth/account_integration_test.go`
  - registration, sessions, auth persistence flows
- `internal/domain/*_test.go`
  - pure validation and semantic rules
- `internal/synctoken/issuer_test.go`
  - sync token issuing behavior

## By Change Type

- HTTP route change:
  - router tests
  - use-case tests if behavior changed
- Auth flow change:
  - auth integration tests
  - router tests if HTTP-visible behavior changed
- Sync upload change:
  - sync upload tests
  - repository integration tests if persistence rules changed
  - router tests if HTTP-visible behavior changed
- Migration or schema change:
  - repository integration tests
  - affected service or auth tests
- Config change:
  - config tests
  - startup or route verification when runtime behavior changes
- Logging or metrics change:
  - router tests

## Minimum Verification Rule

Before considering a change complete:

- run tests closest to the changed boundary
- run broader tests when schema, sync, auth, or startup wiring changed
- update docs when the durable rule changed

## Related Documents

- `../how-to/run-and-verify-locally.md`
- `../how-to/contribution-workflow.md`
- `database-schema-and-migration-rules.md`
