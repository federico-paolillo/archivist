---
id: ARTPROC-005
feature: article-processing
title: Worker Executable Processing Pipeline Orchestration
status: done
depends_on: [ARTPROC-004, TELING-001, TELING-003]
blocks: [MDEXT-005]
parallel: false
exec_plan: null
canonical: true
---

# ARTPROC-005: Worker Executable Processing Pipeline Orchestration

## Objective

Implement the executable Worker processing pipeline that claims article-processing jobs, fetches HTML, stores `snapshot.html`, hands successful jobs to Markdown extraction, and reaches terminal success only after summary-complete processing.

## Story / Context

As the deployed Worker process, I need `archivist-worker process` to invoke the composed article-processing pipeline so a queued Telegram URL proceeds through fetch, snapshot, Markdown extraction, and summary generation before terminal success is recorded.

## Scope

This task includes:

- Job claim integration for article-processing jobs.
- `archivist-worker process`.
- `archivist-worker process --once`.
- `archivist-worker process --idle-sleep <duration>`.
- Startup validation that the processing pipeline is configured with SQLite and `DATA_DIR`.
- A cancellable single-worker polling loop.
- `ProcessOne(ctx)` behavior that reports whether a job was processed.
- Pipeline stages: dequeue, resolve/fetch, snapshot, Markdown handoff, summary-complete terminal transition.
- Snapshot-success behavior that updates `articles.canonical_url`, promotes `snapshot.html`, and continues to Markdown extraction without marking article/job success.
- Terminal success behavior that is reached only after downstream summary generation promotes `summary.md`.
- Failure transaction updating article failure state, job failure state, TTL, and notification row.
- Structured Worker logging for article ID, job ID, URL, final URL, status, duration, and ARC code when applicable.
- Worker tests for executable invocation, one-shot mode, success, failure, cancellation, and transactional behavior.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `ARTPROC-004`.
- Completed `TELING-001` persistence contract.
- Completed `TELING-003` worker terminal notification contract.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker `process` command and orchestration service.
- Transactional terminal state persistence.
- Tests for executable processing, successful snapshot handoff, summary-complete success, and ARC-coded failures.

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
- `docs/ARTIFACTS.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`
- `./ARTPROC-004-worker-url-resolver-and-html-fetcher.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Snapshot success continues to downstream processing
  Given a running article-processing job has fetched valid HTML
  When the Worker stores snapshot.html successfully
  Then articles.canonical_url is the final redirected URL
  And Markdown extraction is invoked
  And articles.status is not set to ready at the snapshot boundary
  And jobs.status is not set to succeeded at the snapshot boundary
  And no success notification row is inserted at the snapshot boundary

Scenario: Process command handles one queued job
  Given the Worker has SQLite and DATA_DIR configured
  And a queued article-processing job exists
  When `archivist-worker process --once` runs
  Then the executable path claims the queued job
  And snapshot.html is written for the article
  And articles.canonical_url is updated

Scenario: Summary-complete processing succeeds
  Given the Worker has stored snapshot.html
  And downstream Markdown extraction and summary generation succeed
  When the Worker completes the pipeline
  Then summary.md is promoted
  And articles.status is ready
  And jobs.status is succeeded
  And one pending success notification row exists for the job

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
- `archivist-worker process` is registered.
- `process --once` processes zero or one queued job and exits.
- Daemon mode polls one job at a time, sleeps only when idle, and exits cleanly on cancellation.
- Missing SQLite or `DATA_DIR` configuration fails startup.
- Snapshot success atomically writes `snapshot.html`, updates `articles.canonical_url`, and continues to Markdown extraction in the current pipeline.
- Snapshot success does not commit article/job success, terminal timestamps/TTL, or a success notification while downstream processing remains.
- Terminal success is committed only after downstream summary generation promotes `summary.md`.
- Failure sets article failed, stores ARC-coded public error, sets job failed, and inserts notification in one transaction.
- Tests cover executable invocation, success, failure, cancellation, canonical URL update, notification creation, and transaction rollback behavior.
- Task status and `PLAN.md` are updated if the task is completed.

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

- `MDEXT-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Summary-complete processing is the final success criterion; snapshot success is only the handoff into Markdown extraction.
