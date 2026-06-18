---
id: UI-004
feature: ui
title: Article destructive actions
depends_on: [UI-003, UIEND-003]
blocks: []
parallel: false
requires_exec_plan: false
canonical: true
---
# UI-004: Article Destructive Actions

## Objective

Implement normal delete and stale force-delete workflows on top of the read-only article master-detail surface.

## Story / Context

As the personal Archivist user, I want explicit confirmation before deleting archived records, and I want stale force-delete recovery only when Gateway says recovery is available.

## Scope

This task includes:

- `Delete` action visible whenever selected article metadata is available, including ready, queued, and failed states.
- Delete confirmation modal with the exact question `Are you sure?` and exact options `Yes` and `Nevermind`.
- Delete cancel and confirm behavior.
- API call to `DELETE ${apiBasePath}/articles/{id}` only after confirmation.
- Successful normal delete route reset to `/articles`, detail clear, and master-list removal.
- Delete failure text in red in the detail pane while preserving the selected URL.
- `Force Delete` action visible only when Gateway detail `canForceDelete` is true.
- Separate destructive confirmation before force delete.
- API call to `DELETE ${apiBasePath}/articles/{id}/force` only after confirmation.
- Successful force delete route reset to `/articles`, detail clear, and master-list removal.
- Force-delete failure text in red in the detail pane while preserving the selected URL.
- Frontend tests for normal delete and force-delete behavior.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Read-only article surface and loaded detail model from `UI-003`.
- `DELETE /articles/{id}`, `DELETE /articles/{id}/force`, and `canForceDelete` from `ui-endpoints`.
- Auth shell, route guards, and API client from `UI-002`.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Delete confirmation calls the API only on `Yes` and resets the route on success.
- Force Delete is available only from `canForceDelete`, uses its own confirmation path, calls the force-delete API only on confirmation, and resets the route on success.
- Delete and force-delete failures keep the selected article route and display red failure text in the detail pane.

## Expected Affected Areas

```text
src/ui/src/
src/ui/src/pages/articles/
src/ui/src/pages/articles/components/
src/ui/package.json
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-ui/SKILL.md`
- `./UI-003-article-master-detail-read-only.md`
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/ui-endpoints/tasks/UIEND-003-gateway-article-delete-api.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Delete confirmation controls API call
  Given an article is selected
  When the user clicks Delete and chooses Nevermind
  Then no DELETE request is sent
  When the user clicks Delete and chooses Yes
  Then DELETE is sent to the configured API base article endpoint
  And the route resets to /articles after success

Scenario: Delete failure leaves selection visible
  Given an article is selected
  When confirmed delete fails
  Then the current URL is unchanged
  And the detail pane shows failure text in red

Scenario: Force delete visibility follows Gateway metadata
  Given the selected article detail has canForceDelete true
  When the article detail renders
  Then Force Delete is visible
  Given the selected article detail has canForceDelete false
  When the article detail renders
  Then Force Delete is not visible

Scenario: Force delete confirmation controls API call
  Given Force Delete is visible
  When the user opens the force-delete confirmation and cancels
  Then no force-delete request is sent
  When the user confirms force delete
  Then DELETE is sent to the configured API base force-delete endpoint
  And the route resets to /articles after success

Scenario: Force delete failure leaves selection visible
  Given Force Delete is visible
  When confirmed force delete fails
  Then the current URL is unchanged
  And the detail pane shows failure text in red
```

## Done When

- Delete modal behavior is implemented.
- Force-delete visibility, modal behavior, success, and failure behavior are implemented.
- Tests cover delete confirm/cancel states, delete success, delete failure, force-delete visibility, force-delete confirm/cancel states, force-delete success, and force-delete failure.
- Required validation passes.

## Validation

Required checks:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual validation:

- Capture `/articles/<article_id>` with delete and force-delete fixtures and compare destructive controls against `docs/design/view.png` and `docs/design/DESIGN.md`.

## Dependencies

Depends on:

- `UI-003`
- `UIEND-003`

Blocks:

- None.

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.
