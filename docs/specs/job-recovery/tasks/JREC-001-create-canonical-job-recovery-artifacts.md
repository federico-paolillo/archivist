---
id: JREC-001
feature: job-recovery
title: Create canonical job recovery artifacts
status: done
depends_on: []
blocks: [JREC-002, JREC-003, JREC-004]
parallel: false
exec_plan: null
canonical: true
---

# JREC-001: Create Canonical Job Recovery Artifacts

## Objective

Create the feature specification, plan, task files, feature index entry, and design decision for stale running job recovery and Worker logging.

## Scope

This task includes:

- `docs/specs/job-recovery/SPEC.md`
- `docs/specs/job-recovery/PLAN.md`
- `docs/specs/job-recovery/DIARY.md`
- `docs/specs/job-recovery/tasks/*.md`
- `docs/specs/INDEX.md` feature row
- `docs/DESIGN.md` durable decision

## Out of Scope

This task does not include:

- Gateway implementation.
- UI implementation.
- Worker implementation.
- Runtime validation beyond documentation sanity checks.

## Inputs

- User-approved feature plan.
- Existing `ui`, `ui-endpoints`, and `summary-generation` feature contracts.

## Outputs

- Canonical feature planning artifacts sufficient for module workers.

## Expected Affected Areas

```text
docs/specs/job-recovery/**
docs/specs/INDEX.md
docs/DESIGN.md
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `AGENTS.md`
- `docs/BOOKKEEPING.md`
- `docs/REBUILD.md`
- `docs/specs/INDEX.md`

## Acceptance Criteria

```gherkin
Scenario: Canonical artifacts exist
  Given the job-recovery feature has been approved
  When the task completes
  Then SPEC.md and PLAN.md define stale force delete and Worker logging requirements
  And task files exist for Gateway, UI, Worker, and integration work
  And docs/specs/INDEX.md includes job-recovery
  And docs/DESIGN.md records the recovery decision
```

## Done When

- Feature artifacts exist.
- Task dependencies are documented.
- The feature index includes `job-recovery`.
- The design decision is recorded.

## Validation

Required checks:

```bash
git diff --check
```

Manual validation, if any:

- Confirm tasks are ready for disjoint module workers.

## Dependencies

Depends on:

- None.

Blocks:

- `JREC-002`
- `JREC-003`
- `JREC-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
