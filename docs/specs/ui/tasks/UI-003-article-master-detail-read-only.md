---
id: UI-003
feature: ui
title: Article master-detail read-only view
depends_on: [UI-002, UIEND-002]
blocks: [UI-004]
parallel: false
requires_exec_plan: false
canonical: true
---
# UI-003: Article Master-Detail Read-Only View

## Objective

Implement the authenticated article list/detail surface, article detail states, safe Markdown rendering, Original action, and article shell controls without destructive actions.

## Story / Context

As the personal Archivist user, I want to scan archived articles, open a selected article, see processing or failure state, and open the source URL before deciding whether any destructive action is needed.

## Scope

This task includes:

- Article list loading from `GET ${apiBasePath}/articles`.
- Master-detail layout matching `docs/design/view.png`.
- Article shell title bar and user icon control containing only `Logout`, wired to the logout behavior from `UI-002`.
- Desktop/tablet article layout where the shell is viewport-framed and the master and detail panes scroll independently.
- Mobile stacked layout for a 430x960 CSS-pixel viewport with a capped, internally scrollable master list and unbounded detail content.
- Single-line footer version label with CSS-only ellipsis for long values such as commit hashes.
- Article row selection and URL update to `/articles/<article_id>`.
- Minimal programmatic selected state for the selected article row.
- Detail loading spinner.
- Blank black detail pane for `/articles` with no selected id.
- Ready, queued/non-ready, failed, and detail-fetch-error states.
- Summary/content Markdown rendering with raw HTML disabled or sanitized.
- `Original` action using `canonicalUrl` when present, otherwise `originalUrl`.
- Frontend tests for read-only article states, routing, shell controls, and Markdown safety.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `GET /articles`, `GET /articles/{id}`, and `canForceDelete` from `ui-endpoints`.
- Auth shell, route guards, logout behavior, and API client from `UI-002`.
- `docs/design/DESIGN.md`
- `docs/design/view.png`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Authenticated `/articles` and `/articles/<article_id>` render the article surface.
- Selecting an article updates the route and detail state.
- Ready, queued/non-ready, failed, and error states match `SPEC.md`.
- Raw article Markdown cannot execute scripts, raw HTML, inline event handlers, or `javascript:` links.
- `canForceDelete` is retained on the loaded detail model for the follow-up destructive-actions task, but this task does not render delete controls.

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
  And the selected row exposes a programmatic selected state
  And a spinner appears until detail loading completes or fails

Scenario: Ready article detail renders safely
  Given the selected article has status ready
  When detail loading succeeds
  Then the detail pane shows title, summary markdown, content markdown, and Original action
  And raw article HTML, scripts, inline event handlers, and javascript links do not execute

Scenario: Queued article detail message
  Given the selected article has status queued
  When detail loading succeeds
  Then the detail pane shows centered white text "Come back later."

Scenario: Failed article shows persisted error
  Given the selected article has status failed
  When detail loading succeeds
  Then the detail pane shows the persisted error message in red and centered

Scenario: Article shell exposes logout control
  Given the user is authenticated on an article route
  When the article shell renders
  Then the title bar shows Archivist
  And the user icon exposes only Logout
```

## Done When

- Article master-detail UI is implemented without destructive controls.
- Article shell controls, responsive layout, route update, no-id, loading, ready, queued, failed, and fetch-error states are implemented.
- Markdown content rendering cannot execute raw article HTML or scripts.
- Tests cover route update, selected state, loading, ready, queued, failed, fetch-error, no-id, Original action, article shell logout control, responsive layout, and Markdown safety.
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

- Capture `/articles` and `/articles/<article_id>` in a browser and compare against `docs/design/view.png` and `docs/design/DESIGN.md`.

## Dependencies

Depends on:

- `UI-002`
- `UIEND-002`

Blocks:

- `UI-004`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.
