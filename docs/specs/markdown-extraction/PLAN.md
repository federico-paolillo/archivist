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
ARTPROC-003 -> MDEXT-002
ARTPROC-005 -> MDEXT-005
MDEXT-002 -> MDEXT-005
MDEXT-003 -> MDEXT-005
MDEXT-004 -> MDEXT-005
MDEXT-005 -> SUMGEN-004
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards


### Phase 2: Worker Extraction Foundations

- `MDEXT-002` extends Worker artifact access with atomic `content.md` writes.
- `MDEXT-003` implements local go-readability v2 extraction and Markdown conversion behind `MarkdownExtractor`.
- `MDEXT-004` implements Jina Reader fallback behind `MarkdownExtractor`, with SDK selection captured in the feature contracts.

### Phase 3: Intermediate Pipeline Integration

- `MDEXT-005` integrates Markdown extraction as the intermediate Worker pipeline stage between snapshotting and summary generation.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `MDEXT-002` | Worker Markdown artifact access | done | `ARTPROC-003` | `MDEXT-005` | yes | - |
| `MDEXT-003` | Worker go-readability extraction | done | - | `MDEXT-005` | yes | - |
| `MDEXT-004` | Worker Jina Reader fallback | done | - | `MDEXT-005` | yes | null |
| `MDEXT-005` | Worker Markdown pipeline integration | done | `ARTPROC-005`, `MDEXT-002`, `MDEXT-003`, `MDEXT-004` | `SUMGEN-004` | no | null |

---

## Concurrency Rules

- `MDEXT-002` must wait for `ARTPROC-003` because it extends the shared artifact access layer created by article processing.
- `MDEXT-005` must wait for HTML snapshot orchestration and all Markdown extraction components, and it blocks summary pipeline integration.
- Worker pipeline orchestration, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- HTML snapshot output from `article-processing`.
- Deterministic artifact paths in `docs/ARTIFACTS.md`.
- ARC error-code catalog in `docs/ERRORS.md`.
- Worker-owned `MarkdownExtractor` contract shared by local and Jina implementations.
- Worker configuration for required `JINA_API_KEY`.
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

- all required tasks are `done`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
- `docs/specs/INDEX.md` reflects the final feature status.
