---
id: MUSER-001
feature: user-id-parameterization
title: Canonical user-id resolution contract
status: done
depends_on: []
blocks: [MUSER-002, MUSER-003, MUSER-004]
parallel: false
exec_plan: ../plans/MUSER-001-canonical-user-id-resolution-contract.execplan.md
canonical: true
---

# MUSER-001: Canonical User-ID Resolution Contract

## Objective

Create the canonical feature contract for runtime `user_id` resolution and update global rebuild documents before code implementation.

## Scope

This task includes:

- Feature `SPEC.md`, `PLAN.md`, task files, and linked ExecPlans.
- `docs/specs/INDEX.md` entry.
- Architecture and design updates defining bootstrap-only hardcoding and database-backed runtime authorization.

## Out of Scope

This task does not include:

- Gateway or Worker source changes.
- User registration or user-management behavior.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../../../ARCHITECTURE.md`
- `../../../DESIGN.md`
- `.agents/skills/archivist-general/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Runtime user-id contract is canonical
  Given the feature artifacts exist
  When implementation workers start
  Then they can follow a documented rule for resolving user_id without hardcoding the personal ULID
```

## Done When

- The feature artifacts and index entry exist.
- Architecture and design docs describe the new ownership rule.
- The task's ExecPlan is accepted.

## Validation

Required checks:

```bash
git diff --check
```

## Dependencies

Depends on:

- None.

Blocks:

- `MUSER-002`
- `MUSER-003`
- `MUSER-004`

## ExecPlan

ExecPlan:

```text
../plans/MUSER-001-canonical-user-id-resolution-contract.execplan.md
```

## Open Questions

- None.
