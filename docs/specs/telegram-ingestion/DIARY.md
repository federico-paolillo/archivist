# Implementation Diary: Telegram Ingestion

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Entry Template

```md
## YYYY-MM-DD — TASK-ID: Task Title

Status:
- completed / partially completed / blocked / skipped

Summary:
- Brief outcome.

Changes:
- Files, schemas, or behavior changed.

Decisions:
- Decisions made during implementation.

Validation:
- Commands and manual checks run.

Follow-ups:
- Remaining work, if any.

Canonical Updates:
- Specs, plans, standards, architecture docs, or design decisions updated.
```

---

## Log

## 2026-05-04 — TELING-DOC: ARC Error Convention Alignment

Status:
- completed

Summary:
- Aligned Telegram ingestion documentation with the shared ARC error-code convention for terminal article-processing failures only.

Changes:
- Updated `SPEC.md`, `PLAN.md`, `TELING-003`, `TELING-004`, and the `TELING-004` ExecPlan.

Decisions:
- ARC codes are transported by Telegram terminal failure notifications when `jobs.error_message` already contains article-processing public failure text.
- Telegram webhook validation replies, authorization failures, acknowledgement failures, and Telegram delivery errors are not ARC-coded.

Validation:
- Inspected Markdown diff and searched for unresolved placeholders.

Follow-ups:
- Implement TELING and ARTPROC tasks according to their updated contracts.

Canonical Updates:
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `docs/specs/telegram-ingestion/plans/TELING-004-telegram-notification-dispatcher.execplan.md`

## 2026-05-04 — TELING-DOC: Summary Notification Alignment

Status:
- completed

Summary:
- Aligned Telegram terminal notification docs with summary-complete final success.

Changes:
- Updated success notification requirements and task/ExecPlan wording to treat snapshot-only replies as interim bridges only.

Decisions:
- Final v0 successful Telegram terminal replies are summary-based and owned by `SUMGEN-005`.

Validation:
- Documentation consistency checked by repository search and review.

Follow-ups:
- Implement `SUMGEN-005` after Worker summary completion exists.

Canonical Updates:
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `docs/specs/telegram-ingestion/plans/TELING-004-telegram-notification-dispatcher.execplan.md`

## 2026-05-06 — DOCS-SANITY: Dispatcher Scope Correction

Status:
- completed

Summary:
- Corrected Telegram notification documentation so `TELING-004` owns dispatcher infrastructure, failure replies, delivery state, truncation, and cleanup only.

Changes:
- Limited `TELING-004` task and ExecPlan success handling to a later summary-generation branch.
- Updated `SPEC.md` dependencies and artifact-read ownership.
- Accepted the `TELING-001` ExecPlan and fixed PLAN/table parallel notation drift.

Decisions:
- Final v0 success notification content is owned by `SUMGEN-005`.
- Gateway summary artifact reads for success replies are not part of `TELING-004`.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no stale model/config drift or snapshot/Markdown terminal-success wording.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement `TELING-004` according to the narrowed dispatcher scope after dependencies are complete.

Canonical Updates:
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `docs/specs/telegram-ingestion/plans/TELING-001-persistence-contracts.execplan.md`
- `docs/specs/telegram-ingestion/plans/TELING-004-telegram-notification-dispatcher.execplan.md`

## 2026-05-09 — TELING-003: Worker Terminal Notification Contract

Status:
- completed

Summary:
- Implemented worker-side SQLite persistence for atomic job claiming and terminal article/job/notification state transitions.

Changes:
- `src/worker/pkg/app/config/config.go`: Added `SqlitePath` and `DataDir` fields to Root config with empty string defaults.
- `src/worker/pkg/app/config/load_test.go`: Added tests for default values and env-var loading of the new config fields.
- `src/worker/pkg/app/app.go`: Added `DB *sql.DB` and `Jobs jobs.Repository` fields to App; `NewApp` conditionally opens and applies schema when `SqlitePath` is set; `Close` closes the DB.
- `src/worker/pkg/app/app_test.go`: Added test for App with SQLite path that verifies DB and Jobs are non-nil.
- `src/worker/pkg/db/db.go`: New package. Opens a CGO-free modernc SQLite database with WAL, foreign keys, busy timeout, single connection.
- `src/worker/pkg/db/schema.go`: New file. `ApplySchema` idempotently creates users, articles, jobs, and notifications tables.
- `src/worker/pkg/jobs/job.go`: New file. `Job` struct with all fields from the TELING-001 schema. `HasTelegramOrigin` returns true when both `telegram_chat_id` and `telegram_message_id` are set.
- `src/worker/pkg/jobs/repository.go`: New file. `Repository` interface with `ClaimQueued` and `CompleteTerminal`. `SQLiteRepository` implements atomic claim via `UPDATE...RETURNING` and atomic terminal transition via a single transaction (article update + job update + conditional notification insert).
- `src/worker/pkg/jobs/repository_test.go`: New file. Tests for ClaimQueued (success, no rows, Telegram fields), CompleteTerminal success and failure for Telegram jobs, ARC-coded error preservation, and non-Telegram jobs (no notification).
- `src/worker/go.mod` / `src/worker/go.sum`: Added `modernc.org/sqlite` and `github.com/oklog/ulid/v2`.

Decisions:
- Used `modernc.org/sqlite` (CGO-free) to satisfy `CGO_ENABLED=0` requirement.
- `ClaimQueued` uses `UPDATE...RETURNING` with a subquery to atomically select and claim one queued job.
- `HasTelegramOrigin` checks for both `telegram_chat_id` and `telegram_message_id` (the reply-target fields), not `telegram_user_id` alone.
- Notification `expires_at` set to 7 days (REQ-029). Job `expires_at` set to 14 days (REQ-028).
- ARC-coded error text written verbatim to both `articles.error_message` and `jobs.error_message` without modification (REQ-024A).
- Schema is inline DDL using `CREATE TABLE IF NOT EXISTS` — idempotent, no migration tool for v0.
- Config fields use configuro struct-field naming: `SqlitePath` maps to env `APP_SQLITEPATH`, `DataDir` maps to `APP_DATADIR`.

Validation:
- `go build ./...` passed.
- `go tool golangci-lint run` passed (all linters clean).
- `go tool golangci-lint run --fix` passed (no formatting changes needed).
- `go test -race -shuffle=on ./...` passed (all packages).

Follow-ups:
- Gateway implementation of TELING-001 must use the same schema DDL or a compatible migration.
- TELING-004 may proceed once TELING-002 is also complete.

Canonical Updates:
- `docs/specs/telegram-ingestion/tasks/TELING-003-worker-terminal-notification-contract.md` (status: done)
- `docs/specs/telegram-ingestion/PLAN.md` (TELING-003 row: done)
