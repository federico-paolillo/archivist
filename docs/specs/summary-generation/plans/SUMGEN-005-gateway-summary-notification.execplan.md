---
id: SUMGEN-005-PLAN
task: ../tasks/SUMGEN-005-gateway-summary-notification.md
status: proposed
canonical: true
---

# ExecPlan: SUMGEN-005 Gateway Summary Notification

## Objective

Implement Gateway summary-complete success notification behavior by reading `summary.md` through a read-only artifact abstraction and sending deterministic Telegram replies for succeeded jobs.

## Linked Task

- `../tasks/SUMGEN-005-gateway-summary-notification.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/SUMGEN-005-gateway-summary-notification.md`

Add only ExecPlan-specific context:

- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `docs/specs/summary-generation/tasks/SUMGEN-004-worker-summary-pipeline-integration.md`

## Assumptions

- `SUMGEN-004` writes `summary.md` before committing article/job success and creating a pending notification.
- `TELING-004` provides pending notification polling, reply targeting, failure reply handling, delivery state updates, truncation, and cleanup.
- Gateway artifact access for notification success content is read-only.

## Non-Goals

- Do not implement Worker summary generation or artifact writes.
- Do not change SQLite schema.
- Do not add notification retries.
- Do not expose write, create, rename, or delete operations through the read-only artifact abstraction.

## Implementation Sequence

1. Reuse or create a Gateway read-only article artifact abstraction scoped to `DATA_DIR`.
2. Implement summary artifact reads for `{DATA_DIR}/articles/{article_id}/summary.md`.
3. Integrate a succeeded-job notification branch into the dispatcher created by `TELING-004`.
4. For succeeded jobs, read `summary.md`, build `Archived. Summary is: <summary>`, and truncate deterministically to Telegram message limits.
5. Send the reply using `jobs.telegram_chat_id` and `jobs.telegram_message_id`.
6. On successful send, mark the notification `sent`, set `sent_at`, and set `expires_at` to 7 days after send.
7. If `summary.md` is missing or unreadable, mark the notification `failed` with an operational error and do not mutate article/job terminal state.
8. Preserve failed-job notification behavior from `TELING-004`, including unchanged ARC-coded `jobs.error_message` subject only to deterministic truncation.
9. Add tests for summary success, read-only abstraction surface, missing summary artifact, unreadable summary artifact, deterministic truncation, ARC-coded failure preservation, and article/job immutability on delivery failure.
10. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual checks:

- Confirm Gateway artifact notification code cannot write, create, rename, or delete article artifacts.
- Confirm missing `summary.md` fails notification delivery without changing article/job terminal state.

## Documentation Updates Required

- Update `../tasks/SUMGEN-005-gateway-summary-notification.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new notification, truncation, or artifact-read behavior to `../SPEC.md` or relevant conventions before relying on it.

## Risks

- Missing summary artifacts must not be silently converted into success replies.
- A write-capable artifact abstraction would weaken Gateway ownership boundaries.
- Long summaries may exceed Telegram limits if truncation is omitted.
- Stripping ARC prefixes from failed-job replies would break the public error contract.

## Rollback / Recovery Notes

- Disabling the summary success branch leaves success notifications pending until the issue is corrected or cleanup policy is explicitly changed.
- Notification delivery failures are recoverable from notification status and error text.

## Completion Criteria

- Gateway tests cover summary success, missing/unreadable summary artifacts, read-only artifact access, truncation, and ARC-coded failure preservation.
- Gateway validation passes.
- Succeeded summary-complete jobs receive Telegram replies with persisted summary text.
