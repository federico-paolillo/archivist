# AGENTS.md

## Purpose

This file defines how AI agents should work in this repository.

The repository uses a lightweight Markdown-based ALM system. Code is treated as a generated artifact. Specifications, plans, standards, and task files are the durable source of truth.

Agents must follow this file together with `docs/BOOKKEEPING.md` and `docs/REBUILD.md`.

---

## Source of Truth

The canonical rebuild artifacts and historical artifacts are defined in `docs/REBUILD.md`. That file is the authoritative list.

Historical logs may inform implementation, but they do not define required behavior unless the relevant decision has been promoted to a canonical file.

---

## Required Reading Order

Before planning or implementing, read:

1. `docs/BOOKKEEPING.md`
2. `docs/REBUILD.md`
3. `docs/ARCHITECTURE.md`
4. `docs/CONVENTIONS.md`
5. `docs/DESIGN.md`
6. `docs/specs/INDEX.md`

Before working on a feature, also read:

1. `docs/specs/<feature>/SPEC.md`
2. `docs/specs/<feature>/PLAN.md`
3. related feature specs listed in `SPEC.md` or `PLAN.md`

Before implementing a task, also read:

1. `docs/specs/<feature>/tasks/<task>.md`
2. the linked ExecPlan, if `exec_plan` is not null

---

## Execution Modes

### Planning Mode

Use planning mode when creating or revising:

- feature specs
- task breakdowns
- feature plans
- task DAGs
- ExecPlans
- rebuild rules
- standards

In planning mode, do not modify production code unless explicitly requested.

### Implementation Mode

Use implementation mode only for tasks marked `ready` or explicitly assigned by the user.

In implementation mode, follow the task file and linked ExecPlan. Do not invent new durable behavior. If required information is missing, update the relevant specification or mark the task blocked.

### Rebuild Mode

Use rebuild mode when recreating the application from canonical documents.

In rebuild mode, ignore previous implementation choices unless they are captured in canonical files.

---

## Task Selection Rules

Agents may work on a task only when:

1. the task status is `ready`, or the user explicitly assigns it;
2. all dependencies in `depends_on` are `done`, or the user explicitly overrides this;
3. the task has clear acceptance criteria or a clear `Done When` section;
4. required context files are available.

Do not work on tasks marked `draft`, `blocked`, `done`, or `skipped` unless explicitly asked.

---

## Concurrency Rules

See `docs/BOOKKEEPING.md` § Dependency and Concurrency Rules for the full conditions. If uncertain, sequence the tasks.

---

## Missing Information Rule

Do not silently invent durable behavior.

When required information is missing:

1. add an open question to the relevant `SPEC.md`, task, or ExecPlan;
2. mark the task `blocked` if implementation cannot continue safely;
3. propose a minimal decision only when it is local and reversible;
4. promote architectural, interface, storage, security, or rebuild-relevant decisions to canonical docs.

---

## Document Update Rules

When implementing a task, update all relevant bookkeeping artifacts:

1. task frontmatter status;
2. task status row in `PLAN.md`;
3. `DIARY.md` with an append-only entry;
4. `SPEC.md`, `PLAN.md`, `CONVENTIONS.md`, `ARCHITECTURE.md`, or `DESIGN.md` if durable decisions changed;
5. `docs/specs/INDEX.md` if feature status or dependencies changed.

---

## ExecPlan Rules

Use an ExecPlan when a task is complex, risky, cross-cutting, or modifies architecture, storage, authentication, public APIs, schema, worker behavior, deployment behavior, or shared conventions. Simple tasks set `exec_plan: null`.

See `docs/BOOKKEEPING.md` § Linking ExecPlans to Tasks for the bidirectional frontmatter format.

---

## Implementation Diary Rules

After completing or materially changing a task, append an entry to the feature `DIARY.md`.

Each entry should include:

- date
- task ID
- status outcome
- summary of changes
- decisions made
- validation performed
- follow-ups
- canonical documents updated

Do not put raw chain-of-thought in `DIARY.md`. Record conclusions, rationale, decisions, and validation.

---

## Validation Rules

Before marking a task `done`, run the validation specified in the task or ExecPlan.

If validation cannot be run, record why in both the task and the diary entry.

A task should not be marked `done` unless its acceptance criteria and `Done When` section are satisfied.

---

## Prohibited Agent Behavior

Agents must not:

- treat existing code as the source of truth;
- create behavior not present in canonical docs without updating the docs;
- use `DIARY.md` as the only source for required behavior;
- ignore task dependencies;
- change public interfaces without updating specs and dependent tasks;
- mark tasks done without validation notes;
- create unlinked ExecPlans;
- hide decisions in implementation comments only;
- make broad cross-feature changes without reading affected feature specs.

---

## Default Response Pattern

When asked to plan a feature:

1. identify or create the feature folder;
2. draft or update `SPEC.md`;
3. decompose into task files;
4. update `PLAN.md` with DAG and status table;
5. identify tasks needing ExecPlans;
6. update `INDEX.md` if needed.

When asked to implement a task:

1. state the task ID;
2. summarize loaded context;
3. implement only the scoped task;
4. run validation;
5. update task status, `PLAN.md`, and `DIARY.md`;
6. promote durable decisions to canonical docs.
