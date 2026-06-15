# Phase 06 Verification

Date: 2026-06-13

## Scope

This verification reviews the current backend against `backend_coding_guide/07-phase-06-account-system.md` and its referenced account, API, validation, security, migration, and compatibility rules.

Phase 06 is expected to add Yasumi's first-party account system: username/email/password registration, login, logout, refresh-token rotation, durable sessions, account-backed authentication middleware, default settings initialization, and account-flow tests.

## Result

Phase 06 verification passed on 2026-06-13.

The current backend implements first-party accounts with username, email, password authentication, Argon2id password hashing, durable refresh sessions, access-token authenticated API requests, refresh-token rotation, logout revocation, disabled-account checks, default settings creation, account-backed sync token issuance, and Docker-based verification.

## Verified Working

- Migration `0002_account_system.up.sql` creates `users`, `user_credentials`, and `user_sessions`.
- Existing synced user-data tables now reference `users.id` through forward-only foreign keys.
- Registration creates `users`, `user_credentials`, `user_sessions`, and default `user_settings` in one transaction.
- Username and email are case-insensitively unique.
- Passwords are stored as Argon2id hashes with parameters, not raw passwords.
- Refresh tokens are stored only as SHA-256 hashes.
- Login accepts username or email.
- Unknown identifiers and wrong passwords both return `invalid_credentials`.
- Refresh rotates the refresh token and invalidates the old token.
- Logout revokes the active session.
- Disabled accounts cannot log in, authenticate existing sessions, or refresh sessions.
- Runtime auth middleware resolves the authenticated user from the account session, not from the old development bearer token.
- `/v1/session` returns account identity and does not return synced `user_settings`.
- `/v1/sync/token` scopes the sync token to the internal `users.id`.

## Evidence

- `internal/app/app.go` wires `AccountService` as both the account service and authenticator for the HTTP router.
- `internal/auth/auth.go` implements registration, login, refresh, logout, access-token parsing, Argon2id hashing, refresh-token hashing, duplicate-account error mapping, and disabled-account enforcement.
- `internal/repository/repository.go` implements account lookup, session creation, session rotation/revocation support, and transactional account creation with default settings.
- `internal/httpapi/router.go` registers `POST /v1/auth/register`, `POST /v1/auth/login`, `POST /v1/auth/logout`, `POST /v1/auth/refresh`, and `GET /v1/session`.
- `internal/auth/account_integration_test.go` covers registration, login by username/email, invalid credentials, duplicate username/email, refresh rotation, logout revocation, disabled accounts, password hash storage, refresh-token hash storage, and default settings creation.
- `internal/httpapi/router_test.go` covers account route behavior and authenticated direct API boundaries.

## Commands Run

Docker-based Go toolchain:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm gofmt -w cmd internal
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go vet ./...
```

Deployment build:

```powershell
docker build -t yasumi-backend-phase56-verify .
```

Runtime probes against the compose API:

```powershell
docker compose -f .\env\docker-compose.yml up -d api
```

Observed runtime results:

- `/healthz` returned `{"status":"ok"}`.
- `POST /v1/auth/register` created a local account and returned access and refresh tokens.
- `GET /v1/session` accepted the account access token and returned the registered user.
- `POST /v1/sync/token` accepted the account access token and returned a sync token scoped to the internal `users.id`.
- `POST /v1/sync/upload` accepted an empty mutation batch under the account user context.
- `/readyz` returned `not_ready` with `database: ok` and `sync: unavailable` because PowerSync was not running in this verification pass.

## Residual Limitations

- Login and registration rate limiting is still not implemented; this remains a documented release-hardening limitation.
- Email is required but not verified in MVP, matching the Phase 06 guide.
- Access tokens use a compact HMAC-signed internal token format rather than JWT. This still satisfies the guide because the backend resolves a trustworthy internal `user_id` through the session.

## Sign-off

Phase 06 is accepted for the current backend state and the backend can proceed to Phase 07 hardening.
