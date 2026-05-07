---
id: MDEXT-004
feature: markdown-extraction
title: Worker Jina Reader Fallback
status: done
depends_on: [MDEXT-001]
blocks: [MDEXT-005]
parallel: true
exec_plan: ../plans/MDEXT-004-worker-jina-reader-fallback.execplan.md
canonical: true
---

# MDEXT-004: Worker Jina Reader Fallback

## Objective

Implement Jina Reader fallback for Markdown extraction behind the Worker-owned `MarkdownExtractor` interface when local go-readability cannot produce Markdown.

## Story / Context

As the Worker, I need a paid fallback extractor for pages that local readability rejects or cannot convert, while keeping the default path local to control cost.

## Scope

This task includes:

- Verifying whether an official Jina Reader Go SDK exists at implementation time.
- Using an official Jina Reader Go SDK if one exists and supports the Reader API.
- Implementing a small internal Reader adapter only if no suitable official SDK exists.
- Implementing the shared Worker-owned `MarkdownExtractor` contract for the Jina extractor.
- Keeping Jina SDK/client or adapter types inside the Jina implementation.
- Calling Jina Reader with the article canonical URL.
- Supporting `JINA_ENABLED`.
- Supporting optional `JINA_API_KEY` for authenticated Reader requests.
- Mapping general Jina failures to `ARC-010`.
- Mapping insufficient balance failures to `ARC-011` when exposed by status, code, or response body.
- Tests for disabled Jina, successful fallback, general failure, timeout/transport failure, and insufficient balance.

## Out of Scope

This task does not include:

- Local go-readability extraction.
- Artifact writes.
- SQLite job state transitions.
- Gateway notification behavior.
- ReaderLM-v2 usage by default.
- Untagged or low-adoption third-party wrapper dependencies.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `MDEXT-001`.
- Jina Reader API documentation.
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker Jina `MarkdownExtractor` implementation.
- Jina Reader client or adapter isolated behind the extractor implementation.
- Provider failure classification for Jina fallback.
- Worker configuration support for optional `JINA_API_KEY`.

## Expected Affected Areas

```text
src/worker/go.mod
src/worker/
Worker configuration
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: official SDK is preferred
  Given an official Jina Reader Go SDK exists
  And it supports the Reader API needed by Archivist
  When the Jina fallback is implemented
  Then the Worker uses that SDK instead of a hand-rolled HTTP client

Scenario: no suitable official SDK exists
  Given no official Jina Reader Go SDK exists
  When the Jina fallback is implemented
  Then the Worker uses a small internal Reader adapter
  And does not depend on untagged third-party wrappers

Scenario: Jina fallback succeeds
  Given local extraction cannot produce Markdown
  And Jina Reader returns Markdown for the canonical URL
  When the fallback client runs
  Then it returns Markdown
  And reports Jina as the selected provider

Scenario: Jina reports insufficient balance
  Given Jina Reader responds with an insufficient balance error
  When the fallback client classifies the error
  Then it returns an ARC-011 failure
```

## Done When

- Jina SDK availability is verified and recorded in the task diary entry.
- Jina fallback implements `MarkdownExtractor`.
- Worker pipeline code does not import Jina SDK/client or adapter types.
- Jina fallback can return Markdown from the canonical URL.
- `JINA_ENABLED` gates fallback usage.
- `JINA_API_KEY` is loaded without being required when unauthenticated Reader use is acceptable.
- Jina failures map to ARC-coded public errors.
- Tests cover success, disabled fallback, general failure, and insufficient balance.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- Confirm current Jina SDK availability from official Jina documentation or repositories before choosing adapter implementation.

## Dependencies

Depends on:

- `MDEXT-001`

Blocks:

- `MDEXT-005`

## ExecPlan

ExecPlan:

```text
../plans/MDEXT-004-worker-jina-reader-fallback.execplan.md
```

## Open Questions

- None.

## Notes

- Current planning verification found `github.com/jina-ai/client-go`, but it targets older Jina client semantics and is not a Reader-specific SDK.
