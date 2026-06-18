---
feature: <feature-slug>
canonical: true
---
# Feature Plan: <Feature Title>

## Purpose

This file is the feature-level rebuild contract for task order, dependencies, concurrency rules, validation sequence, and validation requirements.

---

## Task DAG

```text
<TASK-001> -> <TASK-003>
<TASK-002> -> <TASK-003>
<TASK-004>
```

---

## Execution Phases

### Phase 1: Foundations

- `<TASK-001>` TODO
- `<TASK-002>` TODO

### Phase 2: Implementation

- `<TASK-003>` TODO

### Phase 3: Validation

- `<TASK-004>` TODO

---

## Tasks

| ID | Task | Depends On | Blocks | Parallel | Requires ExecPlan |
|---|---|---|---|---|---|
| `<TASK-001>` | TODO | - | - | yes | no |
| `<TASK-002>` | TODO | - | - | yes | no |
| `<TASK-003>` | TODO | `<TASK-001>`, `<TASK-002>` | - | no | yes |

---

## Concurrency Rules

- TODO: list tasks that may run concurrently.
- TODO: list tasks that must be sequenced.
- TODO: list shared files, schemas, or interfaces that force serialization.

---

## Blocking Interfaces or Schemas

List public contracts, schemas, migrations, shared packages, or executable boundaries that block dependent tasks.

- TODO

---

## Validation Sequence

1. TODO
2. TODO
3. TODO

Validation commands:

```bash
# TODO: add commands
```

---

## Open Planning Questions

- Q-001: TODO

---

## Completion Criteria

The feature is complete when:

- all task acceptance criteria are satisfied;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
