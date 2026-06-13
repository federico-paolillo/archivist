# REBUILD.md

## Purpose

Defines the full-regeneration contract. It states which files are canonical and how an agent should rebuild the system from scratch.

This file is mandatory if the project is intended to be regenerated multiple times.

---

## Canonical Rebuild Artifacts

The following files are authoritative for rebuild:

```text
AGENTS.md
docs/BOOKKEEPING.md
docs/REBUILD.md
docs/ARCHITECTURE.md
docs/ERRORS.md
docs/ARTIFACTS.md
docs/DESIGN.md
docs/design/DESIGN.md
docs/design/login.png
docs/design/view.png
docs/specs/INDEX.md
docs/specs/*/SPEC.md
docs/specs/*/PLAN.md
docs/specs/*/tasks/*.md
docs/specs/*/plans/*.execplan.md
```

---

## Non-Canonical Scaffolding

The following files are reusable scaffolding templates, not rebuild artifacts:

```text
docs/templates/*.md
```

Use templates to create canonical feature artifacts, but do not treat unresolved template placeholders as rebuild requirements.

---

## Rebuild Rule

For a full rebuild:

1. Start from canonical Markdown artifacts.
2. Ignore existing implementation unless explicitly referenced by canonical documents.
3. Recreate the application according to feature specs, feature plans, task files, architecture, shared contracts, and design decisions.
4. Do not infer behavior from old code.
5. Do not add behavior that is not specified.
6. If required behavior is missing, add or update the relevant spec/task before implementing it.

---

## Rebuild Reading Order

Read documents in this order:

1. `AGENTS.md`
2. `docs/BOOKKEEPING.md`
3. `docs/REBUILD.md`
4. `docs/ARCHITECTURE.md`
5. `docs/ERRORS.md`
6. `docs/ARTIFACTS.md`
7. `docs/DESIGN.md`
8. `docs/design/DESIGN.md` and referenced design assets when rebuilding UI behavior
9. `docs/specs/INDEX.md`
10. feature `SPEC.md` files in dependency order
11. feature `PLAN.md` files in dependency order
12. task files in executable order
13. linked ExecPlans when present

Repo-local skills under `.agents/skills/` are development guidance, not rebuild artifacts. They may help an implementation agent work idiomatically, but required rebuild behavior must be present in the canonical files above.

---

## Feature Execution Order

Feature execution order is defined by `docs/specs/INDEX.md` and each feature `PLAN.md`.

Default rules:

1. global foundations before feature implementation;
2. schema and interface tasks before dependent implementation tasks;
3. shared packages before executables that consume them;
4. back-end capabilities before UI that depends on them;
5. validation and integration tasks after dependent implementation tasks.

Project-specific ordering:

```text
1. telegram-ingestion
2. authn
3. article-processing
4. markdown-extraction
5. summary-generation
6. ui-endpoints
7. ui
8. snapshotter
9. otel-observability
```

Task-level cross-feature dependencies in feature `PLAN.md` files further constrain this order. In particular, the shared persistence foundation from `TELING-001` must precede auth password persistence, article processing orchestration, and UI endpoint work; final success notification behavior is completed by `SUMGEN-005`; and the browser UI starts only after auth and UI article endpoint contracts are implemented and validated. Worker runtime configuration, stale-job force deletion, and user-id resolution are current contracts inside the owning features, not standalone rebuild phases.

---

## Task Execution Rule

Agents may execute only tasks marked `ready`, unless explicitly assigned by the user.

Before executing a task, read:

```text
docs/specs/<feature>/SPEC.md
docs/specs/<feature>/PLAN.md
docs/specs/<feature>/tasks/<task>.md
```

If the task has an ExecPlan, also read:

```text
docs/specs/<feature>/plans/<task>.execplan.md
```

---

## Missing Information

If an agent cannot implement a task because required information is missing:

1. add an open question to the relevant task or feature spec;
2. mark the task `blocked` if necessary;
3. update `PLAN.md`;
4. do not invent durable behavior in code.

---

## Validation Requirements

A rebuild is not complete until:

1. all required feature tasks are `done`;
2. all acceptance criteria are satisfied;
3. the validation suite passes;
4. deployment or runtime smoke tests pass, if applicable;
5. durable decisions discovered during the rebuild have been promoted to canonical documents.

Executable and service-boundary rebuild work must include validation through the deployed entrypoint shape, such as a CLI command, hosted-service loop, or HTTP route. Tests that only instantiate internal services do not satisfy executable-boundary acceptance criteria.

Project validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
cd src/ui && npm run format
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
cd src/snapshotter && uv sync --locked --all-extras --dev
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
docker buildx build --file snapshotter.Dockerfile --platform linux/amd64 --load --tag archivist-snapshotter:test .
docker compose --env-file .env.local.example -f docker-compose.yaml -f docker-compose.local.yaml config --quiet
ARCHIVIST_GATEWAY_IMAGE=gateway ARCHIVIST_WORKER_IMAGE=worker ARCHIVIST_UI_IMAGE=ui ARCHIVIST_SNAPSHOTTER_IMAGE=snapshotter docker compose --env-file .env.example -f docker-compose.yaml -f docker-compose.prod.yaml config --quiet
```

---

## Rebuild Completion Criteria

A rebuild is complete when:

- the application can be built from scratch;
- all required executables run;
- configured tests pass;
- feature acceptance criteria are satisfied;
- no required behavior exists only in code;
- `docs/specs/INDEX.md` reflects final feature status.
