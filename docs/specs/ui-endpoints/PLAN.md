---
feature: ui-endpoints
status: draft
canonical: true
---

# Feature Plan: UI Article Endpoints

## Purpose

This file is the feature-level implementation control board for the Gateway APIs consumed by the UI article views.

---

## Task DAG

```text
AUTHN-003 -> UIEND-002
TELING-001 -> UIEND-002
SUMGEN-005 -> UIEND-002
UIEND-001 -> UIEND-002

AUTHN-003 -> UIEND-003
TELING-001 -> UIEND-003
UIEND-001 -> UIEND-003
```

---

## Execution Phases

### Phase 1: Canonical Contracts

- `UIEND-001` creates the feature artifacts and records the Gateway artifact cleanup convention for delete.

### Phase 2: Gateway APIs

- `UIEND-002` implements authenticated list/detail APIs and read-only detail artifact access.
- `UIEND-003` implements authenticated hard delete and artifact directory cleanup.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `UIEND-001` | Create canonical artifacts | done | - | `UIEND-002`, `UIEND-003` | no | - |
| `UIEND-002` | Gateway article read API | blocked | `UIEND-001`, `AUTHN-003`, `TELING-001`, `SUMGEN-005` | - | yes | `plans/UIEND-002-gateway-article-read-api.execplan.md` |
| `UIEND-003` | Gateway article delete API | blocked | `UIEND-001`, `AUTHN-003`, `TELING-001` | - | yes | `plans/UIEND-003-gateway-article-delete-api.execplan.md` |

---

## Concurrency Rules

- `UIEND-002` and `UIEND-003` may run in parallel after dependencies if they do not modify the same Gateway route or DI registration files concurrently.
- Article API route registration, shared DTOs, shared repository interfaces, and shared artifact abstractions must be coordinated.
- Delete behavior must not weaken the read-only artifact abstraction used by detail and notification dispatch.

---

## Blocking Interfaces or Schemas

- `app-cookie` authentication and `ClaimTypes.NameIdentifier`.
- SQLite `articles`, `jobs`, and `notifications`.
- Deterministic article artifact paths under `DATA_DIR`.
- `GET /articles`, `GET /articles/{id}`, and `DELETE /articles/{id}` wire contracts.

---

## Validation Sequence

1. Run Gateway API tests for authentication, list/detail pagination, artifact behavior, delete behavior, and same-origin rejection.
2. Run complete Gateway formatting, build, and test validation.

Validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
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
