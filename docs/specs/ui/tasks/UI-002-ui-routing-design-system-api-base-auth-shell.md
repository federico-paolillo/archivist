---
id: UI-002
feature: ui
title: UI routing, design system, API base config, and auth shell
status: blocked
depends_on: [UI-001, AUTHN-003]
blocks: [UI-003]
parallel: false
exec_plan: ../plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md
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
- API client dependency construction through `deps.ts` and `makeDeps()`.
- `VITE_API_BASE_PATH` defaulting and normalization.
- `GET /auth/session` startup/protected-route checks.
- Login form with large masked password control and `IDENTIFY` submit.
- Login success navigation to `/articles`.
- Invalid login navigation to `/login/failed`.
- Blank black `/login/failed` page.
- Top article shell title bar and user icon menu containing only `Logout`.
- Logout call and navigation behavior.
- Frontend tests for auth shell behavior.

## Out of Scope

This task does not include:

- Article list rendering beyond a shell placeholder.
- Article detail loading.
- Delete workflow.
- Gateway endpoint implementation.
- Retry or requeue behavior.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `POST /login`, `POST /logout`, and `GET /auth/session` from `authn`.
- `VITE_API_BASE_PATH`, default `/api`.
- `docs/design/DESIGN.md`
- `docs/design/login.png`
- `docs/design/view.png`

## Outputs

Expected outputs, files, behavior, or interfaces:

- UI routes exist and render the correct auth/shell states.
- API client uses the configured API base and credentials.
- Password is not persisted outside transient component state.
- Invalid login produces a blank black page at `/login/failed`.

## Expected Affected Areas

```text
src/ui/src/
src/ui/src/app.css
src/ui/src/app.tsx
src/ui/src/main.tsx
src/ui/package.json
src/ui/vite.config.ts
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/UI.md`
- `docs/specs/authn/SPEC.md`
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

Scenario: Logout returns to login
  Given the user is authenticated on /articles
  When the user opens the user menu and clicks Logout
  Then the UI posts to the configured logout endpoint
  And navigates to /login
```

## Done When

- Routes and auth shell are implemented.
- Design primitives match the canonical visual constraints.
- Tests cover API base usage, login success, invalid-login black page, session `401`, and logout.
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

- `UI-001`
- `AUTHN-003`

Blocks:

- `UI-003`

## ExecPlan

ExecPlan:

```text
../plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md
```

## Open Questions

- None.
