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

## 2026-05-09 — TELING-001: Post-Review Corrective Fixes

Status:
- completed

Summary:
- Applied four corrective fixes identified during the 2026-05-08 code review of TELING-001.

Changes:
- FIX-1: Changed `NotificationEntity.ExpiresAt` from `DateTimeOffset?` to non-nullable `DateTimeOffset` in `src/gateway/Archivist.Gateway.Application/Persistence/Entities/NotificationEntity.cs:41`. Matches spec intent; fails fast at insertion if the value is omitted rather than silently accepting null.
- FIX-2: Extended `PersistenceConstants` with `NotificationSent = "sent"` and `NotificationFailed = "failed"` in `src/gateway/Archivist.Gateway.Application/Persistence/PersistenceConstants.cs`. Pre-defines the values needed by TELING-004 so no magic strings are required in the next task.
- FIX-3: Updated `docs/specs/INDEX.md` `telegram-ingestion` status from `draft` to `in_progress`. TELING-001 is done; the feature is actively being implemented.
- FIX-4: Strengthened the duplicate-update idempotency test in `TelegramIngestionRepositoryTest` to assert that the personal user row (`telegram_user_id`) is not mutated on a duplicate enqueue. A regression that overwrites the user row on the idempotent path would now be caught.

Decisions:
- `NotificationEntity.ExpiresAt` is non-nullable by spec; nullability was an accidental omission, not a deliberate design choice.
- `NotificationSent` and `NotificationFailed` constants are added now to unblock TELING-004 without requiring a separate bookkeeping task.

Validation:
- `cd src/gateway && dotnet format`
- `cd src/gateway && dotnet build`
- `cd src/gateway && dotnet test`
- `go tool lefthook run lint && go tool lefthook run test` (Worker, unchanged by these fixes)

Follow-ups:
- TELING-002 can proceed; all TELING-001 corrective items are closed.
- TELING-004 can reference `PersistenceConstants.NotificationSent` and `PersistenceConstants.NotificationFailed` without introducing magic strings.

Canonical Updates:
- `docs/specs/INDEX.md` (telegram-ingestion status: draft → in_progress)

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

## 2026-05-10 — TELING-004: Telegram Notification Dispatcher

Status:
- completed

Summary:
- Implemented gateway-owned background dispatcher polling pending notification rows from SQLite, sending Telegram error replies for failed jobs, and cleaning up expired sent/failed notifications after 7 days.

Changes:
- `src/gateway/Archivist.Gateway.Application/Persistence/TelegramNotificationContracts.cs` — new file; `PendingNotificationRow` record and `ITelegramNotificationRepository` interface.
- `src/gateway/Archivist.Gateway.Application/Persistence/Defaults/EfTelegramNotificationRepository.cs` — new file; EF Core implementation of `ITelegramNotificationRepository`. `GetPendingAsync` joins notifications to jobs. `MarkSentAsync`/`MarkFailedAsync` use fetch-then-update. `DeleteExpiredAsync` uses client-side filtering to avoid EF Core SQLite `DateTimeOffset` translation limits.
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramNotificationDispatcher.cs` — new file; polls pending rows, skips succeeded-job notifications (deferred to SUMGEN-005), sends error reply for failed jobs, truncates to `TelegramMessageMaxLength = 4096`, marks sent/failed, handles delivery failure without retrying.
- `src/gateway/Archivist.Gateway.Application/Telegram/TelegramNotificationDispatcherService.cs` — new file; `BackgroundService` that runs dispatcher every 10 seconds in a scoped DI scope.
- `src/gateway/Archivist.Gateway.Application/Telegram/Extensions/ServiceCollectionExtensions.cs` — added TELING-004 DI registrations (`ITelegramNotificationRepository`, `TelegramNotificationDispatcher`, `TelegramNotificationDispatcherService`).
- `src/gateway/Archivist.Gateway.Application/Archivist.Gateway.Application.csproj` — added `Microsoft.Extensions.Hosting.Abstractions 10.0.0`.
- `src/gateway/Archivist.Gateway.Tests/Telegram/TelegramNotificationDispatcherTest.cs` — new file; 13 tests using `FakeTimeProvider`, `NullLogger`, `FakeTelegramClient`, and real `EfTelegramNotificationRepository` against in-memory SQLite.

Decisions:
- `DeleteExpiredAsync` loads non-pending notification rows into memory before filtering by `ExpiresAt` because EF Core SQLite cannot translate `DateTimeOffset` comparisons in WHERE clauses (local reversible decision).
- `MarkSentAsync`/`MarkFailedAsync` use fetch-then-update (`FindAsync` + `SaveChangesAsync`) because `ExecuteUpdateAsync` chained `SetProperty` calls do not translate for SQLite in this version.
- Succeeded-job notifications are left `pending` with a Debug log message; success content selection is owned by `SUMGEN-005`.
- CA1031 is suppressed on the delivery failure catch site (REQ-026) and in the background service cycle catch.

Validation:
- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed, 0 warnings, 0 errors.
- `cd src/gateway && dotnet test` — passed, 57/57 tests (all passing).

Follow-ups:
- SUMGEN-005 must implement succeeded-job success content selection and flip succeeded-job notifications from `pending` to `sent`.
- Terminal job/notification cleanup TTLs should be verified end-to-end once SUMGEN-005 lands.

Canonical Updates:
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md` (status: done)
- `docs/specs/telegram-ingestion/PLAN.md` (TELING-004 row: done)
- `docs/specs/telegram-ingestion/plans/TELING-004-telegram-notification-dispatcher.execplan.md` (status: completed)

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

## 2026-05-30 — TELING-SEC: Telegram Runtime Security Hardening

Status:
- completed

Summary:
- Hardened Gateway Telegram registration so required runtime secrets are validated during startup and the Telegram Bot API typed HTTP client does not use default `IHttpClientFactory` request logging.

Changes:
- `src/gateway/Archivist.Gateway.Application/Telegram/Extensions/ServiceCollectionExtensions.cs` — bound `TelegramSettings` with validation for nonblank `BotToken`/`WebhookSecret`, positive `AllowedUserId`, `ValidateOnStart()`, and `RemoveAllLoggers()` on the typed Telegram client.
- `src/gateway/Archivist.Gateway.Tests/Api/GatewayConfigurationSourceTest.cs` — added tests for valid binding, startup validation failure cases, typed client registration, and reflected `HttpClientFactoryOptions` logger suppression state.
- `src/gateway/Archivist.Gateway.Tests/IntegrationTest.cs` — added fake Telegram settings to the shared integration-test configuration so unrelated host-start tests satisfy the new startup contract.

Decisions:
- Telegram settings validation is a durable startup behavior and is promoted to `SPEC.md`.
- Telegram Bot API token exposure is mitigated at the typed-client registration by calling Microsoft `IHttpClientBuilder.RemoveAllLoggers()` instead of relying on application-wide logging filters.

Validation:
- `cd src/gateway && dotnet format --verify-no-changes` — passed.
- `cd src/gateway && dotnet build` — passed, 0 warnings, 0 errors.
- `cd src/gateway && dotnet test` — passed, 156/156 tests.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/telegram-ingestion/SPEC.md` — added REQ-030 for Telegram runtime configuration startup validation.

## 2026-05-31 — TELING-002-REVIEW-P2: Authorized non-text Telegram invalid reply

Status:
- completed

Summary:
- Resolved the active P2 review finding where authorized Telegram updates with chat/message identifiers but missing `text` were treated as `NoMessage`.
- Gateway now keeps truly unreplyable updates ignored when chat or message identifiers are missing, but sends `Nope, you must send only an URL` for authorized media-only or caption-shaped messages with a reply target.

Changes:
- `TelegramWebhookHandler` treats missing text as invalid once sender, chat ID, and message ID are present.
- `TelegramUpdateDto.MessageId` is nullable so missing message identifiers remain unreplyable instead of becoming `0`.
- Added handler and endpoint regressions for missing text, media-only, caption-shaped, and missing reply-target paths.

Decisions:
- Captions remain out of scope for URL ingestion and do not count as accepted text.
- No route, schema, Telegram reply text, or authorization behavior changed beyond the invalid-message path required by TELING-002.

Validation:
- Gateway worker: `cd src/gateway && dotnet format` — passed.
- Gateway worker: `cd src/gateway && dotnet build` — passed.
- Gateway worker: `cd src/gateway && dotnet test` — passed: 162 tests.
- Gateway reviewer: `dotnet build && dotnet test` — passed, 162 tests; approved with no findings.
- Coordinator: `cd src/gateway && dotnet format && dotnet build && dotnet test` — passed, 162 tests.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/telegram-ingestion/DIARY.md` — this review-remediation entry.
