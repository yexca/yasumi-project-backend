# Phase 06 Development

Date: 2026-06-13

## Scope

Implemented `backend_coding_guide/07-phase-06-account-system.md`: first-party username, email, and password accounts with durable sessions and account-backed authentication.

## Implemented

- Added forward-only account migration `0002_account_system.up.sql`.
- Added `users`, `user_credentials`, and `user_sessions`.
- Added case-insensitive unique indexes for username and email.
- Added refresh-token hash storage and session expiry/revocation columns.
- Added foreign keys from synced user-owned tables to `users.id`.
- Added development-data backfill for existing synced `user_id` values before applying account ownership foreign keys.
- Added repository methods for account creation, identifier lookup, session creation, active-session lookup, refresh lookup, and session revocation.
- Added Argon2id password hashing.
- Added signed access tokens containing user id and session id.
- Stored refresh tokens only as SHA-256 hashes.
- Added account-backed auth middleware through `AccountService.Authenticate`.
- Added `POST /v1/auth/register`.
- Added `POST /v1/auth/login`.
- Added `POST /v1/auth/logout`.
- Added `POST /v1/auth/refresh`.
- Updated `GET /v1/session` to return account identity fields and no synced settings.
- Registration creates default `user_settings` in the same transaction as the account and credential rows.
- Login accepts username or email.
- Refresh rotates refresh tokens and revokes the previous session.
- Logout revokes the current session.
- Disabled accounts are rejected for login, refresh, and authenticated access.
- Added route tests for account endpoint status and error shapes.
- Added database-backed account integration tests.
- Updated repository integration tests for the new account ownership foreign keys.

## Security Notes

- Passwords are never returned in API responses.
- Password hashes are not returned in API responses.
- Raw refresh tokens are returned only in auth/session responses and accepted only in refresh requests.
- Refresh tokens are stored as hashes.
- Login returns the same `invalid_credentials` code for unknown identifiers and wrong passwords.
- Access tokens are bounded and tied to a revocable session id.
- Rate limiting is not implemented yet and remains a release-hardening limitation.

## Verification

Docker-based checks run during development:

```powershell
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go test ./...
docker compose -f .\env\docker-compose.yml --profile tools run --rm -e YASUMI_POSTGRES_HOST=postgres go-toolchain go test ./internal/repository ./internal/auth -count=1 -v
docker compose -f .\env\docker-compose.yml --profile tools run --rm go-toolchain go vet ./...
docker build -t yasumi-backend-phase6-dev .
docker compose --env-file .\env\local.env.example -f .\env\docker-compose.yml up -d --build api
```

Runtime probes confirmed:

- Register creates an account and returns a session payload.
- Refresh rotates the refresh token.
- Reusing the previous refresh token returns `401`.
- Logout returns `204`.
- Accessing `/v1/session` with a revoked access token returns `401`.

## Known Follow-up

- Phase 05 was rerun after the account change and is now accepted for the current backend state.
- Login and registration rate limiting is still recommended before release.
