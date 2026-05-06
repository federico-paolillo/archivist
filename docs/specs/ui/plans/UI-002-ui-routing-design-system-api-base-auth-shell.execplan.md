---
id: UI-002-PLAN
task: ../tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md
status: proposed
canonical: true
---

# ExecPlan: UI-002 UI Routing, Design System, API Base Config, and Auth Shell

## Objective

Implement the browser routing shell, frontend dependency composition, API base configuration, design-system primitives, password login, blank login-failure route, session checks, and logout.

## Linked Task

- `../tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md` sections for Web UI, Service Boundaries, Runtime Topology, Security Boundaries, and Configuration.
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `docs/design/login.png`
- `docs/design/view.png`

## Assumptions

- `AUTHN-004` has validated the Gateway auth endpoint contract consumed by the UI.
- `VITE_API_BASE_PATH` defaults to `/api` and is same-origin.
- The reverse proxy strips the public `/api` prefix before forwarding to Gateway.

## Non-Goals

- Do not implement article list/detail behavior beyond authenticated shell placeholders.
- Do not change Gateway routes.
- Do not persist passwords or auth cookie values in frontend-accessible storage.
- Do not implement Retry.

## Implementation Sequence

1. Create a `deps.ts` composition root that exposes an API client configured from `import.meta.env.VITE_API_BASE_PATH ?? "/api"`.
2. Normalize the API base by trimming trailing slashes and preserving `/api` as the default.
3. Ensure every API client call uses same-origin `fetch` with `credentials: "include"`.
4. Add auth client methods for `getSession`, `login`, and `logout` using `/auth/session`, `/login`, and `/logout` under the configured API base.
5. Introduce route handling through `preact-iso` for `/login`, `/login/failed`, `/articles`, and `/articles/:articleId`.
6. Add a protected-route/session gate for article routes that calls `GET /auth/session` and sends `401` results to `/login`.
7. Build global CSS design primitives in `app.css`: black surfaces, white text, 0px radius, 2px/4px borders, no shadows, no gradients, Space Grotesk/Public Sans font stacks with safe fallbacks, and fixed-grid spacing.
8. Implement `/login` with the centered ARCHIVIST title, large masked password control, and `IDENTIFY` submit control matching `docs/design/login.png`.
9. Keep the login password only in transient component state; clear it after submit completion.
10. On login success, navigate to `/articles`.
11. On login failure or non-204 response, navigate to `/login/failed`.
12. Implement `/login/failed` as a full-viewport black page with no visible content or controls.
13. Implement the authenticated article shell top bar with `Archivist` title and user icon.
14. Add a user icon menu containing only `Logout`; logout posts to the configured logout endpoint and navigates to `/login` on success or `401`.
15. Add focused tests for API base normalization, credentials usage, login success, invalid login, blank failure route, session `401`, and logout.
16. Update task status, feature plan, and diary after implementation and validation if this task is completed.

## Validation Plan

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

Manual checks:

- `/login` visually matches the screenshot constraints.
- `/login/failed` is blank and black.
- `/articles` unauthenticated access navigates to `/login`.

## Documentation Updates Required

- Update `../tasks/UI-002-ui-routing-design-system-api-base-auth-shell.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append a `DIARY.md` entry with validation results and any durable decisions promoted to canonical docs.
- Promote any new frontend dependency or design decision to `docs/conventions/UI.md` if it must survive rebuild.

## Risks

- API-base normalization mistakes can route UI calls to browser pages instead of Gateway APIs.
- Login failure must not show helpful diagnostics; any visible failure text on `/login/failed` violates the spec.
- Font loading choices must not introduce network dependencies unless explicitly documented.

## Rollback / Recovery Notes

- If routing changes break the app shell, restore the last working shell and keep API client changes isolated behind `deps.ts`.
- If design implementation diverges from screenshots, adjust CSS primitives before changing component structure.

## Completion Criteria

- The auth shell satisfies linked task acceptance criteria.
- Validation commands pass or failures are recorded.
- Manual login route checks are recorded in `DIARY.md`.
