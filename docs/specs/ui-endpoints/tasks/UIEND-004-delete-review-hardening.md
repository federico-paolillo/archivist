---
id: UIEND-004
feature: ui-endpoints
title: Delete review hardening
status: done
depends_on: [UIEND-003]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# UIEND-004: Delete review hardening

## Objective

Close review findings around normal delete ULID normalization and Gateway delete SQLite/filesystem consistency documentation.

## Story / Context

As an authenticated UI user, I want normal delete to treat valid article IDs consistently with detail and force-delete routes. As the operator, I want the known SQLite/filesystem atomicity limitation documented for rebuilds and repairs.

## Scope

This task includes:

- Normalize valid ULID route values before normal delete service calls.
- Add behavior tests for valid non-canonical ULID casing where appropriate.
- Document the accepted v0 delete consistency limitation in `docs/DESIGN.md`, `docs/ARTIFACTS.md`, and affected specs.

## Out of Scope

This task does not include:

- Repair queues, tombstones, or new cleanup jobs.
- Changing the normal delete conflict semantics for running jobs.
- Changing force-delete stale eligibility.

## Inputs

Required context:

- `../SPEC.md`
- `../PLAN.md`
- `docs/DESIGN.md`
- `docs/ARTIFACTS.md`
- `docs/specs/job-recovery/SPEC.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Normal delete normalizes valid route IDs
  Given an existing article ID
  When the user calls DELETE /articles/{id} with a valid non-canonical casing
  Then Gateway deletes the same article that detail and force-delete routes would resolve
```

```gherkin
Scenario: Delete consistency limitation is rebuild-visible
  Given a rebuild from canonical docs
  Then delete behavior documents that SQLite and filesystem cleanup are not atomically rollbackable together
  And operator repair guidance exists for artifact-deleted/commit-failed state
```

## Done When

- Normal delete uses canonical ULID normalization.
- Delete consistency limitation is promoted to canonical docs.
- Gateway validation passes or failures are recorded.
