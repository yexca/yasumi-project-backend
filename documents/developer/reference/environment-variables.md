# Environment Variables Reference

## What This Document Solves

This document lists the backend environment variables that shape current runtime behavior and points to the files that carry local examples.

## When To Read It

- Before changing configuration.
- When wiring local, container, or deployment environments.

## Example Sources

- Root stack example: `.env.example`
- Older local stack example: `env/local.env.example`
- Local environment notes: `env/README.md`

## Core Application

- `YASUMI_APP_ENV`

## HTTP

- `YASUMI_HTTP_HOST`
- `YASUMI_HTTP_PORT`
- `YASUMI_HTTP_ALLOWED_ORIGINS`
- `YASUMI_HTTP_READ_HEADER_TIMEOUT`
- `YASUMI_HTTP_SHUTDOWN_TIMEOUT`
- `YASUMI_HTTP_REQUEST_TIMEOUT`

The root `.env.example` also carries host-port helper values used by Docker Compose:

- `YASUMI_HTTP_HOST_PORT`
- `YASUMI_FRONTEND_HOST_PORT`

## Logging

- `YASUMI_LOG_LEVEL`
- `YASUMI_LOG_FORMAT`

## Auth

- `YASUMI_AUTH_DEV_TOKEN`
- `YASUMI_AUTH_DEV_USER_ID`
- `YASUMI_AUTH_DEV_DISPLAY_NAME`

## Sync Token

- `YASUMI_SYNC_TOKEN_SECRET`
- `YASUMI_SYNC_TOKEN_TTL`

## PostgreSQL

- `YASUMI_POSTGRES_HOST`
- `YASUMI_POSTGRES_PORT`
- `YASUMI_POSTGRES_DB`
- `YASUMI_POSTGRES_USER`
- `YASUMI_POSTGRES_PASSWORD`
- `YASUMI_POSTGRES_SSLMODE`

The root `.env.example` also carries the Docker host-port helper:

- `YASUMI_POSTGRES_HOST_PORT`

## PowerSync

- `YASUMI_POWERSYNC_URL`
- `YASUMI_POWERSYNC_PUBLIC_URL`

The root `.env.example` also carries the Docker host-port helper:

- `YASUMI_POWERSYNC_HOST_PORT`

## Frontend Runtime Values In The Root Stack

These values are not backend config keys, but they appear in the root local stack wiring:

- `VITE_BACKEND_BASE_URL`
- `VITE_POWERSYNC_ENDPOINT`

## Rules

- The Go binary reads process environment variables, not `.env` files directly.
- Keep `.env.example` and `env/local.env.example` aligned with `internal/config/config.go` where applicable.
- Never commit real secrets or production credentials.

## Related Documents

- `../how-to/change-configuration.md`
- `../../README.md`
- `../../env/README.md`
