# Backend MVP Release Checklist

Date: 2026-06-13

## Build and Test

- [ ] `go test ./...` passes in the Docker toolchain.
- [ ] `go vet ./...` passes in the Docker toolchain.
- [ ] Repository and account integration tests pass against PostgreSQL.
- [ ] Docker image builds from the root `Dockerfile`.
- [ ] Migrations apply from an empty database.

## Contracts

- [ ] Backend constants cover `shared_constants/`.
- [ ] Item and deadline validation pass shared fixture compatibility tests.
- [ ] Error codes and field keys remain stable.
- [ ] PowerSync streams are scoped by authenticated `user_id`.
- [ ] No direct MVP business CRUD routes exist.

## Security and Privacy

- [ ] Passwords are never returned.
- [ ] Password hashes are never returned.
- [ ] Refresh tokens are stored only as hashes.
- [ ] Access tokens resolve through active, unrevoked sessions.
- [ ] Disabled accounts cannot log in, refresh, access `/v1/session`, request sync tokens, or upload sync mutations.
- [ ] Logs do not contain item titles, notes, Quick Add source text, parser output, passwords, password hashes, authorization headers, raw access tokens, or raw refresh tokens.
- [ ] `/metrics` is restricted by deployment networking or ingress.

## Sync and Offline

- [ ] Create and edit offline, reconnect, and verify convergence.
- [ ] Delete versus edit keeps tombstone until explicit restore.
- [ ] Restore deleted or archived rows clears metadata and preserves business status unless explicitly reopened.
- [ ] Invalid offline `date_task` writes return recoverable validation details.
- [ ] Duplicate due postponed activation returns the existing accepted result.
- [ ] Settings conflicts converge by server-accepted last write.
- [ ] Another user's rows never appear in sync output.

## Operations

- [ ] `/healthz` returns `200`.
- [ ] `/readyz` returns `200` with PostgreSQL and PowerSync available.
- [ ] `/readyz` fails clearly with generic dependency labels when a dependency is unavailable.
- [ ] `/metrics` exposes request failure, validation rejection, sync upload result, auth failure, and dependency readiness counters.
- [ ] Deployment configuration is documented in `deployment-operations.md`.

## Future-Scope Guardrails

- [ ] No active endpoint implements calendar OAuth.
- [ ] No active migration stores calendar provider tokens or provider event payloads.
- [ ] No active sync stream includes calendar provider events.
- [ ] No operation event type records calendar link history.

## Known Limitations Accepted for MVP

- [ ] Login and registration rate limiting is deferred.
- [ ] Email verification is deferred.
- [ ] Calendar integration is deferred.
- [ ] Field-level conflict review is deferred.
