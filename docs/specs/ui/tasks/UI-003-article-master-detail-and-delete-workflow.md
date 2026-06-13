---
id: UI-003
feature: ui
title: Article master-detail view and delete workflow
status: done
depends_on: [UI-002, UIEND-002, UIEND-003]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# UI-003: Article master-detail view and delete workflow

## Objective

Implement the authenticated article master-detail view, article detail states, safe Markdown rendering, Original action, normal delete confirmation workflow, and stale force-delete workflow.

## Story / Context

As the personal Archivist user, I want to scan archived articles, open a selected article, see processing or failure state, open the source URL, delete records I no longer need, and clean up stale running-job state when Gateway says recovery is available.

## Scope

This task includes:

- Article list loading from `GET ${apiBasePath}/articles`.
- Master-detail layout matching `docs/design/view.png`.
- Desktop/tablet article layout where the shell is viewport-framed and the master and detail panes scroll independently.
- Mobile stacked layout for a 430x960 CSS-pixel viewport with a capped, internally scrollable master list and unbounded detail content.
- Single-line footer version label with CSS-only ellipsis for long values such as commit hashes.
- Article row selection and URL update to `/articles/<article_id>`.
- Detail loading spinner.
- Blank black detail pane for `/articles` with no selected id.
- Ready, queued/non-ready, failed, detail-fetch-error, and delete-error states.
- Summary/content Markdown rendering with raw HTML disabled or sanitized.
- `Original` action using `canonicalUrl` when present, otherwise `originalUrl`.
- `Delete` action and confirmation modal.
- Delete cancel and confirm behavior.
- `Force Delete` action visible only when `canForceDelete` is true.
- Separate force-delete confirmation and API call to `DELETE ${apiBasePath}/articles/{id}/force`.
- Force-delete success and failure behavior.
- List/detail state reset after successful delete.
- Frontend tests for article states, delete behavior, and force-delete behavior.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- `GET /articles`, `GET /articles/{id}`, `DELETE /articles/{id}`, `DELETE /articles/{id}/force`, and `canForceDelete` from `ui-endpoints`.
- Auth shell and API client from `UI-002`.
- `docs/design/DESIGN.md`
- `docs/design/view.png`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Authenticated `/articles` and `/articles/<article_id>` render the article surface.
- Selecting an article updates the route and detail state.
- Ready, queued/non-ready, failed, and error states match `SPEC.md`.
- Delete confirmation calls the API only on `Yes` and resets the route on success.
- Force Delete is available only from `canForceDelete`, uses its own confirmation path, calls the force-delete API only on confirmation, and resets the route on success.

## Expected Affected Areas

```text
src/ui/src/
src/ui/src/app.css
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
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/authn/SPEC.md`
- `docs/design/DESIGN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Articles route without id
  Given the user is authenticated
  When the browser opens /articles
  Then the master list is visible
  And the detail pane is blank and black

Scenario: Article selection loads detail
  Given the article list is loaded
  When the user clicks an article
  Then the URL changes to /articles/{article_id}
  And a spinner appears until detail loading completes or fails

Scenario: Queued article detail message
  Given the selected article has status queued
  When detail loading succeeds
  Then the detail pane shows centered white text "Come back later."

Scenario: Failed article shows persisted error
  Given the selected article has status failed
  When detail loading succeeds
  Then the detail pane shows the persisted error message in red and centered

Scenario: Delete confirmation controls API call
  Given an article is selected
  When the user clicks Delete and chooses Nevermind
  Then no DELETE request is sent
  When the user clicks Delete and chooses Yes
  Then DELETE is sent to the configured API base article endpoint

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
```

## Done When

- Article master-detail UI is implemented.
- Delete modal behavior is implemented.
- Force-delete visibility, modal behavior, success, and failure behavior are implemented.
- Markdown content rendering cannot execute raw article HTML or scripts.
- Tests cover route update, loading, ready, queued, failed, fetch-error, no-id, delete confirm/cancel states, force-delete visibility, force-delete confirm/cancel states, force-delete success, and force-delete failure.
- Validation passes or failures are recorded.

## Validation

Required checks:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual validation:

- Capture `/articles` and `/articles/<article_id>` in a browser and compare against `docs/design/view.png` and `docs/design/DESIGN.md`.

## Dependencies

Depends on:

- `UI-002`
- `UIEND-002`
- `UIEND-003`

Blocks:

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
