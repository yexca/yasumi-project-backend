# Deployment and Operations

Date: 2026-06-13

## Runtime Shape

- Build the API from the root `Dockerfile`.
- Run migrations with `/usr/local/bin/yasumi-migrate` before starting `/usr/local/bin/yasumi-api`.
- Run PostgreSQL with logical WAL enabled for PowerSync.
- Run PowerSync with `env/powersync/service.yaml` and `env/powersync/sync-config.yaml`.
- For local validation, the root `docker-compose.example.yml` builds the image, starts PostgreSQL, runs migrations, starts the API, and can start PowerSync with the `sync` profile.

## Required Configuration

Set these values explicitly outside local development:

- `YASUMI_APP_ENV`
- `YASUMI_HTTP_HOST`
- `YASUMI_HTTP_PORT`
- `YASUMI_HTTP_READ_HEADER_TIMEOUT`
- `YASUMI_HTTP_SHUTDOWN_TIMEOUT`
- `YASUMI_HTTP_REQUEST_TIMEOUT`
- `YASUMI_LOG_LEVEL`
- `YASUMI_LOG_FORMAT`
- `YASUMI_AUTH_DEV_TOKEN`
- `YASUMI_AUTH_DEV_USER_ID`
- `YASUMI_AUTH_DEV_DISPLAY_NAME`
- `YASUMI_SYNC_TOKEN_SECRET`
- `YASUMI_SYNC_TOKEN_TTL`
- `YASUMI_POSTGRES_HOST`
- `YASUMI_POSTGRES_PORT`
- `YASUMI_POSTGRES_DB`
- `YASUMI_POSTGRES_USER`
- `YASUMI_POSTGRES_PASSWORD`
- `YASUMI_POSTGRES_SSLMODE`
- `YASUMI_POWERSYNC_URL`

Do not use the local default sync-token secret outside local development.

Root Docker Compose reads `.env` automatically for interpolation when run from the repository root. The application binary does not load `.env` files directly; the final runtime environment must be supplied by the process manager, container platform, or Docker Compose.

Use `.env.example` only as a local template. Do not store production credentials in repository files.

## Health Checks

- `GET /healthz` confirms the process is alive.
- `GET /readyz` confirms PostgreSQL and PowerSync are reachable.
- `GET /metrics` exposes operational counters and dependency readiness gauges.

Readiness returns `503` until every required dependency is available.

## Operational Checks

- Run migrations from an empty database before deploying a new API image.
- Verify `/readyz` reports `database: ok` and `sync: ok`.
- Verify `/metrics` includes `yasumi_dependency_ready{dependency="database"} 1` and `yasumi_dependency_ready{dependency="sync"} 1` after a readiness probe.
- Register a test account, log in, refresh, log out, and confirm revoked sessions fail.
- Request a sync token and confirm `stream_scope.user_id` is the authenticated internal user id.
- Send an empty sync upload batch and confirm it returns `202`.
- Confirm logs contain request ids, route labels, status codes, and stable error codes only.

## Network Controls

- Restrict `/metrics` to trusted infrastructure.
- Serve the API behind TLS in non-local environments.
- Keep PostgreSQL and PowerSync service ports private unless explicitly required by the deployment platform.

## MVP Deferred Scope

- Calendar OAuth, provider token storage, event import, event export, and calendar links are intentionally absent.
- Direct item, area, recurring template, user settings, and operation-history CRUD endpoints are intentionally absent.
- Hard delete for normal user data is intentionally absent.
