---
id: MDEXT-005-PLAN
task: ../tasks/MDEXT-005-worker-markdown-pipeline-integration.md
status: completed
canonical: true
---

# ExecPlan: MDEXT-005 Worker Markdown Pipeline Integration

## Objective

Integrate Markdown extraction into the Worker orchestration slice so `content.md` is atomically written before final v0 processing continues to summary generation, or fails with ARC-coded errors when local extraction and Jina fallback cannot produce Markdown.

## Linked Task

- `../tasks/MDEXT-005-worker-markdown-pipeline-integration.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/MDEXT-005-worker-markdown-pipeline-integration.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`

## Assumptions

- `ARTPROC-005` has provided the Worker pipeline that snapshots HTML and persists snapshot-related failures.
- `MDEXT-002` provides atomic `content.md` writes.
- `MDEXT-003` provides a local go-readability `MarkdownExtractor`.
- `MDEXT-004` provides a Jina Reader `MarkdownExtractor` with ARC-coded failure mapping.
- Markdown success is an intermediate stage in final v0 because summary generation is part of the canonical pipeline.

## Non-Goals

- Do not implement provider adapters or artifact-store primitives in this task.
- Do not implement summary generation.
- Do not add extraction telemetry columns.
- Do not add retry states or automatic retry scheduling.
- Do not call Telegram APIs from the Worker.

## Implementation Sequence

1. Identify the existing Worker pipeline stage after `snapshot.html` is promoted.
2. Replace snapshot-only terminal success with a Markdown extraction stage.
3. Read snapshot HTML through the artifact layer.
4. Call the local `MarkdownExtractor` with snapshot bytes and canonical URL.
5. On local success, retain the Markdown and selected provider metadata.
6. On local unreadable or local extraction failure, log the fallback reason and call the Jina `MarkdownExtractor` when enabled.
7. On Jina success, retain the Markdown and selected provider metadata.
8. On Jina failure, map the terminal public error to the Jina ARC code; use `ARC-011` for insufficient balance.
9. Atomically write `content.md` before any downstream summary-generation handoff.
10. Hand off to the summary stage after `content.md` is promoted and do not mark the article/job succeeded at the Markdown boundary.
11. Persist Markdown failure in one transaction: set article `failed`, set ARC-coded `articles.error_message`, mark job `failed`, persist job error context, set completion/TTL fields, and insert one pending notification.
12. Add structured logs for provider attempts, fallback reason, selected provider, `article_id`, `job_id`, canonical URL, duration, ARC code, and artifact write result.
13. Add tests for local success handoff, local unreadable plus Jina success handoff, local failure plus Jina success handoff, Jina general failure, Jina insufficient balance, Markdown write failure, failure notification creation, abstraction boundaries, and rollback behavior.
14. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Inspect one local successful fixture run to confirm `snapshot.html` and `content.md` exist and no summary artifacts exist.
- Inspect SQLite state to confirm Markdown success is non-terminal and failures atomically update article/job/notification state.
- Inspect captured logs for provider attempt, fallback decision, selected provider, and artifact write result fields.

## Documentation Updates Required

- Update `../tasks/MDEXT-005-worker-markdown-pipeline-integration.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new processing behavior, failure code, configuration, logging field, or state transition to `../SPEC.md`, `docs/conventions/ERRORS.md`, or the relevant convention file.

## Risks

- Advancing the pipeline before `content.md` rename would create an invalid downstream handoff.
- Missing fallback logs would make provider decisions hard to diagnose.
- Mapping Jina balance errors to generic failures would hide the operator action required to restore processing.
- Persisting raw provider details in `articles.error_message` would violate the ARC convention.
- Gateway notification dispatch must not treat snapshot or Markdown completion as final success once this task is complete.

## Rollback / Recovery Notes

- Failed jobs are terminal in v0; manual requeue is performed by sending the URL again.
- A failed Markdown write must not promote partial files to `content.md`.
- If notification insertion fails, the terminal state transaction must roll back.
- Existing `snapshot.html` remains available for manual diagnosis or future reprocessing.

## Completion Criteria

- Worker tests cover Markdown success handoff and ARC-coded provider/artifact failures.
- Terminal failure article/job/notification state changes are atomic.
- Critical extraction decisions are logged.
- Worker validation passes.
- No summary or retry behavior is introduced.
