---
feature: markdown-extraction
canonical: true
---
# Feature Plan: Markdown Extraction With Fallbacks

## Purpose

This file is the feature-level implementation control board for Markdown extraction. It defines task order, dependencies, concurrency rules, validation sequence, and validation requirements.

---

## Task DAG

```text
ARTPROC-003 -> MDEXT-002
MDEXT-001 -> MDEXT-003
MDEXT-001 -> MDEXT-004
ARTPROC-005 -> MDEXT-005
MDEXT-002 -> MDEXT-005
MDEXT-003 -> MDEXT-005
MDEXT-004 -> MDEXT-005
MDEXT-002 -> SUMGEN-002
MDEXT-005 -> SUMGEN-004
```

---

## Execution Phases

### Phase 1: Extractor Contract

- `MDEXT-001` defines the Worker-owned `MarkdownExtractor` contract and result taxonomy shared by local extraction, Jina fallback, and pipeline orchestration.

### Phase 2: Worker Extraction Foundations

- `MDEXT-002` extends Worker artifact access with atomic `content.md` writes.
- `MDEXT-003` implements local go-readability v2 extraction and Markdown conversion behind the contract from `MDEXT-001`.
- `MDEXT-004` implements Jina Reader fallback behind the contract from `MDEXT-001`, with SDK selection captured in the feature contracts.

### Phase 3: Intermediate Pipeline Integration

- `MDEXT-005` integrates Markdown extraction as the intermediate Worker pipeline stage between snapshotting and the summary-continuation boundary.

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `MDEXT-001` | Worker Markdown extractor contract | - | `MDEXT-003`, `MDEXT-004` | yes | no |
| `MDEXT-002` | Worker Markdown artifact access | `ARTPROC-003` | `MDEXT-005`, `SUMGEN-002` | yes | no |
| `MDEXT-003` | Worker go-readability extraction | `MDEXT-001` | `MDEXT-005` | yes | no |
| `MDEXT-004` | Worker Jina Reader fallback | `MDEXT-001` | `MDEXT-005` | yes | no |
| `MDEXT-005` | Worker Markdown pipeline integration | `ARTPROC-005`, `MDEXT-002`, `MDEXT-003`, `MDEXT-004` | `SUMGEN-004` | no | no |

---

## Concurrency Rules

- `MDEXT-002` must wait for `ARTPROC-003` because it extends the shared artifact access layer created by article processing.
- `MDEXT-003` and `MDEXT-004` must wait for `MDEXT-001` because both providers must implement the same result taxonomy before orchestration consumes them.
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
5. Validate provider implementations against the shared extractor result taxonomy.
6. Run Worker pipeline transaction and logging tests.
7. Run complete Worker and Gateway verification.

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

- all task acceptance criteria are satisfied;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
