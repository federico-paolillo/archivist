---
id: SUMGEN-003
feature: summary-generation
title: Summarizer Provider Adapter
status: done
depends_on: [SUMGEN-001]
blocks: [SUMGEN-004]
parallel: true
exec_plan: ../plans/SUMGEN-003-summarizer-provider-adapter.execplan.md
canonical: true
---

# SUMGEN-003: Summarizer Provider Adapter

## Objective

Implement the Worker-owned `SummarizerService` abstraction and the first Anthropic Claude provider adapter using official Anthropic SDKs when suitable.

## Story / Context

As the Worker, I need a summarizer boundary that can call Claude now without tying pipeline orchestration to Anthropic SDK types.

## Scope

This task includes:

- Defining `SummarizerService` and provider-neutral request/result types.
- Implementing an Anthropic adapter behind `SummarizerService`.
- Using `github.com/anthropics/anthropic-sdk-go` in Go if it remains suitable at implementation time.
- Supporting `LLM_PROVIDER=anthropic`.
- Supporting required `LLM_API_KEY` for Anthropic.
- Supporting default `LLM_MODEL=claude-3-5-haiku-20241022`.
- Sending Markdown source with a fixed system prompt requesting text-only output.
- Mapping Anthropic billing failures to `ARC-015`.
- Mapping request-too-large or context overflow failures to `ARC-014`.
- Mapping generic provider, API, timeout, transport, permission, authentication, rate-limit, overloaded, and malformed-output failures to `ARC-013` unless a more specific code applies.
- Tests for configuration, success, text-only output handling, and error classification.

## Out of Scope

This task does not include:

- Reading or writing article artifacts.
- SQLite job state transitions.
- Gateway notification dispatch.
- Multiple summarizer providers beyond the abstraction.
- Chunked or truncated summarization.
- Structured summary JSON.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `SUMGEN-001`.
- Anthropic Client SDK documentation.
- Anthropic Go SDK documentation.
- Anthropic API error documentation.
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker `SummarizerService` contract.
- Worker Anthropic summarizer adapter.
- Worker configuration support for summary provider settings.
- Provider failure classification for summarization.

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
- `../plans/SUMGEN-003-summarizer-provider-adapter.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: official Anthropic SDK is used
  Given a suitable official Anthropic Go SDK exists
  When the Anthropic summarizer adapter is implemented
  Then the Worker uses that SDK instead of a hand-rolled HTTP client
  And Anthropic SDK types do not leak into pipeline orchestration

Scenario: Claude summary succeeds
  Given Markdown source text
  And Anthropic returns text content
  When the summarizer runs
  Then it returns text-only summary content

Scenario: Anthropic billing fails
  Given Anthropic returns HTTP 402 billing_error
  When the adapter classifies the error
  Then it returns an ARC-015 failure

Scenario: Anthropic request is too large
  Given Anthropic returns HTTP 413 request_too_large
  When the adapter classifies the error
  Then it returns an ARC-014 failure
```

## Done When

- `SummarizerService` exists and is provider-neutral.
- Anthropic adapter uses an official SDK when suitable.
- Anthropic SDK types stay inside the adapter.
- Configuration loads provider, model, and API key correctly.
- Error mapping covers `ARC-013`, `ARC-014`, and `ARC-015`.
- Tests cover success, configuration, and failure classification.
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

- Confirm current Anthropic SDK suitability from official Anthropic documentation or repositories before implementation.

## Dependencies

Depends on:

- `SUMGEN-001`

Blocks:

- `SUMGEN-004`

## ExecPlan

ExecPlan:

```text
../plans/SUMGEN-003-summarizer-provider-adapter.execplan.md
```

## Open Questions

- None.

## Notes

- Do not implement chunking or source truncation in this task.
- `AnthropicAdapter` does not accept a logger and must not emit `slog.Info` or `slog.Error` calls. Structured logging for provider, model, request id, ARC code, duration, and article context is owned by SUMGEN-004 pipeline orchestration per `.agents/skills/archivist-worker/SKILL.md`.
- `SummarizerRequest` carries `ArticleID`, `JobID`, and `URL` metadata fields so orchestration can thread article context into log entries without a second interface change.
- Worker runtime configuration key reconciliation is corrected by `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`; SUMGEN-004 must consume LLM settings from that canonical config shape.
