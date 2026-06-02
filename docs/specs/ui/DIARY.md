# Implementation Diary: Final Browser UI

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-05-06 - UI-001: Create canonical UI artifacts

Status:
- completed

Summary:
- Created the canonical ALM structure for the final browser UI feature.

Changes:
- Added the UI feature specification, plan, implementation tasks, and ExecPlans.
- Linked the feature from the global feature index.
- Recorded cross-feature ownership so browser UI rendering belongs to `ui`, while auth endpoints remain in `authn` and article APIs remain in `ui-endpoints`.

Decisions:
- The UI consumes Gateway routes through `VITE_API_BASE_PATH`, default `/api`.
- Invalid login navigates to `/login/failed`, which renders a blank black page.
- Retry is out of scope for v0.

Validation:
- Documentation-only change. Markdown artifact consistency was checked manually.

Follow-ups:
- Execute `UI-002` after `AUTHN-003` is done.

Canonical Updates:
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui/PLAN.md`
- `docs/specs/ui/tasks/*.md`
- `docs/specs/ui/plans/*.execplan.md`
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/UI.md`
- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `docs/specs/ui-endpoints/SPEC.md`

## 2026-05-06 - DOCS-SANITY: UI Rebuild And DTO Correction

Status:
- completed

Summary:
- Corrected UI docs to consume explicit lower-camel article API DTOs and canonical design assets.

Changes:
- Updated `SPEC.md`, `PLAN.md`, `UI-002`, `UI-003`, and their ExecPlans.
- Added `AUTHN-004` as a dependency of `UI-002`.
- Fixed the UI test glob and ensured `VITE_API_BASE_PATH` is part of canonical configuration.

Decisions:
- `docs/design/DESIGN.md`, `docs/design/login.png`, and `docs/design/view.png` are canonical rebuild inputs.
- UI consumes `originalUrl`, `canonicalUrl`, `errorMessage`, and `createdAt` over the JSON wire.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no snake_case UI/UI endpoint wire field names.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement UI routing after `AUTHN-004`; implement article views after UI endpoints are complete.

Canonical Updates:
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui/PLAN.md`
- `docs/specs/ui/tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`
- `docs/specs/ui/tasks/UI-003-article-master-detail-and-delete-workflow.md`
- `docs/specs/ui/plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md`
- `docs/specs/ui/plans/UI-003-article-master-detail-and-delete-workflow.execplan.md`
- `docs/conventions/UI.md`
- `docs/REBUILD.md`

## 2026-05-20 - UI-002: UI Routing, Design System, API Base Config, And Auth Shell

Status:
- completed

Summary:
- Implemented the Preact auth shell for the final browser UI.

Changes:
- Added `deps.ts` with `makeDeps()`, API base normalization, and auth API methods for `GET /auth/session`, `POST /login`, and `POST /logout`.
- Wired `main.tsx` through the composition root.
- Implemented `preact-iso` routes for `/login`, `/login/failed`, `/articles`, and `/articles/:articleId`.
- Added the protected article-route session gate, login success navigation, invalid-login blank black route, and top article shell bar with a user icon menu containing only `Logout`.
- Added brutalist CSS primitives: black surfaces, stark white text, heavy borders, 0 radius, and no shadows, gradients, blur, or external font loading.
- Added frontend tests for API base normalization, credentials usage, login success, invalid-login blank route, session `401`, and logout.

Decisions:
- Explicit `VITE_API_BASE_PATH=/` is normalized to the unprefixed Gateway route base; unset or blank values still default to `/api`.
- Article list/detail/delete behavior remains deferred to `UI-003`; UI-002 renders only authenticated shell placeholders.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 8 tests.
- Browser smoke validation passed for `/login`, `/login/failed`, and unauthenticated `/articles`.

Follow-ups:
- Implement article master-detail and delete workflow in `UI-003` after its dependencies are satisfied.

Canonical Updates:
- `docs/specs/ui/tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`
- `docs/specs/ui/plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md`
- `docs/specs/ui/PLAN.md`

## 2026-05-20 - UI-002: Review Follow-Up

Status:
- completed

Summary:
- Addressed UI-002 review findings for API base normalization and frontend test conventions.

Changes:
- Collapsed protocol-relative and multiple-leading-slash API base inputs into same-origin paths, so `//api` and `///api///` normalize to `/api`.
- Added tests for double-leading and multiple-leading slash API base inputs.
- Added Testing Library and `@testing-library/user-event` dev dependencies.
- Reworked route tests to use Testing Library queries and user-event interactions instead of raw DOM event dispatch.

Decisions:
- Multiple leading slashes are normalized to a single same-origin leading slash rather than rejected, preserving tolerant configuration handling while preventing protocol-relative fetch URLs.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 8 tests.

Follow-ups:
- None.

Canonical Updates:
- None.

## 2026-05-20 - UI-002: Page Structure Refactor

Status:
- completed

Summary:
- Refactored the UI auth shell into the canonical page/component directory structure without changing route behavior or API contracts.

Changes:
- Kept `src/ui/src/app.tsx` as the route composition root.
- Moved login, login-failed, and articles page behavior under `src/ui/src/pages/<pagename>/<pagename>.tsx`.
- Added page-local components under `pages/login/components/` and `pages/articles/components/`.
- Added the shared protected-route session gate under `src/ui/src/components/`.
- Updated future UI-003 guidance to extend the existing `pages/articles` page and page-local components.

Decisions:
- `login-failed` is the page directory name for the `/login/failed` route.
- No generic button/input components were introduced because the current implementation only reuses CSS primitives.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed.

Follow-ups:
- Future UI pages must follow `src/ui/src/pages/<pagename>/<pagename>.tsx` and keep page-specific components under that page's `components/` directory.

Canonical Updates:
- `docs/conventions/UI.md`
- `docs/specs/ui/tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`
- `docs/specs/ui/tasks/UI-003-article-master-detail-and-delete-workflow.md`
- `docs/specs/ui/plans/UI-003-article-master-detail-and-delete-workflow.execplan.md`

## 2026-05-28 - UI-003: Article Master-Detail And Delete Workflow

Status:
- completed

Summary:
- Implemented the authenticated article master-detail route surface and destructive delete workflow.

Changes:
- Added article list/detail/delete methods to the UI API client using the configured API base path and credentialed same-origin fetch.
- Replaced the article shell placeholder with a route-driven master list, blank no-id detail state, loading spinner, ready detail view, queued/non-terminal `Come back later.` state, failed persisted-error state, detail fetch failure state, and delete failure state.
- Added `Original` action behavior using `canonicalUrl` before `originalUrl`.
- Added the delete confirmation modal with `Are you sure?`, `Yes`, and `Nevermind`.
- Added `markdown-it` and `@types/markdown-it`; configured rendering with raw HTML disabled, linkification disabled, default unsafe-link validation, and `noopener noreferrer` links.
- Expanded frontend tests for article route selection, no-id detail, loading, ready, queued, failed, fetch-error, article API `401`, Original link behavior, delete cancel, delete success, and delete failure.

Decisions:
- The Markdown renderer/sanitizer policy is now canonical in `docs/conventions/UI.md`.
- `UI-004` is ready after `UI-003` completion; the overall UI feature remains `in_progress`.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 19 tests.
- Manual browser validation used the built UI with a temporary same-origin mock API. `/articles` rendered the master list with blank black detail, and `/articles/01HREADY000000000000000000` rendered the ready detail with Delete/Original, no Retry, and no unsafe Markdown link or raw script/image nodes.

Follow-ups:
- `UI-004` should run the final full UI validation pass against integrated Gateway data or an agreed deployment-like test fixture.

Canonical Updates:
- `docs/conventions/UI.md`
- `docs/specs/ui/PLAN.md`
- `docs/specs/ui/tasks/UI-003-article-master-detail-and-delete-workflow.md`
- `docs/specs/ui/plans/UI-003-article-master-detail-and-delete-workflow.execplan.md`

## 2026-05-29 - UI-004: Final UI Validation Pass

Status:
- completed

Summary:
- Completed the final UI validation pass for the v0 browser UI using automated frontend checks and Gateway-seeded browser validation.

Changes:
- Added frontend coverage for password non-persistence across URL, cookies, local storage, session storage, and IndexedDB use.
- Added the canonical login placeholder text visible in `docs/design/login.png`.
- Adjusted the UI CSS to better match the canonical login and article-view screenshots, keep the article shell constrained to the viewport, keep the footer visible, and prevent action/title overlap at common desktop widths.
- Marked `UI-004` and the `ui` feature complete.

Decisions:
- Gateway-seeded browser validation used a local HTTPS same-origin proxy that stripped `/api/*` before forwarding to Gateway and sent forwarded `https` host context. Auth semantics, Secure cookies, and same-origin checks were not disabled.
- Browser automation established the authenticated session through the real Gateway `/login` endpoint, then captured authenticated UI routes against seeded SQLite and artifact data.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 21 tests.
- Gateway-seeded browser validation passed for `/login`, `/login/failed`, `/articles`, and `/articles/01HY0000000000000000000003`.
- The ready detail rendered title, summary Markdown, content Markdown, `Delete`, and `Original`; `Retry` was absent.
- The 1366x768 layout check found no action/title overlap, visible footer, no document scrolling, and internal pane scrolling.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui/tasks/UI-004-final-ui-validation-pass.md`
- `docs/specs/ui/PLAN.md`
- `docs/specs/INDEX.md`

## 2026-05-30 - UI Status Coherence Correction

Status:
- completed

Summary:
- Corrected the canonical UI feature status in `SPEC.md` from `draft` to `done` so it matches the completed UI plan and global feature index.

Changes:
- Updated only `docs/specs/ui/SPEC.md` frontmatter status.
- Verified `docs/specs/ui/PLAN.md` and `docs/specs/INDEX.md` already record the UI feature as `done`.
- Made no behavior, source code, task, or review document changes.

Decisions:
- Treated this as a documentation coherence fix, not a feature behavior change.

Validation:
- Non-mutating `rg`/`nl` checks verified UI `SPEC.md`, UI `PLAN.md`, and `docs/specs/INDEX.md` agree on `done` status.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui/SPEC.md`

## 2026-06-02 - UI Layout Refinement

Status:
- completed

Summary:
- Refined the article shell layout so desktop and tablet routes keep solid black header/footer chrome while the master/detail workspace owns scrolling.
- Updated the mobile stacked layout for a 430x960 CSS-pixel viewport with a capped, internally scrollable master list and unbounded detail content in normal document flow.
- Preserved the visible password textarea behavior for the pasted 2048-character bearer secret.

Decisions:
- Desktop/tablet article routes keep the document fixed to the viewport and make `.article-workspace` the single scroll container.
- Mobile article routes use document scrolling. The master list is capped to a reasonable viewport-relative height and scrolls internally; detail content grows without an internal scroll container.
- The login password control may be a visible textarea. The password value remains transient and must not be persisted or logged.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 28 tests.
- Browser validation used the built UI with a temporary same-origin mock API.
- Desktop 1280x720 `/articles/01HREADY000000000000000000`: document scroll height matched the viewport, header/footer backgrounds were solid black, `.article-workspace` was the only scroll container, and master/detail panes used visible overflow.
- Mobile 430x960 `/articles/01HREADY000000000000000000`: document scrolling was enabled, horizontal scroll was absent, the master list max height resolved to 384px and scrolled internally, and detail content was unbounded in normal page flow.
- `/login`: password control rendered as `TEXTAREA`.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui/tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`
- `docs/specs/ui/tasks/UI-003-article-master-detail-and-delete-workflow.md`
- `docs/specs/ui/plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md`

## 2026-06-02 - Desktop Scroll Correction And Footer Ellipsis

Status:
- completed

Summary:
- Corrected the desktop/tablet article layout so the master and detail panes scroll independently inside the viewport-framed shell.
- Preserved the mobile stacked layout with a capped, internally scrollable master list and unbounded detail content.
- Added CSS-only single-line ellipsis behavior for long article footer version labels such as full commit hashes.

Decisions:
- Desktop/tablet article routes keep document scrolling disabled and use independent scroll containers for `.article-master` and `.article-detail`; `.article-workspace` clips the panes instead of scrolling itself.
- Mobile article routes keep document scrolling enabled and keep the detail pane unbounded in page flow.
- Version labels remain full values in markup and are visually truncated only by CSS when the viewport cannot fit them.

Validation:
- `cd src/ui && npm run format`: passed.
- `cd src/ui && npm run lint`: passed.
- `cd src/ui && VITE_VERSION_LABEL=667b7a95fb7aa204d5b812fc8f9b9ea4d0e5d094 npm run build`: passed.
- `cd src/ui && npm run test`: passed, 2 test files and 28 tests.
- Browser validation used the built UI with a temporary same-origin mock API and the long commit-style version label.
- Desktop 1280x720 `/articles/01HREADY000000000000000000`: document scroll height matched the viewport, `.article-workspace` had `overflow: hidden`, `.article-master` and `.article-detail` had `overflow: auto` and scrollable content, and header/footer backgrounds were solid black.
- Mobile 430x960 `/articles/01HREADY000000000000000000`: document scrolling was enabled, horizontal scroll was absent, the master list max height resolved to 384px and scrolled internally, detail content was unbounded in normal page flow, and the footer used `overflow: hidden`, `white-space: nowrap`, and `text-overflow: ellipsis`.
- `/login`: password control still rendered as `TEXTAREA`.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui/tasks/UI-003-article-master-detail-and-delete-workflow.md`
