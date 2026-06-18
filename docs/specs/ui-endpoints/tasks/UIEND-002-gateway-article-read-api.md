---
id: UIEND-002
feature: ui-endpoints
title: Gateway article read API
depends_on: [AUTHN-004, TELING-001, SUMGEN-005]
blocks: [UIEND-003, UI-003]
parallel: false
requires_exec_plan: false
canonical: true
---
# UIEND-002: Gateway Article Read API

## Objective

Implement authenticated Gateway article list and detail APIs for the UI.

## Scope

This task includes:

- `GET /articles`
- `GET /articles/{id}`
- Article ULID cursor validation and route normalization.
- Fixed 25-item list pages.
- Lower-camel JSON DTOs for list and detail responses.
- Authenticated session user scoping.
- Read-only artifact reads for `summary.md` and `content.md`.
- Server-computed `canForceDelete` article detail metadata through the shared Gateway force-delete eligibility service or predicate also used by `UIEND-003`.
- Gateway integration tests for auth, pagination, detail, and artifact behavior.


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
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/ARTIFACTS.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/summary-generation/SPEC.md`

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

Scenario: Article detail reports force-delete metadata
  Given the authenticated user owns an article with only stale running jobs
  When the browser requests GET /articles/{id}
  Then the response contains canForceDelete true

Scenario: Article detail uses the shared force-delete rule
  Given Gateway evaluates article job state for canForceDelete
  When the detail response is produced
  Then it uses the same application service or predicate as force-delete enforcement

Scenario: Article detail enforces ownership
  Given user "U1" owns an article
  And user "U2" is authenticated
  When user "U2" requests GET /articles/{id}
  Then the response status is 404
```

## Done When

- `GET /articles` and `GET /articles/{id}` are implemented.
- Detail responses include `canForceDelete`.
- Tests cover unauthenticated access, ownership scoping, pagination, invalid cursors, missing article, malformed IDs, ULID normalization, ready article artifacts, queued/failed nullable artifacts, force-delete eligibility metadata, and shared predicate/service coverage proving detail `canForceDelete` uses the same rules as force-delete enforcement.
- Required validation passes.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:
- `AUTHN-004`
- `TELING-001`
- `SUMGEN-005`

Blocks:

- `UIEND-003`
- `UI-003`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.
