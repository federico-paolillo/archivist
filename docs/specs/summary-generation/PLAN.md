---
feature: summary-generation
canonical: true
---
# Feature Plan: Summary Generation

## Purpose

This file is the feature-level implementation control board for summary generation. It defines task order, dependencies, concurrency rules, validation sequence, and validation requirements.

---

## Task DAG

```text
MDEXT-002 -> SUMGEN-002
MDEXT-005 -> SUMGEN-004
SUMGEN-002 -> SUMGEN-004
SUMGEN-003 -> SUMGEN-004
SUMGEN-004 -> SUMGEN-005
TELING-004 -> SUMGEN-005
SUMGEN-005 -> UIEND-002
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards


### Phase 2: Worker Summary Foundations

- `SUMGEN-002` extends Worker artifact access with `content.md` reads and atomic `summary.md` writes.
- `SUMGEN-003` implements `SummarizerService` and the Anthropic SDK-backed adapter.

### Phase 3: Final Pipeline And Notifications

- `SUMGEN-004` integrates summary generation into Worker terminal processing and makes summary success the terminal success point.
- `SUMGEN-005` owns summary-success notification body construction and summary artifact reads through read-only Gateway artifact access.

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `SUMGEN-002` | Worker summary artifact access | `MDEXT-002` | `SUMGEN-004` | yes | no |
| `SUMGEN-003` | Summarizer provider adapter | - | `SUMGEN-004` | yes | no |
| `SUMGEN-004` | Worker summary pipeline integration | `MDEXT-005`, `SUMGEN-002`, `SUMGEN-003` | `SUMGEN-005` | no | no |
| `SUMGEN-005` | Gateway summary notification | `SUMGEN-004`, `TELING-004` | `UIEND-002` | no | no |

---

## Concurrency Rules

- `SUMGEN-002` and `SUMGEN-003` may run in parallel after their dependencies are satisfied because they own separate artifact and provider-adapter surfaces.
- `SUMGEN-002` must wait for Markdown artifact access because it extends deterministic article artifact reads and writes.
- `SUMGEN-004` must wait for `MDEXT-005` because summary generation starts after the Markdown pipeline stage promotes `content.md` and reaches the summary-continuation boundary.
- `SUMGEN-004` owns Worker summary-complete terminal state.
- `SUMGEN-005` owns Gateway summary-complete notification content.
- Worker pipeline orchestration, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- Markdown output from `markdown-extraction`.
- Deterministic artifact paths in `docs/ARTIFACTS.md`.
- ARC error-code catalog in `docs/ERRORS.md`.
- Worker-owned `SummarizerService` contract.
- Worker configuration for `LLM_PROVIDER`, `LLM_API_KEY`, and `LLM_MODEL`.
- Gateway read-only artifact abstraction for summary notification.

---

## Validation Sequence

1. Validate canonical docs and task dependencies.
2. Run Worker summary artifact access tests.
3. Run Worker summarizer adapter tests.
4. Run Worker pipeline transaction and logging tests.
5. Run Gateway read-only artifact and notification tests.
6. Run final cross-feature pipeline validation after `SUMGEN-005`: Telegram webhook enqueue -> Worker process through snapshot, Markdown, and summary -> pending notification creation -> Gateway dispatch of the summary reply.
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

### Final Cross-Feature Validation Boundary

After `SUMGEN-005`, validate the completed ingestion/article/Markdown/summary notification path as one boundary:

```text
Telegram webhook enqueue
-> Worker process through snapshot.html
-> Worker Markdown extraction through content.md
-> Worker summary generation through summary.md
-> pending notification row
-> Gateway dispatcher sends summary Telegram reply
```

This boundary is intentionally after `SUMGEN-005` because `TELING-004` owns dispatcher infrastructure and delivery semantics, while `SUMGEN-005` owns summary-success body construction and `summary.md` reads.

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
