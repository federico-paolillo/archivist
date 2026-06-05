---
id: MUSER-005-PLAN
task: ../tasks/MUSER-005-integration-validation.md
status: completed
canonical: true
---

# ExecPlan: MUSER-005 Integration Validation

## Objective

Integrate Gateway, Worker, and documentation slices and validate the cross-module ownership contract.

## Linked Task

- `../tasks/MUSER-005-integration-validation.md`

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../tasks/MUSER-005-integration-validation.md`
- Worker and Gateway implementation reports.
- Reviewer findings.

## Assumptions

- Gateway and Worker slices are developed on disjoint file scopes.
- Coordinator owns final ALM status updates.

## Non-Goals

- Adding new product behavior after review starts.

## Implementation Sequence

1. Inspect Gateway and Worker diffs for scope and contract alignment.
2. Run targeted searches for remaining personal-user hardcoding outside allowed bootstrap/test contexts.
3. Run Gateway validation.
4. Run Worker validation.
5. Run `git diff --check` and Compose config validation when Docker is available.
6. Record validation and any unavailable commands in `DIARY.md`.

## Validation Plan

```bash
git diff --check
cd src/gateway && dotnet format && dotnet build && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
docker compose --env-file .env.example config --quiet
```

## Documentation Updates Required

- `../DIARY.md`
- `../PLAN.md`
- Task status files as validation completes.

## Risks

- Full validation can expose pre-existing environment failures unrelated to the feature; record exact output.

## Rollback / Recovery Notes

If integration fails after worker slices land, revert the failing slice and return its task to `ready` or `blocked` with findings.

## Completion Criteria

- Validation is complete or blocked with recorded reasons.
- Reviewer findings are resolved or explicitly waived by coordinator.
