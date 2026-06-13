---
id: TELING-001
feature: telegram-ingestion
title: Persistence contracts
status: done
depends_on: []
blocks: [TELING-002, TELING-003]
parallel: false
exec_plan: null
canonical: true
---

# TELING-001: Persistence Contracts

## Objective

Define and implement the SQLite schema and repository contracts needed for Telegram URL ingestion, users, articles, jobs, notifications, Telegram update idempotency, and deterministic article artifact paths.

## Story / Context

As the gateway and worker, we need a shared database contract so ingestion, worker processing, and terminal notification behavior can be implemented independently without inventing storage semantics.

## Scope

This task includes:

- `users` table with existing bootstrapped user ID `01ASB2XFCZJY7WHZ2FNRTMQJCT`, nullable unique `telegram_user_id`, and preservation of auth-owned `password_hash`.
- `articles` table with only durable article state: URL, optional canonical URL/title, status, error, and created timestamp.
- `jobs` table for worker queue state and Telegram origin metadata.
- `notifications` table for gateway delivery state linked to jobs.
- Telegram update idempotency through unique `jobs.telegram_update_id`.
- Deterministic article artifact path convention derived from `DATA_DIR` and `article_id`.
- TTL fields and cleanup contracts for terminal jobs and notifications.
- Gateway and worker repository interfaces or equivalent persistence boundaries.
- Persistence tests for schema constraints, idempotency, enqueue atomicity, terminal notification insertion, and cleanup eligibility.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/ARTIFACTS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- SQLite schema/migration or schema initialization for users, articles, jobs, and notifications.
- Persistence code to resolve an existing user by `telegram_user_id`, then atomically create an article and queued job for that user.
- Persistence code to atomically update terminal article/job state and create notification rows during worker terminal transitions.
- Repository support for deterministic artifact path resolution, or a documented path builder used by gateway and worker.
- Tests proving the shared persistence contract.

## Expected Affected Areas

```text
src/gateway/
src/worker/
SQLite schema or migrations
docs/specs/telegram-ingestion/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/ARTIFACTS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

Do not load unrelated feature folders unless required by discovered dependencies.

## Acceptance Criteria

```gherkin
Scenario: Valid Telegram URL is persisted atomically
  Given a Telegram update_id has not been processed
  And users.telegram_user_id maps the Telegram sender to an existing user row
  When the gateway records a valid URL ingestion
  Then one article is created with that user's id and status "queued"
  And one queued job is created for that article with that user's id
  And the Telegram update_id is recorded
  And the Telegram sender user ID is recorded on the job as telegram_user_id
  And no users row is created, updated, or reassigned by ingestion
  And the transaction commits atomically

Scenario: Existing Telegram sender mapping controls ownership
  Given users.telegram_user_id maps Telegram sender 12345 to an existing user row
  When the gateway records the ingestion
  Then the article and queued job use that user row's id
  And the users table is unchanged

Scenario: Duplicate Telegram update is ignored
  Given a Telegram update_id has already been processed
  When the gateway records the same update_id again
  Then no duplicate article is created
  And no duplicate job is created

Scenario: Worker records terminal success notification
  Given a job originated from Telegram
  And the job succeeds with a summary
  When the terminal transition is persisted
  Then the article status is "ready"
  And the job status is "succeeded"
  And one pending notification row is created for that job

Scenario: Worker records terminal failure notification
  Given a job originated from Telegram
  And the job fails with an error message
  When the terminal transition is persisted
  Then the article status is "failed"
  And the job status is "failed"
  And the job error_message contains the final error
  And one pending notification row is created for that job

Scenario: Deterministic article artifacts are resolved without path columns
  Given an article id exists
  When gateway or worker needs an artifact path
  Then the path is derived from DATA_DIR and article_id
  And no artifact path column is required in SQLite
```

## Done When

- SQLite schema and persistence contracts support all state required by the feature spec.
- Idempotency prevents duplicate articles/jobs for duplicate Telegram `update_id` values.
- Accepted Telegram ingestions persist `telegram_user_id` separately from `chat_id`.
- Accepted Telegram ingestions resolve ownership from an existing `users.telegram_user_id` mapping.
- Telegram ingestion does not create, update, or reassign `users` rows.
- Telegram ingestion preserves the auth-owned `users.password_hash` column.
- Article schema omits summary, domain, artifact path columns, extraction telemetry, and processed timestamp.
- Terminal notification rows can be created atomically with terminal article/job state.
- Job and notification states exclude retry states.
- TTL cleanup eligibility is represented for terminal jobs and sent/failed notifications.
- Gateway and worker tests cover the persistence contract.
- Task status and `PLAN.md` are updated if the task is completed.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- Inspect the resulting schema to confirm required keys and constraints exist.

## Dependencies

Depends on:

- None.

Blocks:

- `TELING-002`
- `TELING-003`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Use ULIDs for identifiers; do not delegate ID generation to SQLite.
- Do not introduce automatic retries in this task.
