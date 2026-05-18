---
name: archivist-worker
description: Use when implementing, reviewing, or planning Archivist Worker module changes under src/worker, including Go code, Worker CLI commands, configuration, provider adapters, artifact writes, pipeline orchestration, jobs, errors, logging, and Worker validation.
allowed-tools: Read Write Edit Grep Glob Bash
---

# Archivist Worker

## Purpose

Use this skill for repository-specific Worker work in Archivist. This skill does not define generic Go style; use the Go Developer Agent for portable Go standards. This skill enforces Archivist's canonical Worker conventions and ALM rules.

## Required Context

Start with the repository orientation bundle:

```text
AGENTS.md
docs/REBUILD.md
docs/specs/INDEX.md
```

For any Worker implementation, review, or standards work, also read:

```text
docs/conventions/GENERAL.md
docs/conventions/WORKER.md
```

Load additional context by trigger:

- `docs/BOOKKEEPING.md`: changing task status, `PLAN.md`, `DIARY.md`, specs, ExecPlans, or dependency/concurrency state.
- `docs/ARCHITECTURE.md`: changing executables, service boundaries, storage, integrations, deployment, runtime topology, or configuration semantics.
- `docs/DESIGN.md`: relying on or changing durable decisions or rebuild-relevant tradeoffs.
- `docs/conventions/ERRORS.md`: changing persisted article-processing failure behavior, ARC codes, public messages, or error classification.
- `docs/conventions/ARTIFACTS.md`: changing artifact paths, filenames, atomic writes, or article data layout.
- Relevant feature `SPEC.md`, `PLAN.md`, task files, and linked ExecPlans before implementing a task.

Do not load unrelated feature folders unless dependencies or canonical docs require them.

## Source-Of-Truth Rules

- Canonical Markdown controls required behavior. Existing Go code is implementation evidence, not source of truth.
- Work only on tasks marked `ready` or explicitly assigned by the user.
- Do not invent durable behavior. If the task lacks required behavior, update the relevant spec/task or mark the task blocked.
- Promote durable changes to canonical docs, not just code comments or diary entries.
- After material task implementation, update task frontmatter, the feature `PLAN.md`, and append to the feature `DIARY.md` when AGENTS/BOOKKEEPING require it.

## Worker-Specific Checkpoints

Apply the relevant checkpoints from `docs/conventions/WORKER.md` rather than copying its full text:

- Composition root: `pkg/app.NewApp` owns Worker wiring. Use explicit constructor injection. Test new `App` fields or service creation logic in `pkg/app/app_test.go`.
- CLI commands: `internal/app/program.go` owns `urfave/cli/v3` registration and typed flag extraction. Command action functions live in command-named files and must not accept or reference CLI types. Production behavior changed through a CLI command requires executable-surface tests through command registration.
- Configuration: use `pkg/app/config` and configuro with the `ARCHIVIST_` prefix and canonical nested shape. Production Worker code outside config must not read environment variables directly.
- Provider boundaries: orchestration depends on Archivist-owned interfaces. Pipeline code must not import or expose external provider SDK request/response types.
- HTTP: outbound Worker HTTP goes through an injected `*req.Client`. Do not use `req.C()`. Direct outbound `net/http` is forbidden except for documented SDK bridge cases; tests may use `httptest`.
- Logging: pipeline orchestration owns structured logs for article-processing stages. Provider adapters return data and errors; they do not emit info/error logs. Secrets and full article/summary content must never be logged.
- Errors: package error infrastructure belongs in `errors.go`. ARC sentinels and public messages belong in `internal/arc`; persistence must use ARC public messages, not raw diagnostic errors.
- Filesystem/artifacts: use traversal-resistant APIs when operating under `DATA_DIR`, especially article artifact paths. Artifact writes under `/data` must be atomic.
- Configuration changes: update `docs/conventions/GENERAL.md`, `docs/ARCHITECTURE.md`, and affected feature specs/tasks when adding Worker configuration keys.

## Validation

Run Worker validation from `src/worker/` unless the task or ExecPlan specifies a narrower or broader command set:

```bash
go tool lefthook run build
go tool lefthook run format
go tool lefthook run lint
go tool lefthook run test
```

Before marking a task done, record validation in the task and diary according to repository bookkeeping rules. If validation cannot run, record the exact reason in both places.

## Output

When reporting Worker work, include:

- task ID when applicable;
- loaded context summary;
- changed Worker areas;
- canonical docs updated or why none were needed;
- validation commands and results;
- blockers or follow-ups.
