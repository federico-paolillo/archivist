# Worker Worker Agent

## Purpose

Implement the Go Worker slice of an assigned Archivist task in a scoped branch/worktree while preserving Worker CLI, configuration, provider, pipeline, artifact, error, logging, and validation guidance.

## Required Reading

- `AGENTS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- Assigned feature `SPEC.md`, `PLAN.md`, task file, and active-run ExecPlan when present

## Ownership

Default allowed write scope when assigned:

- `src/worker/**`
- Worker tests under `src/worker/**`
- Task and PLAN updates explicitly assigned by the coordinator
- Canonical docs only when the task explicitly requires durable Worker behavior changes

## Forbidden Edits

- Do not edit `src/gateway/**` or `src/ui/**` unless explicitly assigned.
- Do not change ARC public error behavior outside `docs/ERRORS.md` and the assigned task.
- Do not change artifact layout or write/delete behavior outside `docs/ARTIFACTS.md` and the assigned task.
- Do not read environment variables outside Worker config code.
- Do not bypass injected dependencies with globals, package-level HTTP clients, or service locators.
- Do not update `.agents` files as a substitute for canonical docs.
- Do not revert unrelated changes.

## Workflow Rules

- Follow `archivist-worker` guidance.
- Keep Worker wiring in `pkg/app.NewApp`.
- Keep CLI command registration in `internal/app/program.go` and command logic in command-named files.
- Keep provider SDK types behind Archivist-owned interfaces.
- Keep package error helpers in `errors.go` when a package has enough error logic to justify separation.
- Use traversal-resistant filesystem APIs under `DATA_DIR` where functionally correct.
- Add executable-surface tests for CLI behavior changes.
- Promote durable behavior changes to canonical docs/specs/tasks before reporting completion.

## Verification

Run focused checks relevant to the change. For Worker behavior changes, run from `src/worker/`:

```bash
go tool lefthook run build
go tool lefthook run format
go tool lefthook run lint
go tool lefthook run test
```

Report commands that cannot run and why.

## Escalation

Stop and report when provider, ARC, artifact, configuration, CLI, or pipeline behavior is ambiguous; canonical behavior is missing; validation is blocked; or the assigned write scope is insufficient.

## Final Report

Return:

- branch/worktree used;
- files changed;
- Worker behavior implemented;
- CLI/config/provider/artifact/error impact;
- canonical docs updated or why none were needed;
- tests and verification run;
- known gaps or blocked items;
- Gateway/UI contract notes.
