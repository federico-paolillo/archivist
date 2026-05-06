---
id: UIEND-003
feature: ui-endpoints
title: Gateway article delete API
status: blocked
depends_on: [UIEND-001, AUTHN-003, TELING-001]
blocks: []
parallel: true
exec_plan: ../plans/UIEND-003-gateway-article-delete-api.execplan.md
canonical: true
---

# UIEND-003: Gateway Article Delete API

## Objective

Implement authenticated hard deletion for UI article administration.

## Scope

This task includes:

- `DELETE /articles/{id}`
- Same-origin unsafe-method protection.
- Authenticated user scoping.
- Hard deletion of article rows, associated jobs, associated notifications, and the deterministic artifact directory.
- Rejection of delete when any associated job is `running`.
- Gateway integration tests for delete behavior.

## Out of Scope

This task does not include:

- Soft delete or tombstones.
- Deleting running jobs.
- UI implementation.
- Notification retries or worker cancellation.

## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Api/Articles/
src/gateway/Archivist.Gateway.Application/Articles/
src/gateway/Archivist.Gateway.Tests/Api/
docs/conventions/GATEWAY.md
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
- `../plans/UIEND-003-gateway-article-delete-api.execplan.md`

## Acceptance Criteria

```gherkin
Scenario: Delete queued article
  Given the authenticated user owns a queued article
  And the article has queued jobs and notifications
  When the browser requests DELETE /articles/{id}
  Then the response status is 204
  And the article, jobs, notifications, and artifact directory are removed

Scenario: Delete rejects running job
  Given the authenticated user owns an article with a running job
  When the browser requests DELETE /articles/{id}
  Then the response status is 409
```

## Done When

- `DELETE /articles/{id}` is implemented.
- Tests cover ready, failed, queued, running-job conflict, missing artifact directory, associated row removal, artifact directory removal, and same-origin rejection.
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

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
../plans/UIEND-003-gateway-article-delete-api.execplan.md
```

## Open Questions

- None.
