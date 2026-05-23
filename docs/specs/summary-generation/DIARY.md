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

---

## 2026-05-09 — refactor: adopt imroc/req/v3

**Status:** completed

**Summary:** Switched `AnthropicAdapter` to accept an injected `*req.Client` (`github.com/imroc/req/v3`). The underlying `*http.Client` is extracted via `client.GetClient()` and passed to the Anthropic SDK via `option.WithHTTPClient`, keeping the SDK integration intact. The HTTP client is constructed once in the composition root (`pkg/app/NewApp`) and passed to all adapters. Convention promoted to `docs/conventions/WORKER.md`.

**Validation:** `go tool lefthook run build`, `format`, `lint`, `test` — all passed.

**Canonical Updates:**
- `docs/conventions/WORKER.md` (new "HTTP client" section; expanded "Composition Root and Poor Man's DI" section)

---

## 2026-05-09 — refactor: promote orchestration-owns-logging and error-helpers conventions

**Status:** completed

**Summary:** Promoted two cross-cutting decisions to canonical documents following the code review of SUMGEN-003:

1. **Orchestration-owns-logging**: `AnthropicAdapter` must not accept a `*slog.Logger` and must not emit `slog.Info` or `slog.Error` calls. Structured log entries for provider, model, provider request id, ARC code, `article_id`, `job_id`, canonical URL, duration, and artifact write result are owned exclusively by SUMGEN-004 pipeline orchestration. The existing logger field and all `a.logger.*` calls have been removed from the adapter. Information previously logged (model, request_id, status_code, error_type) is now returned in `SummarizerResult` fields (`RequestID`, `StatusCode`) for orchestration to consume.

2. **Error helpers in `errors.go`**: error-building infrastructure (ARC constants, `classifyError`, `classifyAPIError`, `isBillingError`, `isTooLargeError`) moved from `anthropic.go` to `src/worker/internal/summary/errors.go`.

3. **`SummarizerRequest` metadata fields**: `ArticleID`, `JobID`, and `URL` added to `SummarizerRequest` so orchestration can thread article context into log entries. These fields are unused by SUMGEN-003; SUMGEN-004 is responsible for populating them.

REVIEW.md findings SUMGEN-003-FIX-2 ("add duration logging") and SUMGEN-003-FIX-4 ("add article_id/job_id/url") are resolved. Duration is measured by SUMGEN-004 orchestration, not by the adapter.

**Canonical Updates:**
- `docs/conventions/WORKER.md` (new "Structured Logging" and "Error helpers" sections)
- `docs/conventions/ARTIFACTS.md` (artifact access layer logging responsibility)
- `docs/specs/summary-generation/tasks/SUMGEN-003-summarizer-provider-adapter.md` (Notes)
- `docs/specs/summary-generation/tasks/SUMGEN-004-worker-summary-pipeline-integration.md` (Notes)

## 2026-05-09 — SUMGEN-003: Post-Review Corrective Fixes

Status:
- completed

Summary:
- Closed the two remaining corrective items from the 2026-05-08 code review of SUMGEN-003: config struct-tag alignment and context-cancellation test coverage.

Changes:
- SUMGEN-003-FIX-1 (closed): Added explicit `config:"LLM_PROVIDER"`, `config:"LLM_MODEL"`, and `config:"LLM_API_KEY"` struct tags to the `LLM` struct in `src/worker/pkg/app/config/config.go`. Previously, configuro derived `APP_LLM_APIKEY` for the API key field, diverging from the uppercase snake_case convention used by all other config fields (`APP_JINA_API_KEY`, etc.). `load_test.go` updated to use the correct env var name `APP_LLM_API_KEY` in `TestLLMConfigurationLoadsFromEnvVars`.
- SUMGEN-003-FIX-3 (closed): Added `TestAnthropicAdapterContextCancellationIsARC013` to `src/worker/internal/summary/anthropic_test.go`. Cancels the context before calling `Summarize`; asserts `ResultStatusFailure` with `ErrorCodeProviderFailure` (`ARC-013`). No real network calls; test is self-contained.

Decisions:
- Explicit `config:` struct tags are now required for all fields in every config struct. The prior undocumented behaviour of configuro-derived names for nested structs is not relied upon anywhere in the codebase.

Validation:
- `go tool lefthook run lint && go tool lefthook run test` — passed.

Follow-ups:
- All open SUMGEN-003 corrective items are closed. SUMGEN-004 pipeline integration can proceed once SUMGEN-002 is done.

Canonical Updates:
- None. The config-tag requirement is already implicit in the GENERAL.md uppercase snake_case convention; no new canonical update is required beyond the code fix.

## 2026-05-17 — SUMGEN-003: Composition Root Summarizer Wiring

Status:
- completed

Summary:
- Corrected the remaining SUMGEN-003 completion gap by wiring the Anthropic summarizer adapter into `pkg/app.NewApp` and exposing it as `App.Summarizer`.
- Kept provider selection out of `app.go`; `config.Root.Validate()` remains responsible for rejecting unsupported `LLM_PROVIDER` values and missing Anthropic API keys.

Changes:
- Added `summary.SummarizerService` to the Worker composition root.
- Constructed `summary.NewAnthropicAdapter` unconditionally from the shared `req.Client`, `cfg.LLM.API.Key`, and `cfg.LLM.Model`.
- Extended `app_test.go` coverage to assert summarizer construction, Anthropic provider selection, and app-level unsupported-provider rejection.

Decisions:
- No new configuration keys, schemas, public APIs, or provider routing behavior were introduced.
- `anthropic` remains the only supported v0 provider.

Validation:
- `go test ./pkg/app/config ./pkg/app ./internal/summary` — passed.
- `go tool lefthook run build` — passed.
- `go tool lefthook run format` — passed.
- `go tool lefthook run lint` — passed.
- `go tool lefthook run test` — passed.

Follow-ups:
- SUMGEN-004 can consume `App.Summarizer` when pipeline integration proceeds.

Canonical Updates:
- None. The canonical task/spec already required this behavior; this entry records the corrective implementation.

## 2026-05-17 — SUMGEN-003: Anthropic Context Overflow Classification

Status:
- completed

Summary:
- Corrected Anthropic summary adapter error classification so provider signals for context-window overflow and request/token-size overflow map to `ARC-014` instead of falling through to generic `ARC-013`.

Changes:
- Extended `src/worker/internal/summary/errors.go` to classify Anthropic HTTP 400 `invalid_request_error` responses as `ARC-014` only when the decoded provider error message indicates context-window, prompt-length, request-size, or token-limit overflow.
- Extended `src/worker/internal/summary/anthropic.go` to classify successful responses with `stop_reason = "model_context_window_exceeded"` as `ARC-014`.
- Added adapter regression tests for invalid-request context overflow, unrelated invalid-request errors, and response-level context-window stop reason.

Decisions:
- No public interfaces, configuration keys, ARC codes, schemas, provider abstractions, task statuses, or feature plan rows changed.
- The invalid-request mapping is intentionally conservative: it requires Anthropic `invalid_request_error` plus size/context wording.

Validation:
- `go test ./internal/summary` — passed.
- `go tool lefthook run build` — passed.
- `go tool lefthook run format` — passed.
- `go tool lefthook run lint` — passed.
- `go tool lefthook run test` — passed.

Follow-ups:
- None.

Canonical Updates:
- None. `SPEC.md`, `SUMGEN-003`, and the linked ExecPlan already required context-window overflow and request-too-large failures to map to `ARC-014`.

## 2026-05-20 — SUMGEN-002: Worker Summary Artifact Access

Status:
- done

Summary:
- Extended the Worker artifact store with atomic `summary.md` writes using the existing rooted article artifact access and temp-file promotion path.
- Preserved existing deterministic `content.md` read behavior through `OpenMarkdown`.

Changes:
- Added `Store.WriteSummary(articleID, io.Reader)` for `{DATA_DIR}/articles/{article_id}/summary.md`.
- Added a `.summary.md.*.tmp` temp-file pattern and reused the existing article-root `MkdirAll`, `OpenRoot`, temp write, cleanup, and rename machinery.
- Wrapped summary write failures so callers can match `arc.ErrSummaryWrite` / `ARC-016` while still extracting `artifacts.StoreError` metadata.
- Added artifact-store tests for content read, missing content read, deterministic summary path, atomic summary promotion, failed summary cleanup, traversal rejection, symlink escape rejection, and absence of `summary.json`.

Decisions:
- No new artifact paths, configuration keys, schemas, public filesystem path APIs, or summary JSON behavior were introduced.
- The summary write operation maps store failures to `ARC-016`; content reads continue to surface `fs.ErrNotExist` through the existing `StoreError` pattern.

Validation:
- `go test ./internal/artifacts` — passed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed after concurrent UI test changes completed.

Follow-ups:
- `SUMGEN-004` can now consume deterministic `content.md` reads and atomic `summary.md` writes.

Canonical Updates:
- `docs/specs/summary-generation/tasks/SUMGEN-002-worker-summary-artifact-access.md` — status: done, validation recorded.
- `docs/specs/summary-generation/PLAN.md` — SUMGEN-002 row: done.
