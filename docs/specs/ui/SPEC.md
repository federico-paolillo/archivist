---
id: UI
slug: ui
title: Final Browser UI
status: done
owner: null
depends_on: [authn, ui-endpoints]
impacts: [ui, gateway, deployment]
canonical: true
---

# Feature: Final Browser UI

## Intent

Provide the final v0 browser interface for password-only login and authenticated article review.

## Motivation

Archivist needs a private, minimal UI for a single user to authenticate, inspect archived articles, open originals, and delete records. The UI must follow the durable brutalist design system and consume Gateway contracts without making existing Gateway routes ambiguous with browser routes.

## Scope

In scope:

- Preact/Vite browser routes `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- Login page matching `docs/design/login.png` and `docs/design/DESIGN.md`.
- Article master-detail page matching `docs/design/view.png` and `docs/design/DESIGN.md`.
- Vite build-time API base path configuration through `VITE_API_BASE_PATH`, defaulting to `/api`.
- Auth session checks, login, logout, article list/detail loading, original-link navigation, and delete confirmation.
- Client-side rendering of `summaryMarkdown` and `contentMarkdown` without executing raw article HTML or scripts.
- Frontend automated tests and browser/screenshot validation against the design assets.

## Out of Scope

Not included:

- Gateway auth endpoint implementation, owned by `authn`.
- Gateway article list/detail/delete endpoint implementation, owned by `ui-endpoints`.
- Changing Gateway route contracts to include an `/api` prefix.
- Retry or requeue actions.
- Search, filtering, sorting controls, tagging controls, structured summary fields, account management, password rotation, password reset, roles, tenants, PWA/offline behavior, or full-text search.
- Reading SQLite or `/data` artifacts directly from the UI.

## Users / Actors

- Personal Archivist user.
- Preact/Vite UI.
- Gateway API behind the configured API base path.
- Reverse proxy that maps public `/api/*` traffic to Gateway unprefixed routes.

## Requirements

- REQ-001: The UI must expose browser routes `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- REQ-002: Visual implementation must draw constraints from `docs/design/DESIGN.md`, `docs/design/login.png`, and `docs/design/view.png`.
- REQ-003: UI surfaces must use the monochrome brutalist style: black backgrounds, stark white text, heavy borders, 0px radius, no shadows, no gradients, and no soft blur.
- REQ-004: The UI API client must read `VITE_API_BASE_PATH` at build time and default to `/api` when it is unset.
- REQ-005: The API base path must be normalized so calls do not produce double slashes, and all UI API calls must use same-origin `fetch` with credentials included.
- REQ-006: Public deployment must reverse-proxy `${VITE_API_BASE_PATH}/*` to the Gateway's unprefixed route contracts.
- REQ-007: The login page must be available at `/login`.
- REQ-008: The login page must show only the product title, a large masked password control, and an `IDENTIFY` submit control, plus design footer chrome when implemented.
- REQ-009: The password control must support pasting the 2048-character secret and must not store the password in local storage, session storage, IndexedDB, cookies, or URL state.
- REQ-010: Login submit must call `POST ${apiBasePath}/login` with JSON `{ "password": string }`.
- REQ-011: Successful login must navigate to `/articles`.
- REQ-012: Failed login must navigate to `/login/failed`.
- REQ-013: `/login/failed` must render only a blank black page with no text, status explanation, retry hint, or interactive controls.
- REQ-014: Article routes must require a valid auth session. A `401` from `GET ${apiBasePath}/auth/session` or a protected API call must navigate to `/login`.
- REQ-015: `/articles` and `/articles/<article_id>` must share one authenticated master-detail application shell.
- REQ-016: The articles shell must include a top title bar with `Archivist` on the left and a user icon on the right.
- REQ-017: Clicking the user icon must open a menu containing only `Logout`.
- REQ-018: Logout must call `POST ${apiBasePath}/logout`; success or `401` must navigate to `/login`.
- REQ-019: The master pane must load article metadata from `GET ${apiBasePath}/articles` and render ledger-like article rows.
- REQ-020: Clicking an article row must update the URL to `/articles/<article_id>` and show a design-compatible loading spinner in the detail pane while detail loading is in progress.
- REQ-021: When no `<article_id>` is present, the detail pane must be blank and black.
- REQ-022: Detail loading must call `GET ${apiBasePath}/articles/{id}`.
- REQ-023: A failed detail load must show the error message in red, centered in the detail pane.
- REQ-024: A ready article detail must show title, summary markdown, content markdown, and the original/canonical link using the screenshot layout as the target.
- REQ-025: For original-link behavior, use `canonicalUrl` when present, otherwise `originalUrl`; open the URL in a new tab or window with `noopener` and `noreferrer`.
- REQ-026: A queued article, or any non-ready/non-failed article state, must show centered white text exactly `Come back later.` in the detail pane.
- REQ-027: A failed article must show the persisted `errorMessage` in red, centered in the detail pane.
- REQ-028: The detail view for a selected article must expose `Delete` whenever article metadata is available, including ready, queued, and failed states.
- REQ-029: The detail view must expose `Original` when an original or canonical URL is available.
- REQ-030: `Retry` must not be implemented or displayed in v0.
- REQ-031: Delete must first show a modal with the exact question `Are you sure?` and the exact options `Yes` and `Nevermind`.
- REQ-032: Choosing `Nevermind` must close the modal without calling the API.
- REQ-033: Choosing `Yes` must call `DELETE ${apiBasePath}/articles/{id}`.
- REQ-034: Successful delete must reset the URL to `/articles`, clear the detail pane, and remove the deleted article from the master list.
- REQ-035: Delete failure must leave the current URL selected and show the failure text in red in the detail pane.
- REQ-036: Markdown rendering must not execute raw article HTML, scripts, inline event handlers, or `javascript:` links.
- REQ-037: Frontend tests must cover API base path usage, login success, invalid-login black page navigation, session `401`, logout, route update on article selection, no-id detail state, loading, ready, queued, failed, fetch-error, and delete confirm/cancel behavior.
- REQ-038: Manual browser validation must capture `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>` and compare them against `docs/design/DESIGN.md` plus the two screenshots.

## Acceptance Criteria

```gherkin
Feature: Final Browser UI

Scenario: Login succeeds
  Given the browser is on /login
  When the user submits the correct password
  Then the UI posts the password to the configured API base /login endpoint
  And navigates to /articles
  And does not persist the password in browser storage

Scenario: Login fails silently
  Given the browser is on /login
  When the login endpoint returns 401
  Then the UI navigates to /login/failed
  And the page is blank and black

Scenario: Session expires
  Given the browser is on /articles
  When the session check or an authenticated API request returns 401
  Then the UI navigates to /login

Scenario: Article route without an id
  Given the browser is authenticated
  When the user opens /articles
  Then the master list is visible
  And the detail pane is blank and black

Scenario: Selecting an article
  Given the browser is authenticated
  And the article list is loaded
  When the user clicks an article row
  Then the URL changes to /articles/{article_id}
  And the detail pane shows a loading spinner until detail loading completes or fails

Scenario: Ready article detail
  Given the selected article has status ready
  And summaryMarkdown and contentMarkdown are available
  When detail loading succeeds
  Then the detail pane shows the article title, summary, content, Delete action, and Original action

Scenario: Queued article detail
  Given the selected article has status queued
  When detail loading succeeds
  Then the detail pane shows centered white text "Come back later."
  And Delete remains available for the selected article

Scenario: Failed article detail
  Given the selected article has status failed
  When detail loading succeeds
  Then the detail pane shows the persisted error message in red and centered
  And Delete remains available for the selected article

Scenario: Detail load fails
  Given the browser is authenticated
  When GET /articles/{article_id} fails
  Then the detail pane shows the error message in red and centered

Scenario: Delete is cancelled
  Given an article is selected
  When the user clicks Delete and chooses Nevermind
  Then no delete request is sent
  And the current URL is unchanged

Scenario: Delete succeeds
  Given an article is selected
  When the user clicks Delete and confirms Yes
  Then the UI sends DELETE to the configured API base article endpoint
  And navigates to /articles
  And clears the detail pane
```

## Data and State

The UI owns only browser state:

- current route and selected `article_id`;
- article list page returned by `GET /articles`;
- current detail loading, success, and error state;
- modal open/closed state for delete confirmation;
- transient password input state on `/login`.

The UI must not persist the password or auth cookie value in JavaScript-accessible storage. The server-issued `__Host-app-auth` cookie is `HttpOnly` and is managed by Gateway responses.

Article list rows consume the `ui-endpoints` metadata contract:

- `id`
- `title`
- `originalUrl`
- `canonicalUrl`
- `status`
- `errorMessage`
- `createdAt`

Article detail consumes the list metadata plus:

- `summaryMarkdown`
- `contentMarkdown`

## Interfaces

Browser routes:

- `/login`
- `/login/failed`
- `/articles`
- `/articles/<article_id>`

Build-time configuration:

- `VITE_API_BASE_PATH`: public same-origin API base path. Defaults to `/api`.

API calls:

- `GET ${apiBasePath}/auth/session`
- `POST ${apiBasePath}/login`
- `POST ${apiBasePath}/logout`
- `GET ${apiBasePath}/articles`
- `GET ${apiBasePath}/articles/{id}`
- `DELETE ${apiBasePath}/articles/{id}`

All API calls must include credentials.

## Dependencies

Depends on:

- `authn` for `POST /login`, `POST /logout`, `GET /auth/session`, cookie authentication, and `401` behavior.
- `ui-endpoints` for article list/detail/delete contracts.
- `docs/design/DESIGN.md`
- `docs/design/login.png`
- `docs/design/view.png`

Implementation agents should use `.agents/skills/archivist-ui/SKILL.md` for UI coding guidance. The skill is not a feature dependency or rebuild source of truth.

Impacts:

- `src/ui/` application shell, router, API client, styling, and tests.
- Deployment/reverse proxy expectations for the public `/api` API base.
- Browser validation workflow for UI rebuilds.

## Rebuild Notes

- Browser routes and Gateway API routes intentionally overlap in path names. The browser UI uses `/articles` as a page route, while API calls use the configured API base path, default `/api`.
- Do not add `/api` prefixes to Gateway route contracts. The reverse proxy owns mapping public `/api/*` requests to Gateway unprefixed routes.
- `docs/design/DESIGN.md` and the screenshots under `docs/design/` are canonical design inputs for this feature.
- Implement the visible browser UI as the first screen for its routes. Do not replace it with a landing page.
- Do not implement Retry until a future feature defines a backend retry/requeue contract.
- Markdown rendering must treat article content as untrusted input.

## Security / Privacy Notes

- The password is a 2048-character bearer secret and must never be logged or persisted by frontend code.
- Authenticated requests rely on the server-issued `HttpOnly` cookie and must use same-origin `fetch` with credentials included.
- Raw article Markdown is untrusted. Render Markdown with raw HTML disabled or sanitized so article content cannot execute script.
- Delete is a destructive operation and must require explicit modal confirmation.

## Observability / Logging Notes

- Frontend logging, if any, must not include passwords, auth cookie values, article content, or `Set-Cookie` headers.
- Automated and manual validation should record route, viewport, and result, not private article content beyond fixture data.

## Open Questions

- None.

## Related Documents

- `./PLAN.md`
- `./DIARY.md`
- `./tasks/UI-001-create-canonical-ui-artifacts.md`
- `./tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`
- `./tasks/UI-003-article-master-detail-and-delete-workflow.md`
- `./tasks/UI-004-final-ui-validation-pass.md`
- `./plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md`
- `./plans/UI-003-article-master-detail-and-delete-workflow.execplan.md`
