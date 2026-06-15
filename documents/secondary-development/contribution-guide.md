# Contribution Guide

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

## Action Meanings

- `feat`: new capability or user-visible behavior.
- `fix`: defect fix or behavior correction.
- `refactor`: structural change without intended external behavior change.
- `chore`: configuration, scripts, environment, or maintenance work.
- `docs`: documentation changes.
- `perf`: performance-focused change.

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
docs(documents): reorganize development round guides
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
2. `documents/development-rounds/round-02-real-sync-and-structure/` for active round guidance.
3. `documents/secondary-development/` for durable backend development guidance.

## Minimum Checks

- The change stays inside clear module boundaries.
- Handler, service, repository, and domain responsibilities remain separated.
- New rules live in `domain` or `service` instead of being duplicated in several places.
- New configuration, endpoints, or data contracts are documented.
- The commit message follows `action(module): summary`.
