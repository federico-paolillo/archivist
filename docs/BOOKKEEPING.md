# BOOKKEEPING.md

## Purpose

This repository uses a lightweight Markdown-based ALM system. The goal is to make the specification and planning artifacts the durable source of truth, while treating source code as a rebuildable side effect.

The system is designed for humans and AI agents working together on a monorepo. It supports feature-level specifications, task decomposition, dependency-aware execution, per-task implementation plans, and a historical implementation diary.

The primary objective is repeatable full regeneration of the application from repository documents.

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
      DIARY.md
      EXECPLAN.md
      PLAN.md
      SPEC.md
      TASK.md
    specs/
      INDEX.md
      <feature-slug>/
        SPEC.md
        PLAN.md
        DIARY.md
        tasks/
          <task-id>-<task-slug>.md
        plans/
          <task-id>-<task-slug>.execplan.md
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

Explains the ALM system itself: structure, rules, workflow, canonical status, and how agents should interpret the files.

This file is normative for process, not for product behavior.

### `AGENTS.md`

Defines how AI agents should work in the repository. It should tell agents what to read, how to select tasks, how to update documents, how to handle missing information, and what not to do.

This file is an execution contract.

### `docs/REBUILD.md`

Defines the full-regeneration contract. It states which files are canonical, which files are historical, and how an agent should rebuild the system from scratch.

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

Global feature index and navigation map. It should list all features, status, dependencies, and links to their specs and plans.

This file prevents scattered feature folders from becoming unmanageable.

### `docs/specs/<feature-slug>/SPEC.md`

Feature-level specification. It defines motivation, intent, scope, out-of-scope items, requirements, interfaces, data implications, acceptance criteria, dependencies, and rebuild notes.

This is the main canonical file for a feature.

### `docs/specs/<feature-slug>/PLAN.md`

Feature execution plan. It contains the task DAG, task board, dependency rules, concurrency rules, execution order, and readiness state.

This is the main control board for implementation.

### `docs/specs/<feature-slug>/tasks/<task>.md`

A task is an executable unit of work. It is usually lighter than a full user story but may contain story-like framing and Gherkin acceptance criteria.

Each task must have a stable ID and frontmatter metadata.

### `docs/specs/<feature-slug>/plans/<task>.execplan.md`

An ExecPlan is a detailed implementation plan for a complex task. Not every task needs one.

ExecPlans are linked bidirectionally: the task points to its ExecPlan, and the ExecPlan points back to the task.

ExecPlan status values are:

```text
proposed     drafted but not yet authoritative for execution
accepted     approved and ready to guide implementation
in_progress  currently being executed
completed    task completed using this plan
superseded   retained for history but replaced by another plan or task update
```

ExecPlans are canonical planning records even when `status: proposed`, but proposed ExecPlans are not authoritative execution guidance. A task marked `ready` that links an ExecPlan must link an `accepted` ExecPlan; during execution the ExecPlan may move to `in_progress`, and after completion it may move to `completed`.

### `docs/specs/<feature-slug>/DIARY.md`

Append-only implementation log for the feature. It records what changed, why, validation performed, decisions discovered, and follow-ups.

`DIARY.md` is historical by default. It is not canonical for rebuild unless a decision is promoted into `SPEC.md`, `PLAN.md`, `docs/DESIGN.md`, or another canonical document.

### `.agents/skills/<skill-name>/SKILL.md`

A skill file defines a reusable agent workflow for a specific kind of task. Skills are not executed automatically; they must be explicitly invoked.

The planner agent skill (`planner-agent/SKILL.md`) defines how to turn a feature idea into the full set of ALM artifacts: SPEC, PLAN, DIARY, tasks, and optional ExecPlans.

Skill files are not canonical rebuild artifacts. They are tooling for the humans and agents working on the project.

### `docs/templates/*.md`

Templates are non-canonical scaffolding files. They are reusable starting points for generated ALM artifacts, not rebuild inputs.

Generated feature files become canonical only when they are created under a canonical path listed in `docs/REBUILD.md` and all template placeholders have been resolved.

---

## Canonical vs Historical Artifacts

### Canonical

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
docs/specs/*/plans/*.execplan.md
```

### Historical

These files describe what happened during previous implementation runs:

```text
docs/specs/*/DIARY.md
```

Historical files may contain useful rationale, but they must not be the only place where durable behavior or constraints are documented.

---

## Feature Lifecycle

### 1. Feature Creation

A human and/or planner agent creates a new feature folder:

```text
docs/specs/<feature-slug>/
  SPEC.md
  PLAN.md
  DIARY.md
  tasks/
  plans/
```

The feature starts with a coarse-grained `SPEC.md`.

### 2. Feature Specification

`SPEC.md` should define:

- motivation
- intent
- scope
- out of scope
- requirements
- acceptance criteria
- interfaces
- data implications
- dependencies
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
- status
- dependencies
- whether tasks may run concurrently
- blocking relationships
- required ExecPlans
- validation sequence

### 5. ExecPlan Creation

Create an ExecPlan only when a task is complex, risky, cross-cutting, or touches architectural boundaries.

Simple tasks can be executed directly from the task file.

### 6. Implementation

Agents execute tasks marked `ready`. They must load the required context bundle, perform the task, update status, run validation, and append a diary entry.

### 7. Promotion of Decisions

If implementation discovers a durable decision, it must be promoted from `DIARY.md` into one of:

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

## Task Status Vocabulary

Use a small fixed vocabulary.

```text
draft        task is incomplete or not yet actionable
ready        task is actionable and dependencies are satisfied
blocked      task is valid but waiting on dependencies or decisions
in_progress  task is currently being executed
done         task is complete and validated
skipped      task is intentionally not implemented
```

Avoid custom statuses unless `docs/BOOKKEEPING.md` is updated.

---

## Task Metadata

Each task should use frontmatter:

```yaml
---
id: AUTHN-001
feature: authn
title: Implement login
status: ready
depends_on: []
blocks: [AUTHN-003]
parallel: true
exec_plan: ../plans/AUTHN-001-implement-login.execplan.md
canonical: true
---
```

### Frontmatter Field Schema

| Field | Required | Values | Notes |
|---|---|---|---|
| `id` | yes | `FEATURE-NNN` string | Stable. Do not renumber. |
| `feature` | yes | kebab-case slug | Must match the feature folder name. |
| `title` | yes | free text | Short task title. |
| `status` | yes | see Task Status Vocabulary | One of: `draft`, `ready`, `blocked`, `in_progress`, `done`, `skipped`. |
| `depends_on` | yes | list of task IDs or `[]` | Tasks that must be `done` before this task is `ready`. |
| `blocks` | yes | list of task IDs or `[]` | Tasks that cannot start until this task is `done`. |
| `parallel` | yes | `true` \| `false` | Whether this task may run concurrently with other parallel-safe tasks. |
| `exec_plan` | yes | relative path or `null` | Link to the ExecPlan file, or `null` if none. |
| `canonical` | yes | `true` | Marks the file as a canonical rebuild artifact. Always `true` for task files. |

The same `canonical: true` field appears in SPEC, PLAN, and ExecPlan frontmatter and carries the same meaning: the file must be treated as authoritative for rebuild.

`id` must be stable. Filenames may change, but IDs should not.

---

## Linking ExecPlans to Tasks

A task links to its ExecPlan through frontmatter:

```yaml
exec_plan: ../plans/AUTHN-001-implement-login.execplan.md
```

The ExecPlan links back:

```yaml
task: ../tasks/AUTHN-001-implement-login.md
```

This bidirectional link is required for complex tasks.

If a task is `ready` and has an ExecPlan, the linked ExecPlan must be `accepted`. Do not execute a ready task from a `proposed` ExecPlan.

If a task does not need an ExecPlan, set:

```yaml
exec_plan: null
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

Task files and linked ExecPlans define the required execution context.

The default implementation context is:

```text
.agents/skills/archivist-general/SKILL.md
.agents/skills/<relevant-skill>/SKILL.md
docs/specs/<feature>/SPEC.md
docs/specs/<feature>/PLAN.md
docs/specs/<feature>/tasks/<task>.md
```

Read only the skills relevant to the task unless the task spans modules.

If the task has an ExecPlan, also read:

```text
docs/specs/<feature>/plans/<task>.execplan.md
```

Load global docs by trigger, not by default:

- `docs/BOOKKEEPING.md`: ALM artifact creation or update, task status changes, dependency or concurrency questions, diary rules, ExecPlan rules.
- `docs/ARCHITECTURE.md`: executables, service boundaries, storage, runtime topology, integrations, authentication boundaries, deployment assumptions.
- `docs/DESIGN.md`: durable cross-task decisions, decision changes, rebuild-relevant rationale.
- `docs/ERRORS.md`: persisted public ARC error-code changes.
- `docs/ARTIFACTS.md`: artifact path, filename, access, write, or delete contracts.
- Related feature specs: only when listed in `docs/specs/INDEX.md`, the feature `SPEC.md`, the feature `PLAN.md`, task dependencies, or task required context.

---

## Handling Missing Information

Agents must not silently invent durable behavior.

When required information is missing:

1. If the missing information blocks implementation, mark the task `blocked` and add an open question.
2. If a reasonable local decision is possible and reversible, record it in the task and diary.
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

The authoritative feature-level work board is `PLAN.md`.

Each task also has its own status in frontmatter. The task file is the local status record; `PLAN.md` is the feature summary.

When changing a task status, update both:

1. the task frontmatter
2. the task table in `PLAN.md`

If automation is later introduced, `PLAN.md` may be generated from task frontmatter.

---

## Implementation Diary Rules

`DIARY.md` is append-only.

Each entry should include:

- date
- task ID
- status outcome
- summary of changes
- decisions made
- validation performed
- follow-ups
- decisions promoted to canonical docs

Do not use `DIARY.md` as the backlog or source of requirements.

---

## Quality Gates

Before a feature is considered complete:

1. All required tasks are `done` or explicitly `skipped`.
2. Acceptance criteria in `SPEC.md` and task files are satisfied.
3. Validation steps have run successfully.
4. Durable decisions from `DIARY.md` have been promoted.
5. `PLAN.md` reflects final task status.
6. `docs/specs/INDEX.md` reflects feature status.
7. Rebuild implications are captured in `docs/REBUILD.md` or feature rebuild notes.

Tasks that modify executable or service boundaries must validate the public executable or service entrypoint. Internal service tests are required but not sufficient for behavior that must be reachable from a deployed binary, hosted service, CLI command, or HTTP route.

---

## Anti-Patterns

Avoid these:

- keeping requirements only in code
- keeping architecture only in comments
- using `DIARY.md` as canonical requirements
- creating tasks without stable IDs
- creating ExecPlans without linking them to tasks
- running agents without bounded context
- loading the entire `docs/` tree without a declared trigger
- letting agents infer cross-feature dependencies
- decomposing features into one-file mechanical tasks with no validation boundary
- duplicating contradictory rules across docs
- allowing `docs/` and `specs/` to disagree without resolution

---

## Minimal Viable Use

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
docs/specs/<feature>/DIARY.md
```

ExecPlans and design decisions can be added when complexity requires them.
