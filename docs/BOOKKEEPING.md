# BOOKKEEPING.md

## Purpose

This repository uses a lightweight Markdown-based ALM system. The goal is to make the specification and planning artifacts the durable source of truth, while treating source code as a rebuildable side effect.

The system is designed for humans and AI agents working together on a monorepo. It supports feature-level specifications, task decomposition, dependency-aware execution, and per-task implementation plans.

The primary objective is repeatable full regeneration of the application from repository documents.

ALM documents describe current intended repository behavior. Work that is not currently intended must be removed from feature plans or captured only as an open question, assumption, constraint, or risk that affects current behavior.

---

## Core Principle

The codebase is not the canonical artifact.

Canonical rebuild knowledge must live in `AGENTS.md` or Markdown files under `docs/` and `docs/specs/`. If a behavior, constraint, architectural decision, interface, or acceptance criterion must survive a rebuild, it must be written in one of the canonical files.

Implementation may discover missing decisions. When that happens, the decision must be promoted back into the appropriate document before or alongside code changes.

---

## Recommended Repository Structure

```text
/
  AGENTS.md

  docs/
    BOOKKEEPING.md
    REBUILD.md
    ARCHITECTURE.md
    ERRORS.md
    ARTIFACTS.md
    DESIGN.md
    templates/
      EXECPLAN.md
      PLAN.md
      SPEC.md
      TASK.md
    specs/
      INDEX.md
      <feature-slug>/
        SPEC.md
        PLAN.md
        tasks/
          <task-id>-<task-slug>.md
  .agents/
    skills/
      archivist-general/
        SKILL.md
      archivist-gateway/
        SKILL.md
      archivist-ui/
        SKILL.md
      archivist-worker/
        SKILL.md
      planner-agent/
        SKILL.md
```

The exact names of architecture and design documents may vary, but their authority must be explicit in `docs/REBUILD.md`.

---

## File Roles

### `docs/BOOKKEEPING.md`

Explains the ALM system itself: structure, rules, workflow, canonical rebuild contracts, and how agents should interpret the files.

This file is normative for process, not for product behavior.

### `AGENTS.md`

Defines how AI agents should work in the repository. It should tell agents what to read, how to select tasks, how to update documents, how to handle missing information, and what not to do.

This file is an execution contract.

### `docs/REBUILD.md`

Defines the full-regeneration contract. It states which files are canonical and how an agent should rebuild the system from scratch.

This file is mandatory if the project is intended to be regenerated multiple times.

### `docs/ARCHITECTURE.md`

Describes global system architecture: executables, services, boundaries, data ownership, storage, runtime topology, and integration patterns.

Architecture decisions that constrain all features belong here or in `docs/DESIGN.md`.

### `docs/ERRORS.md`

Defines shared persisted public error-code contracts. `docs/ERRORS.md` is canonical because ARC codes and public messages are shared by Worker, Gateway, UI, Telegram notification behavior, persisted state, and rebuild documentation.

### `docs/ARTIFACTS.md`

Defines deterministic filesystem artifact paths, artifact access boundaries, and artifact write/delete contracts. `docs/ARTIFACTS.md` is canonical because artifact layout is shared by Worker, Gateway, UI, backups, storage behavior, and rebuild documentation.

### `.agents/skills`

Defines repo-local development guidance: coding practice, testing guidance, naming style, module workflow, review posture, and validation commands.

Skills are not canonical rebuild artifacts. They may restate or point to canonical contracts, but a requirement that must survive rebuild belongs in `docs/`, `docs/specs/`, or root `AGENTS.md`.

### `docs/DESIGN.md`

Records durable decisions that must survive rebuilds. This can be ADR-like but does not need heavy ceremony.

A decision discovered during implementation should be promoted here if it affects more than one task or must remain true across rebuilds.

### `docs/specs/INDEX.md`

Global feature index and navigation map. It should list all features, dependencies, and links to their specs and plans.

This file prevents scattered feature folders from becoming unmanageable.

### `docs/specs/<feature-slug>/SPEC.md`

Feature-level specification. It defines motivation, intent, current intended behavior, current negative requirements, requirements, constraints, assumptions, risks, interfaces, data implications, acceptance criteria, dependencies, validation, rebuild notes, and open questions.

This is the main canonical file for a feature.

### `docs/specs/<feature-slug>/PLAN.md`

Feature execution plan. It contains the task DAG, task board, dependency rules, concurrency rules, execution order, validation sequence, and ExecPlan requirements.

This is the main control board for implementation.

### `docs/specs/<feature-slug>/tasks/<task>.md`

A task is an executable unit of work. It is usually lighter than a full user story but may contain story-like framing and Gherkin acceptance criteria.

Each task must have a stable ID and frontmatter metadata.

### ExecPlans

An ExecPlan is a detailed active-run implementation plan for a complex task. Not every task needs one.

ExecPlans are not canonical rebuild artifacts and are not retained as rebuild history by default. A canonical task records only whether it requires an ExecPlan through `requires_exec_plan: true` or `false`.

An ExecPlan may guide implementation only for the active run. It must not add requirements beyond the current feature `SPEC.md`, feature `PLAN.md`, and linked task file. If planning exposes missing behavior, update the spec, plan, or task before using the ExecPlan.

### `.agents/skills/<skill-name>/SKILL.md`

A skill file defines a reusable agent workflow for a specific kind of task. Skills are not executed automatically; they must be explicitly invoked.

The planner agent skill (`planner-agent/SKILL.md`) defines how to turn a feature idea into the full set of ALM artifacts: SPEC, PLAN, tasks, and optional ExecPlans.

Skill files are not canonical rebuild artifacts. They are tooling for the humans and agents working on the project.

### `docs/templates/*.md`

Templates are non-canonical scaffolding files. They are reusable starting points for generated ALM artifacts, not rebuild inputs.

Generated feature files become canonical only when they are created under a canonical path listed in `docs/REBUILD.md` and all template placeholders have been resolved.

---

## Canonical Artifacts

The authoritative canonical artifact list is in `docs/REBUILD.md`. The files below are the current set for reference:

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
```

---

## Feature Lifecycle

### 1. Feature Creation

A human and/or planner agent creates a new feature folder:

```text
docs/specs/<feature-slug>/
  SPEC.md
  PLAN.md
  tasks/
```

The feature starts with a coarse-grained `SPEC.md`.

### 2. Feature Specification

`SPEC.md` should define:

- motivation
- intent
- current intended behavior
- current negative requirements
- requirements
- constraints
- assumptions
- risks
- acceptance criteria
- interfaces
- data implications
- dependencies
- validation
- rebuild notes
- open questions

The spec should be human-readable and direct. It should not be a hidden prompt.

### 3. Task Breakdown

The feature is decomposed into task files under `tasks/`.

Tasks should be independently executable where possible. A task should be large enough to produce meaningful progress and small enough to validate.

Avoid decomposing into trivial mechanical steps unless the step has an independent validation boundary.

### 4. Feature Planning

`PLAN.md` defines the task DAG and execution board.

It should specify:

- task IDs
- dependencies
- whether tasks may run concurrently
- blocking relationships
- whether tasks require ExecPlans
- validation sequence

### 5. ExecPlan Creation

Create an ExecPlan only when a task is complex, risky, cross-cutting, or touches architectural boundaries.

Simple tasks can be executed directly from the task file.

### 6. Implementation

Agents execute tasks selected by explicit user assignment or by non-canonical run state initialized from the canonical DAG. They must load the required context bundle, perform the task, update non-canonical run state when one is in use, and run validation.

### 7. Promotion of Decisions

If implementation discovers a durable decision, it must be recorded in one of:

- the feature `SPEC.md`
- the feature `PLAN.md`
- the relevant task file
- `docs/DESIGN.md`
- `docs/ARCHITECTURE.md`
- `docs/ERRORS.md`
- `docs/ARTIFACTS.md`

If the discovery is only a reusable development practice and not rebuild-critical behavior, update the relevant `.agents/skills` file instead.

### 8. Rebuild

A rebuild starts from canonical documents, not from prior code. The rebuild order is defined in `docs/REBUILD.md`.

---

## Run State

Canonical rebuild artifacts do not store execution state.

External run tracking may use this vocabulary:

```text
pending      task exists but dependencies are not satisfied in this run
ready        task dependencies are satisfied in this run
in_progress  task is currently being executed in this run
blocked      task cannot continue in this run without a decision or external state
done         task acceptance criteria and validation requirements are satisfied in this run
```

This vocabulary belongs to non-canonical run state only. Do not write it into canonical feature, plan, task, or index metadata.

---

## Task Metadata

Each task should use frontmatter:

```yaml
---
id: TASK-001
feature: authn
title: Implement login
depends_on: []
blocks: [TASK-003]
parallel: true
requires_exec_plan: false
canonical: true
---
```

### Frontmatter Field Schema

| Field | Required | Values | Notes |
|---|---|---|---|
| `id` | yes | `FEATURE-NNN` string | Stable. Do not renumber. |
| `feature` | yes | kebab-case slug | Must match the feature folder name. |
| `title` | yes | free text | Short task title. |
| `depends_on` | yes | list of task IDs or `[]` | Tasks whose contracts must be satisfied before this task can execute. |
| `blocks` | yes | list of task IDs or `[]` | Tasks that require this task's contract first. |
| `parallel` | yes | `true` \| `false` | Whether this task may run concurrently with other parallel-safe tasks. |
| `requires_exec_plan` | yes | `true` \| `false` | Whether an active-run ExecPlan is required before execution. |
| `canonical` | yes | `true` | Marks the file as a canonical rebuild artifact. Always `true` for task files. |

The same `canonical: true` field appears in SPEC and PLAN frontmatter. ExecPlans are active-run artifacts and should use `canonical: false` if frontmatter is present.

`id` must be stable. Filenames may change, but IDs should not.

---

## ExecPlan Requirements

A task declares whether it requires an ExecPlan through frontmatter:

```yaml
requires_exec_plan: true
```

The active-run ExecPlan should link back to the task:

```yaml
task: ../tasks/TASK-001-implement-login.md
canonical: false
```

If a task does not need an ExecPlan, set:

```yaml
requires_exec_plan: false
```

---

## Dependency and Concurrency Rules

Tasks may run concurrently only when all of the following are true:

1. Neither task depends on the other.
2. They do not modify the same public interface, schema, migration, executable boundary, or shared package contract.
3. Their expected file areas do not significantly overlap.
4. Their validation steps can be run independently or in a clearly sequenced way.
5. `PLAN.md` does not explicitly prohibit concurrency.

Schema, API contract, shared package, and configuration changes are usually blocking tasks.

---

## Task-Scoped Context Loading

Task files define the required execution context. An active-run ExecPlan may add implementation context, but it does not add requirements beyond current specs and tasks.

The default implementation context is:

```text
.agents/skills/archivist-general/SKILL.md
.agents/skills/<relevant-skill>/SKILL.md
docs/specs/<feature>/SPEC.md
docs/specs/<feature>/PLAN.md
docs/specs/<feature>/tasks/<task>.md
```

Read only the skills relevant to the task unless the task spans modules.

If `requires_exec_plan: true`, create or read the active-run ExecPlan before implementation.

Load global docs by trigger, not by default:

- `docs/BOOKKEEPING.md`: ALM artifact creation or update, dependency or concurrency questions, ExecPlan rules.
- `docs/ARCHITECTURE.md`: executables, service boundaries, storage, runtime topology, integrations, authentication boundaries, deployment assumptions.
- `docs/DESIGN.md`: durable cross-task decisions, decision changes, rebuild-relevant rationale.
- `docs/ERRORS.md`: persisted public ARC error-code changes.
- `docs/ARTIFACTS.md`: artifact path, filename, access, write, or delete contracts.
- Related feature specs: only when listed in `docs/specs/INDEX.md`, the feature `SPEC.md`, the feature `PLAN.md`, task dependencies, or task required context.

---

## Handling Missing Information

Agents must not silently invent durable behavior.

When required information is missing:

1. If the missing information blocks implementation, record the blocker in non-canonical run state and add an open question to the relevant task or spec when it affects durable behavior.
2. If a reasonable local decision is possible and reversible, record it in the task.
3. If the decision affects architecture, interfaces, storage, security, data, or rebuild reproducibility, promote it to a canonical document.

---

## Acceptance Criteria

Use Gherkin-like acceptance criteria when useful, especially for user-visible behavior or observable system behavior.

Example:

```gherkin
Scenario: Successful login
  Given I have the correct secret
  When I submit the login form
  Then I receive an authenticated session cookie
  And I can access protected pages
```

Acceptance criteria should be testable. Avoid vague criteria such as “works well” or “is user-friendly” without concrete observable conditions.

---

## Work Tracking

`PLAN.md` is the authoritative feature-level rebuild board. It defines the DAG, task metadata, concurrency rules, and validation sequence. It is not a run-state board.

Run tracking must live outside the canonical rebuild artifact set. A future MCP task board or similar non-canonical mechanism may track `pending`, `ready`, `in_progress`, `blocked`, and `done` for a specific execution. Do not list run-state files in `docs/REBUILD.md`.

Run state is initialized by reading `docs/specs/INDEX.md`, each feature `PLAN.md`, and task frontmatter, constructing the dependency graph from `depends_on` and `blocks`, and computing current readiness from the run's satisfied tasks and dependency satisfaction.

---

## Quality Gates

Before a feature is considered complete:

1. All task acceptance criteria are satisfied.
2. Acceptance criteria in `SPEC.md` and task files are satisfied.
3. Validation steps have run successfully.
4. Rebuild implications are captured in `docs/REBUILD.md` or feature rebuild notes.

Tasks that modify executable or service boundaries must validate the public executable or service entrypoint. Internal service tests are required but not sufficient for behavior that must be reachable from a deployed binary, hosted service, CLI command, or HTTP route.

---

## Anti-Patterns

Avoid these:

- keeping requirements only in code
- keeping architecture only in comments
- creating tasks without stable IDs
- creating ExecPlans without linking them to tasks
- running agents without bounded context
- loading the entire `docs/` tree without a declared trigger
- letting agents infer cross-feature dependencies
- decomposing features into one-file mechanical tasks with no validation boundary
- duplicating contradictory rules across docs
- allowing `docs/` and `specs/` to disagree without resolution

---

## Minimum Useful Structure

For a small project, the minimum useful version is:

```text
AGENTS.md
docs/BOOKKEEPING.md
docs/REBUILD.md
docs/ARCHITECTURE.md
.agents/skills/archivist-general/SKILL.md
docs/ERRORS.md
docs/ARTIFACTS.md
docs/specs/INDEX.md
docs/specs/<feature>/SPEC.md
docs/specs/<feature>/PLAN.md
docs/specs/<feature>/tasks/*.md
```

ExecPlans and design decisions can be added when complexity requires them.
