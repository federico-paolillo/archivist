---
feature: summary-generation
status: in_progress
canonical: true
---

# Feature Plan: Summary Generation

## Purpose

This file is the feature-level implementation control board for summary generation. It defines task order, dependencies, concurrency rules, validation sequence, and execution status.

---

## Task DAG

```text
SUMGEN-001 -> SUMGEN-002
SUMGEN-001 -> SUMGEN-003
MDEXT-005 -> SUMGEN-002
WCFG-001 -> SUMGEN-002
WCFG-002 -> SUMGEN-002
SUMGEN-002 -> SUMGEN-004
SUMGEN-003 -> SUMGEN-004
WCFG-001 -> SUMGEN-004
WCFG-002 -> SUMGEN-004
SUMGEN-004 -> SUMGEN-005
TELING-004 -> SUMGEN-005
```

---

## Execution Phases

### Phase 1: Canonical Planning And Standards

- `SUMGEN-001` creates the feature ALM artifacts and updates canonical architecture, design, artifact, configuration, logging, and error conventions.

### Phase 2: Worker Summary Foundations

- `SUMGEN-002` extends Worker artifact access with `content.md` reads and atomic `summary.md` writes.
- `SUMGEN-003` implements `SummarizerService` and the Anthropic SDK-backed adapter.

### Phase 3: Final Pipeline And Notifications

- `SUMGEN-004` integrates summary generation into Worker terminal processing and makes summary success the final v0 success point.
- `SUMGEN-005` replaces Markdown-complete success notifications with summary-based Telegram replies through read-only Gateway artifact access.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `SUMGEN-001` | Create feature artifacts and contracts | done | - | `SUMGEN-002`, `SUMGEN-003` | no | - |
| `SUMGEN-002` | Worker summary artifact access | done | `SUMGEN-001`, `MDEXT-005`, `WCFG-001`, `WCFG-002` | `SUMGEN-004` | yes | - |
| `SUMGEN-003` | Summarizer provider adapter | done | `SUMGEN-001` | `SUMGEN-004` | yes | `plans/SUMGEN-003-summarizer-provider-adapter.execplan.md` |
| `SUMGEN-004` | Worker summary pipeline integration | done | `SUMGEN-002`, `SUMGEN-003`, `WCFG-001`, `WCFG-002` | `SUMGEN-005` | no | `plans/SUMGEN-004-worker-summary-pipeline-integration.execplan.md` (completed) |
| `SUMGEN-005` | Gateway summary notification | blocked | `SUMGEN-004`, `TELING-004` | - | no | `plans/SUMGEN-005-gateway-summary-notification.execplan.md` |

---

## Concurrency Rules

- `SUMGEN-002` and `SUMGEN-003` may run in parallel after their dependencies are done because they own separate artifact and provider-adapter surfaces.
- `SUMGEN-002` must wait for `MDEXT-005` because it extends the Markdown-complete artifact/pipeline boundary.
- Remaining Worker summary tasks must use canonical Worker config from `worker-runtime-configuration/WCFG-001` and non-optional Worker composition from `worker-runtime-configuration/WCFG-002`.
- `SUMGEN-004` is complete and owns Worker summary-complete terminal state.
- `SUMGEN-005` has completed task dependencies but remains blocked until its proposed ExecPlan is accepted or updated.
- Worker pipeline orchestration, SQLite terminal-transition code, and Gateway dispatcher behavior must not be modified concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- Existing SQLite `articles`, `jobs`, and `notifications` contracts from `telegram-ingestion`.
- Markdown output from `markdown-extraction`.
- Deterministic artifact paths in `docs/conventions/ARTIFACTS.md`.
- ARC error-code catalog in `docs/conventions/ERRORS.md`.
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
