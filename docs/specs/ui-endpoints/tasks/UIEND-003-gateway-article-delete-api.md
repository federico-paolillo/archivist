---
id: UIEND-003
feature: ui-endpoints
title: Gateway article delete API
status: done
depends_on: [AUTHN-004, TELING-001, UIEND-002]
blocks: [UI-003]
parallel: false
exec_plan: null
canonical: true
---

# UIEND-003: Gateway Article Delete API

## Objective

Implement authenticated normal hard deletion and stale force deletion for UI article administration.

## Scope

This task includes:

- `DELETE /articles/{id}`
- `DELETE /articles/{id}/force`
- Same-origin unsafe-method protection.
- Authenticated session user scoping.
- ULID route normalization for normal delete and force delete.
- Hard deletion of article rows, associated jobs, associated notifications, and the deterministic artifact directory.
- Rejection of normal delete when any associated job is `running`.
- Force deletion only when all associated running jobs are stale.
- SQLite write-transaction serialization with worker job claim.
- SQLite/filesystem consistency limitation handling documented in `SPEC.md`.
- Gateway integration tests for delete behavior.


## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Api/Articles/
src/gateway/Archivist.Gateway.Application/Articles/
src/gateway/Archivist.Gateway.Tests/Api/
.agents/skills/archivist-gateway/SKILL.md
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

Scenario: Force delete removes stale running job state
  Given the authenticated user owns an article with a running job started more than 2 hours ago
  When the browser requests DELETE /articles/{id}/force
  Then the response status is 204
  And the article, jobs, notifications, and artifact directory are removed

Scenario: Force delete rejects active running job
  Given the authenticated user owns an article with a running job started less than 2 hours ago
  When the browser requests DELETE /articles/{id}/force
  Then the response status is 409
  And the article, job, notifications, and artifact directory remain

Scenario: Delete enforces ownership
  Given user "U1" owns an article
  And user "U2" is authenticated
  When user "U2" requests DELETE /articles/{id}
  Then the response status is 404
```

## Done When

- `DELETE /articles/{id}` and `DELETE /articles/{id}/force` are implemented.
- Tests cover ready, failed, queued, running-job conflict, stale force delete, active running-job force-delete conflict, missing artifact directory, associated row removal, artifact directory removal, ownership scoping, ULID normalization, same-origin rejection, and artifact cleanup rollback.
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
- `AUTHN-004`
- `TELING-001`
- `UIEND-002`

Blocks:

- `UI-003`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
