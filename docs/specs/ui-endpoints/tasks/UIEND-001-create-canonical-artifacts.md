---
id: UIEND-001
feature: ui-endpoints
title: Create canonical artifacts
status: done
depends_on: []
blocks: [UIEND-002, UIEND-003]
parallel: false
exec_plan: null
canonical: true
---

# UIEND-001: Create Canonical Artifacts

## Objective

Create the feature specification, implementation plan, task files, ExecPlans, diary, and index entry for UI article endpoints.

## Scope

This task includes:

- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/ui-endpoints/PLAN.md`
- `docs/specs/ui-endpoints/tasks/*.md`
- `docs/specs/ui-endpoints/plans/*.execplan.md`
- `docs/specs/ui-endpoints/DIARY.md`
- `docs/specs/INDEX.md` update.
- Gateway convention update for the narrow article-delete artifact cleanup operation.

## Out of Scope

This task does not include:

- Gateway route implementation.
- UI implementation.
- SQLite schema changes.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/BOOKKEEPING.md`
- `docs/conventions/GATEWAY.md`

## Acceptance Criteria

```gherkin
Scenario: UI endpoint artifacts exist
  Given the UI endpoint feature is planned
  When repository documentation is inspected
  Then the feature has a SPEC, PLAN, DIARY, task files, and linked ExecPlans
  And the feature appears in docs/specs/INDEX.md
```

## Done When

- Canonical UI endpoint artifacts exist.
- Gateway convention distinguishes read-only artifact access from article delete cleanup.
- Task status and `PLAN.md` record this task as done.

## Validation

Required checks:

```bash
git diff -- docs/specs/ui-endpoints docs/specs/INDEX.md docs/conventions/GATEWAY.md
```

## Dependencies

Depends on:

- None.

Blocks:

- `UIEND-002`
- `UIEND-003`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
