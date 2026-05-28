---
id: SUMGEN-004-PLAN
task: ../tasks/SUMGEN-004-worker-summary-pipeline-integration.md
status: completed
canonical: true
---

# ExecPlan: SUMGEN-004 Worker Summary Pipeline Integration

## Objective

Integrate summary generation into Worker orchestration so article-processing jobs succeed only after `summary.md` is atomically written, or fail with ARC-coded errors when summarization or summary artifact persistence fails.

## Linked Task

- `../tasks/SUMGEN-004-worker-summary-pipeline-integration.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/SUMGEN-004-worker-summary-pipeline-integration.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`

## Assumptions

- `MDEXT-005` has provided Markdown extraction and `content.md` promotion.
- `SUMGEN-002` provides deterministic `content.md` reads and atomic `summary.md` writes.
- `SUMGEN-003` provides `SummarizerService` and ARC-coded summarizer failure mapping.
- Summary success is the final v0 completion point.

## Non-Goals

- Do not implement provider adapters or artifact-store primitives in this task.
- Do not add summary JSON or SQLite summary columns.
- Do not add retry states or automatic retry scheduling.
- Do not call Telegram APIs from the Worker.

## Implementation Sequence

1. Identify the Worker pipeline point immediately after `content.md` promotion.
2. Remove or bypass Markdown-boundary terminal success behavior for final v0.
3. Read `content.md` through the artifact layer.
4. Call `SummarizerService` with article metadata and Markdown source.
5. On summarizer failure, retain the ARC-coded public error.
6. Atomically write `summary.md` before opening the terminal success transaction.
7. On summary write failure, map to `ARC-016`.
8. Persist summary success in one transaction: set article `ready`, clear article error, mark job `succeeded`, set completion/TTL fields, and insert one pending notification.
9. Persist summary failure in one transaction: set article `failed`, set ARC-coded `articles.error_message`, mark job `failed`, persist job error context, set completion/TTL fields, and insert one pending notification.
10. Add structured logs for summarizer provider, model, request ID when available, `article_id`, `job_id`, canonical URL, duration, ARC code, and artifact write result.
11. Add tests for summary success, Markdown-not-terminal behavior, generic provider failure, context/request-too-large failure, billing failure, summary write failure, notification creation, and rollback behavior.
12. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Coordinator / Worker / Review Workflow

Execution must use a coordinator-led multi-agent loop:

1. The coordinator owns task state, ALM consistency, validation decisions, and final acceptance. The coordinator does not switch into the Worker or Reviewer role.
2. A `worker` subagent with medium reasoning effort implements the Worker code and focused tests under `src/worker/`. The worker must not modify unrelated Gateway or UI surfaces.
3. After worker completion, a separate `review` subagent with high reasoning effort reviews the implementation and ALM updates. The review must prioritize idiomatic Go, consistency with new and existing code, repository conventions, and ALM document consistency.
4. The coordinator sends confirmed findings back to the worker. The worker fixes only confirmed findings and reports changed files and validation results.
5. The coordinator performs final integration review, runs required validation, and only then updates the task, feature plan, diary, ExecPlan status, and derived masterplan to completion.

Review gates:

- Implementation must match the canonical task, feature spec, Worker conventions, artifact conventions, ARC error conventions, and this ExecPlan.
- Summary success must be terminal only after `summary.md` is atomically promoted.
- Markdown and snapshot stages must remain non-terminal.
- Logs must not include full Markdown, summary text, provider payloads, API keys, or other secrets.
- ALM updates must keep task frontmatter, feature `PLAN.md`, ExecPlan status, `DIARY.md`, and `docs/MASTERPLAN.md` consistent.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Inspect one local successful fixture run to confirm `snapshot.html`, `content.md`, and `summary.md` exist and no `summary.json` exists.
- Inspect SQLite state for article/job/notification atomic success and failure transitions.
- Inspect captured logs for summarizer provider, model, ARC code, and artifact write result fields.

## Documentation Updates Required

- Update `../tasks/SUMGEN-004-worker-summary-pipeline-integration.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new processing behavior, failure code, configuration, logging field, or state transition to `../SPEC.md`, `docs/conventions/ERRORS.md`, or the relevant convention file.

## Risks

- Committing job success before `summary.md` rename would create false success.
- Leaving Markdown-boundary success active would violate final v0 done semantics.
- Persisting raw provider details in `articles.error_message` would violate the ARC convention.
- Logging full Markdown or summary content would expose private article data.

## Rollback / Recovery Notes

- Failed jobs are terminal in v0; manual requeue is performed by sending the URL again.
- A failed summary write must not promote partial files to `summary.md`.
- If notification insertion fails, the terminal state transaction must roll back.
- Existing `snapshot.html` and `content.md` remain available for manual diagnosis or future reprocessing.

## Completion Criteria

- Worker tests cover summary success and ARC-coded provider/artifact failures.
- Terminal article/job/notification state changes are atomic.
- Markdown-boundary success is no longer final v0 success.
- Worker validation passes.
