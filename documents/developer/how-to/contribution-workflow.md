# Contribution Workflow

## What This Document Solves

This document explains the current workflow expectations for contributing backend changes and keeping code, tests, and docs aligned.

## When To Read It

- Before preparing a backend change for review.
- When deciding which documents and checks must move with the code.

## Current Workflow

1. Understand the current boundary you are changing.
2. Make the smallest complete code change that fits that boundary.
3. Run the verification layers that match the blast radius.
4. Update durable documentation if the rule changed.
5. Use a clear commit message tied to the real module.

## Commit Format

Use this format:

```text
action(module): summary
```

Allowed `action` values:

- `feat`
- `fix`
- `refactor`
- `chore`
- `docs`
- `perf`

## Module Naming

Keep `module` short and tied to the actual boundary. Prefer:

- `documents`
- `auth`
- `httpapi`
- `sync`
- `repository`
- `domain`
- `config`
- `migrations`
- `telemetry`
- `docker`

When a change crosses several directories, choose the module that best describes the behavior being changed instead of broad names such as `core` or `misc`.

## Examples

```text
feat(sync): support recurring template upload validation
fix(auth): reject expired refresh sessions
refactor(repository): split account query helpers
docs(documents): reorganize backend developer guides
chore(docker): align local compose defaults
perf(httpapi): reduce metrics label allocations
```

## Documentation Updates

Update documentation when a change affects:

- public or authenticated endpoints
- configuration or environment variables
- migrations or data constraints
- sync upload contracts
- recommended development workflow

Prefer this order:

1. Root `README.md` or `README.zh-cn.md` for operator-facing behavior.
2. `documents/developer/` for durable backend development guidance.
3. `documents/original/` only for historical index updates, path fixes, errata, or redaction.

## Minimum Checks

- The change stays inside clear module boundaries.
- Handler, auth or service, repository, and domain responsibilities remain separated.
- New rules live in `domain`, `auth`, or `service` instead of being duplicated in several places.
- New configuration, endpoints, or data contracts are documented.
- The commit message follows `action(module): summary`.

## Related Documents

- `run-and-verify-locally.md`
- `../reference/test-matrix.md`
- `../reference/documentation-rules.md`
