# Round 02: Backend Real Sync and Structure

This directory contains backend-local guidance for Round 02. It should be read together with the deployment repository's `docs/development-rounds/round-02-real-sync-and-structure/` directory.

Round 02 is about making the running app sync real data through PowerSync, removing fixture-first behavior from the main path, and reducing structural pressure in the sync-facing backend files without changing the public contract casually.

## Summary

Round 02 keeps the backend contract steady while the full app moves from simulated sync to real PowerSync-backed data. Backend work should focus on synced onboarding rows, contract-compatible upload handling, and incremental file splits around the existing sync and repository pressure points.

## Backend Goals

1. Keep the existing sync upload contract stable while the frontend begins sending real mutations.
2. Ensure new accounts start with useful synced onboarding rows instead of depending on frontend fixture data.
3. Split high-pressure backend files incrementally when a change touches their concern area.
4. Keep root Docker Compose as the primary full-stack verification path.

## Documents

- `backend-round-plan.md`
  - Backend workstreams, implementation boundaries, and delivery order.
- `onboarding-seed-data.md`
  - Rules for default synced items created during account registration.
- `structure-guidance.md`
  - Backend structure cleanup guidance, including the earlier optimization assessment integrated into the current Round 02 plan.
- `backend-acceptance-checklist.md`
  - Backend-focused acceptance checks for this round.

## Current Constraints

- Do not replace PowerSync or the PostgreSQL sync tables during this round.
- Do not add a parallel fixture path for registered users.
- Do not move business rules into HTTP handlers or SQL helpers.
- Prefer small file splits over new packages unless an existing package boundary is clearly wrong.
- Keep backend changes compatible with the deployment repository's Round 02 acceptance criteria.
