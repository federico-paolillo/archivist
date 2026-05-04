---
id: ARTPROC-005-PLAN
task: ../tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md
status: proposed
canonical: true
---

# ExecPlan: ARTPROC-005 Worker Snapshot Pipeline Orchestration

## Objective

Implement the Worker orchestration slice that claims article-processing jobs, fetches HTML, stores `snapshot.html`, and commits terminal article/job/notification state transactionally.

## Linked Task

- `../tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`

## Assumptions

- `TELING-001` has provided the shared SQLite schema and repository contracts.
- `TELING-003` has provided or defined atomic terminal job transitions and notification insertion.
- `ARTPROC-004` provides a fetch result containing final URL and HTML bytes, or an ARC-coded public failure.
- Snapshot success is an interim completion point until the v0 extraction/summarization feature supersedes it.

## Non-Goals

- Do not implement extraction, scoring, Markdown generation, or summarization.
- Do not create summary or content placeholder artifacts.
- Do not add retry states or automatic retry scheduling.
- Do not call Telegram APIs from the Worker.

## Implementation Sequence

1. Identify the Worker composition root and job-processing loop boundaries.
2. Wire the existing queued-job claim contract into an article-processing handler.
3. Load the article original URL for the claimed job.
4. Call the `ARTPROC-004` resolver/fetcher.
5. On fetch success, call the `ARTPROC-003` artifact layer to atomically write `snapshot.html`.
6. Keep extraction and rating as explicit no-op stages with no side effects.
7. Persist snapshot success in one transaction: update `articles.canonical_url`, set article `ready`, clear article error, mark job `succeeded`, set completion/TTL fields, and insert one pending notification.
8. On any ARC-coded fetch or snapshot failure, persist failure in one transaction: set article `failed`, set ARC-coded `articles.error_message`, mark job `failed`, persist job error context, set completion/TTL fields, and insert one pending notification.
9. On unknown implementation failures, map public article error to `ARC-999` while logging operational details.
10. Add structured logs for `article_id`, `job_id`, original URL, final URL when available, status, duration, and ARC code.
11. Add tests for success, fetch failures, snapshot failure, notification creation, canonical URL update, and rollback behavior.
12. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Inspect one local successful fixture run to confirm `snapshot.html` exists and no placeholder artifacts exist.
- Inspect SQLite state for article/job/notification atomic success and failure transitions.

## Documentation Updates Required

- Update `../tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new processing behavior, failure code, or state transition to `../SPEC.md` or `docs/conventions/ERRORS.md`.

## Risks

- Committing article/job state before snapshot rename would create false success; snapshot write must complete before terminal success transaction.
- Missing transaction boundaries could create a terminal job without a notification row.
- Persisting raw HTTP details in `articles.error_message` would violate the ARC convention.
- Gateway notification dispatch must handle snapshot-only success until summary artifacts exist.

## Rollback / Recovery Notes

- Failed jobs are terminal in v0; manual requeue is performed by sending the URL again.
- A failed snapshot write must not promote partial files to `snapshot.html`.
- If notification insertion fails, the terminal state transaction must roll back.

## Completion Criteria

- Worker tests cover successful snapshot processing and ARC-coded failures.
- Terminal article/job/notification state changes are atomic.
- Worker validation passes.
- No extraction, scoring, Markdown, summary, or retry behavior is introduced.
