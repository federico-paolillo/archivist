---
id: MUSER-005
feature: user-id-parameterization
title: Integration validation
status: done
depends_on: [MUSER-002, MUSER-003, MUSER-004]
blocks: [MUSER-006]
parallel: false
exec_plan: ../plans/MUSER-005-integration-validation.execplan.md
canonical: true
---

# MUSER-005: Integration Validation

## Objective

Integrate reviewed Gateway, Worker, and documentation slices, then run cross-module validation.

## Scope

This task includes:

- Reviewing worker reports and diffs.
- Resolving conflicts.
- Running Gateway and Worker validation.
- Checking that source and canonical docs agree.

## Out of Scope

This task does not include new behavior beyond `MUSER-002`, `MUSER-003`, and `MUSER-004`.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/MUSER-005-integration-validation.execplan.md`
- `.agents/skills/archivist-integrator/SKILL.md`
- `.agents/skills/archivist-general/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Integrated feature validates
  Given Gateway and Worker slices are implemented
  When integration validation runs
  Then all required module checks pass or exact failures are recorded
```

## Done When

- Gateway and Worker validations pass or failures are recorded.
- `git diff --check` passes.
- Cross-module contract checks find no remaining runtime personal-user hardcoding outside auth bootstrap.

## Validation

Required checks:

```bash
git diff --check
cd src/gateway && dotnet format && dotnet build && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
```

## Dependencies

Depends on:

- `MUSER-002`
- `MUSER-003`
- `MUSER-004`

Blocks:

- `MUSER-006`

## ExecPlan

ExecPlan:

```text
../plans/MUSER-005-integration-validation.execplan.md
```

## Open Questions

- None.
