---
id: SUMGEN-003-PLAN
task: ../tasks/SUMGEN-003-summarizer-provider-adapter.md
status: completed
canonical: true
---

# ExecPlan: SUMGEN-003 Summarizer Provider Adapter

## Objective

Implement the `SummarizerService` abstraction and first Anthropic Claude adapter while keeping Anthropic SDK types out of Worker pipeline orchestration.

## Linked Task

- `../tasks/SUMGEN-003-summarizer-provider-adapter.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/SUMGEN-003-summarizer-provider-adapter.md`

Add only ExecPlan-specific context:

- `docs/conventions/WORKER.md`
- `docs/conventions/ERRORS.md`
- Anthropic Client SDK documentation
- Anthropic Go SDK repository or documentation
- Anthropic API error documentation

## Assumptions

- `LLM_PROVIDER=anthropic` is the only supported v0 provider value.
- `LLM_MODEL` defaults to `claude-3-5-haiku-20241022`.
- `LLM_API_KEY` is required when using Anthropic.
- Planning verification found `github.com/anthropics/anthropic-sdk-go` as the official Go SDK; implementation must re-check suitability at execution time.

## Non-Goals

- Do not integrate the Worker pipeline.
- Do not read or write article artifacts.
- Do not create summary JSON.
- Do not chunk or truncate Markdown.
- Do not add a second provider implementation.

## Implementation Sequence

1. Re-check official Anthropic SDK availability and suitability for Go.
2. Add the official Anthropic Go SDK dependency if still suitable.
3. Define provider-neutral `SummarizerService` request/result types.
4. Define a fixed system prompt for text-only article summary output.
5. Implement Anthropic adapter construction from Worker configuration.
6. Call Claude with configured model and Markdown source content.
7. Extract text-only output from the provider response and reject empty output as `ARC-013`.
8. Map HTTP 402 `billing_error` to `ARC-015`.
9. Map HTTP 413 `request_too_large` and context-window overflow to `ARC-014`.
10. Map generic provider/API/transport/auth/rate-limit/overloaded/malformed-output failures to `ARC-013`.
11. Keep Anthropic SDK types private to the adapter package.
12. Add configuration and adapter tests for success and failure mapping.
13. Record SDK version and selection rationale in `DIARY.md` when implementation is performed.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Inspect imports from Worker pipeline code to confirm no Anthropic SDK types leak outside the adapter.
- Confirm provider request IDs are captured for logs when the SDK exposes them.

## Documentation Updates Required

- Update `../tasks/SUMGEN-003-summarizer-provider-adapter.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new provider configuration, prompt contract, or error mapping to canonical docs before relying on it.

## Risks

- Leaking SDK types into orchestration would make provider replacement harder.
- Treating billing or context failures as generic provider failures would hide actionable operator/user feedback.
- Logging full Markdown or summary content would expose private article data.

## Rollback / Recovery Notes

- Failed summary generation is terminal in v0; manual requeue is performed by sending the URL again.
- Disabling or changing `LLM_PROVIDER` without a supported provider should fail application startup or Worker configuration validation.

## Completion Criteria

- `SummarizerService` exists and is provider-neutral.
- Anthropic adapter uses official SDK when suitable.
- ARC mappings are tested.
- Worker validation passes.
