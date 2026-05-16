---
id: ARTPROC-007-PLAN
task: ../tasks/ARTPROC-007-worker-executable-processing-command.md
status: completed
canonical: true
---

# ExecPlan: ARTPROC-007 Worker Executable Processing Command

## Objective

Make the Worker executable process queued article jobs through an explicit `process` command.

## Linked Task

- `../tasks/ARTPROC-007-worker-executable-processing-command.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/ARTPROC-007-worker-executable-processing-command.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/conventions/WORKER.md`
- `../tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`

## Assumptions

- v0 uses a single Worker instance.
- The Worker command must be explicit; the root command must not process jobs.
- `SnapshotPipeline` is the processing boundary the executable must invoke.
- Configuration key reconciliation is a separate corrective task.

## Non-Goals

- Do not introduce concurrent processing.
- Do not introduce retries, schedules, or backoff policy.
- Do not move processing to the root command.
- Do not change Telegram or Gateway behavior.

## Implementation Sequence

1. Add a `process` command to the Worker CLI registry in `src/worker/internal/app/program.go`.
2. Put the command entrypoint and loop helper in `src/worker/internal/app/process.go`, without depending on `urfave/cli/v3` types.
3. Add `--once` and `--idle-sleep` flags to the `process` command.
4. Validate `App.SnapshotPipeline != nil` before entering the loop.
5. Change `SnapshotPipeline.ProcessOne(ctx)` to return `(processed bool, err error)`.
6. Implement one-shot mode by calling `ProcessOne` at most once and returning nil when no job exists.
7. Implement daemon mode by immediately continuing after processed work and sleeping only when no job is available.
8. Treat context cancellation as graceful shutdown.
9. Add command/action tests for one-shot processing, missing pipeline configuration, and idle cancellation.
10. Update ALM docs and diary after validation.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- `cd src/worker && go run ./cmd/app process --help`

## Documentation Updates Required

- Update `../SPEC.md` with the executable processing requirement.
- Update `../PLAN.md` with `ARTPROC-007`.
- Add this task and ExecPlan.
- Update `docs/ARCHITECTURE.md`.
- Update `docs/conventions/WORKER.md`.
- Update `docs/BOOKKEEPING.md` and `docs/REBUILD.md` executable-boundary validation rules.
- Append `../DIARY.md`.

## Risks

- Leaving no queued job indistinguishable from processed work would force a busy loop or unnecessary sleeps.
- Running processing from the root command would make accidental execution too easy.
- Testing only `SnapshotPipeline` would not prove production binary behavior.

## Rollback / Recovery Notes

- Reverting `process` command registration returns the binary to diagnostic-only behavior and must be treated as a blocking regression.
- Failed jobs remain terminal in v0 and are manually requeued by sending the URL again.

## Completion Criteria

- `archivist-worker process --once` reaches the composed processing path in tests.
- Missing pipeline configuration fails clearly.
- Idle daemon mode exits cleanly on cancellation.
- Worker validation passes.
