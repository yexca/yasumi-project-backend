# Yasumi Backend Developer Guide

This directory is the durable backend development guide for the current codebase.

Use it when you need to change today's backend behavior, not when you need to reconstruct how a delivery round happened. Historical round records remain in `../original/`.

## What This Guide Solves

- Understand the current backend structure and dependency direction.
- Find the right change path for common backend tasks.
- Check stable rules for routes, configuration, migrations, tests, and documentation updates.

## When To Read This Directory

- Before adding or changing backend behavior in `cmd/` or `internal/`.
- Before changing sync upload rules, migrations, or environment variables.
- Before updating developer-facing backend documentation.

## Reading Paths

### I need to understand the backend before editing it

1. `concepts/architecture.md`
2. `concepts/sync-model.md`
3. `concepts/data-and-migrations.md`

### I need to implement a common change

- Add or change an authenticated endpoint:
  - `how-to/add-authenticated-endpoint.md`
- Extend sync upload support for a synced table:
  - `how-to/extend-sync-upload.md`
- Add a database migration:
  - `how-to/add-migration.md`
- Change configuration or environment variables:
  - `how-to/change-configuration.md`
- Run backend checks locally:
  - `how-to/run-and-verify-locally.md`
- Prepare and document a contribution:
  - `how-to/contribution-workflow.md`

### I need stable facts or checklists

- Current route surface:
  - `reference/api-surface.md`
- Environment variables:
  - `reference/environment-variables.md`
- Database and migration rules:
  - `reference/database-schema-and-migration-rules.md`
- Test matrix:
  - `reference/test-matrix.md`
- Documentation maintenance rules:
  - `reference/documentation-rules.md`

## Task Index

- Add a route: `how-to/add-authenticated-endpoint.md`
- Extend `POST /v1/sync/upload`: `how-to/extend-sync-upload.md`
- Add or change SQL schema: `how-to/add-migration.md`
- Add a config key: `how-to/change-configuration.md`
- Verify a backend change end to end: `how-to/run-and-verify-locally.md`
- Check what must be documented after a code change: `reference/documentation-rules.md`

## Maintenance Rules

- Keep this directory organized by the reader's problem, not by delivery round.
- Move historical process notes, acceptance logs, and round plans to `../original/`.
- When a round document contains a lasting rule, extract that rule here and leave the original round document intact.
- Do not store real secrets, tokens, personal data, or production connection details in these documents.

## Related Documents

- Backend document index: `../README.md`
- Historical archive: `../original/README.md`
- Repository overview: `../../README.md`
