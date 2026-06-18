---
name: archivist-general
description: Use for cross-module Archivist implementation practice, validation discipline, dependency restraint, configuration hygiene, naming guidance, security reminders, and non-contract coding rules.
---

# Archivist General

Use this skill for implementation, review, or planning work that spans more than one Archivist module or does not fit a narrower Gateway, Worker, or UI skill.

## Source Of Truth

- Canonical rebuild behavior lives in `AGENTS.md`, `docs/REBUILD.md`, `docs/ARCHITECTURE.md`, `docs/DESIGN.md`, `docs/ERRORS.md`, `docs/ARTIFACTS.md`, and `docs/specs/**`.
- Skills are implementation guidance only. If a decision affects runtime behavior, public interfaces, persistence, security, deployment, artifact layout, user-visible errors, validation requirements, or rebuild behavior, promote it to a canonical document.
- `docs/ERRORS.md` owns persisted public ARC error codes and messages.
- `docs/ARTIFACTS.md` owns deterministic artifact paths and artifact access contracts.

## Stack

- Gateway/API: ASP.NET Core Minimal API under `src/gateway/`.
- Worker: Go under `src/worker/`.
- UI: Preact with Vite under `src/ui/`.
- Storage: SQLite for metadata and queue state; filesystem artifacts under `DATA_DIR`.

## Implementation Practice

- Keep dependencies minimal. Add external dependencies only when they replace non-trivial custom implementation or are required by an accepted architecture/design decision.
- Use ULIDs when a new identifier is needed. Do not use GUIDs or delegate identifier generation to the database.
- Keep source layout and naming consistent with the target module skill.
- Treat configuration keys as canonical behavior. New configuration keys must be documented in `docs/ARCHITECTURE.md` and affected specs/tasks.
- Do not commit secrets. Treat API keys, auth bootstrap passwords, Telegram secrets, and provider credentials as secret material.
- Prefer structured logs for observable runtime behavior. Never log secrets or full article/summary content.
- Artifact writes under `DATA_DIR` must follow `docs/ARTIFACTS.md`.
- Persisted user-facing article-processing failures must follow `docs/ERRORS.md`.

## Validation

Run validation from the affected module unless the task or active-run ExecPlan gives a narrower or broader command set.

Gateway:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Worker:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

UI:

```bash
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
```

For documentation-only changes, run `git diff --check` and targeted stale-reference checks. Do not run formatters that rewrite code unless source code changed.

## Output

Report:

- task ID when applicable;
- canonical context loaded;
- affected modules;
- canonical documents updated or why none were needed;
- validation commands and results;
- blockers or follow-ups.
