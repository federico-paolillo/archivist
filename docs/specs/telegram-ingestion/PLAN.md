---
feature: telegram-ingestion
status: draft
canonical: true
---

# Feature Plan: Telegram Ingestion

## Purpose

This file is the feature-level implementation control board for Telegram ingestion. It defines task order, dependencies, concurrency rules, validation sequence, and execution status.

---

## Task DAG

```text
TELING-001 -> TELING-002
TELING-001 -> TELING-003
TELING-002 -> TELING-004
TELING-003 -> TELING-004
```

---

## Execution Phases

### Phase 1: Persistence Contracts

- `TELING-001` defines the SQLite schema and repository contracts for users, articles, jobs, notifications, deterministic artifact paths, and cleanup rules.

### Phase 2: Ingestion And Terminal State

- `TELING-002` implements gateway webhook ingestion after persistence contracts are available.
- `TELING-003` implements worker terminal notification writes after persistence contracts are available.

### Phase 3: Notification Dispatch

- `TELING-004` implements gateway terminal notification dispatch after ingestion and worker terminal notification contracts exist.

---

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `TELING-001` | Persistence contracts | done | - | `TELING-002`, `TELING-003` | no | `plans/TELING-001-persistence-contracts.execplan.md` |
| `TELING-002` | Telegram webhook ingestion | done | `TELING-001` | `TELING-004` | yes | - |
| `TELING-003` | Worker terminal notification contract | done | `TELING-001` | `TELING-004` | yes | - |
| `TELING-004` | Telegram notification dispatcher | blocked | `TELING-002`, `TELING-003` | - | no | `plans/TELING-004-telegram-notification-dispatcher.execplan.md` |

---

## Concurrency Rules

- `TELING-001` must run first because it defines shared schema and repository contracts.
- `TELING-002` and `TELING-003` may run concurrently after `TELING-001` is done if they do not both modify the same migration/repository files.
- `TELING-004` must run after `TELING-002` and `TELING-003` because it depends on stored job Telegram metadata, article result state, and terminal notification rows.
- Schema, public repository interfaces, and Telegram message fixtures must not be changed concurrently by multiple tasks.

---

## Blocking Interfaces or Schemas

- SQLite tables for users, articles, jobs, and notifications.
- Deterministic article artifact path convention under `DATA_DIR`.
- Gateway repository interfaces used by webhook ingestion.
- Worker repository interfaces used for terminal job transitions and notification creation.
- ARC-coded public article-processing failure convention in `docs/conventions/ERRORS.md`; terminal failure notification dispatch must preserve `[ARC-NNN]` prefixes from `jobs.error_message`.
- Telegram reply text fixtures:
  - `Nope, you must send only an URL`
  - `Ok, I will have a look`

---

## Validation Sequence

1. Complete persistence contract tests.
2. Run gateway webhook ingestion tests.
3. Run worker terminal job/notification tests.
4. Run gateway notification dispatcher and cleanup tests.
5. Run complete gateway and worker verification.

Validation commands:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

---

## Open Planning Questions

- None.

---

## Completion Criteria

The feature is complete when:

- all required tasks are `done` or explicitly `skipped`;
- acceptance criteria in `SPEC.md` are satisfied;
- validation sequence passes;
- durable implementation decisions have been promoted to canonical documents;
- `DIARY.md` contains final implementation notes;
- `docs/specs/INDEX.md` reflects the final feature status.
