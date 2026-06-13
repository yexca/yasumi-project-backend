# Phase 07 Development

Date: 2026-06-13

## Scope

Implemented `backend_coding_guide/08-phase-07-hardening-release.md`: backend hardening, release-readiness checks, compatibility tests, observability, and operational documentation.

## Implemented

- Added production-safe request logging with request id, method, route label, status, and duration only.
- Added `/metrics` with text counters for request failures, API errors, auth failures, validation rejections, sync upload results, and dependency readiness.
- Added readiness metrics for PostgreSQL and PowerSync dependency health.
- Added shared-constants compatibility tests against `dev_documents/shared_constants`.
- Added contract fixture validation tests against `contracts/fixtures/data/item-shapes-and-deadlines.json`.
- Added PowerSync rule tests verifying every synced MVP table is scoped by `auth.user_id()`.
- Added future-scope guardrail test that active backend code and environment files do not implement calendar integration.
- Added offline conflict tests for restore, delete-versus-edit tombstone precedence, duplicate postponed activation, and settings last-write-wins coverage already present in sync upload tests.
- Added database index verification for expected MVP query paths.
- Updated the Docker toolchain environment to mount `dev_documents` read-only at `/dev_documents` and set `YASUMI_CONTRACTS_DIR`.
- Documented deployment configuration, operational checks, known limitations, and the MVP release checklist.

## Security and Privacy Notes

- Request logs do not include request bodies, query strings, item titles, notes, Quick Add source text, parser output, passwords, raw access tokens, raw refresh tokens, password hashes, or authorization headers.
- API rejection logs include stable error code and retryability only.
- Metrics labels use fixed route names, stable codes, dependency names, and result labels.
- `/readyz` reports generic dependency labels only.

## Verification Focus

- Backend constants cover shared contract values.
- Backend validation matches item/deadline contract fixtures.
- PowerSync rules are user-scoped for every synced MVP table.
- Account auth and session tests remain covered by Phase 06 integration tests.
- Sync conflict assumptions are covered at the service boundary.
- Database indexes exist for expected MVP query paths.
- No active MVP endpoint, migration, or environment config implements calendar integration.

## Known Limitations

- Login and registration rate limiting is still not implemented.
- `/metrics` is unauthenticated in the local MVP service; deployment should restrict it at the network or ingress layer.
- PowerSync readiness depends on `YASUMI_POWERSYNC_URL` being reachable.
- Calendar OAuth, import/export, provider token storage, and calendar links remain out of scope.
