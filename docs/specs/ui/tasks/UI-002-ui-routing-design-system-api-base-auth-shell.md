---
id: UI-002
feature: ui
title: UI routing, design system, API base config, and auth shell
status: done
depends_on: [AUTHN-004]
blocks: [UI-003]
parallel: false
exec_plan: null
canonical: true
---

# UI-002: UI routing, design system, API base config, and auth shell

## Objective

Implement the Preact routing shell, brutalist design primitives, configured API base client, login flow, login failure page, session checks, and logout behavior.

## Story / Context

As the personal Archivist user, I want to authenticate with the password-only browser UI and enter the private article surface without the UI disclosing login failure details.

## Scope

This task includes:

- Preact router for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- Design-system CSS primitives based on `docs/design/DESIGN.md`.
- Shared application layout containing header, main content, and footer regions for `/login`, `/articles`, and `/articles/<article_id>`.
- Header with `Archivist` brand link and pluggable right-side header content for authenticated route controls.
- Footer rendering `VERSION: {import.meta.env.VITE_VERSION_LABEL}` with fixed 40px chrome-row height and CSS-only truncation for long values.
- API client dependency construction through `deps.ts` and `makeDeps()`.
- `VITE_API_BASE_PATH` defaulting and normalization.
- `GET /auth/session` startup/protected-route checks.
- Login form with large visible password textarea/control and `IDENTIFY` submit.
- Login success navigation to `/articles`.
- Invalid login navigation to `/login/failed`.
- Blank black `/login/failed` page.
- Login route rendered inside the shared visible layout.
- Top article shell title bar and user icon menu containing only `Logout`.
- Logout call and navigation behavior.
- Mobile layout where header and footer scroll naturally with the page.
- Mobile login content in normal page flow rather than vertically centered.
- Frontend tests for auth shell behavior.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- `POST /login`, `POST /logout`, and `GET /auth/session` from `authn`.
- Validated browser auth client contract from `AUTHN-004`.
- `VITE_API_BASE_PATH`, default `/api`.
- `docs/design/DESIGN.md`
- `docs/design/login.png`
- `docs/design/view.png`

## Outputs

Expected outputs, files, behavior, or interfaces:

- UI routes exist and render the correct auth/shell states.
- `/login`, `/articles`, and `/articles/<article_id>` render through shared app layout.
- `/login/failed` remains blank black with no text or chrome.
- API client uses the configured API base and credentials.
- Password is not persisted outside transient component state.
- Invalid login produces a blank black page at `/login/failed`.

## Expected Affected Areas

```text
src/ui/src/
src/ui/src/app.css
src/ui/src/app.tsx
src/ui/src/main.tsx
src/ui/src/components/
src/ui/src/pages/login/
src/ui/src/pages/login-failed/
src/ui/src/pages/articles/
src/ui/package.json
src/ui/vite.config.ts
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-ui/SKILL.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `docs/design/DESIGN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Login form posts through configured API base
  Given VITE_API_BASE_PATH is /api
  When the user submits the login form
  Then the UI sends POST /api/login with credentials included

Scenario: Invalid login navigates to blank black page
  Given the login endpoint returns 401
  When the user submits the login form
  Then the browser route is /login/failed
  And the rendered page contains no visible text
  And no shared app header or footer is rendered

Scenario: Login uses shared app chrome
  Given the browser opens /login
  Then the page renders the shared Archivist header
  And the login form is visible inside the main content region
  And the footer renders the configured version label
  And no authenticated user menu is visible

Scenario: Logout returns to login
  Given the user is authenticated on /articles
  When the user opens the user menu and clicks Logout
  Then the UI posts to the configured logout endpoint
  And navigates to /login

Scenario: Mobile layout scrolls naturally
  Given a 430x960 CSS-pixel viewport
  When the browser opens /login or an article route
  Then the header and footer participate in normal document flow
  And the login form is not vertically centered
```

## Done When

- Routes and auth shell are implemented.
- Shared layout is implemented and used by login and article routes.
- `/login/failed` and session-check placeholders remain blank.
- Design primitives match the canonical visual constraints.
- Tests cover API base usage, login success, invalid-login black page, session `401`, logout, shared layout behavior, configured footer label, and mobile layout behavior.
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

- Capture `/login` and `/login/failed` in a browser and compare against `docs/design/login.png` and `docs/design/DESIGN.md`.

## Dependencies

Depends on:
- `AUTHN-004`

Blocks:

- `UI-003`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.
