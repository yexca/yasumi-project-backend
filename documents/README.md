# Yasumi Backend Documents

This directory keeps backend-local development notes. The active structure mirrors the deployment repository so round-level guidance can be followed without guessing which document is current.

## Structure

- `secondary-development/`
  - Backend architecture, extension notes, and contribution guidance for ongoing work.
- `development-rounds/round-01-mvp-backend-delivery/`
  - Archived Round 01 MVP backend delivery records.
- `development-rounds/round-02-real-sync-and-structure/`
  - Active backend guidance for Round 02: real sync, structure cleanup, and onboarding seed data.

## Reading Order

1. Start with `development-rounds/round-02-real-sync-and-structure/README.md` for current round scope.
2. Read `secondary-development/architecture-and-extension.md` before changing backend boundaries.
3. Check `secondary-development/contribution-guide.md` before submitting changes.
4. Use `development-rounds/round-01-mvp-backend-delivery/` only for historical context.

## Rules

- Do not store real secrets, tokens, personal data, or production connection details in documents.
- Historical round documents should only receive link fixes, redaction, or necessary errata.
- Current implementation guidance belongs under `development-rounds/round-02-real-sync-and-structure/` while Round 02 is active.
