---
id: SNAP-003
feature: snapshotter
title: Implement snapshot archive capture
depends_on:
  - SNAP-002
blocks:
  - SNAP-005
parallel: false
requires_exec_plan: false
canonical: true
---
# SNAP-003: Implement Snapshot Archive Capture

## Objective

Implement staged `/data` capture, SQLite online backup, best-effort artifact copy, manifest generation, tar.gz creation, and cleanup.

## Acceptance Criteria

```gherkin
Scenario: Snapshot archive contains a safe database backup
  Given a source /data directory with archive.db and artifact files
  When Snapshotter captures an archive
  Then the staged archive contains data/archive.db from SQLite online backup
  And it excludes live archive.db-wal and archive.db-shm sidecars
  And it contains manifest.json
```

## Done When

- Capture tests cover SQLite backup, sidecar omission, disappearing file skip behavior, manifest fields, and archive name format.
- Capture code does not coordinate with Gateway or Worker.

## Validation

```bash
cd src/snapshotter && uv run pytest
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../../ARTIFACTS.md`

## Open Questions

- None.
