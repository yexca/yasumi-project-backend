# Onboarding Seed Data

## Problem

Before Round 02, the visible "default tasks" were effectively frontend fixture data. That is not enough for real sync because a newly registered account needs backend-owned synced rows that can appear on another device through PowerSync.

## Decision

Account registration creates a minimal set of onboarding `items` in the same database transaction as:

- `users`
- `user_credentials`
- `user_settings`
- `user_sessions`

The current seed set contains one item for each supported item type:

- `inbox`
  - Demonstrates quick capture before organizing.
- `date_task`
  - Demonstrates a task scheduled for a specific day.
- `deadline_task`
  - Demonstrates planned work plus a due date.
- `idea`
  - Demonstrates a revisit-able idea that does not become overdue like a task.

## Content Guidelines

Seed tasks should:

- be clearly instructional
- avoid fake personal obligations
- avoid references to a specific company, person, or private scenario
- use stable product concepts such as Inbox, date task, deadline task, and idea
- stay small enough that the user can delete or complete them quickly

Seed tasks should not:

- create operation history rows for actions that have not happened
- create archived, deleted, completed, or abandoned rows
- depend on frontend fixture IDs
- require a recurring template unless the recurring workflow is intentionally added to onboarding later

## Data Rules

- All seed rows use the registered user's real `user_id`.
- `created_by_device_id` and `updated_by_device_id` use `account-registration`.
- `revision` starts at `1`.
- `created_at`, `updated_at`, `client_updated_at`, and `server_updated_at` use the registration timestamp.
- Date-based examples derive from the registration date in the app's default Tokyo timezone.
- `pressure_metadata` is an empty JSON object.

## Extension Rule

If onboarding expands later, keep it in the auth/account registration boundary unless the setup becomes large enough to justify a dedicated service. Do not add a separate frontend-only onboarding fixture path for registered accounts.
