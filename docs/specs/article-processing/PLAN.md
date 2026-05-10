---
feature: article-processing
status: draft
canonical: true
---

# Feature Plan: URL-To-Article Processing Pipeline

## Purpose

This file is the feature-level implementation control board for article processing. It defines task order, dependencies, concurrency rules, validation sequence, and execution status.

---

## Task DAG

```text
ARTPROC-001 -> ARTPROC-002
ARTPROC-001 -> ARTPROC-003
ARTPROC-002 -> ARTPROC-004
ARTPROC-003 -> ARTPROC-004
ARTPROC-004 -> ARTPROC-005
ARTPROC-005 -> ARTPROC-006
ARTPROC-005 -> MDEXT-005
```

Cross-feature dependency:

```text
TELING-001 -> ARTPROC-005
TELING-003 -> ARTPROC-005
TELING-004 -> ARTPROC-006
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards

- `ARTPROC-001` creates the feature ALM artifacts.
- `ARTPROC-002` defines the shared ARC error-code catalog.

### Phase 2: Worker Foundations

- `ARTPROC-003` builds the reusable Worker filesystem artifact access layer.
- `ARTPROC-004` implements URL resolution, HTML fetching, limits, and ARC failure mapping.

### Phase 3: Processing And Notifications

- `ARTPROC-005` implements Worker orchestration and transactional terminal state changes.
- `ARTPROC-006` remains skipped because downstream features supersede snapshot-stage success; final v0 success notification is owned by `SUMGEN-005`.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `ARTPROC-001` | Create feature spec and plan artifacts | done | - | `ARTPROC-002`, `ARTPROC-003` | no | - |
| `ARTPROC-002` | Define shared ARC error-code convention | done | `ARTPROC-001` | `ARTPROC-004` | yes | - |
| `ARTPROC-003` | Worker filesystem artifact access layer | done | `ARTPROC-001` | `ARTPROC-004` | yes | - |
| `ARTPROC-004` | Worker URL resolver and HTML fetcher | done | `ARTPROC-002`, `ARTPROC-003` | `ARTPROC-005` | no | - |
| `ARTPROC-005` | Worker snapshot pipeline orchestration | done | `ARTPROC-004`, `TELING-001`, `TELING-003` | `ARTPROC-006`, `MDEXT-005` | no | `plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md` |
| `ARTPROC-006` | Gateway snapshot success notification bridge | skipped | `ARTPROC-005`, `TELING-004` | - | no | - |

---

## Concurrency Rules

- `ARTPROC-003` may run after `ARTPROC-001` because it only owns Worker artifact access code.
- `ARTPROC-004` must wait for `ARTPROC-002` and `ARTPROC-003` because fetch failures must map to canonical ARC codes and snapshot size handling depends on artifact boundaries.
- `ARTPROC-005` must run after Worker fetch and after Telegram ingestion persistence/outbox contracts are implemented.
- `ARTPROC-006` is skipped when downstream pipeline stages are planned before snapshot-only notification work, because `SUMGEN-005` owns the final success notification contract.
- Worker repository, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- Deterministic article artifact path convention under `DATA_DIR`.
- ARC error-code catalog in `docs/conventions/ERRORS.md`.
- Worker HTTP fetch policy using `github.com/imroc/req/v3`.
- Snapshot-only Gateway success notification is superseded by summary-complete notification in `summary-generation`.

---

## Validation Sequence

1. Validate canonical docs and task dependencies.
2. Run Worker filesystem artifact tests.
3. Run Worker fetcher tests.
4. Run Worker pipeline transaction tests.
5. Run complete Worker and Gateway verification.

Validation commands:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
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
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
