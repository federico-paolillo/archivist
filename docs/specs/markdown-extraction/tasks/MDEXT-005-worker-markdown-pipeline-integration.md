---
id: MDEXT-005
feature: markdown-extraction
title: Worker Markdown Pipeline Integration
status: blocked
depends_on: [ARTPROC-005, MDEXT-002, MDEXT-003, MDEXT-004]
blocks: [MDEXT-006]
parallel: false
exec_plan: ../plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md
canonical: true
---

# MDEXT-005: Worker Markdown Pipeline Integration

## Objective

Integrate Markdown extraction into the Worker pipeline through the `MarkdownExtractor` abstraction so `content.md` is atomically written before the pipeline continues to summary generation or, until summary generation exists, commits interim terminal success.

## Story / Context

As the Worker, I need the article-processing pipeline to progress from HTML snapshotting into Markdown extraction before marking the article ready and the job succeeded.

## Scope

This task includes:

- Pipeline sequence after HTML snapshotting: local extractor, optional Jina extractor fallback, Markdown artifact write, and either summary-generation handoff or interim terminal transition.
- Calling only the Worker-owned `MarkdownExtractor` contract from orchestration code.
- Structured logs for provider attempts, fallback decisions, selected provider, ARC code, and artifact write result.
- Markdown-success transaction updating article status, job success, TTL, and notification row.
- Markdown-failure transaction updating article failure state, job failure state, TTL, and notification row.
- Tests for abstraction usage, success, fallback, provider failure, artifact failure, logging, and transactional behavior.

## Out of Scope

This task does not include:

- Local extractor implementation.
- Jina fallback implementation.
- Artifact access implementation.
- Gateway notification dispatch implementation.
- LLM summarization.
- Automatic retries.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `ARTPROC-005`.
- Completed `MDEXT-002`.
- Completed `MDEXT-003`.
- Completed `MDEXT-004`.
- `../plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker pipeline orchestration with Markdown extraction through `MarkdownExtractor`.
- Transactional terminal state persistence after Markdown success or failure.
- Tests for Markdown-complete pipeline behavior.

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
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `./MDEXT-002-worker-markdown-artifact-access.md`
- `./MDEXT-003-worker-go-readability-extraction.md`
- `./MDEXT-004-worker-jina-reader-fallback.md`
- `../plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown success is committed transactionally
  Given a running article-processing job has stored snapshot.html
  And local extraction produces Markdown
  When the Worker stores content.md successfully
  Then articles.status is ready
  And articles.error_message is null
  And jobs.status is succeeded
  And one pending notification row exists for the job
  And those database changes commit in one transaction

Scenario: local unreadable result falls back to Jina
  Given local extraction returns unreadable
  And Jina Reader returns Markdown
  When the Worker processes Markdown extraction
  Then it logs the fallback reason
  And stores content.md
  And commits success

Scenario: extraction failure is committed transactionally
  Given local extraction cannot produce Markdown
  And Jina Reader fails with ARC-010
  When the Worker records terminal failure
  Then articles.status is failed
  And articles.error_message starts with "[ARC-010]"
  And jobs.status is failed
  And one pending notification row exists for the job
  And those database changes commit in one transaction
```

## Done When

- Worker pipeline calls extractor abstractions only; provider-specific SDK types do not enter orchestration.
- Worker pipeline marks Markdown-stage success only after `content.md` is promoted.
- Final v0 success remains blocked on summary generation once `summary-generation` is implemented.
- Failure sets article failed, stores ARC-coded public error, sets job failed, and inserts notification in one transaction.
- Logs capture provider attempt, fallback reason, selected provider, ARC code, and artifact write result.
- Tests cover local success, Jina fallback success, provider failure, Markdown write failure, notification creation, and transaction rollback behavior.
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

- `ARTPROC-005`
- `MDEXT-002`
- `MDEXT-003`
- `MDEXT-004`

Blocks:

- `MDEXT-006`

## ExecPlan

ExecPlan:

```text
../plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md
```

## Open Questions

- None.

## Notes

- The `summary-generation` feature supersedes Markdown success as the final processing completion criterion.
