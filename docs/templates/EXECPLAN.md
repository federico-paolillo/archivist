---
id: <TASK-ID>-PLAN
task: ../tasks/<TASK-ID>-<task-slug>.md
canonical: false
---
# ExecPlan: <TASK-ID> <Task Title>

ExecPlans are active-run planning artifacts, not canonical rebuild artifacts. Create one when the linked task has `requires_exec_plan: true` or when the execution coordinator needs stepwise implementation guidance. This plan must not add requirements beyond the linked task, feature spec, and feature plan.

## Objective

Describe the task objective and the expected implementation outcome.

## Linked Task

- `../tasks/<TASK-ID>-<task-slug>.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/<TASK-ID>-<task-slug>.md`

Add only ExecPlan-specific context:

- TODO: additional `docs/ARCHITECTURE.md` sections, only if needed
- TODO: additional `docs/DESIGN.md` decisions, only if needed
- TODO: additional `.agents/skills/<relevant-skill>/SKILL.md`, only if needed

## Task Requirements Addressed

- TODO: summarize requirements already present in the linked task/spec/plan. Do not introduce new requirements here.

## Constraints

- TODO

## Assumptions

- TODO

## Implementation Sequence

1. TODO
2. TODO
3. TODO

## Validation Plan

```bash
# TODO: add commands
```

Manual checks:

- TODO

## Documentation Updates Required

- TODO

## Risks

- TODO

## Rollback / Recovery Notes

- TODO

## Completion Criteria

- TODO
