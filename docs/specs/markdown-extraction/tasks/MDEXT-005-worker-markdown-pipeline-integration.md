---
id: MDEXT-005
feature: markdown-extraction
title: Worker Markdown Pipeline Integration
depends_on: [ARTPROC-005, MDEXT-002, MDEXT-003, MDEXT-004]
blocks: [SUMGEN-004]
parallel: false
requires_exec_plan: false
canonical: true
---
# MDEXT-005: Worker Markdown Pipeline Integration

## Objective

Integrate Markdown extraction as the intermediate Worker pipeline stage through the `MarkdownExtractor` abstraction so `content.md` is atomically written before the summary-continuation boundary is reached.

## Story / Context

As the Worker, I need the article-processing pipeline to progress from HTML snapshotting into Markdown extraction, then expose the continuation point consumed by summary generation without implementing summary generation in this task.

## Scope

This task includes:

- Pipeline sequence after HTML snapshotting: local extractor, mandatory Jina extractor fallback, Markdown artifact write, and summary-continuation boundary.
- Calling only the Worker-owned `MarkdownExtractor` contract from orchestration code.
- Structured logs for provider attempts, fallback decisions, selected provider, ARC code, and artifact write result.
- Markdown-success continuation boundary after `content.md` is promoted.
- Markdown-failure transaction updating article failure state, job failure state, TTL, and notification row.
- Tests for abstraction usage, success, fallback, provider failure, artifact failure, logging, and transactional behavior.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- Requires `ARTPROC-005`.
- Requires `MDEXT-002`.
- Requires `MDEXT-003`.
- Requires `MDEXT-004`.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker pipeline orchestration with Markdown extraction through `MarkdownExtractor`.
- Pipeline continuation boundary after Markdown success and transactional terminal state persistence after Markdown failure.
- Tests for Markdown-stage handoff behavior.

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
- `docs/ARTIFACTS.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `./MDEXT-001-worker-markdown-extractor-contract.md`
- `./MDEXT-002-worker-markdown-artifact-access.md`
- `./MDEXT-003-worker-go-readability-extraction.md`
- `./MDEXT-004-worker-jina-reader-fallback.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown success reaches summary continuation
  Given a running article-processing job has stored snapshot.html
  And local extraction produces Markdown
  When the Worker stores content.md successfully
  Then the summary-continuation boundary is reached
  And articles.status is not set to ready at the Markdown boundary
  And jobs.status is not set to succeeded at the Markdown boundary
  And no success notification row exists for the job at the Markdown boundary

Scenario: local unreadable result falls back to Jina
  Given local extraction returns unreadable
  And Jina Reader returns Markdown
  When the Worker processes Markdown extraction
  Then it logs the fallback reason
  And stores content.md
  And reaches the summary-continuation boundary without terminal success at the Markdown boundary

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
- Worker pipeline promotes Markdown-stage success only after `content.md` is promoted.
- Final success is not committed at the Markdown boundary; summary generation owns the terminal success transition.
- Failure sets article failed, stores ARC-coded public error, sets job failed, and inserts notification in one transaction.
- Logs capture provider attempt, fallback reason, selected provider, ARC code, and artifact write result.
- Tests cover local success, Jina fallback success, provider failure, Markdown write failure, notification creation, and transaction rollback behavior.

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

- `SUMGEN-004`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.

## Notes

- Summary-complete processing is the final success criterion; Markdown success is only the continuation boundary into summary generation.
- MDEXT-005 is the sole owner of structured log entries for the Markdown extraction pipeline: provider attempt, fallback reason, selected provider, ARC code on failure, `article_id`, `job_id`, canonical URL, duration (measured by orchestration around the adapter call), and artifact write result. Adapters do not log; they return result types with sufficient fields. See `.agents/skills/archivist-worker/SKILL.md` §Structured Logging.
