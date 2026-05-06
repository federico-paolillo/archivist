---
id: UIEND-003-PLAN
task: ../tasks/UIEND-003-gateway-article-delete-api.md
status: proposed
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

## Non-Goals

- Do not add worker cancellation.
- Do not add tombstones or soft-delete state.
- Do not delete user rows.

## Implementation Sequence

1. Add delete service behavior that checks article ownership and running-job conflicts.
2. Remove associated notifications, jobs, article row, and artifact directory as one handler operation.
3. Keep delete artifact cleanup separate from read-only artifact access.
4. Map `DELETE /articles/{id}` with `RequireAuthorization` and same-origin enforcement.
5. Add integration tests for success, conflict, not found, malformed IDs, missing artifacts, row cleanup, directory cleanup, and cross-site rejection.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Documentation Updates Required

- Update task status, feature plan, and diary after implementation.

## Risks

- Filesystem and SQLite changes are not an atomic distributed transaction. The v0 handler performs both in one operation and fails before commit when artifact cleanup fails.

## Rollback / Recovery Notes

- If delete is disabled after deployment, existing articles and artifacts remain readable through list/detail.

## Completion Criteria

- Delete endpoint and tests are implemented.
- Validation passes or failures are recorded.
