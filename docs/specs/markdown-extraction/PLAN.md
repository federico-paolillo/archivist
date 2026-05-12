---
feature: markdown-extraction
status: done
canonical: true
---

# Feature Plan: Markdown Extraction With Fallbacks

## Purpose

This file is the feature-level implementation control board for Markdown extraction. It defines task order, dependencies, concurrency rules, validation sequence, and execution status.

---

## Task DAG

```text
MDEXT-001 -> MDEXT-002
MDEXT-001 -> MDEXT-003
MDEXT-001 -> MDEXT-004
ARTPROC-003 -> MDEXT-002
ARTPROC-005 -> MDEXT-005
MDEXT-002 -> MDEXT-005
MDEXT-003 -> MDEXT-005
MDEXT-004 -> MDEXT-005
MDEXT-005 -> SUMGEN-002
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards

- `MDEXT-001` creates the feature ALM artifacts and updates canonical architecture, design, artifact, configuration, logging, and error conventions.

### Phase 2: Worker Extraction Foundations

- `MDEXT-002` extends Worker artifact access with atomic `content.md` writes.
- `MDEXT-003` implements local go-readability v2 extraction and Markdown conversion behind `MarkdownExtractor`.
- `MDEXT-004` implements Jina Reader fallback behind `MarkdownExtractor`, with SDK selection captured in an ExecPlan.

### Phase 3: Pipeline Integration And Notifications

- `MDEXT-005` integrates Markdown extraction into Worker terminal processing.
- `MDEXT-006` is skipped once `summary-generation` is planned because `SUMGEN-005` owns the final success notification path.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `MDEXT-001` | Create feature artifacts and contracts | done | - | `MDEXT-002`, `MDEXT-003`, `MDEXT-004` | no | - |
| `MDEXT-002` | Worker Markdown artifact access | done | `MDEXT-001`, `ARTPROC-003` | `MDEXT-005` | yes | - |
| `MDEXT-003` | Worker go-readability extraction | done | `MDEXT-001` | `MDEXT-005` | yes | - |
| `MDEXT-004` | Worker Jina Reader fallback | done | `MDEXT-001` | `MDEXT-005` | yes | `plans/MDEXT-004-worker-jina-reader-fallback.execplan.md` |
| `MDEXT-005` | Worker Markdown pipeline integration | done | `ARTPROC-005`, `MDEXT-002`, `MDEXT-003`, `MDEXT-004` | `SUMGEN-002` | no | `plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md` |
| `MDEXT-006` | Gateway Markdown success notification | skipped | `MDEXT-005`, `TELING-004` | - | no | - |

---

## Concurrency Rules

- `MDEXT-003` and `MDEXT-004` may run in parallel after `MDEXT-001` because they own separate provider adapters.
- `MDEXT-002` must wait for `ARTPROC-003` because it extends the shared artifact access layer created by article processing.
- `MDEXT-005` must wait for HTML snapshot orchestration and all Markdown extraction components.
- `MDEXT-006` is skipped because summary generation supersedes Markdown-complete terminal notifications before implementation.
- Worker pipeline orchestration, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- HTML snapshot output from `article-processing`.
- Deterministic artifact paths in `docs/conventions/ARTIFACTS.md`.
- ARC error-code catalog in `docs/conventions/ERRORS.md`.
- Worker-owned `MarkdownExtractor` contract shared by local and Jina implementations.
- Worker configuration for `JINA_ENABLED` and optional `JINA_API_KEY`.
- Gateway terminal notification content selection for summary-complete success in `summary-generation`.

---

## Validation Sequence

1. Validate canonical docs and task dependencies.
2. Run Worker artifact access tests.
3. Run Worker go-readability extraction tests.
4. Run Worker Jina fallback tests.
5. Run Worker pipeline transaction and logging tests.
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

- all required tasks are `done` or explicitly `skipped`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
