---
feature: ui
status: done
canonical: true
---

# Feature Plan: Final Browser UI

## Purpose

This file is the feature-level implementation control board for the final Preact browser UI.

---

## Task DAG

```text
AUTHN-004 -> UI-002
UIEND-002 -> UI-003
UIEND-003 -> UI-003
```

---

## Execution Phases

### Phase 1: Authenticated Shell And Shared Layout

- `UI-002` implements routing, shared application chrome, brutalist design primitives, configured API base handling, session checks, login, login failure, and logout.

### Phase 2: Article Workflows

- `UI-003` implements the article master-detail view, detail states, Markdown rendering, Original action, normal Delete confirmation workflow, and stale Force Delete workflow driven by `canForceDelete`.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `UI-002` | UI routing, design system, API base config, and auth shell | done | `AUTHN-004` | `UI-003` | no | null |
| `UI-003` | Article master-detail view and delete workflow | done | `UI-002`, `UIEND-002`, `UIEND-003` | - | no | null |

---

## Concurrency Rules

- UI implementation tasks are sequenced because they share the router, composition root, API client, global styles, and top-level application shell.
- `UI-002` must wait for `AUTHN-004` because it consumes the final auth endpoint, cookie, effective-origin, and client contract.
- `UI-003` must wait for `UIEND-002` and `UIEND-003` because it consumes the article read, normal delete, force-delete, and `canForceDelete` contracts.

---

## Blocking Interfaces or Schemas

- Browser route ownership for `/login`, `/login/failed`, `/articles`, and `/articles/<article_id>`.
- Shared application layout for `/login`, `/articles`, and `/articles/<article_id>`.
- `VITE_API_BASE_PATH`, default `/api`.
- Reverse proxy mapping from public `/api/*` to Gateway unprefixed routes.
- `POST /login`, `POST /logout`, and `GET /auth/session` from `authn`.
- `GET /articles`, `GET /articles/{id}`, `DELETE /articles/{id}`, `DELETE /articles/{id}/force`, and `canForceDelete` from `ui-endpoints`.
- Design assets under `docs/design/`.

---

## Validation Sequence

1. Run frontend format, lint, build, and tests.
2. Run automated UI tests for auth, API base, routing, shared layout, article states, delete modal behavior, and force-delete modal behavior.
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

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `docs/specs/INDEX.md` reflects the final feature status.
