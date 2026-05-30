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
- Keeping Jina callable whenever local extraction needs fallback, without a runtime toggle.
- Supporting required `JINA_API_KEY` for authenticated Reader requests.
- Mapping general Jina failures to `ARC-010`.
- Mapping insufficient balance failures to `ARC-011` when exposed by status, code, or response body.
- Tests for successful fallback, general failure, timeout/transport failure, insufficient balance, bounded response reads, and non-text success responses.

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
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker Jina `MarkdownExtractor` implementation.
- Jina Reader client or adapter isolated behind the extractor implementation.
- Provider failure classification for Jina fallback.
- Worker configuration support for required `JINA_API_KEY`.

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
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

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
- Jina fallback has no runtime toggle.
- `JINA_API_KEY` is loaded as required Worker configuration.
- Jina failures map to ARC-coded public errors.
- Tests cover success, general failure, and insufficient balance.
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
- `JinaExtractor` does not accept a logger and must not emit structured log entries. Structured logging for provider attempt, selected provider, ARC code, duration, and artifact write result is owned by MDEXT-005 pipeline orchestration per `.agents/skills/archivist-worker/SKILL.md`.
- Worker runtime configuration key reconciliation is corrected by `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`; rebuilds must use the canonical `ARCHIVIST_` mapping there instead of historical implementation notes.
- Jina Reader responses must be read through hard limits: successful Markdown responses are capped at 10 MiB and non-OK diagnostic bodies are capped at 64 KiB. Successful responses must have a text Markdown-compatible content type (`text/plain`, `text/markdown`, or `text/x-markdown`). Oversized bodies, missing or invalid success content types, and non-text success content types map to `ARC-010`.
- Jina insufficient balance must map to `ARC-011` when exposed by HTTP 402 or by bounded non-OK response text/JSON containing known insufficient-balance markers. Other non-OK responses remain `ARC-010`.
