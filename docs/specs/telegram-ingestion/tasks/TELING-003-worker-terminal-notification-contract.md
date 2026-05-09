---
id: TELING-003
feature: telegram-ingestion
title: Worker terminal notification contract
status: done
depends_on: [TELING-001]
blocks: [TELING-004]
parallel: true
exec_plan: null
canonical: true
---

# TELING-003: Worker Terminal Notification Contract

## Objective

Implement the worker-side persistence behavior that atomically updates article/job terminal state and creates notification rows when Telegram-originated jobs succeed or fail.

## Story / Context

As the worker, I need to report terminal job outcomes through SQLite so the gateway can send Telegram replies without direct worker-to-Telegram or worker-to-gateway calls.

## Scope

This task includes:

- Claiming queued jobs atomically with `UPDATE ... RETURNING`.
- Detecting when a terminal job originated from Telegram.
- Updating article status to `ready` or `failed`.
- Updating job status to `succeeded` or `failed`.
- Persisting final job error state on failure.
- Preserving ARC-coded public article-processing failures on the article/job error fields used by downstream notification dispatch.
- Creating one pending notification row for the terminal job.
- Setting terminal job `expires_at` to 14 days after completion.
- Ensuring article update, job update, and notification insert are atomic.
- Worker tests for success, failure, and non-Telegram jobs.

## Out of Scope

This task does not include:

- Telegram API calls.
- Gateway notification dispatch.
- The article fetching/extraction/summarization implementation except where tests need fixtures for terminal state.
- Gateway webhook ingestion.
- Automatic retry behavior.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `TELING-001` persistence contract.
- Worker job terminal state model.
- Text summary artifact contract from summary generation when available.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker repository/service methods that claim queued jobs and write terminal article/job/notification state.
- Worker tests proving atomic claim and atomic terminal transition plus notification creation.

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
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `./TELING-001-persistence-contracts.md`

Do not load unrelated feature folders unless required by discovered dependencies.

## Acceptance Criteria

```gherkin
Scenario: Worker claims a queued job
  Given a queued job exists
  When the worker claims work
  Then the job status changes to running using UPDATE RETURNING
  And no locked_at or locked_by fields are required

Scenario: Telegram-originated job succeeds
  Given a job has Telegram origin metadata
  And the worker has a generated summary
  When the worker marks the job succeeded
  Then the article status is ready
  And the job status is succeeded
  And the job expires_at is 14 days after completion
  And one pending notification row is created for the job

Scenario: Telegram-originated job fails
  Given a job has Telegram origin metadata
  And the job has a latest error message
  When the worker marks the job failed
  Then the article status is failed
  And the job status is failed
  And the job error_message stores that error
  And the job expires_at is 14 days after completion
  And one pending notification row is created for the job

Scenario: Telegram-originated article-processing job fails with an ARC-coded error
  Given a job has Telegram origin metadata
  And the worker article-processing failure is "[ARC-003] The URL was not found."
  When the worker marks the job failed
  Then the article status is failed
  And the article error_message stores "[ARC-003] The URL was not found."
  And the job status is failed
  And the job error_message stores "[ARC-003] The URL was not found."
  And one pending notification row is created for the job

Scenario: Non-Telegram job reaches terminal state
  Given a job has no Telegram origin metadata
  When the worker marks the job terminal
  Then no Telegram notification row is created
```

## Done When

- Worker claims queued jobs atomically with `UPDATE ... RETURNING`.
- Worker terminal persistence writes notification rows only for Telegram-originated jobs.
- Worker terminal persistence updates article, job, and notification state in one transaction.
- Success writes deterministic artifacts and marks article/job success.
- Failure persists final error and marks article/job failure.
- Article-processing failures preserve ARC-coded public error text for downstream Gateway notification dispatch.
- No worker retry state or retry scheduling exists.
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

- None.

## Dependencies

Depends on:

- `TELING-001`

Blocks:

- `TELING-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Worker code must not import or depend on Telegram Bot API clients.
- ARC coding is required only for persisted public article-processing failures, not for Telegram protocol, authorization, acknowledgement, or delivery errors.
