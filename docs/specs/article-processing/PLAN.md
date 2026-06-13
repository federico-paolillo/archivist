---
feature: article-processing
status: done
canonical: true
---

# Feature Plan: URL-To-Article Processing Pipeline

## Purpose

This file is the feature-level implementation control board for article processing. It defines task order, dependencies, concurrency rules, validation sequence, and execution status.

---

## Task DAG

```text
ARTPROC-002 -> ARTPROC-004
ARTPROC-003 -> ARTPROC-004
ARTPROC-004 -> ARTPROC-005
ARTPROC-005 -> MDEXT-005
```

Cross-feature dependency:

```text
TELING-001 -> ARTPROC-005
TELING-003 -> ARTPROC-005
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards

- `ARTPROC-002` defines the shared ARC error-code catalog.

### Phase 2: Worker Foundations

- `ARTPROC-003` builds the reusable Worker filesystem artifact access layer.
- `ARTPROC-004` implements URL resolution, HTML fetching, Worker SSRF policy, direct guarded dialing, conservative limits, and ARC failure mapping.

### Phase 3: Processing Pipeline

- `ARTPROC-005` exposes `archivist-worker process`, claims queued jobs, writes `snapshot.html`, hands successful jobs to Markdown extraction, and leaves terminal success to summary-complete processing.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `ARTPROC-002` | Define shared ARC error-code convention | done | - | `ARTPROC-004` | yes | - |
| `ARTPROC-003` | Worker filesystem artifact access layer | done | - | `ARTPROC-004` | yes | - |
| `ARTPROC-004` | Worker URL resolver and HTML fetcher | done | `ARTPROC-002`, `ARTPROC-003` | `ARTPROC-005` | no | - |
| `ARTPROC-005` | Worker executable processing pipeline orchestration | done | `ARTPROC-004`, `TELING-001`, `TELING-003` | `MDEXT-005` | no | null |

---

## Concurrency Rules

- `ARTPROC-004` must wait for `ARTPROC-002` and `ARTPROC-003` because fetch failures must map to canonical ARC codes and guarded fetch size handling depends on artifact boundaries.
- `ARTPROC-005` must run after Worker fetch and after Telegram ingestion persistence/outbox contracts are implemented.
- Worker repository, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- Deterministic article artifact path convention under `DATA_DIR`.
- ARC error-code catalog in `docs/ERRORS.md`.
- Worker HTTP fetch and SSRF policy using `github.com/imroc/req/v3` and `src/worker/internal/ssrf`.
- Worker executable command surface: `archivist-worker process`.
- Snapshot success hands off to Markdown extraction; summary-complete notification is owned by `summary-generation`.

---

## Validation Sequence

1. Validate canonical docs and task dependencies.
2. Run Worker filesystem artifact tests.
3. Run Worker fetcher tests.
4. Run Worker pipeline transaction tests.
5. Run Worker executable processing command tests.
6. Run complete Worker and Gateway verification.

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

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
- `docs/specs/INDEX.md` reflects the final feature status.
