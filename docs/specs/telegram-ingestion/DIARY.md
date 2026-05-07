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

## 2026-05-07 — TELING-001: Persistence Contracts

Status:
- completed

Summary:
- Implemented the shared SQLite persistence foundation for Telegram ingestion and Worker terminal transitions.

Changes:
- Added Gateway EF Core SQLite entities, schema constraints, Telegram ingestion repository contract, idempotent enqueue repository, ULID generation, and deterministic article artifact path resolver.
- Added Worker pure-Go SQLite persistence using `modernc.org/sqlite`, schema initialization, atomic `UPDATE ... RETURNING` job claim, terminal success/failure transitions, pending notification creation, TTL cleanup eligibility queries, and deterministic artifact path resolver.
- Added Gateway and Worker tests for enqueue atomicity, duplicate Telegram update idempotency, auth password hash preservation, canonical schema shape, artifact path derivation, worker claim, terminal notification creation, and cleanup eligibility.

Decisions:
- No new durable product decisions were made; implementation follows `SPEC.md`, `PLAN.md`, `ARTIFACTS.md`, `ARCHITECTURE.md`, and DSGN-011/DSGN-014.

Validation:
- `cd src/gateway && dotnet format`
- `cd src/gateway && dotnet build`
- `cd src/gateway && dotnet test`
- `cd src/worker && go tool lefthook run build`
- `cd src/worker && go tool lefthook run format`
- `cd src/worker && go tool lefthook run lint`
- `cd src/worker && go tool lefthook run test`
- Validation setup: ran `npm ci` in `src/ui` because root lefthook build/format/lint/test hooks include UI commands even when invoked from `src/worker`.

Follow-ups:
- `TELING-002` can consume Gateway persistence for webhook ingestion.
- `TELING-003` can consume Worker queue and terminal persistence for processing completion.

Canonical Updates:
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/plans/TELING-001-persistence-contracts.execplan.md`

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
