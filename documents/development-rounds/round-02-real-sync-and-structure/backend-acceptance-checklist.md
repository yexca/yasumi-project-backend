# Backend Acceptance Checklist

Use this checklist for backend-local Round 02 work. The full cross-device acceptance pass still belongs to the deployment repository's Round 02 playbook.

## Registration

- [ ] Registration creates `users`, `user_credentials`, `user_settings`, onboarding `items`, and the initial `user_sessions` row in one transaction.
- [ ] Onboarding items include exactly the intended item types unless the onboarding document is updated.
- [ ] Onboarding rows pass current synced-table constraints.
- [ ] Failed onboarding item creation rolls back the account registration transaction.

## Sync Token

- [ ] Authenticated users can request `/v1/sync/token`.
- [ ] Issued tokens are scoped to the authenticated user and active device.
- [ ] Disabled or expired account states still fail with the stable auth error behavior.

## Sync Upload

- [ ] Valid non-empty `items` mutations are accepted.
- [ ] Valid `areas`, `recurring_task_templates`, `operation_history`, and `user_settings` mutations keep existing acceptance behavior.
- [ ] Invalid item shapes produce stable validation errors.
- [ ] Invalid semantic transitions remain rejected.
- [ ] Duplicate semantic operations converge through idempotency where the contract requires it.

## Persistence

- [ ] Accepted writes appear in PostgreSQL synced tables.
- [ ] Server metadata such as `server_updated_at`, `revision`, and device IDs remains consistent with existing rules.
- [ ] Operation history remains append-only.

## Structure

- [ ] Any split of sync upload logic keeps aggregate-specific rules readable in isolation.
- [ ] Any split of repository SQL keeps account/session concerns separate from synced-data concerns.
- [ ] Public package usage remains stable unless a round document explicitly justifies the change.

## Verification

- [ ] Relevant Go tests pass.
- [ ] Backend readiness still reports database and sync dependencies correctly in the root stack.
- [ ] The deployment stack can use the backend for a real Round 02 sync acceptance run.
