# Secondary Development Architecture and Extension Notes

## Current Structure

The backend is a lightweight Go service built mostly on the standard library and explicit PostgreSQL access.

- `cmd/yasumi-api`
  - API process entrypoint.
- `cmd/yasumi-migrate`
  - Database migration entrypoint.
- `internal/app`
  - Application assembly for configuration, logging, database connections, routes, and services.
- `internal/httpapi`
  - HTTP routing, middleware, request parsing, and response writing.
- `internal/auth`
  - Account registration, login, sessions, access tokens, and password handling.
- `internal/service`
  - Use-case logic, currently centered on sync upload acceptance.
- `internal/repository`
  - PostgreSQL transaction and SQL boundary.
- `internal/domain`
  - Domain constants, validation rules, state transitions, and stable errors.
- `internal/synctoken`
  - PowerSync-compatible sync token issuing.
- `internal/telemetry`
  - Logging and metrics support.
- `internal/migrations`
  - Embedded SQL migrations.

## Dependency Direction

Keep the current chain:

1. `cmd/*` starts processes only.
2. `internal/app` assembles dependencies only.
3. `internal/httpapi` converts HTTP input and output.
4. `internal/service` owns use-case rules and orchestration.
5. `internal/domain` owns pure validation and semantic constraints.
6. `internal/repository` owns persistence details.

This keeps business behavior in `service` and `domain`, keeps HTTP separate from SQL, and makes tests easier to aim at the right boundary.

## Development Principles

### Cohesion

- Keep one feature's protocol, use case, validation, and persistence changes grouped around a clear domain concept.
- Prefer existing module boundaries over new catch-all helpers.

### Low Coupling

- `httpapi` should not write SQL directly.
- `service` should not depend on HTTP request or response types.
- `repository` should not decide business state transitions.
- New capabilities should enter through small interfaces when a boundary is needed.

### Lightweight Changes

- Reuse the existing directory and package style.
- Add abstractions only when they remove real duplication or isolate real complexity.
- Build the smallest complete behavior first, then refine the structure.

## Common Extension Paths

### Add an Authenticated Endpoint

1. Add validation or error semantics in `internal/domain` when needed.
2. Add the use-case method in `internal/service`.
3. Add repository reads or writes in `internal/repository`.
4. Add an HTTP handler in `internal/httpapi/handlers.go`.
5. Register the route in `internal/httpapi/router.go`.
6. Add focused service, repository, or router tests based on the blast radius.

### Extend Sync Upload Support

1. Define the table's domain validation rules.
2. Add the repository record and transaction method.
3. Add aggregate-specific decode, normalize, validate, and accept logic in `internal/service`.
4. Define server ownership, revision, device, timestamp, and idempotency behavior.
5. Cover acceptance, rejection, and error mapping in tests.

### Add Configuration

1. Update `internal/config/config.go`.
2. Update `.env.example`.
3. Update README or round documentation when the operator-facing behavior changes.

### Add Database Structure

1. Add an incremental migration under `internal/migrations/sql`.
2. Add only fields or indexes needed by the current use case.
3. Update matching repository records and SQL.
4. Update sync contract or round documentation if synced data shape changes.

## Avoid

- Business logic inside handlers.
- Hidden state-transition decisions inside repository methods.
- Service methods that accept HTTP-specific request or response objects.
- Broad `util` packages with unclear semantics.
- Large directory rewrites that invalidate historical verification notes.
