---
id: JREC-003
feature: job-recovery
title: UI force delete workflow
status: done
depends_on: [JREC-001, JREC-002]
blocks: [JREC-005]
parallel: true
exec_plan: null
canonical: true
---

# JREC-003: UI Force Delete Workflow

## Objective

Expose a distinct force-delete action for article detail views only when the Gateway says stale force deletion is available.

## Scope

This task includes:

- API client support for `DELETE /articles/{id}/force`.
- Article detail type support for `canForceDelete`.
- Force Delete action visibility.
- Separate destructive confirmation.
- UI success, failure, and auth-expiry behavior.
- UI tests.

## Out of Scope

This task does not include:

- Gateway implementation.
- Worker logging.
- UI-computed stale eligibility from timestamps.
- Retry or reprocess controls.

## Inputs

- `../SPEC.md`
- Gateway detail field `canForceDelete`.
- Gateway force-delete route `DELETE /articles/{id}/force`.
- Existing delete UI workflow.

## Outputs

- UI force-delete workflow with tests.

## Expected Affected Areas

```text
src/ui/src/deps.ts
src/ui/src/pages/articles/**
src/ui/src/**/*.test.tsx
src/ui/src/**/*.test.ts
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./JREC-003-ui-force-delete-workflow.md`
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-ui/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Force delete is visible when allowed
  Given article detail returns canForceDelete true
  When the selected article detail renders
  Then the Force Delete action is visible

Scenario: Force delete is hidden when not allowed
  Given article detail returns canForceDelete false
  When the selected article detail renders
  Then the Force Delete action is not visible

Scenario: Force delete succeeds
  Given Force Delete is visible
  When the user confirms the force-delete dialog
  Then the UI calls DELETE /articles/{id}/force
  And navigates to /articles
  And removes the deleted article from the list
```

## Done When

- UI consumes `canForceDelete`.
- UI exposes a distinct Force Delete action only when allowed.
- Force delete has a separate confirmation path.
- Tests cover visibility, success, failure, cancel, and auth expiry.

## Validation

Required checks:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual validation, if any:

- Browser validation if a local server is started during integration.

## Dependencies

Depends on:

- `JREC-001`
- `JREC-002`

Blocks:

- `JREC-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Implemented and reviewed through the multi-agent UI worker/reviewer workflow.
- Validation: `npm run format`, `npm run lint`, `npm run build`, and `npm run test` passed; integrated `go tool lefthook run test` passed the UI vitest target with 27 tests.
