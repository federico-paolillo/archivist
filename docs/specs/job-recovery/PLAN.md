---
feature: job-recovery
status: done
canonical: true
---

# Feature Plan: Job Recovery And Worker Logging

## Purpose

This file controls implementation of stale running job recovery and Worker logging improvements.

---

## Task DAG

```text
JREC-001 -> JREC-002
JREC-001 -> JREC-003
JREC-001 -> JREC-004
JREC-002 -> JREC-003
JREC-002 -> JREC-005
JREC-003 -> JREC-005
JREC-004 -> JREC-005
```

---

## Execution Phases

### Phase 1: Canonical Contracts

- `JREC-001` creates the feature artifacts, index entry, and design decision.

### Phase 2: Module Implementation

- `JREC-002` implements Gateway force delete and article detail force-delete metadata.
- `JREC-003` implements the UI force-delete workflow after the Gateway contract exists.
- `JREC-004` improves Worker process and pipeline logs.

### Phase 3: Integration

- `JREC-005` integrates reviewed module work and runs final validation.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `JREC-001` | Create canonical job recovery artifacts | done | - | `JREC-002`, `JREC-003`, `JREC-004` | no | - |
| `JREC-002` | Gateway force delete API | done | `JREC-001` | `JREC-003`, `JREC-005` | yes | - |
| `JREC-003` | UI force delete workflow | done | `JREC-001`, `JREC-002` | `JREC-005` | yes | - |
| `JREC-004` | Worker logging improvements | done | `JREC-001` | `JREC-005` | yes | - |
| `JREC-005` | Integration validation | done | `JREC-002`, `JREC-003`, `JREC-004` | - | no | - |

---

## Concurrency Rules

- `JREC-002` and `JREC-004` may run concurrently because their write scopes are Gateway and Worker only.
- `JREC-003` may run concurrently once it uses the `canForceDelete` and `forceDeleteArticle` contract defined in `SPEC.md`; if Gateway response shape changes, coordinate before editing UI.
- `JREC-005` must wait for Gateway, UI, and Worker slices and their reviews.
- Do not run two workers against the same module scope.
- Coordinator owns ALM status and diary updates after worker reports.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` tables.
- Existing hard delete artifact cleanup contract.
- New Gateway route `DELETE /articles/{id}/force`.
- Article detail JSON field `canForceDelete`.
- UI API client method `forceDeleteArticle(id)`.
- Worker stdout `slog` structured logs.

---

## Validation Sequence

1. Gateway tests for force delete, normal delete preservation, auth, same-origin, ownership, and artifact cleanup rollback.
2. UI tests for force-delete visibility, confirmation, API call, success, failure, and auth-expiry handling.
3. Worker tests for process-loop, claim, stage, terminal, and terminal-persistence-failure logs.
4. Full module validation.

Validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
git diff --check
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
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
