# AGENTS.md

## Purpose

This file defines how AI agents should work in this repository.

The repository uses a lightweight Markdown-based ALM system. Code is treated as a generated artifact. Specifications, plans, standards, and task files are the durable source of truth.

Agents must follow this file together with `docs/BOOKKEEPING.md` and `docs/REBUILD.md`.

---

## Source of Truth

The canonical rebuild artifacts are defined in `docs/REBUILD.md`. That file is the authoritative list.

---

## Repo-Local Multi-Agent Tooling

Significant parallel feature work may use `.agents/skills/archivist-agentic-feature/SKILL.md`.

Repo-local role profiles live under `.agents/agents/`. Workflow handoff templates live under `.agents/workflows/templates/`.

These files are non-canonical tooling. Durable behavior must still be captured in canonical docs, feature specs, feature plans, or task files. Active-run ExecPlans may guide implementation only when a canonical task has `requires_exec_plan: true`, but they do not add requirements beyond current specs and tasks.

---

## Context Loading Policy

Start each request by reading only the orientation bundle:

1. `AGENTS.md`
2. `docs/REBUILD.md`
3. `docs/specs/INDEX.md`

Then classify the request before loading more context:

- planning
- implementation
- rebuild
- standards change
- review
- question-answering

Identify the affected feature, module or modules, and durable surfaces before loading more files. Durable surfaces include architecture, design decisions, schema, API contracts, storage, authentication, deployment, artifact contracts, user-visible error contracts, and ALM process rules.

Load additional context by trigger:

- `docs/BOOKKEEPING.md`: creating or updating specs, plans, tasks, or ExecPlans; resolving dependency or concurrency questions; applying ExecPlan rules.
- `docs/ARCHITECTURE.md`: changing or depending on executables, service boundaries, storage, runtime topology, integrations, authentication boundaries, or deployment assumptions.
- `docs/DESIGN.md`: relying on or changing durable cross-task decisions, decision rationale, or rebuild-relevant tradeoffs.
- `docs/ERRORS.md`: changing persisted public ARC error codes, public messages, or error classification contracts.
- `docs/ARTIFACTS.md`: changing artifact paths, filenames, access, write, or delete contracts.
- `.agents/skills/archivist-general/SKILL.md` and the relevant module skill: implementation practice, validation guidance, or module-specific development workflow.
- Feature `SPEC.md`, `PLAN.md`, task files, and ExecPlans: work on that feature or task.
- Related feature specs: only when listed in `docs/specs/INDEX.md`, the feature `SPEC.md`, the feature `PLAN.md`, task dependencies, or task required context.

Do not load unrelated feature folders or the full `docs/` tree unless the request is explicitly a rebuild, audit, or cross-cutting standards change.

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

Use implementation mode only for tasks explicitly assigned by the user or selected by non-canonical run state initialized from the canonical task DAG.

In implementation mode, follow the task file and any active-run ExecPlan required by `requires_exec_plan: true`. Do not invent new durable behavior. If required information is missing, update the relevant specification or record the run blocker outside the canonical rebuild artifact set.

### Rebuild Mode

Use rebuild mode when recreating the application from canonical documents.

In rebuild mode, ignore previous implementation choices unless they are captured in canonical files.

---

## Task Selection Rules

Agents may work on a task only when:

1. the user explicitly assigns it or non-canonical run state selects it;
2. all dependencies in `depends_on` are satisfied in the current run state, or the user explicitly overrides this;
3. the task has clear acceptance criteria or a clear `Done When` section;
4. required context files are available.

Do not treat canonical task files as execution state. Current run state must live outside the canonical rebuild artifact set.

---

## Concurrency Rules

See `docs/BOOKKEEPING.md` § Dependency and Concurrency Rules for the full conditions. If uncertain, sequence the tasks.

---

## Missing Information Rule

Do not silently invent durable behavior.

When required information is missing:

1. add an open question to the relevant `SPEC.md` or task;
2. record the blocker in non-canonical run state if implementation cannot continue safely;
3. propose a minimal decision only when it is local and reversible;
4. promote architectural, interface, storage, security, or rebuild-relevant decisions to canonical docs.

---

## Document Update Rules

When implementing a task, update all relevant bookkeeping artifacts:

1. `SPEC.md`, `PLAN.md`, task files, `ARCHITECTURE.md`, `DESIGN.md`, `docs/ERRORS.md`, or `docs/ARTIFACTS.md` if durable decisions changed;
2. `docs/specs/INDEX.md` if feature dependencies changed.

Update `.agents/skills` only when non-canonical development guidance itself changes. Do not use skills as the only record of rebuild-critical behavior.

---

## ExecPlan Rules

Use an ExecPlan when a task is complex, risky, cross-cutting, or modifies architecture, storage, authentication, public APIs, schema, worker behavior, deployment behavior, or shared conventions. Canonical tasks record this with `requires_exec_plan: true`; simple tasks set `requires_exec_plan: false`.

ExecPlans are active-run planning artifacts, not rebuild history. They may guide implementation only for the current run, and no ExecPlan may add requirements beyond the current feature spec, plan, and task file.

See `docs/BOOKKEEPING.md` § Linking ExecPlans to Tasks for the bidirectional frontmatter format.

---

## Validation Rules

Before marking a task complete in non-canonical run state, run the validation specified in the task or active-run ExecPlan.

If validation cannot be run, record why in the task and any affected plan or canonical document.

A task should not be marked complete in run state unless its acceptance criteria and `Done When` section are satisfied.

---

## Prohibited Agent Behavior

Agents must not:

- treat existing code as the source of truth;
- create behavior not present in canonical docs without updating the docs;
- ignore task dependencies;
- change public interfaces without updating specs and dependent tasks;
- mark tasks complete in run state without validation evidence;
- create ExecPlans that are not tied to an active run and a canonical task;
- hide decisions in implementation comments only;
- make broad cross-feature changes without reading affected feature specs.

---

## Default Response Pattern

When asked to plan a feature:

1. identify or create the feature folder;
2. draft or update `SPEC.md`;
3. decompose into task files;
4. update `PLAN.md` with DAG and task table;
5. identify tasks needing ExecPlans;
6. update `INDEX.md` if needed.

When asked to implement a task:

1. state the task ID;
2. summarize loaded context;
3. implement only the scoped task;
4. run validation;
5. update non-canonical run state if one is in use;
6. promote durable decisions to canonical docs.
