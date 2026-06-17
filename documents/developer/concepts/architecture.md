# Backend Architecture

## What This Document Solves

This document explains the current backend structure, the dependency direction that keeps it maintainable, and the boundaries that should remain clear when new work is added.

## When To Read It

- Before changing code across more than one backend package.
- Before introducing a new route, migration, or sync rule.
- Before reorganizing files in `internal/`.

## Current Structure

The backend is a small Go service built mostly on the standard library, explicit PostgreSQL access, and a narrow HTTP surface.

- `cmd/yasumi-api`
  - API process entrypoint.
- `cmd/yasumi-migrate`
  - Migration process entrypoint.
- `internal/app`
  - Application assembly for config, logging, repository wiring, routes, services, and readiness checks.
- `internal/httpapi`
  - HTTP routing, middleware, request parsing, response writing, and HTTP-visible error mapping.
- `internal/auth`
  - Account registration, login, refresh, logout, profile updates, password changes, access token authentication, and onboarding seed creation.
- `internal/service`
  - Use-case logic outside the auth boundary, currently centered on sync upload acceptance.
- `internal/repository`
  - PostgreSQL transaction shell, row records, and explicit SQL.
- `internal/domain`
  - Stable domain errors, validation rules, and semantic state-transition helpers.
- `internal/config`
  - Startup configuration parsing and validation.
- `internal/migrations`
  - Embedded migration runner and SQL files.
- `internal/synctoken`
  - PowerSync-compatible sync token issuing.
- `internal/telemetry`
  - Logging and metrics support.
- `env/`
  - Local environment and toolchain support files.

## Dependency Direction

Keep the current chain:

1. `cmd/*` starts processes only.
2. `internal/app` assembles dependencies only.
3. `internal/httpapi` converts HTTP input and output.
4. `internal/auth` and `internal/service` own use-case behavior.
5. `internal/domain` owns reusable validation and semantic rules.
6. `internal/repository` owns persistence details.

This keeps business behavior out of handlers, keeps SQL out of route code, and keeps tests aimed at the correct layer.

## Current Architectural Shape

Two use-case centers matter today:

- `internal/auth`
  - account lifecycle and initial synced account data
- `internal/service`
  - sync upload acceptance across synced tables

That split is healthy for the current codebase. Do not force all business behavior into one catch-all package just for symmetry.

## Development Principles

### Cohesion

- Keep protocol, use-case, validation, and persistence changes grouped around a clear backend concern.
- Prefer existing package boundaries before inventing a new package.

### Low Coupling

- `httpapi` should not write SQL directly.
- `service` and `auth` should not depend on HTTP request or response types.
- `repository` should not decide business state transitions.
- `app` should stay thin and assembly-focused.

### Lightweight Changes

- Reuse the current directory and package style.
- Add abstractions only when they remove real duplication or isolate real complexity.
- Prefer local file splitting over package rewrites.

## Current Pressure Points

The codebase already addressed earlier round pressure by splitting:

- sync upload behavior into `internal/service/sync_upload_*.go`
- repository SQL by concern in `internal/repository/*_repository.go`

Keep extending those local structures instead of rebuilding them into a broader framework.

## Avoid

- Business logic inside handlers.
- Hidden state-transition decisions inside repository methods.
- Service methods that accept HTTP-specific request or response objects.
- Broad `util` packages with unclear semantics.
- Large structural rewrites that do not pay for themselves in current work.

## Related Documents

- `sync-model.md`
- `data-and-migrations.md`
- `../how-to/add-authenticated-endpoint.md`
- `../how-to/extend-sync-upload.md`
