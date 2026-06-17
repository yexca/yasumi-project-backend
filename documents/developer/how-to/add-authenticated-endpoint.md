# Add An Authenticated Endpoint

## What This Document Solves

This document shows how to add or change an authenticated HTTP endpoint without collapsing backend boundaries.

## When To Read It

- Before editing `internal/httpapi` for a new route.
- Before adding a backend capability that should be exposed over HTTP.

## Current Rule

Follow the current dependency direction:

1. `internal/httpapi` parses requests and writes responses.
2. `internal/auth`, `internal/service`, or another use-case boundary owns behavior.
3. `internal/domain` owns reusable validation and semantic errors when needed.
4. `internal/repository` owns SQL and persistence.

Do not place business rules or SQL directly in handlers.

## Change Path

1. Decide whether the new behavior belongs under `internal/auth` or `internal/service`.
2. Add domain validation or stable error semantics in `internal/domain` if the rule is reusable.
3. Add or extend repository methods if persistence changes are needed.
4. Implement the use-case method.
5. Add the handler in `internal/httpapi/handlers.go`.
6. Register the route in `internal/httpapi/router.go`.
7. Protect the route with `requireAuth` when it is authenticated.
8. Update route and behavior documentation.

## Implementation Notes

- Keep request decoding close to the handler.
- Keep response mapping in `internal/httpapi`.
- Keep authenticated user scope explicit.
- Reuse the existing API error shape instead of inventing a route-specific one.
- If a route changes operator-facing behavior, update the root README as well as developer docs.

## How To Verify

- Add or update router tests in `internal/httpapi/router_test.go`.
- Add service or auth tests when the use-case behavior changes materially.
- Verify the route returns stable error codes and field keys.
- Verify unauthenticated access still fails correctly.

## Related Documents

- `../concepts/architecture.md`
- `../reference/api-surface.md`
- `../reference/test-matrix.md`
- `contribution-workflow.md`
