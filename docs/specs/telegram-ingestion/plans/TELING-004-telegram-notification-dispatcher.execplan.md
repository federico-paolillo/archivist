---
id: TELING-004-PLAN
task: ../tasks/TELING-004-telegram-notification-dispatcher.md
status: proposed
canonical: true
---

# ExecPlan: TELING-004 Telegram Notification Dispatcher

## Objective

Implement gateway-owned dispatch of pending Telegram notifications from SQLite, deriving reply target and content from jobs/articles and marking delivery errors without automatic retry.

## Linked Task

- `../tasks/TELING-004-telegram-notification-dispatcher.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/TELING-004-telegram-notification-dispatcher.md`

Add only ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GATEWAY.md`

## Assumptions

- `TELING-002` has introduced or can share a Telegram client abstraction.
- `TELING-003` writes pending notification rows for terminal Telegram-originated jobs.
- Telegram terminal replies are best-effort in v0; delivery failure is persisted but not retried automatically.
- Success reply content is not selected by TELING-004.
- Summary-complete success notification content is owned by `SUMGEN-005`.
- Failed article-processing replies preserve ARC-coded `jobs.error_message` exactly, subject only to deterministic Telegram length truncation.

## Non-Goals

- Do not implement immediate queued acknowledgements here.
- Do not modify worker terminal transition behavior except through already-defined notification contracts.
- Do not introduce retry scheduling, backoff, an external scheduler, queue, or broker.

## Implementation Sequence

1. Reuse or introduce a gateway Telegram client abstraction capable of sending a reply to a `telegram_chat_id` and `telegram_message_id`.
2. Implement notification polling for `pending` rows.
3. For each pending notification, join `notifications -> jobs -> articles`.
4. For succeeded jobs, leave the notification pending unless a later feature has registered or implemented a success-content branch.
5. For failed jobs, read `jobs.error_message` and truncate to Telegram message limits.
6. For failed article-processing jobs, preserve the `[ARC-NNN]` prefix and public message from `jobs.error_message`; do not reinterpret it as a Telegram, Gateway, or delivery error.
7. Send failure Telegram replies using job Telegram reply-target metadata.
8. On successful failure-reply send, mark the notification `sent`, set `sent_at`, and set `expires_at` to 7 days after send.
9. On Telegram delivery failure, mark the notification `failed`, persist operational delivery `error_message`, and set `expires_at` to 7 days after failure.
10. Implement gateway cleanup for `sent` and `failed` notifications whose `expires_at` is in the past.
11. Add dispatcher tests for pending succeeded-job behavior, ARC-coded failed article-processing reply, generic failed job reply, Telegram delivery failure, truncation, and cleanup.
12. Update task status, `PLAN.md`, and `DIARY.md` after validation if implementation is completed.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual checks:

- Confirm notification delivery failures never mutate terminal article/job state.
- Confirm Telegram delivery failure messages stored on notifications are operational Gateway errors, not ARC-coded article-processing errors.

## Documentation Updates Required

- Update `../tasks/TELING-004-telegram-notification-dispatcher.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new delivery or cleanup behavior to `../SPEC.md` if it becomes durable behavior.

## Risks

- Success notifications remain pending until a downstream success-content feature such as `SUMGEN-005` implements succeeded-job dispatch.
- Long failure messages may exceed Telegram limits if truncation is omitted.
- Stripping or rewording ARC-coded `jobs.error_message` would break the shared public failure contract.
- Joining through jobs/articles requires job rows to outlive notification dispatch; cleanup must respect the 7-day notification and 14-day terminal job TTLs.

## Rollback / Recovery Notes

- Dispatcher can be disabled without deleting queued or terminal jobs; pending notifications remain in SQLite until cleanup policy handles them.
- Terminal notification delivery failures are recoverable by inspecting notification status and error text.

## Completion Criteria

- Dispatcher tests cover pending succeeded-job behavior, ARC-coded failed article-processing reply, generic failed job reply, Telegram delivery failure, truncation, and cleanup.
- Gateway validation passes.
- The feature's terminal notification acceptance criteria are satisfied.
