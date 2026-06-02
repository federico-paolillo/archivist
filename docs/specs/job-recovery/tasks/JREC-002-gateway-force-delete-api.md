---
id: JREC-002
feature: job-recovery
title: Gateway force delete API
status: done
depends_on: [JREC-001]
blocks: [JREC-003, JREC-005]
parallel: true
exec_plan: null
canonical: true
---

# JREC-002: Gateway Force Delete API

## Objective

Implement authenticated force deletion for articles whose associated running jobs are stale, and expose article detail metadata for UI force-delete eligibility.

## Scope

This task includes:

- `DELETE /articles/{id}/force`.
- Same-origin unsafe-method protection.
- Authenticated ownership checks.
- SQLite write transaction with stale-running-job recheck.
- Article detail `canForceDelete`.
- Gateway tests.

## Out of Scope

This task does not include:

- UI implementation.
- Worker logging.
- Schema migrations.
- Automatic retries or requeue behavior.

## Inputs

- `../SPEC.md`
- Existing `DELETE /articles/{id}` implementation.
- Existing article detail API implementation.
- Existing artifact deletion abstraction.

## Outputs

- Gateway route and service behavior for stale force delete.
- Updated article detail response contract.
- Tests covering force delete and normal delete preservation.

## Expected Affected Areas

```text
src/gateway/Archivist.Gateway.Api/Articles/**
src/gateway/Archivist.Gateway.Application/Articles/**
src/gateway/Archivist.Gateway.Application/Persistence/**
src/gateway/Archivist.Gateway.Tests/**
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./JREC-002-gateway-force-delete-api.md`
- `docs/ARCHITECTURE.md`
- `docs/ARTIFACTS.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Force delete removes stale running job state
  Given an authenticated user owns an article with a running job started more than 2 hours ago
  When DELETE /articles/{id}/force is called with a same-origin request
  Then the response is 204
  And the article, jobs, notifications, and artifact directory are removed

Scenario: Force delete rejects active running jobs
  Given an authenticated user owns an article with a running job started less than 2 hours ago
  When DELETE /articles/{id}/force is called
  Then the response is 409
  And no database or artifact state is removed

Scenario: Article detail reports force delete eligibility
  Given an authenticated user owns an article with only stale running jobs
  When GET /articles/{id} is called
  Then the response includes canForceDelete true
```

## Done When

- Force delete route is implemented and authenticated.
- Detail response includes `canForceDelete`.
- Normal delete still rejects running jobs.
- Gateway tests cover success and failure paths.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual validation, if any:

- None.

## Dependencies

Depends on:

- `JREC-001`

Blocks:

- `JREC-003`
- `JREC-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Implemented and reviewed through the multi-agent Gateway worker/reviewer workflow.
- Validation: `dotnet format` passed; `dotnet build` passed; focused non-host rollback test passed; integrated `go tool lefthook run test` passed the Gateway test target with 179 tests.
- Direct `dotnet test` and serialized `dotnet test --no-build -- RunConfiguration.MaxCpuCount=1` stalled after test discovery in this environment and were terminated with `killall dotnet`; the same WebApplicationFactory/testhost hang appeared during Gateway worker/reviewer validation and did not report a failed assertion.
