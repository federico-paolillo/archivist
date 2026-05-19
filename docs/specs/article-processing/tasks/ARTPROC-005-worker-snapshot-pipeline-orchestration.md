---
id: ARTPROC-005
feature: article-processing
title: Worker Snapshot Pipeline Orchestration
status: done
depends_on: [ARTPROC-004, TELING-001, TELING-003]
blocks: [ARTPROC-006, ARTPROC-007, MDEXT-005]
parallel: false
exec_plan: ../plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md
canonical: true
---

# ARTPROC-005: Worker Snapshot Pipeline Orchestration

## Objective

Implement the Worker pipeline that claims article-processing jobs, fetches HTML, stores `snapshot.html`, and either commits interim terminal state when no downstream pipeline exists or hands off to Markdown extraction in final v0.

## Story / Context

As the Worker, I need to expose the first complete URL-to-article processing pipeline slice so a queued Telegram URL becomes either a stored HTML snapshot with success notification intent or an ARC-coded terminal failure.

## Scope

This task includes:

- Job claim integration for article-processing jobs.
- Pipeline stages: dequeue, resolve/fetch, snapshot, extraction slot no-op, rating slot no-op, terminal transition.
- Snapshot-success transaction updating article status, canonical URL, job success, TTL, and notification row only when no downstream pipeline exists.
- Final v0 handoff behavior where snapshot success continues to Markdown extraction and does not mark article/job success.
- Failure transaction updating article failure state, job failure state, TTL, and notification row.
- Structured Worker logging for article ID, job ID, URL, final URL, status, duration, and ARC code when applicable.
- Worker tests for success, failure, and transactional behavior.

## Out of Scope

This task does not include:

- Jina.ai extraction.
- go-readability extraction.
- Candidate scoring.
- Markdown or summary artifact creation.
- Gateway notification dispatch implementation.
- Automatic retries.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `ARTPROC-004`.
- Completed `TELING-001` persistence contract.
- Completed `TELING-003` worker terminal notification contract.
- `../plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker orchestration service.
- Transactional terminal state persistence.
- Tests for successful snapshot processing and ARC-coded failures.

## Expected Affected Areas

```text
src/worker/
SQLite repository code
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`
- `./ARTPROC-004-worker-url-resolver-and-html-fetcher.md`
- `../plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Snapshot success continues to downstream processing in final v0
  Given a running article-processing job has fetched valid HTML
  When the Worker stores snapshot.html successfully
  Then articles.canonical_url is the final redirected URL
  And Markdown extraction is invoked when markdown-extraction is implemented
  And articles.status is not set to ready at the snapshot boundary
  And jobs.status is not set to succeeded at the snapshot boundary
  And no success notification row is inserted at the snapshot boundary

Scenario: Fetch failure is committed transactionally
  Given a running article-processing job fails with ARC-004
  When the Worker records terminal failure
  Then articles.status is failed
  And articles.error_message starts with "[ARC-004]"
  And jobs.status is failed
  And one pending notification row exists for the job
  And those database changes commit in one transaction
```

## Done When

- Worker pipeline claims and processes queued article-processing jobs.
- Snapshot success atomically writes `snapshot.html`, updates `articles.canonical_url`, and continues to Markdown extraction in final v0.
- Final v0 snapshot success does not commit article/job success, terminal timestamps/TTL, or a success notification.
- Failure sets article failed, stores ARC-coded public error, sets job failed, and inserts notification in one transaction.
- Extraction and rating slots are explicit no-ops or documented extension points.
- Tests cover success, failure, canonical URL update, notification creation, and transaction rollback behavior.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- Inspect SQLite state and artifact directory for one successful local fixture job.

## Dependencies

Depends on:

- `ARTPROC-004`
- `TELING-001`
- `TELING-003`

Blocks:

- `ARTPROC-006`
- `MDEXT-005`

## ExecPlan

ExecPlan:

```text
../plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md
```

## Open Questions

- None.

## Notes

- The `summary-generation` feature supersedes snapshot success as the final processing completion criterion.
