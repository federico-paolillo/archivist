---
id: TELING-001-PLAN
task: ../tasks/TELING-001-persistence-contracts.md
status: proposed
canonical: true
---

# ExecPlan: TELING-001 Persistence Contracts

## Objective

Create the shared SQLite persistence foundation for Telegram ingestion: users, articles, jobs, notifications, Telegram idempotency, terminal TTL cleanup, and deterministic article artifact paths.

## Linked Task

- `../tasks/TELING-001-persistence-contracts.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/TELING-001-persistence-contracts.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GATEWAY.md`
- `docs/conventions/WORKER.md`

## Assumptions

- SQLite remains the only queue and metadata store.
- IDs are generated as ULIDs by application code.
- Gateway and worker may use language-specific persistence implementations, but the schema contract must stay compatible.
- `telegram_update_id` is globally unique enough to key Telegram ingestion idempotency.
- The v0 personal user row has `id = 01ASB2XFCZJY7WHZ2FNRTMQJCT`.
- `users.telegram_user_id` is nullable at rest and unique when present so auth bootstrap can create the row before Telegram ingestion.
- `users.password_hash` is owned by `authn` and must be preserved by Telegram persistence code.
- Article artifact paths are derived from `DATA_DIR` and `article_id`, not stored in SQLite.

## Non-Goals

- Do not implement Telegram webhook handling.
- Do not implement Telegram API calls.
- Do not implement article extraction, summarization, or artifact writes.
- Do not introduce tenant ownership, roles, or per-user routing.
- Do not add worker retries, notification retries, retry backoff, or an external queue.
- Do not persist extraction telemetry fields.

## Implementation Sequence

1. Inspect existing gateway and worker persistence scaffolding and choose the smallest schema initialization approach consistent with current project structure.
2. Define `users` with `id`, nullable unique `telegram_user_id`, and nullable `password_hash`; seed or ensure the personal user row through the gateway ingestion path without overwriting `password_hash`.
3. Define `articles` with durable article state only: `id`, `user_id`, `original_url`, nullable `canonical_url`, nullable `title`, `status`, nullable `error_message`, and `created_at`.
4. Define deterministic artifact path construction from `DATA_DIR` and `article_id`, without artifact path columns.
5. Define `jobs` with user/article links, v0 states `queued`, `running`, `succeeded`, `failed`, Telegram origin metadata, error/timestamp/TTL fields, and unique `telegram_update_id`.
6. Define `notifications` with `job_id`, v0 states `pending`, `sent`, `failed`, error/timestamp/TTL fields, and a unique terminal notification per job.
7. Implement gateway persistence for atomic user/article/job creation and Telegram idempotency.
8. Implement worker persistence for atomic job claim with `UPDATE ... RETURNING`.
9. Implement worker persistence for atomic terminal article/job update plus pending notification insert.
10. Add tests for schema constraints, duplicate `telegram_update_id`, valid enqueue, deterministic artifact paths, job claim, terminal success notification, terminal failure notification, and TTL eligibility.
11. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Verify all schema changes are represented in rebuild-canonical docs if durable behavior changes during implementation.

## Documentation Updates Required

- Update `../tasks/TELING-001-persistence-contracts.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any schema behavior not already in `../SPEC.md` back into `../SPEC.md`.

## Risks

- Divergent C# and Go views of the SQLite schema can create runtime incompatibility.
- Idempotency that is not transactionally tied to enqueueing can create duplicate jobs or lost updates.
- Conflating Telegram sender user ID with chat ID can corrupt future identity correlation.
- Overwriting `users.password_hash` during Telegram user upsert would break UI/API authentication.
- Storing artifact paths in SQLite would contradict the deterministic artifact path contract.
- Adding retries opportunistically would contradict the v0 no-retry decision.

## Rollback / Recovery Notes

- Schema changes should be introduced in a way that can be recreated from canonical docs.
- Failed implementation attempts should leave task status as `blocked` or `draft` with the missing decision documented.

## Completion Criteria

- Persistence tests pass in gateway and worker.
- Shared schema supports all data required by `../SPEC.md`.
- `TELING-002` and `TELING-003` can consume the persistence contract without making new durable schema decisions.
