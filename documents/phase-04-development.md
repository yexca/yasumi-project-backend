# Phase 04 Development Notes

Date: 2026-06-13

## Scope

Phase 04 implements the MVP direct backend API boundary: authentication context, sync token issuance, health, readiness, request middleware, and stable HTTP error responses.

This phase intentionally does not add direct CRUD routes for synced business data.

## Implemented

- Bearer authentication boundary with a replaceable authenticator interface and local development token implementation.
- Request ID and request timeout middleware.
- Stable API error DTO shape with centralized HTTP status mapping.
- `GET /healthz` public liveness endpoint.
- `GET /readyz` dependency readiness endpoint with generic `database` and `sync` checks.
- `GET /v1/session` authenticated session context endpoint.
- `POST /v1/sync/token` authenticated sync authorization payload endpoint scoped to the authenticated user.
- Sync token issuer interface with bounded-lifetime HMAC local implementation.
- API startup now opens the PostgreSQL pool used by readiness.
- Local environment examples for development auth and sync token configuration.
- Route-level tests for unauthenticated requests, authenticated session, sync token scope, validation error shape, readiness, and absent MVP CRUD routes.

## Authentication Stance

The current authenticator is a development bearer-token boundary configured through environment variables. It is intentionally small and replaceable so a production auth provider can be added without changing the stable direct API contract.

`POST /v1/sync/token` ignores client-supplied `user_id` and scopes the token response to the authenticated user.

## Out of Scope

- Production identity provider integration.
- Full PowerSync upload validation.
- Item, area, recurring template, operation history, or user settings CRUD routes.
- Calendar OAuth or provider token storage.

## Verification

Run:

```powershell
go fmt ./...
go test ./...
go vet ./...
```

Docker toolchain equivalent:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
```
