# Documentation Rules

## What This Document Solves

This document explains where backend documentation should live and how to keep it maintainable as the repository changes.

## When To Read It

- Before adding or rewriting backend documents.
- When deciding whether a note belongs in `developer/` or `original/`.

## Current Rule

Document by the reader's problem, not by the development round.

Use:

- `documents/developer/`
  - current and future durable guidance
- `documents/original/`
  - historical round records, plans, acceptance logs, release checklists

## What Belongs In Developer

- current architecture and boundaries
- common change paths
- stable route, config, migration, and test rules
- documentation maintenance rules

## What Belongs In Original

- round plans
- phase-by-phase development logs
- acceptance checklists tied to a past delivery round
- release checklists kept for historical traceability

## Writing Rules For Developer Documents

Each how-to or reference document should be easy to scan and should answer:

- what problem it solves
- when to read it
- current rule
- change path
- how to verify
- related documents

## Update Order

When a code change affects durable behavior, prefer this update order:

1. Root `README.md` or `README.zh-cn.md` for operator-facing behavior.
2. `documents/developer/` for backend development guidance.
3. `documents/original/` only when a historical index, path correction, erratum, or redaction is needed.

## Related Documents

- `../README.md`
- `../original/README.md`
- `../../README.zh-cn.md`
