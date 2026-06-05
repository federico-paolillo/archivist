---
id: MUSER-006
feature: user-id-parameterization
title: Review and ALM closure
status: done
depends_on: [MUSER-005]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# MUSER-006: Review And ALM Closure

## Objective

Record final reviews, update ALM status, append diary entries, and close the feature.

## Scope

This task includes:

- Gateway and Worker reviewer outcomes.
- Final task and plan statuses.
- Feature diary entries.
- Feature index status.

## Out of Scope

This task does not include new runtime behavior.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../../../BOOKKEEPING.md`
- `.agents/skills/archivist-reviewer/SKILL.md`
- `.agents/skills/archivist-general/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Feature closure is recorded
  Given the implementation is validated
  When ALM closure is performed
  Then task, plan, diary, and index state reflect the final outcome
```

## Done When

- All `MUSER` tasks are updated.
- `DIARY.md` records implementation and validation.
- `docs/specs/INDEX.md` reflects the final feature status.

## Validation

Required checks:

```bash
git diff --check
```

## Dependencies

Depends on:

- `MUSER-005`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
