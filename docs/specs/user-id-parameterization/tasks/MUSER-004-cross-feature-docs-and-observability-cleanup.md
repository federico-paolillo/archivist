---
id: MUSER-004
feature: user-id-parameterization
title: Cross-feature docs and observability cleanup
status: done
depends_on: [MUSER-001]
blocks: [MUSER-005]
parallel: true
exec_plan: null
canonical: true
---

# MUSER-004: Cross-Feature Docs And Observability Cleanup

## Objective

Align existing completed feature specs with the user-id parameterization contract without changing product scope.

## Scope

This task includes updates to existing canonical specs that still describe runtime personal-user hardcoding or omit `user_id` telemetry.

## Out of Scope

This task does not include source code changes.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../../../ARCHITECTURE.md`
- `../../../DESIGN.md`
- `../../authn/SPEC.md`
- `../../telegram-ingestion/SPEC.md`
- `../../ui-endpoints/SPEC.md`
- `../../otel-observability/SPEC.md`
- `.agents/skills/archivist-general/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Existing specs do not contradict the new contract
  Given the user-id parameterization feature is canonical
  When a rebuild agent reads existing auth, Telegram, UI endpoint, and observability specs
  Then those specs do not require runtime personal-user hardcoding
```

## Done When

- Existing specs align with `MUSER` requirements.
- Snapshotter remains explicitly excluded from `user_id` telemetry.

## Validation

Required checks:

```bash
git diff --check
```

## Dependencies

Depends on:

- `MUSER-001`

Blocks:

- `MUSER-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
