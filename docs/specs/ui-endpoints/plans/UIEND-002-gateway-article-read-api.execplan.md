---
id: UIEND-002-PLAN
task: ../tasks/UIEND-002-gateway-article-read-api.md
status: completed
canonical: true
---

# ExecPlan: UIEND-002 Gateway Article Read API

## Objective

Implement authenticated article list/detail routes backed by SQLite and deterministic read-only artifact access.

## Linked Task

- `../tasks/UIEND-002-gateway-article-read-api.md`

## Required Context

- `../tasks/UIEND-002-gateway-article-read-api.md`
- `../SPEC.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/ARTIFACTS.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/summary-generation/SPEC.md`

## Assumptions

- Article ULID lexical order is the v0 list order.
- `SQLITE_PATH` points at a database with the `articles` table defined by `telegram-ingestion`.
- `DATA_DIR` points at the deterministic article artifact root.

## Non-Goals

- Do not add search, filters, tags, key points, or structured summaries.
- Do not expose artifact write/delete operations through the read API.

## Implementation Sequence

1. Add lower-camel article API DTOs for list, cursors, detail, and error responses.
2. Add application article query services that use SQLite with authenticated user scoping.
3. Add read-only artifact access for `summary.md` and `content.md`.
4. Map `GET /articles` and `GET /articles/{id}` with `RequireAuthorization`.
5. Add integration tests for authentication, pagination, detail, malformed IDs, not found, and artifact strictness.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Documentation Updates Required

- Update task status, feature plan, and diary after implementation.

## Risks

- The current codebase may not yet contain the auth and persistence dependencies; any local bridge must stay within the documented contracts and avoid implementing login/logout behavior here.

## Rollback / Recovery Notes

- Removing the route mapping disables the read API without mutating stored article state.

## Completion Criteria

- Read endpoints and tests are implemented.
- Validation passes or failures are recorded.
