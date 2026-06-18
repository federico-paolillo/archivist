---
id: SNAP-005
feature: snapshotter
title: Implement interval daemon behavior
depends_on:
  - SNAP-003
  - SNAP-004
blocks:
  - SNAP-006
parallel: false
requires_exec_plan: false
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
- Validation reaches the real `archivist-snapshotter` executable entrypoint through a controlled subprocess or integration test path. A `--help` invocation alone is not sufficient executable-boundary validation.

## Validation

```bash
cd src/snapshotter && uv run pytest
```

The pytest suite must include a bounded subprocess or integration test that invokes the real `archivist-snapshotter` entrypoint with controlled configuration and prevents unbounded daemon execution through test-controlled sleep, snapshot, or upload boundaries.

## Required Context

- `../SPEC.md`
- `../PLAN.md`

## Open Questions

- None.
