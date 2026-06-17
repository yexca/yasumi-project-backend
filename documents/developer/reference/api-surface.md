# API Surface Reference

## What This Document Solves

This document lists the backend HTTP surface that exists today and highlights the routes that are intentionally absent.

## When To Read It

- Before adding or changing routes.
- When updating operator-facing README content.
- When checking whether a capability belongs in HTTP or only in the sync path.

## Public Infrastructure Routes

- `GET /healthz`
  - process liveness
- `GET /readyz`
  - dependency readiness for database and sync
- `GET /metrics`
  - Prometheus-style operational metrics

## Auth And Session Routes

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/logout`
- `POST /v1/auth/refresh`
- `GET /v1/session`

## Authenticated Account Routes

- `POST /v1/profile`
- `POST /v1/profile/password`

## Other Authenticated Routes

- `GET /v1/weather`
- `POST /v1/sync/token`
- `POST /v1/sync/upload`

## Important Contract Rules

- `POST /v1/sync/token` scopes the issued token to the authenticated user.
- `POST /v1/sync/upload` is the current sync-facing write boundary.
- Stable API error shape matters as much as HTTP status codes.

## Intentionally Absent Direct CRUD Routes

The MVP backend intentionally does not expose direct CRUD routes for:

- `items`
- `areas`
- `recurring_task_templates`
- `operation_history`
- `user_settings`

If you add one of these later, document why the sync boundary is no longer sufficient.

## Related Documents

- `../how-to/add-authenticated-endpoint.md`
- `../concepts/sync-model.md`
- `../../README.md`
