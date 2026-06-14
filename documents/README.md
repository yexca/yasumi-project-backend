# Yasumi Backend Documents

This directory records backend implementation notes, verification results, and release-readiness checks produced during development.

Documents in this directory must not include real credentials, private keys, access tokens, personal data, or production connection strings.

## Development Notes

- `phase-01-acceptance.md`
- `phase-02-development.md`
- `phase-03-development.md`
- `phase-04-development.md`
- `phase-04-verification.md`
- `phase-05-development.md`
- `phase-05-verification.md`
- `phase-06-development.md`
- `phase-06-verification.md`
- `phase-07-development.md`
- `phase-07-verification.md`

## Local Runtime References

- Root `../docker-compose.example.yml` is the preferred example for building and running PostgreSQL, migrations, the API, and optional PowerSync from the repository root.
- Root `../.env.example` documents the Docker Compose interpolation values used by the root compose example.
- `../env/docker-compose.yml` remains available for the project-local Go toolchain and the older local stack workflow.

Docker Compose reads the root `.env` file automatically when commands are run from the repository root. The Go binaries do not load `.env` files directly; runtime configuration is still provided through environment variables.
