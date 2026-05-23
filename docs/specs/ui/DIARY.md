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
