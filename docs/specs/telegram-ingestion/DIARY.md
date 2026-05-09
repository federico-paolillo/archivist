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

## 2026-05-09 — TELING-002: Telegram Webhook Ingestion

Status:
- completed

Summary:
- Implemented `POST /telegram/webhook` gateway endpoint with webhook secret validation, allowed-user authorization, strict http/https URL-only validation, atomic persistence via the TELING-001 contract, idempotent duplicate update_id handling, and immediate Telegram reply for both valid and invalid authorized messages. Integration tests cover all acceptance criteria paths.

Changes:
- `src/gateway/Archivist.Gateway.Api/Telegram/Endpoints.cs` — route group mapping POST /telegram/webhook
- `src/gateway/Archivist.Gateway.Api/Telegram/Handlers.cs` — static handler extracting secret header and building TelegramWebhookCommand
- `src/gateway/Archivist.Gateway.Api/Telegram/Models/TelegramUpdateDto.cs` — DTOs for update, message, chat, user
- `src/gateway/Archivist.Gateway.Application/Telegram/ITelegramClient.cs` — interface for sending Telegram replies
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramOptions.cs` — options for bot token, webhook secret, allowed user id
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramWebhookCommand.cs` — handler input record
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramWebhookHandler.cs` — application service implementing all validation and persistence logic
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramWebhookResult.cs` — outcome enum and result record
- `src/gateway/Archivist.Gateway.Application/Telegram/Defaults/HttpTelegramClient.cs` — HTTP Bot API client implementation
- `src/gateway/Archivist.Gateway.Application/Telegram/Extensions/ServiceCollectionExtensions.cs` — AddTelegram DI registration
- `src/gateway/Archivist.Gateway.Api/Program.cs` — updated to register and map Telegram conditionally on SQLITE_PATH
- `src/gateway/Archivist.Gateway.Application/Archivist.Gateway.Application.csproj` — added Microsoft.Extensions.Http package
- `src/gateway/Archivist.Gateway.Tests/Api/TelegramWebhookEndpointTest.cs` — 13 integration tests covering bad secret, missing secret, unauthorized sender, invalid URL (4 variants), valid URL, sender ID vs chat ID distinctness, duplicate update, acknowledgement failure, and no-message update
- Also included TELING-001 persistence foundation files (entities, DbContext, repository, extensions, constants) which were missing from the worktree starting checkpoint.

Decisions:
- Telegram and persistence registration are conditional on SQLITE_PATH to avoid DI validation failures in environments without SQLite.
- `TelegramWebhookHandler` is a sealed partial class to support LoggerMessage source generation.
- CA1031 (catch general Exception) is suppressed with explanation at the acknowledgement-failure catch site and fire-and-forget reply helper, because the spec requires acknowledgement failure to not roll back ingestion.
- The API `Handlers` and `Endpoints` classes are `internal` to avoid CA1724 type-name conflicts with .NET framework namespace names.
- `SenderUserId` is read from `message.from.id`, not from `message.chat.id`, as required by the spec (telegram_user_id is sender identity metadata, not reply-target metadata).

Validation:
- `cd src/gateway && dotnet format` — passed, reformatted some files.
- `cd src/gateway && dotnet build` — passed, 0 warnings, 0 errors.
- `cd src/gateway && dotnet test` — passed, 18/18 tests (1 ping, 4 persistence, 13 webhook integration).

Follow-ups:
- TELING-004 (notification dispatcher) can now proceed as TELING-002 is done.
- TELING-003 (worker terminal notification contract) is independent and can proceed concurrently.

Canonical Updates:
- `docs/specs/telegram-ingestion/tasks/TELING-002-telegram-webhook-ingestion.md` — status: done
- `docs/specs/telegram-ingestion/PLAN.md` — TELING-002 status: done
