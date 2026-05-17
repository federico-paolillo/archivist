---
id: ARTPROC-007
feature: article-processing
title: Worker Executable Processing Command
status: done
depends_on: [ARTPROC-005]
blocks: []
parallel: false
exec_plan: ../plans/ARTPROC-007-worker-executable-processing-command.execplan.md
canonical: true
---

# ARTPROC-007: Worker Executable Processing Command

## Objective

Expose article-processing through the Worker executable with an explicit `process` command.

## Story / Context

As the deployed Worker process, I need `archivist-worker process` to invoke the composed pipeline so queued jobs are actually dequeued and processed.

## Scope

This task includes:

- `archivist-worker process`.
- `archivist-worker process --once`.
- `archivist-worker process --idle-sleep <duration>`.
- Startup validation that the snapshot pipeline is configured.
- A cancellable single-worker polling loop.
- `SnapshotPipeline.ProcessOne(ctx)` reporting whether a job was processed.
- Executable-surface tests for the process command.

## Out of Scope

This task does not include:

- Root-command processing behavior.
- Worker parallelism.
- Automatic retries.
- Scheduling or backoff beyond idle polling.
- Worker configuration key reconciliation.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `ARTPROC-005`.
- `../plans/ARTPROC-007-worker-executable-processing-command.execplan.md`
- Worker composition root that wires `SnapshotPipeline`.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker CLI `process` command.
- One-shot processing mode for executable validation.
- Daemon polling loop for production processing.
- Tests proving the executable command reaches the processing path.

## Expected Affected Areas

```text
src/worker/internal/app/
src/worker/internal/pipeline/
docs/specs/article-processing/
docs/conventions/WORKER.md
docs/ARCHITECTURE.md
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `./ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `../plans/ARTPROC-007-worker-executable-processing-command.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Process command handles one queued job
  Given the Worker has SQLite and DATA_DIR configured
  And a queued article-processing job exists
  When `archivist-worker process --once` runs
  Then the command claims the queued job through the executable path
  And snapshot.html is written for the article
  And articles.canonical_url is updated

Scenario: Process command rejects missing pipeline configuration
  Given the Worker has no configured snapshot pipeline
  When `archivist-worker process` starts
  Then it exits with an error explaining that SQLITE_PATH and DATA_DIR are required

Scenario: Process loop exits on cancellation
  Given no queued jobs are available
  When the process loop context is cancelled while idle
  Then the loop exits without an error
```

## Done When

- `archivist-worker process` is registered.
- `process --once` processes zero or one queued job and exits.
- Daemon mode polls one job at a time, sleeps only when idle, and exits cleanly on cancellation.
- Missing `SnapshotPipeline` fails startup.
- Pipeline tests and command tests cover the new `ProcessOne(ctx) (processed, error)` contract.
- Task status and `PLAN.md` are updated.
- `DIARY.md` records the root cause and validation.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- `cd src/worker && go run ./cmd/app process --help`

## Dependencies

Depends on:

- `ARTPROC-005`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
../plans/ARTPROC-007-worker-executable-processing-command.execplan.md
```

## Open Questions

- None.

## Notes

- This task was added retroactively after review found that completed pipeline code was not reachable from the Worker executable.
- The configuration key reconciliation listed as out of scope here is corrected by `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`.
