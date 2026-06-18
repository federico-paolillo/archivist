---
feature: ui-endpoints
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
ARTPROC-005 -> UIEND-003
UIEND-002 -> UIEND-003
```

---

## Execution Phases

### Phase 1: Gateway Read APIs

- `UIEND-002` implements authenticated list/detail APIs, ownership-scoped reads, read-only detail artifact access, ULID normalization, and `canForceDelete` detail metadata through the shared Gateway force-delete eligibility predicate.

### Phase 2: Gateway Delete APIs

- `UIEND-003` implements authenticated normal hard delete, authenticated stale force delete using the shared Gateway force-delete eligibility predicate, artifact directory cleanup, SQLite write-transaction serialization with Worker job claim, ownership checks, same-origin checks, and the documented SQLite/filesystem consistency limitation.

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `UIEND-002` | Gateway article read API | `AUTHN-004`, `TELING-001`, `SUMGEN-005` | `UIEND-003`, `UI-003` | no | no |
| `UIEND-003` | Gateway article delete API | `AUTHN-004`, `TELING-001`, `ARTPROC-005`, `UIEND-002` | `UI-004` | no | no |

---

## Concurrency Rules

- `UIEND-002` and `UIEND-003` must not run concurrently because they share Gateway article route registration, DTOs, repository interfaces, artifact abstractions, ownership scoping, and tests.
- `UIEND-002` must wait for `AUTHN-004` so article reads consume the final post-forwarding authenticated user contract.
- `UIEND-003` must wait for `AUTHN-004` so delete routes consume final same-origin and authenticated user semantics.
- `UIEND-003` must wait for `ARTPROC-005` because delete/claim serialization is a cross-executable SQLite write-transaction contract with Worker queued-job claim behavior.
- `UIEND-003` must wait for `UIEND-002` because force-delete availability extends the article detail response contract.
- Delete behavior must not weaken the read-only artifact abstraction used by detail and notification dispatch.

---

## Blocking Interfaces or Schemas

- `app-cookie` authentication and `ClaimTypes.NameIdentifier`.
- SQLite `articles`, `jobs`, and `notifications`.
- Deterministic article artifact paths under `DATA_DIR`.
- `GET /articles`, `GET /articles/{id}`, `DELETE /articles/{id}`, and `DELETE /articles/{id}/force` wire contracts.
- Article detail JSON field `canForceDelete`.
- Shared Gateway force-delete eligibility predicate used by both detail `canForceDelete` and force-delete enforcement.
- ULID route normalization shared by detail, normal delete, and force delete.

---

## Validation Sequence

1. Run Gateway API tests for authentication, ownership scoping, list/detail pagination, ULID normalization, artifact behavior, `canForceDelete`, normal delete, stale force delete, active running-job rejection, same-origin rejection, shared force-delete eligibility predicate behavior, delete/claim serialization, and SQLite/filesystem consistency behavior.
2. Run complete Gateway formatting, build, and test validation.
3. Run Worker tests that cover queued-job claim behavior when validating delete/claim serialization changes.

Validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run test
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
