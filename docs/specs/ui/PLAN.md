---
feature: ui
status: draft
canonical: true
---

# Feature Plan: Final Browser UI

## Purpose

This file is the feature-level implementation control board for the final Preact browser UI.

---

## Task DAG

```text
UI-001 -> UI-002 -> UI-003 -> UI-004
AUTHN-004 -> UI-002
UIEND-002 -> UI-003
UIEND-003 -> UI-003
```

---

## Execution Phases

### Phase 1: Canonical Contracts

- `UI-001` creates the UI feature artifacts and records cross-feature ownership, route, design, and API-base decisions.

### Phase 2: Authenticated Shell

- `UI-002` implements routing, design-system primitives, configured API base handling, session checks, login, login failure, and logout.

### Phase 3: Article Workflows

- `UI-003` implements the article master-detail view, detail states, Markdown rendering, Original action, and Delete confirmation workflow.

### Phase 4: Validation

- `UI-004` runs final automated and browser validation against the feature spec and design assets.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `UI-001` | Create canonical UI artifacts | done | - | `UI-002` | no | - |
| `UI-002` | UI routing, design system, API base config, and auth shell | blocked | `UI-001`, `AUTHN-004` | `UI-003` | no | `plans/UI-002-ui-routing-design-system-api-base-auth-shell.execplan.md` |
| `UI-003` | Article master-detail view and delete workflow | blocked | `UI-002`, `UIEND-002`, `UIEND-003` | `UI-004` | no | `plans/UI-003-article-master-detail-and-delete-workflow.execplan.md` |
| `UI-004` | Final UI validation pass | blocked | `UI-003` | - | no | - |

---

## Concurrency Rules

- UI implementation tasks are sequenced because they share the router, composition root, API client, global styles, and top-level application shell.
- `UI-002` must wait for `AUTHN-004` because it consumes the validated auth endpoint and client contract.
- `UI-003` must wait for `UIEND-002` and `UIEND-003` because it consumes the article read/delete contracts.
- `UI-004` runs after the UI is feature-complete.

---

## Blocking Interfaces or Schemas

- Browser route ownership for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- `VITE_API_BASE_PATH`, default `/api`.
- Reverse proxy mapping from public `/api/*` to Gateway unprefixed routes.
- `POST /login`, `POST /logout`, and `GET /auth/session` from `authn`.
- `GET /articles`, `GET /articles/{id}`, and `DELETE /articles/{id}` from `ui-endpoints`.
- Design assets under `docs/design/`.

---

## Validation Sequence

1. Run frontend format, lint, build, and tests.
2. Run automated UI tests for auth, API base, routing, article states, and delete modal behavior.
3. Run browser validation for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
4. Compare browser captures against `docs/design/DESIGN.md`, `docs/design/login.png`, and `docs/design/view.png`.

Validation commands:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

---

## Open Planning Questions

- None.

---

## Completion Criteria

The feature is complete when:

- all required tasks are `done` or explicitly `skipped`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
