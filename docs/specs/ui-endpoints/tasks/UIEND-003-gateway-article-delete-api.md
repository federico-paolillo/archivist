---
id: UIEND-003
feature: ui-endpoints
title: Gateway article delete API
depends_on: [AUTHN-004, TELING-001, ARTPROC-005, UIEND-002]
blocks: [UI-004]
parallel: false
requires_exec_plan: false
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
- Force deletion only when the shared Gateway force-delete eligibility service or predicate, also used by `UIEND-002` detail `canForceDelete`, reports that all associated running jobs are stale.
- SQLite write-transaction serialization with Worker job claim behavior owned by `ARTPROC-005`.
- SQLite/filesystem consistency limitation handling documented in `SPEC.md`.
- Gateway integration tests for delete behavior.


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
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`

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

Scenario: Force delete uses the shared Gateway eligibility rule
  Given Gateway evaluates article job state for force-delete enforcement
  When the request reaches delete enforcement
  Then it uses the same application service or predicate as detail canForceDelete

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

Scenario: Delete serializes with Worker claim
  Given a queued job exists for an article
  When Gateway delete and Worker job claim race
  Then SQLite write transactions serialize the operations
  And Worker claim does not claim a job whose article row has been deleted
```

## Done When

- `DELETE /articles/{id}` and `DELETE /articles/{id}/force` are implemented.
- Tests cover ready, failed, queued, running-job conflict, stale force delete, active running-job force-delete conflict, missing artifact directory, associated row removal, artifact directory removal, ownership scoping, ULID normalization, same-origin rejection, artifact cleanup rollback, shared predicate/service coverage proving force-delete enforcement uses the same rules as detail `canForceDelete`, and delete/Worker-claim serialization.
- Required validation passes.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run test
```

## Dependencies

Depends on:
- `AUTHN-004`
- `TELING-001`
- `ARTPROC-005`
- `UIEND-002`

Blocks:

- `UI-004`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.
