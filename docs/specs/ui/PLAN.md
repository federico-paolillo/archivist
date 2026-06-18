---
feature: ui
canonical: true
---
# Feature Plan: Final Browser UI

## Purpose

This file is the feature-level implementation control board for the final Preact browser UI.

---

## Task DAG

```text
AUTHN-004 -> UI-002
UI-002 -> UI-003
UIEND-002 -> UI-003
UI-003 -> UI-004
UIEND-003 -> UI-004
```

---

## Execution Phases

### Phase 1: Authenticated Shell And Shared Layout

- `UI-002` implements routing, shared application chrome, configured API base handling, session checks, route guards, login, login failure, and logout behavior.

### Phase 2: Read-Only Article Surface

- `UI-003` implements the article master-detail view, article shell controls, detail states, Markdown-safe rendering, and Original action.

### Phase 3: Destructive Article Workflows

- `UI-004` implements normal Delete confirmation workflow and stale Force Delete workflow driven by Gateway `canForceDelete`.

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `UI-002` | UI routing, shared layout, API base config, and auth shell | `AUTHN-004` | `UI-003` | no | no |
| `UI-003` | Article master-detail read-only view | `UI-002`, `UIEND-002` | `UI-004` | no | no |
| `UI-004` | Article destructive actions | `UI-003`, `UIEND-003` | - | no | no |

---

## Concurrency Rules

- UI implementation tasks are sequenced because they share the router, composition root, API client, global styles, and top-level application shell.
- `UI-002` must wait for `AUTHN-004` because it consumes the final auth endpoint, cookie, effective-origin, and client contract.
- `UI-003` must wait for `UI-002` and `UIEND-002` because it consumes the auth shell, API client, article read contract, and `canForceDelete` detail metadata without invoking destructive routes.
- `UI-004` must wait for `UI-003` and `UIEND-003` because it adds normal delete and force-delete workflows to the read-only article surface.

---

## Blocking Interfaces or Schemas

- Browser route ownership for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- Shared application layout for `/login`, `/articles`, and `/articles/<article_id>`.
- `VITE_API_BASE_PATH`, default `/api`.
- Reverse proxy mapping from public `/api/*` to Gateway unprefixed routes.
- `POST /login`, `POST /logout`, and `GET /auth/session` from `authn`.
- `GET /articles`, `GET /articles/{id}`, and `canForceDelete` from `ui-endpoints`.
- `DELETE /articles/{id}` and `DELETE /articles/{id}/force` from `ui-endpoints`.
- Design assets under `docs/design/`.

---

## Validation Sequence

1. Run frontend format, lint, build, and tests.
2. Run automated UI tests for auth, API base, routing, shared layout, read-only article states, Markdown safety, delete modal behavior, and force-delete modal behavior.
3. Run browser validation for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
4. Compare browser captures against `docs/design/DESIGN.md`, `docs/design/login.png`, and `docs/design/view.png`, with the login screenshot superseded for the intentionally added shared header chrome.

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

- all task acceptance criteria are satisfied;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
