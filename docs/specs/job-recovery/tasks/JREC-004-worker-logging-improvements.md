---
id: JREC-004
feature: job-recovery
title: Worker logging improvements
status: done
depends_on: [JREC-001]
blocks: [JREC-005]
parallel: true
exec_plan: null
canonical: true
---

# JREC-004: Worker Logging Improvements

## Objective

Improve Worker process-loop and pipeline structured logs so claimed jobs, stage boundaries, terminal outcomes, and terminal persistence failures are diagnosable.

## Scope

This task includes:

- Process-loop iteration start logs.
- Idle/no-job logs.
- Post-claim/pre-URL-load logs.
- Stage start and result logs for fetch, snapshot, canonical URL update, Markdown, summary, terminal success, terminal failure, and terminal-persistence failure.
- Tests that assert important log entries and fields.

## Out of Scope

This task does not include:

- Gateway force delete.
- UI force delete.
- New persisted telemetry.
- Provider adapter info/error logging.
- Secrets or full content logging.

## Inputs

- Existing Worker process command.
- Existing `SnapshotPipeline`, `MarkdownExtractionHandoff`, and `SummaryGenerationHandoff`.
- Existing `slog` test helpers.

## Outputs

- Worker code with structured logs and tests.

## Expected Affected Areas

```text
src/worker/internal/app/**
src/worker/internal/pipeline/**
src/worker/pkg/jobs/**
src/worker/internal/testutils/**
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./JREC-004-worker-logging-improvements.md`
- `docs/ARCHITECTURE.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/summary-generation/SPEC.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Claim is logged before URL loading
  Given a queued job exists
  When the Worker claims it
  Then a log entry includes stage "claim", status "claimed", job_id, and article_id

Scenario: Terminal persistence failure is logged
  Given a job reaches terminal state
  And terminal persistence fails
  When the Worker handles the failure
  Then a log entry includes status "terminal_persist_failed", job_id, article_id, stage, and the diagnostic error
```

## Done When

- Required log entries are present.
- Tests cover the critical zombie-diagnostic paths.
- Logs avoid secrets and full article/provider content.

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

- `JREC-001`

Blocks:

- `JREC-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Implemented and reviewed through the multi-agent Worker worker/reviewer workflow.
- Validation: `go tool lefthook run build`, `go tool lefthook run format`, `go tool lefthook run lint`, and integrated `go tool lefthook run test` passed.
