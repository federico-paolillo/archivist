---
feature: ui-endpoints
status: done
canonical: true
---

# Feature Plan: UI Article Endpoints

## Purpose

This file is the feature-level implementation control board for the Gateway APIs consumed by the UI article views.

---

## Task DAG

```text
AUTHN-004 -> UIEND-002
TELING-001 -> UIEND-002
SUMGEN-005 -> UIEND-002

AUTHN-004 -> UIEND-003
TELING-001 -> UIEND-003
UIEND-002 -> UIEND-003
```

---

## Execution Phases

### Phase 1: Gateway Read APIs

- `UIEND-002` implements authenticated list/detail APIs, ownership-scoped reads, read-only detail artifact access, ULID normalization, and `canForceDelete` detail metadata.

### Phase 2: Gateway Delete APIs

- `UIEND-003` implements authenticated normal hard delete, authenticated stale force delete, artifact directory cleanup, SQLite write-transaction serialization, ownership checks, same-origin checks, and the documented SQLite/filesystem consistency limitation.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `UIEND-002` | Gateway article read API | done | `AUTHN-004`, `TELING-001`, `SUMGEN-005` | `UIEND-003`, `UI-003` | no | null |
| `UIEND-003` | Gateway article delete API | done | `AUTHN-004`, `TELING-001`, `UIEND-002` | `UI-003` | no | null |

---

## Concurrency Rules

- `UIEND-002` and `UIEND-003` must not run concurrently because they share Gateway article route registration, DTOs, repository interfaces, artifact abstractions, ownership scoping, and tests.
- `UIEND-002` must wait for `AUTHN-004` so article reads consume the final post-forwarding authenticated user contract.
- `UIEND-003` must wait for `AUTHN-004` so delete routes consume final same-origin and authenticated user semantics.
- `UIEND-003` must wait for `UIEND-002` because force-delete availability extends the article detail response contract.
- Delete behavior must not weaken the read-only artifact abstraction used by detail and notification dispatch.

---

## Blocking Interfaces or Schemas

- `app-cookie` authentication and `ClaimTypes.NameIdentifier`.
- SQLite `articles`, `jobs`, and `notifications`.
- Deterministic article artifact paths under `DATA_DIR`.
- `GET /articles`, `GET /articles/{id}`, `DELETE /articles/{id}`, and `DELETE /articles/{id}/force` wire contracts.
- Article detail JSON field `canForceDelete`.
- ULID route normalization shared by detail, normal delete, and force delete.

---

## Validation Sequence

1. Run Gateway API tests for authentication, ownership scoping, list/detail pagination, ULID normalization, artifact behavior, `canForceDelete`, normal delete, stale force delete, active running-job rejection, same-origin rejection, and SQLite/filesystem consistency behavior.
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

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes or failures are recorded;
- durable implementation decisions have been promoted to canonical documents;
- `docs/specs/INDEX.md` reflects the final feature status.
