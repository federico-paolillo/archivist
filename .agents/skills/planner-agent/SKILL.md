---
name: planner-agent
description: Turn a feature idea or product note into the repository's Markdown ALM structure. Produces SPEC.md, PLAN.md, task files, optional ExecPlans, and updates to the feature INDEX.
when_to_use: Use when creating a feature specification, decomposing a feature into tasks, producing an implementation plan, identifying task dependencies, deciding on concurrency, creating ExecPlans, or preparing artifacts for full rebuilds.
allowed-tools: Read Write Edit Grep Glob Bash(ls *)
effort: high
---

# Planner Agent Skill

## Purpose

Use this skill to help a human turn a feature idea, product note, or rough plan into the repository's Markdown ALM structure.

The planner agent produces or updates:

- `docs/specs/<feature-slug>/SPEC.md`
- `docs/specs/<feature-slug>/tasks/*.md`
- `docs/specs/<feature-slug>/PLAN.md`
- optional `docs/specs/<feature-slug>/plans/*.execplan.md`
- optional updates to `docs/specs/INDEX.md`
- optional updates to `docs/ARCHITECTURE.md`, `docs/CONVENTIONS.md`, or `docs/DESIGN.md`

The planner does not implement code unless explicitly instructed.

---

## When To Use

Use this skill when the user asks to:

- create a feature specification;
- decompose a feature into tasks;
- produce a feature implementation plan;
- identify dependencies between tasks;
- decide which tasks can run concurrently;
- create ExecPlans for complex tasks;
- assess whether a feature is sufficiently specified for implementation;
- prepare artifacts for repeated full rebuilds.

---

## Required Context

Before planning, read:

```text
AGENTS.md
docs/BOOKKEEPING.md
docs/REBUILD.md
docs/ARCHITECTURE.md
docs/CONVENTIONS.md
docs/DESIGN.md
docs/specs/INDEX.md
```

If planning an existing feature, also read:

```text
docs/specs/<feature-slug>/SPEC.md
docs/specs/<feature-slug>/PLAN.md
docs/specs/<feature-slug>/tasks/*.md
docs/specs/<feature-slug>/DIARY.md
```

Read only related feature specs listed in `INDEX.md`, `SPEC.md`, or `PLAN.md`. Do not load unrelated feature folders by default.

---

## Inputs

The planner may receive any of the following:

- raw feature idea;
- product requirement;
- existing design note;
- existing `PLAN.md` or `SPEC.md`;
- user constraints;
- codebase context;
- cross-feature change request.

If the input is incomplete, proceed with best-effort planning and mark unknowns as open questions. Do not invent durable behavior silently.

---

## Output Principles

Outputs must be:

- human-readable;
- Markdown-native;
- stable under repeated rebuilds;
- explicit about dependencies;
- explicit about acceptance criteria;
- explicit about open questions;
- bounded enough for agents to execute without loading the entire repository.

Prefer small numbers of strong tasks over many trivial mechanical tasks.

---

## Planning Workflow

### 1. Determine Feature Slug

Choose a concise kebab-case feature slug.

Examples:

```text
authn
article-ingestion
article-processing
summary-generation
admin-ui
```

If the feature already exists, reuse its folder.

### 2. Classify Scope

Separate:

- in scope;
- out of scope;
- future work;
- open questions;
- global standards or conventions discovered during planning.

If the feature contains unrelated capabilities, propose splitting it into multiple feature folders.

### 3. Draft or Update `SPEC.md`

The feature spec must include:

- intent;
- motivation;
- scope;
- out of scope;
- requirements;
- acceptance criteria;
- data and interface implications;
- dependencies;
- rebuild notes;
- open questions.

Use Gherkin-like scenarios for observable behavior.

### 4. Identify Standards and Global Decisions

Before creating tasks, extract anything that should not be local to the feature.

Promote these to global docs when needed:

- architectural constraints → `docs/ARCHITECTURE.md`
- coding or testing conventions → `docs/CONVENTIONS.md`
- durable decisions → `docs/DESIGN.md`

Do not bury standards inside tasks.

### 5. Decompose Into Tasks

Create tasks that are independently executable and testable.

A good task has:

- one objective;
- clear inputs and outputs;
- acceptance criteria;
- `Done When` conditions;
- expected affected areas;
- dependencies;
- validation steps.

Reject task boundaries that are merely mechanical steps unless the step has an independent validation boundary.

### 6. Assign Stable Task IDs

Use feature-prefix IDs:

```text
AUTHN-001
AUTHN-002
INGEST-001
PROCESS-001
UI-001
```

Task IDs are stable. Do not renumber existing tasks unless the user explicitly requests cleanup.

### 7. Build the Task DAG

Determine dependencies between tasks.

Classify tasks as:

- independent;
- blocking;
- blocked;
- parallel-safe;
- requires sequencing.

Schema, public interfaces, shared packages, migrations, authentication, and configuration tasks usually block dependent implementation tasks.

### 8. Create `PLAN.md`

The plan must include:

- task DAG;
- task table;
- statuses;
- dependencies;
- concurrency rules;
- execution phases;
- validation sequence;
- tasks requiring ExecPlans.

### 9. Decide Whether ExecPlans Are Needed

Create an ExecPlan for a task when it is:

- complex;
- risky;
- cross-cutting;
- architectural;
- affects storage, auth, deployment, APIs, schemas, or shared contracts;
- likely to need stepwise implementation and rollback guidance.

Do not create ExecPlans for simple isolated tasks.

### 10. Update Index

Update `docs/specs/INDEX.md` when a feature is created, renamed, changes status, or changes dependencies.

---

## Task Quality Checklist

Each task must answer:

1. What is the objective?
2. What feature does it belong to?
3. What does it depend on?
4. What does it block?
5. What files or areas are likely affected?
6. What acceptance criteria must pass?
7. What validation is required?
8. Does it need an ExecPlan?
9. Can it run concurrently with other tasks?
10. What durable decisions, if any, must be promoted?

---

## Acceptance Criteria Guidance

Use Gherkin-like criteria for behavior:

```gherkin
Scenario: User submits a valid URL
  Given the user is authorized
  When the user sends a URL
  Then an article record is created
  And a processing job is queued
```

Use checklist criteria for non-behavioral tasks:

```md
## Acceptance Criteria

- Migration creates the required table.
- Required indexes exist.
- Existing tests pass.
- New repository tests cover insert and lookup behavior.
```

---

## ExecPlan Output Rules

Use `docs/templates/EXECPLAN.md` as the base. An ExecPlan must not contradict the task. If a contradiction is found, update the task first.

See `AGENTS.md` § ExecPlan Rules and `docs/BOOKKEEPING.md` § Linking ExecPlans to Tasks for the structure and frontmatter format.

---

## Handling Cross-Feature Changes

For cross-feature changes:

1. identify all affected feature specs;
2. update dependencies in `INDEX.md`;
3. add or update tasks in each affected feature;
4. avoid hiding cross-feature requirements in a single feature folder;
5. create an ExecPlan for coordination when shared contracts change.

---

## Output Format When Responding To The User

When proposing a breakdown, present:

1. feature slug;
2. spec summary;
3. task list with IDs;
4. task DAG;
5. tasks requiring ExecPlans;
6. open questions;
7. files to create or update.

When writing files directly, use the templates in `docs/templates/` and preserve the repository structure described in `docs/BOOKKEEPING.md`.

---

## Prohibited Planning Behavior

Do not:

- skip `SPEC.md` and go directly to tasks;
- put global architecture decisions only in task files;
- create tasks without stable IDs;
- create unlinked ExecPlans;
- use the implementation diary as a source of requirements;
- infer cross-feature dependencies without documenting them;
- generate many tiny tasks with no independent validation boundary;
- treat existing code as canonical unless the user explicitly asks for a brownfield reconciliation.
