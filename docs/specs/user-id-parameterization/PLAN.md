---
feature: user-id-parameterization
status: done
canonical: true
---

# Feature Plan: User ID Parameterization

## Purpose

This file controls the implementation of runtime `user_id` resolution across Gateway, Worker, SQLite ownership, and observability.

## Task DAG

```text
MUSER-001 -> MUSER-002
MUSER-001 -> MUSER-003
MUSER-001 -> MUSER-004
MUSER-002 -> MUSER-005
MUSER-003 -> MUSER-005
MUSER-004 -> MUSER-005
MUSER-005 -> MUSER-006
```

## Execution Phases

### Phase 1: Canonical Contracts

- `MUSER-001` creates the feature artifacts and user-id resolution contract.

### Phase 2: Parallel Module Implementation

- `MUSER-002` implements Gateway auth, Telegram mapping, ownership, and telemetry behavior.
- `MUSER-003` implements Worker CLI default-user existence checking, job/article ownership, and telemetry behavior.
- `MUSER-004` aligns cross-feature observability and existing canonical docs.

### Phase 3: Integration And Review

- `MUSER-005` integrates Gateway and Worker slices and runs cross-module validation.
- `MUSER-006` records reviews, final validation, and ALM closure.

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `MUSER-001` | Canonical user-id resolution contract | done | - | `MUSER-002`, `MUSER-003`, `MUSER-004` | no | [`ExecPlan`](./plans/MUSER-001-canonical-user-id-resolution-contract.execplan.md) |
| `MUSER-002` | Gateway user-id resolution and telemetry | done | `MUSER-001` | `MUSER-005` | yes | [`ExecPlan`](./plans/MUSER-002-gateway-user-id-resolution-and-telemetry.execplan.md) |
| `MUSER-003` | Worker user-id propagation and telemetry | done | `MUSER-001` | `MUSER-005` | yes | [`ExecPlan`](./plans/MUSER-003-worker-user-id-propagation-and-telemetry.execplan.md) |
| `MUSER-004` | Cross-feature docs and observability cleanup | done | `MUSER-001` | `MUSER-005` | yes | - |
| `MUSER-005` | Integration validation | done | `MUSER-002`, `MUSER-003`, `MUSER-004` | `MUSER-006` | no | [`ExecPlan`](./plans/MUSER-005-integration-validation.execplan.md) |
| `MUSER-006` | Review and ALM closure | done | `MUSER-005` | - | no | - |

## Concurrency Rules

- Gateway and Worker implementation may run in parallel after `MUSER-001`.
- Gateway workers must not edit Worker files.
- Worker workers must not edit Gateway files.
- The coordinator owns canonical docs, feature task status, feature plan status, diary entries, and final integration.
- Reviewer passes must happen before `MUSER-006` is marked done.

## Blocking Interfaces or Schemas

- `users.id`, `users.telegram_user_id`, and `users.password_hash`.
- `articles.user_id` and `jobs.user_id`.
- Password login candidate-loading and exact-one-match contract for `POST /login`.
- Telegram ingestion command and repository contract.
- Worker job repository ownership APIs.
- Log and span attribute key `user_id`.

## Validation Sequence

```bash
git diff --check
cd src/gateway && dotnet format && dotnet build && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
docker compose --env-file .env.example config --quiet
```

## Open Planning Questions

- None.

## Completion Criteria

The feature is complete when:

- all tasks are `done`;
- Gateway and Worker no longer hardcode runtime ownership except for auth bootstrap's accepted personal user and Telegram sender seeds and Worker CLI enqueue's accepted `jobs.DefaultUserID` exception;
- Telegram authorization is database mapping based;
- Gateway and Worker attach `user_id` telemetry when known;
- validation passes or failures are recorded;
- durable behavior is captured in canonical docs;
- `DIARY.md` records final implementation outcomes.
