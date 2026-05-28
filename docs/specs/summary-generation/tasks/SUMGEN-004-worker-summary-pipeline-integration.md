---
id: SUMGEN-004
feature: summary-generation
title: Worker Summary Pipeline Integration
status: done
depends_on: [SUMGEN-002, SUMGEN-003, WCFG-001, WCFG-002]
blocks: [SUMGEN-005]
parallel: false
exec_plan: ../plans/SUMGEN-004-worker-summary-pipeline-integration.execplan.md
canonical: true
---

# SUMGEN-004: Worker Summary Pipeline Integration

## Objective

Integrate summary generation into the Worker pipeline so final article/job success is committed only after `summary.md` is atomically written.

## Story / Context

As the Worker, I need the article-processing pipeline to progress from Markdown extraction into summary generation before marking the article ready and the job succeeded.

## Scope

This task includes:

- Pipeline sequence after Markdown extraction: read `content.md`, summarize through `SummarizerService`, write `summary.md`, terminal transition.
- Replacing Markdown-complete terminal success with summary-complete terminal success.
- Summary-success transaction updating article status, job success, TTL, and notification row.
- Summary-failure transaction updating article failure state, job failure state, TTL, and notification row.
- Structured logs for summarizer provider, model, request ID when available, ARC code, and artifact write result.
- Tests proving snapshot and Markdown stages do not mark success once summary generation exists.
- Tests for success, provider failure, billing failure, context/request-too-large failure, artifact failure, logging, and transactional behavior.

## Out of Scope

This task does not include:

- Summarizer provider implementation.
- Artifact access implementation.
- Gateway notification dispatch implementation.
- Automatic retries.
- Summary JSON or schema validation.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `SUMGEN-002`.
- Completed `SUMGEN-003`.
- Completed `WCFG-001`.
- Completed `WCFG-002`.
- `../plans/SUMGEN-004-worker-summary-pipeline-integration.execplan.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker pipeline orchestration with summary generation.
- Transactional terminal state persistence after summary success or failure.
- Tests for summary-complete pipeline behavior.

## Expected Affected Areas

```text
src/worker/
SQLite repository code
Worker logging
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
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-002-non-optional-worker-composition.md`
- `./SUMGEN-002-worker-summary-artifact-access.md`
- `./SUMGEN-003-summarizer-provider-adapter.md`
- `../plans/SUMGEN-004-worker-summary-pipeline-integration.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Summary success is committed transactionally
  Given a running article-processing job has stored content.md
  And the summarizer produces text
  When the Worker stores summary.md successfully
  Then articles.status is ready
  And articles.error_message is null
  And jobs.status is succeeded
  And one pending notification row exists for the job
  And those database changes commit in one transaction

Scenario: Markdown success is not terminal
  Given summary generation is implemented
  When the Worker promotes content.md
  Then articles.status is not set to ready at the Markdown boundary
  And jobs.status is not set to succeeded at the Markdown boundary
  And no success notification is inserted before summary.md is promoted

Scenario: summarizer failure is committed transactionally
  Given the summarizer fails with ARC-013
  When the Worker records terminal failure
  Then articles.status is failed
  And articles.error_message starts with "[ARC-013]"
  And jobs.status is failed
  And one pending notification row exists for the job
  And those database changes commit in one transaction
```

## Done When

- Worker pipeline marks success only after `summary.md` is promoted.
- Snapshot and Markdown stages are intermediate in the final v0 pipeline.
- Failure sets article failed, stores ARC-coded public error, sets job failed, and inserts notification in one transaction.
- Logs capture summarizer provider, model, request ID when available, ARC code, and artifact write result.
- Tests cover summary success, provider failure, context/request-too-large failure, billing failure, summary write failure, notification creation, and transaction rollback behavior.
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

Validation performed on 2026-05-23:

- `cd src/worker && go test ./...` — passed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed.
- SQLite terminal state and `summary.md` artifact outcomes were validated by Worker pipeline and executable-surface tests.

## Dependencies

Depends on:

- `SUMGEN-002`
- `SUMGEN-003`
- `WCFG-001`
- `WCFG-002`

Blocks:

- `SUMGEN-005`

## ExecPlan

ExecPlan:

```text
../plans/SUMGEN-004-worker-summary-pipeline-integration.execplan.md
```

## Open Questions

- None.

## Notes

- Do not call Telegram APIs from the Worker.
- Status is `done`; implementation, review fixes, and required Worker validation are complete.
- SUMGEN-004 is the sole owner of structured log entries for the summarization pipeline: provider, model, provider request id, ARC code on failure, `article_id`, `job_id`, canonical URL, duration (measured by orchestration around the adapter call), and artifact write result. Adapters do not log. See `docs/conventions/WORKER.md` §Structured Logging.
