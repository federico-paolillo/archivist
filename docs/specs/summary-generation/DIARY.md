# Implementation Diary: Summary Generation

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-05-04 — SUMGEN-001: Create Feature Artifacts And Contracts

Status:
- completed

Summary:
- Created the summary-generation feature planning artifacts and promoted summary completion, text-only summary artifacts, provider SDK boundaries, and final v0 success semantics to canonical docs.

Changes:
- Added `SPEC.md`, `PLAN.md`, task files, and ExecPlans for summary generation.
- Updated feature index, architecture, design decisions, artifact conventions, error catalog, and Worker/Gateway conventions.
- Amended Markdown extraction planning to use `MarkdownExtractor` and added a Jina fallback ExecPlan.

Decisions:
- Final v0 success is summary-complete, not snapshot-complete or Markdown-complete.
- v0 summary output is text-only and persisted as `summary.md`.
- Provider SDKs are required when official and suitable; custom HTTP adapters are fallback only.

Validation:
- Planned validation is documentation-focused for this task. Production validation belongs to implementation tasks.

Follow-ups:
- Implement blocked dependency tasks before Worker summary pipeline integration.
- Re-check Jina SDK availability during `MDEXT-004` execution.

Canonical Updates:
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/markdown-extraction/PLAN.md`

## 2026-05-06 — DOCS-SANITY: Summary Final Success Contract

Status:
- completed

Summary:
- Completed the summary-generation docs so summary completion is the only final v0 success path.

Changes:
- Added and linked the `SUMGEN-005` ExecPlan.
- Replaced the invalid Anthropic model ID with `claude-3-5-haiku-20241022` while retaining `LLM_MODEL` override support.
- Accepted the ready `SUMGEN-003` ExecPlan.

Decisions:
- Summary-complete processing owns final article/job success and Gateway success notification content.
- Snapshot and Markdown stages are intermediate handoffs.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no invalid model ID and no ready task linked to a proposed ExecPlan.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement `SUMGEN-005` after Worker summary completion and the narrowed Telegram dispatcher exist.

Canonical Updates:
- `docs/specs/summary-generation/SPEC.md`
- `docs/specs/summary-generation/PLAN.md`
- `docs/specs/summary-generation/tasks/SUMGEN-001-create-feature-artifacts-and-contracts.md`
- `docs/specs/summary-generation/tasks/SUMGEN-003-summarizer-provider-adapter.md`
- `docs/specs/summary-generation/tasks/SUMGEN-005-gateway-summary-notification.md`
- `docs/specs/summary-generation/plans/SUMGEN-003-summarizer-provider-adapter.execplan.md`
- `docs/specs/summary-generation/plans/SUMGEN-005-gateway-summary-notification.execplan.md`

## 2026-05-07 — SUMGEN-003: Summarizer Provider Adapter

Status:
- completed

Summary:
- Implemented the Worker-owned `SummarizerService` abstraction and the Anthropic Claude provider adapter using the official `github.com/anthropics/anthropic-sdk-go` SDK at v1.38.0.

Changes:
- Created `src/worker/internal/summary/contract.go`: `SummarizerService` interface, `SummarizerRequest`, `SummarizerResult`, `Provider`, `ResultStatus`, and `ErrorCode` types mirroring the MarkdownExtractor pattern.
- Created `src/worker/internal/summary/anthropic.go`: `AnthropicAdapter` with private SDK types, compile-time `var _ SummarizerService = (*AnthropicAdapter)(nil)` assertion, `NewAnthropicAdapter`, `NewAnthropicAdapterWithBaseURL` (for test injection), error classification into ARC-013/ARC-014/ARC-015.
- Created `src/worker/internal/summary/anthropic_test.go`: httptest-server-based tests for success path, empty output (ARC-013), generic API error (ARC-013), HTTP 413 request too large (ARC-014), HTTP 402 billing error (ARC-015), and SDK isolation assertion.
- Extended `src/worker/pkg/app/config/config.go`: Added `LLM` struct with `Provider` (default "anthropic"), `Model` (default "claude-3-5-haiku-20241022"), `APIKey` (no default, never logged).
- Updated `src/worker/pkg/app/config/load_test.go`: Added LLM config default and env var loading tests.
- Updated `src/worker/pkg/app/app.go`: Added `Summarizer summary.SummarizerService` field to `App`; added `createSummarizer` factory; wired into `NewApp`. Unsupported provider fails at startup.
- Updated `src/worker/pkg/app/app_test.go`: Added assertion that `Summarizer` is non-nil; added test for unsupported provider failing startup.
- Ran `go mod tidy` to add new transitive dependencies for `anthropic-sdk-go`.

Decisions:
- SDK selection: `github.com/anthropics/anthropic-sdk-go` v1.38.0 confirmed suitable at implementation time. Already present in go.mod/go.sum from prior dependency resolution.
- Error classification: HTTP 413 → ARC-014; HTTP 402 or `billing_error` error type → ARC-015; all others (including empty output, transport failures, auth failures) → ARC-013. Context window overflow cannot be reliably distinguished from other invalid_request_error responses without inspecting the message body, so it falls to ARC-013.
- API key validation: missing API key does NOT fail `NewApp`. The adapter is created with an empty key and will return ARC-013 (via HTTP 401) if called. Unsupported provider DOES fail startup. This preserves backward compatibility with existing runner tests that use default config without an API key.
- `ireturn` linter: `createSummarizer` returns `*summary.AnthropicAdapter` (concrete type) to satisfy the `ireturn` linter. The `App.Summarizer` field holds the `summary.SummarizerService` interface.
- Config env var mapping: configuro with `APP_` prefix maps `cfg.LLM.APIKey` to `APP_LLM_APIKEY` (not `APP_LLM_API_KEY`). The `config:` struct tag does not work as expected for nested structs with underscored logical names. Tests use `APP_LLM_APIKEY`.
- HTTP status codes: replaced magic numbers (402, 413) with `net/http` constants to satisfy `usestdlibvars` linter.
- `NewAnthropicAdapterWithBaseURL` is exported to allow httptest server injection in tests. It is a deliberate test-support constructor, not part of the production API.

Validation:
- `go build ./...` passed.
- `go test -race -shuffle=on ./...` passed: all packages including `internal/summary`, `pkg/app`, `pkg/app/config`, `internal/runner`.
- `go tool lefthook run lint` passed for golangci (Go linter); biome (UI) and dotnet build also passed. biome failure is pre-existing toolchain absence.
- `go tool lefthook run format` passed for golangci --fix and dotnet format.
- `go tool lefthook run test` passed for gotest and dotnet test.

Follow-ups:
- SUMGEN-004 (Worker summary pipeline integration) can proceed once SUMGEN-002 is also done.
- The config env var naming (APP_LLM_APIKEY vs APP_LLM_API_KEY) is a known discrepancy with the GENERAL.md convention. A future task should audit all env var mappings or switch to a config library that handles underscores more intuitively.

Canonical Updates:
- `docs/specs/summary-generation/tasks/SUMGEN-003-summarizer-provider-adapter.md` (status: done)
- `docs/specs/summary-generation/plans/SUMGEN-003-summarizer-provider-adapter.execplan.md` (status: completed)
- `docs/specs/summary-generation/PLAN.md` (SUMGEN-003 row: done)
