---
id: UI-003
feature: ui
title: Article master-detail view and delete workflow
status: blocked
depends_on: [UI-002, UIEND-002, UIEND-003]
blocks: [UI-004]
parallel: false
exec_plan: ../plans/UI-003-article-master-detail-and-delete-workflow.execplan.md
canonical: true
---

# UI-003: Article master-detail view and delete workflow

## Objective

Implement the authenticated article master-detail view, article detail states, safe Markdown rendering, Original action, and delete confirmation workflow.

## Story / Context

As the personal Archivist user, I want to scan archived articles, open a selected article, see processing or failure state, open the source URL, and delete records I no longer need.

## Scope

This task includes:

- Article list loading from `GET ${apiBasePath}/articles`.
- Master-detail layout matching `docs/design/view.png`.
- Article row selection and URL update to `/articles/<article_id>`.
- Detail loading spinner.
- Blank black detail pane for `/articles` with no selected id.
- Ready, queued/non-ready, failed, detail-fetch-error, and delete-error states.
- Summary/content Markdown rendering with raw HTML disabled or sanitized.
- `Original` action using `canonical_url` when present, otherwise `original_url`.
- `Delete` action and confirmation modal.
- Delete cancel and confirm behavior.
- List/detail state reset after successful delete.
- Frontend tests for article states and delete behavior.

## Out of Scope

This task does not include:

- Auth endpoint implementation.
- Article endpoint implementation.
- Retry or requeue behavior.
- Search, filtering, sorting controls, or pagination UI beyond the first fixed page.
- Direct SQLite or filesystem access.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `GET /articles`, `GET /articles/{id}`, and `DELETE /articles/{id}` from `ui-endpoints`.
- Auth shell and API client from `UI-002`.
- `docs/design/DESIGN.md`
- `docs/design/view.png`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Authenticated `/articles` and `/articles/<article_id>` render the article surface.
- Selecting an article updates the route and detail state.
- Ready, queued/non-ready, failed, and error states match `SPEC.md`.
- Delete confirmation calls the API only on `Yes` and resets the route on success.

## Expected Affected Areas

```text
src/ui/src/
src/ui/src/app.css
src/ui/package.json
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/UI-003-article-master-detail-and-delete-workflow.execplan.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/UI.md`
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

Scenario: Queued article says to return later
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
```

## Done When

- Article master-detail UI is implemented.
- Delete modal behavior is implemented.
- Markdown content rendering cannot execute raw article HTML or scripts.
- Tests cover route update, loading, ready, queued, failed, fetch-error, no-id, and delete confirm/cancel states.
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

- `UI-004`

## ExecPlan

ExecPlan:

```text
../plans/UI-003-article-master-detail-and-delete-workflow.execplan.md
```

## Open Questions

- None.
