---
id: UI-005
feature: ui
title: Shared app layout refactor
status: done
depends_on: [UI-004]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# UI-005: Shared app layout refactor

## Objective

Refactor the browser UI so `/login`, `/articles`, and `/articles/<article_id>` use one shared header, main content, and footer layout.

## Story / Context

As the personal Archivist user, I want the visible application frame to remain consistent across login and authenticated article routes on desktop and mobile.

## Scope

This task includes:

- Shared `AppLayout` component under `src/ui/src/components/`.
- Header with `Archivist` brand link.
- Pluggable right-side header content for authenticated route controls.
- Footer rendering `VERSION: {import.meta.env.VITE_VERSION_LABEL}`.
- Login route rendered inside the same visible layout as article routes.
- Article route chrome moved from the article shell into the shared layout.
- Mobile layout where header and footer scroll naturally with the page.
- Mobile login content in normal page flow rather than vertically centered.
- Frontend tests for shared layout behavior.

## Out of Scope

This task does not include:

- Gateway or API contract changes.
- Auth behavior changes.
- Article workflow changes.
- Retry, requeue, search, filtering, or sorting controls.
- Updating binary design screenshot assets.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `UI-004`.
- `../SPEC.md`
- `../PLAN.md`
- `docs/design/DESIGN.md`
- Existing login and article page implementations.

## Outputs

Expected outputs, files, behavior, or interfaces:

- `/login`, `/articles`, and `/articles/<article_id>` render through shared app layout.
- `/login/failed` remains blank black with no text or chrome.
- Login footer version label uses `VITE_VERSION_LABEL`, not a hard-coded CSS string.
- Existing auth, routing, article detail, delete, and Markdown behavior remains unchanged.

## Expected Affected Areas

```text
docs/specs/ui/SPEC.md
docs/specs/ui/PLAN.md
docs/specs/ui/tasks/UI-005-shared-app-layout-refactor.md
docs/specs/ui/DIARY.md
src/ui/src/components/
src/ui/src/pages/login/
src/ui/src/pages/articles/
src/ui/src/app.css
src/ui/src/app.test.tsx
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-ui/SKILL.md`
- `docs/design/DESIGN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Login uses shared app chrome
  Given the browser opens /login
  Then the page renders the shared Archivist header
  And the login form is visible inside the main content region
  And the footer renders the configured version label
  And no authenticated user menu is visible

Scenario: Login failure remains blank
  Given the browser opens /login/failed
  Then the page contains no visible text
  And no shared app header or footer is rendered

Scenario: Article routes preserve authenticated chrome
  Given the user is authenticated
  When the browser opens /articles or /articles/{article_id}
  Then the page renders the shared Archivist header
  And the authenticated user menu is visible
  And the footer renders the configured version label
  And existing article route behavior is unchanged

Scenario: Mobile layout scrolls naturally
  Given a 430x960 CSS-pixel viewport
  When the browser opens /login or an article route
  Then the header and footer participate in normal document flow
  And the login form is not vertically centered
```

## Done When

- Shared layout is implemented and used by login and article routes.
- `/login/failed` and session-check placeholders remain blank.
- Tests cover the shared layout behavior.
- Validation passes or failures are recorded.
- `PLAN.md` and `DIARY.md` are updated.

## Validation

Required checks:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual validation:

- Capture and review `/login`.
- Capture and review `/login/failed`.
- Capture and review `/articles`.
- Capture and review `/articles/<article_id>`.
- Check desktop and 430x960 mobile layout behavior.

Result:

- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 29 tests.
- Browser validation used the built UI with a temporary same-origin mock API and `VITE_VERSION_LABEL=browser-check`.
- Desktop `/login`: rendered shared header, login form, no user menu, and configured footer.
- Desktop `/login/failed`: rendered blank black route with no shared app layout.
- Desktop `/articles`: rendered shared authenticated chrome, blank detail pane, configured footer, and no document scrolling.
- Desktop `/articles/01HREADY000000000000000000`: rendered ready detail with independent master/detail pane scrolling and no document scrolling.
- Mobile 430x960 `/login`: shared header and footer used static positioning in normal page flow, and the login form top measured 120px from the viewport top rather than centered.
- Mobile 430x960 `/articles/01HREADY000000000000000000`: shared header and footer used static positioning in normal page flow, document scrolling was enabled, detail overflow was visible, and the master list max height resolved to 384px.
- Follow-up correction on 2026-06-05 fixed the mobile shared footer row at 40px on short routes such as `/login` while keeping header and footer in normal document flow.

## Dependencies

Depends on:

- `UI-004`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
