---
id: SNAP-003-PLAN
task: ../tasks/SNAP-003-capture.md
status: completed
canonical: true
---

# ExecPlan: SNAP-003 Capture

## Objective

Implement archive capture using SQLite online backup, best-effort artifact copy, manifest generation, and tar.gz output.

## Linked Task

- `../tasks/SNAP-003-capture.md`

## Implementation Sequence

1. Build timestamp and archive naming helpers.
2. Stage a fresh work directory for each attempt.
3. Use Python `sqlite3.Connection.backup()` to copy the configured database to staged `data/archive.db`.
4. Copy non-database files from `ARCHIVIST_DATA_DIR` into staged `data/`, skipping configured database sidecars and files that disappear during traversal.
5. Write `manifest.json` at archive root.
6. Create the `.tar.gz` archive from staged root and remove the staging directory after upload or failure.

## Validation Plan

```bash
cd src/snapshotter && uv run pytest
```

## Risks

- Artifact copy is intentionally not transactionally consistent with SQLite; this limitation must remain visible in the manifest.

## Completion Criteria

- Tests prove database backup, sidecar exclusion, manifest fields, and archive naming.
