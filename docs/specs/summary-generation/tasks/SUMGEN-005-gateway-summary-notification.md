---
id: SUMGEN-005
feature: summary-generation
title: Gateway Summary Notification
depends_on: [SUMGEN-004, TELING-004]
blocks: [UIEND-002]
parallel: false
requires_exec_plan: false
canonical: true
---
# SUMGEN-005: Gateway Summary Notification

## Objective

Implement Gateway summary-success notification body construction and summary artifact reads so succeeded summary-complete jobs reply with the persisted summary through the dispatcher infrastructure from `TELING-004`.

## Story / Context

As the authorized Telegram user, I need the terminal success reply to include the generated summary and reply to my original message.

## Scope

This task includes:

- Gateway read-only article artifact abstraction scoped to `DATA_DIR`.
- Reading `{DATA_DIR}/articles/{article_id}/summary.md` for succeeded jobs.
- Telegram success reply text starting with `Archived. Summary is:`.
- Supplying summary-success reply bodies to the dispatcher target-delivery path from `TELING-004`.
- Deterministic truncation to Telegram message length limits through the `TELING-004` dispatcher contract.
- Failure behavior when `summary.md` is missing or unreadable.
- Tests proving Gateway artifact access cannot write, create, rename, or delete files.
- Tests covering summary-complete success notification and ARC-coded failure notification preservation.


## Inputs

Required inputs, existing files, interfaces, or decisions:

- Requires `SUMGEN-004`.
- Requires `TELING-004`.
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/ARTIFACTS.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway read-only artifact access.
- Gateway summary-success body construction for summary-complete jobs.
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
- `docs/ARTIFACTS.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/specs/telegram-ingestion/SPEC.md`
- `docs/specs/telegram-ingestion/PLAN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-004-telegram-notification-dispatcher.md`
- `./SUMGEN-004-worker-summary-pipeline-integration.md`

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
- Markdown-complete and snapshot-only success notifications are no longer the final success notification path.
- Gateway artifact abstraction is read-only.
- Missing or unreadable `summary.md` fails notification delivery without mutating article/job state.
- Tests cover summary success, read-only access, missing summary artifact, deterministic truncation, and ARC-coded failure preservation.

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

- `UIEND-002`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.

## Notes

- Gateway must not write article artifacts.
- Gateway success notifications now read `summary.md` through a read-only artifact abstraction.
- `TELING-004` owns dispatcher infrastructure, reply target delivery, delivery failure handling, truncation mechanics, and notification cleanup. This task owns summary-success body construction and summary artifact reads.
