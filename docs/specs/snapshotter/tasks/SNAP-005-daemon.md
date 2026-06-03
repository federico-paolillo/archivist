---
id: SNAP-005
feature: snapshotter
title: Implement interval daemon behavior
status: done
depends_on:
  - SNAP-003
  - SNAP-004
blocks:
  - SNAP-006
parallel: false
exec_plan: ../plans/SNAP-005-daemon.execplan.md
canonical: true
---

# SNAP-005: Implement Interval Daemon Behavior

## Objective

Implement the executable daemon loop that sleeps before the first snapshot, snapshots/uploads once per configured interval, cleans temporary work files, logs failures, and continues after failure.

## Acceptance Criteria

```gherkin
Scenario: Daemon sleeps first and continues after failure
  Given Snapshotter starts with a configured interval
  When the first upload fails
  Then it logs the failure
  And it waits for the next interval instead of exiting
```

## Done When

- Tests cover interval validation and failure continuation with injectable sleep/snapshot/upload functions.
- Executable entrypoint runs through the service composition boundary.

## Validation

```bash
cd src/snapshotter && uv run pytest
cd src/snapshotter && uv run archivist-snapshotter --help
```

## Required Context

- `../SPEC.md`
- `../PLAN.md`

## Open Questions

- None.
