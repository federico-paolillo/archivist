# REBUILD.md

## Purpose

Defines the full-regeneration contract. It states which files are canonical, which files are historical, and how an agent should rebuild the system from scratch.

This file is mandatory if the project is intended to be regenerated multiple times.

---

## Canonical Rebuild Artifacts

The following files are authoritative for rebuild:

```text
AGENTS.md
docs/BOOKKEEPING.md
docs/REBUILD.md
docs/ARCHITECTURE.md
docs/CONVENTIONS.md
docs/DESIGN.md
docs/specs/INDEX.md
docs/specs/*/SPEC.md
docs/specs/*/PLAN.md
docs/specs/*/tasks/*.md
docs/specs/*/plans/*.execplan.md
```

---

## Historical Artifacts

The following files are historical by default:

```text
docs/specs/*/DIARY.md
```

Historical artifacts may explain prior implementation choices, but they must not be the only place where required behavior is defined.

If a diary entry contains a durable decision, promote it to a canonical document before relying on it for rebuild.

---

## Rebuild Rule

For a full rebuild:

1. Start from canonical Markdown artifacts.
2. Ignore existing implementation unless explicitly referenced by canonical documents.
3. Recreate the application according to feature specs, feature plans, task files, standards, and design decisions.
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
5. `docs/CONVENTIONS.md`
6. `docs/DESIGN.md`
7. `docs/specs/INDEX.md`
8. feature `SPEC.md` files in dependency order
9. feature `PLAN.md` files in dependency order
10. task files in executable order
11. linked ExecPlans when present

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
TODO: define feature execution order here.
```

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

1. all required feature tasks are `done` or explicitly `skipped`;
2. all acceptance criteria are satisfied;
3. the validation suite passes;
4. deployment or runtime smoke tests pass, if applicable;
5. feature diaries record implementation outcomes;
6. durable decisions discovered during the rebuild have been promoted to canonical documents.

Project validation commands:

```bash
# TODO: add validation commands
```

---

## Rebuild Completion Criteria

A rebuild is complete when:

- the application can be built from scratch;
- all required executables run;
- configured tests pass;
- feature acceptance criteria are satisfied;
- no required behavior exists only in code or diary entries;
- `docs/specs/INDEX.md` reflects final feature status.
