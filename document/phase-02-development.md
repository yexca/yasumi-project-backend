# Phase 02 Development Notes

Date: 2026-06-13

## Scope

Phase 02 implements a pure backend domain baseline that later database, service, and sync work can depend on.

The implementation intentionally avoids HTTP, PostgreSQL, PowerSync, logging, and framework dependencies.

## Implemented

- Typed constants matching the shared constants reference.
- Allow-list validators for shared enum-like values.
- Stable domain error shape with error codes, field keys, and retryability.
- Item shape validation for inbox items, date tasks, deadline tasks, and ideas.
- Deadline mode validation for date-only, floating, and fixed deadlines.
- Status transition validation using the shared transition table.
- Metadata precedence helpers for deleted, archived, hidden, and status states.
- Settings defaults and validation for MVP language, locale, week start, timezone, and recommendation windows.
- Idempotency key builders for ordinary semantic actions, recurrence actions, recurrence generation, and due postponed activation.

## Fixture Coverage

Tests cover the available contract fixture scenarios for:

- Valid item rows for all MVP item types.
- Valid deadline modes.
- Invalid deadline mode combinations.
- Allowed and rejected status transitions.
- Metadata precedence.
- Language-specific settings defaults.
- Recurrence idempotency key formats.

The tests are table-driven and keep fixture-derived values explicit in the backend repository so the package remains portable when cloned by itself.

