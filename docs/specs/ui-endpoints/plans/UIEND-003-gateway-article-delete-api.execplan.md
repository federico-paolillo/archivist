---
id: UIEND-003-PLAN
task: ../tasks/UIEND-003-gateway-article-delete-api.md
status: completed
canonical: true
---

# ExecPlan: UIEND-003 Gateway Article Delete API

## Objective

Implement authenticated article hard deletion with SQLite cleanup and deterministic artifact directory removal.

## Linked Task

- `../tasks/UIEND-003-gateway-article-delete-api.md`

## Required Context

- `../tasks/UIEND-003-gateway-article-delete-api.md`
- `../SPEC.md`
- `docs/conventions/GATEWAY.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`

## Assumptions

- Deleting queued jobs is allowed.
- Deleting running jobs is not allowed and returns `409`.
- Missing artifact directories are already-cleaned state and do not fail deletion.
- Delete and worker claim serialize through SQLite write transactions; whichever transition commits first determines the visible outcome.

## Non-Goals

- Do not add worker cancellation.
- Do not add tombstones or soft-delete state.
- Do not delete user rows.

## Implementation Sequence

1. Add delete service behavior that starts a SQLite write transaction, checks article ownership, and rechecks running-job conflicts inside that transaction.
2. Ensure worker job claim only claims queued jobs whose article row still exists, so a delete that commits first leaves no claimable job.
3. Remove associated notifications, jobs, article row, and artifact directory as one handler operation; commit the database delete only after artifact cleanup succeeds or is known to be unnecessary.
4. Keep delete artifact cleanup separate from read-only artifact access.
5. Map `DELETE /articles/{id}` with `RequireAuthorization` and same-origin enforcement.
6. Add integration tests for success, delete/claim race ordering, conflict, not found, malformed IDs, missing artifacts, row cleanup, directory cleanup, and cross-site rejection.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Documentation Updates Required

- Update task status, feature plan, and diary after implementation.

## Risks

- Filesystem and SQLite changes are not an atomic distributed transaction. The v0 handler deletes database rows inside an open SQLite write transaction, performs artifact cleanup before commit, and rolls back the database deletes if cleanup fails. Cleanup failure returns `500` and leaves article state intact.

## Rollback / Recovery Notes

- If delete is disabled after deployment, existing articles and artifacts remain readable through list/detail.

## Completion Criteria

- Delete endpoint and tests are implemented.
- Validation passes or failures are recorded.

Completion note:

- Completed on 2026-05-12 under explicit user assignment override from proposed guidance.
