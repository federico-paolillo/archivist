---
id: SUMGEN-005
feature: summary-generation
title: Gateway Summary Notification
status: blocked
depends_on: [SUMGEN-004, TELING-004]
blocks: []
parallel: false
exec_plan: ../plans/SUMGEN-005-gateway-summary-notification.execplan.md
canonical: true
---

# SUMGEN-005: Gateway Summary Notification

## Objective

Update Gateway terminal notification dispatch so succeeded summary-complete jobs reply with the persisted summary through read-only artifact access.

## Story / Context

As the authorized Telegram user, I need the terminal success reply to include the generated summary and reply to my original message.

## Scope

This task includes:

- Gateway read-only article artifact abstraction scoped to `DATA_DIR`.
- Reading `{DATA_DIR}/articles/{article_id}/summary.md` for succeeded jobs.
- Telegram success reply text starting with `Archived. Summary is:`.
- Deterministic truncation to Telegram message length limits.
- Failure behavior when `summary.md` is missing or unreadable.
- Tests proving Gateway artifact access cannot write, create, rename, or delete files.
- Tests covering summary-complete success notification and ARC-coded failure notification preservation.

## Out of Scope

This task does not include:

- Worker summary generation.
- Worker artifact writes.
- SQLite schema changes.
- UI article detail rendering.
- Notification retries.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `SUMGEN-004`.
- Completed `TELING-004`.
- `docs/conventions/GATEWAY.md`
- `docs/conventions/ARTIFACTS.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway read-only artifact access.
- Gateway dispatcher success branch for summary-complete jobs.
- Gateway tests for summary success, missing summary artifact handling, truncation, and ARC-coded failure notification behavior.

## Expected Affected Areas

```text
src/gateway/
Gateway notification dispatcher
Gateway tests
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `./SUMGEN-004-worker-summary-pipeline-integration.md`
- `../plans/SUMGEN-005-gateway-summary-notification.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Summary-complete success notification is sent
  Given a pending notification exists for a succeeded job
  And summary.md exists for the article
  When the Gateway dispatches the notification
  Then it replies to the original Telegram message with text starting "Archived. Summary is:"
  And it includes the summary text
  And it marks the notification sent if Telegram accepts the reply

Scenario: Gateway artifact access is read-only
  Given Gateway has an article artifact abstraction
  When application code uses the abstraction
  Then it can read summary.md
  And it cannot write, create, rename, or delete artifacts

Scenario: Summary artifact is missing
  Given a pending notification exists for a succeeded job
  And summary.md is missing
  When the Gateway dispatches the notification
  Then the notification is marked failed
  And article and job terminal state are not changed
```

## Done When

- Summary-complete success notifications are supported.
- Markdown-complete and snapshot-only success notifications are no longer the final v0 success notification path.
- Gateway artifact abstraction is read-only.
- Missing or unreadable `summary.md` fails notification delivery without mutating article/job state.
- Tests cover summary success, read-only access, missing summary artifact, deterministic truncation, and ARC-coded failure preservation.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual validation, if any:

- None.

## Dependencies

Depends on:

- `SUMGEN-004`
- `TELING-004`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
../plans/SUMGEN-005-gateway-summary-notification.execplan.md
```

## Open Questions

- None.

## Notes

- Gateway must not write article artifacts.
