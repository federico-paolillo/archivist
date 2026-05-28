---
id: UIEND-002
feature: ui-endpoints
title: Gateway article read API
status: done
depends_on: [UIEND-001, AUTHN-003, TELING-001, SUMGEN-005]
blocks: [UI-003]
parallel: false
exec_plan: ../plans/UIEND-002-gateway-article-read-api.execplan.md
canonical: true
---

# UIEND-002: Gateway Article Read API

## Objective

Implement authenticated Gateway article list and detail APIs for the UI.

## Scope

This task includes:

- `GET /articles`
- `GET /articles/{id}`
- Article ULID cursor validation.
- Fixed 25-item list pages.
- Lower-camel JSON DTOs for list and detail responses.
- Authenticated user scoping.
- Read-only artifact reads for `summary.md` and `content.md`.
- Gateway integration tests for auth, pagination, detail, and artifact behavior.

## Out of Scope

This task does not include:

- Delete behavior.
- UI implementation.
- Login/logout implementation.
- SQLite schema creation or migrations.

## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Api/Articles/
src/gateway/Archivist.Gateway.Application/Articles/
src/gateway/Archivist.Gateway.Tests/Api/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/summary-generation/SPEC.md`
- `../plans/UIEND-002-gateway-article-read-api.execplan.md`

## Acceptance Criteria

```gherkin
Scenario: Article list requires authentication
  Given no valid auth cookie is present
  When the browser requests GET /articles
  Then the response status is 401

Scenario: Article detail returns Markdown artifacts
  Given the authenticated user owns a ready article
  And summary.md and content.md exist
  When the browser requests GET /articles/{id}
  Then the response contains summaryMarkdown and contentMarkdown
```

## Done When

- `GET /articles` and `GET /articles/{id}` are implemented.
- Tests cover unauthenticated access, pagination, invalid cursors, missing article, malformed IDs, ready article artifacts, and queued/failed nullable artifacts.
- Validation passes or failures are recorded.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `UIEND-001`
- `AUTHN-003`
- `TELING-001`
- `SUMGEN-005`

Blocks:

- `UI-003`

## ExecPlan

ExecPlan:

```text
../plans/UIEND-002-gateway-article-read-api.execplan.md
```

## Open Questions

- None.
