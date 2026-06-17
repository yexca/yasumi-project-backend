# Change Configuration

## What This Document Solves

This document shows how to add or change backend configuration without leaving code, local examples, and runtime docs out of sync.

## When To Read It

- Before editing `internal/config/config.go`.
- Before adding, renaming, or removing environment variables.

## Current Rule

Configuration is loaded once at startup from environment variables. The application binary does not load `.env` files by itself.

Current operator-facing examples live in:

- `.env.example`
- `env/local.env.example`
- `env/README.md`
- root `README.md`
- root `README.zh-cn.md`

## Change Path

1. Update `internal/config/config.go`.
2. Add validation if the new setting must fail fast.
3. Update `.env.example` if the root Docker Compose workflow uses the setting.
4. Update `env/local.env.example` if the older local environment also needs it.
5. Update `env/README.md` and root README files when local or operator behavior changes.
6. Update developer reference docs for durable configuration changes.
7. Add or update tests in `internal/config/config_test.go`.

## Rules

- Keep defaults intentional and reviewable.
- Fail fast for invalid or missing required settings.
- Trim and parse values once during startup.
- Do not store production secrets in repository files.
- If a new setting affects HTTP behavior, sync behavior, or migrations, document its verification path.

## How To Verify

- Run config tests.
- Start the API with expected local settings.
- Verify route, readiness, or sync behavior if the setting affects runtime behavior.
- Verify example environment files still match the code.

## Related Documents

- `../reference/environment-variables.md`
- `../reference/documentation-rules.md`
- `run-and-verify-locally.md`
