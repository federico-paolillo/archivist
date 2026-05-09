# Implementation Diary: Markdown Extraction With Fallbacks

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-05-04 — MDEXT-001: Create Feature Artifacts And Contracts

Status:
- completed

Summary:
- Created the Markdown extraction feature specification, task plan, task files, and Worker pipeline ExecPlan.

Changes:
- Added canonical Markdown extraction behavior with go-readability v2 first and Jina Reader fallback.
- Added artifact-path conventions for article artifacts.
- Extended architecture, design, configuration, logging, and error-code conventions.

Decisions:
- Markdown extraction, not HTML snapshotting, is the current terminal success point until the future summary feature supersedes it.
- `content.md` remains under `{DATA_DIR}/articles/{article_id}/` following the original artifact path convention.
- v0 extraction uses deterministic local-first fallback instead of candidate scoring.
- Jina integration should use an official Reader Go SDK if one exists at implementation time; otherwise use a small internal adapter.

Validation:
- Documentation consistency checked by repository search and review.

Follow-ups:
- Implement `MDEXT-002` through `MDEXT-006` when dependencies are satisfied.

Canonical Updates:
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/markdown-extraction/PLAN.md`
- `docs/specs/markdown-extraction/tasks/*.md`
- `docs/specs/markdown-extraction/plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

## 2026-05-04 — MDEXT-DOC: Provider Boundary And Summary Supersession

Status:
- completed

Summary:
- Amended Markdown extraction planning so local and Jina extraction run behind `MarkdownExtractor`, and summary generation supersedes Markdown-complete terminal success.

Changes:
- Updated `SPEC.md`, `PLAN.md`, `MDEXT-003`, `MDEXT-004`, `MDEXT-005`, and the `MDEXT-005` ExecPlan.
- Added the `MDEXT-004` Jina fallback ExecPlan.
- Marked `MDEXT-006` skipped because `SUMGEN-005` owns final success notifications.

Decisions:
- Pipeline orchestration depends on `MarkdownExtractor`, not go-readability or Jina SDK/client types.
- Jina uses an official suitable SDK when one exists; otherwise the Worker uses a small internal Reader adapter.
- Markdown completion is intermediate once summary generation is implemented.

Validation:
- Documentation consistency checked by repository search and review.

Follow-ups:
- Implement `MDEXT-004` with SDK availability verification at execution time.
- Implement `SUMGEN-002` after `MDEXT-005` is done.

Canonical Updates:
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/markdown-extraction/PLAN.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-003-worker-go-readability-extraction.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-004-worker-jina-reader-fallback.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-006-gateway-markdown-success-notification.md`
- `docs/specs/markdown-extraction/plans/MDEXT-004-worker-jina-reader-fallback.execplan.md`
- `docs/specs/markdown-extraction/plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`

## 2026-05-06 — DOCS-SANITY: Markdown Handoff Correction

Status:
- completed

Summary:
- Corrected Markdown extraction docs so Markdown success is only a handoff to summary generation in final v0.

Changes:
- Updated `SPEC.md`, `PLAN.md`, `MDEXT-004`, `MDEXT-005`, `MDEXT-006`, and the `MDEXT-004`/`MDEXT-005` ExecPlans.
- Fixed `MDEXT-005` to block `SUMGEN-002` and linked `MDEXT-004` to its accepted ExecPlan.

Decisions:
- Markdown success writes `content.md` and keeps the job running for summary generation.
- Markdown success does not mark the article ready, mark the job succeeded, or create success notifications in final v0.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no stale terminal-success wording at the Markdown boundary.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement `SUMGEN-002` only after `MDEXT-005` is complete.

Canonical Updates:
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/markdown-extraction/PLAN.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-004-worker-jina-reader-fallback.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-006-gateway-markdown-success-notification.md`
- `docs/specs/markdown-extraction/plans/MDEXT-004-worker-jina-reader-fallback.execplan.md`
- `docs/specs/markdown-extraction/plans/MDEXT-005-worker-markdown-pipeline-integration.execplan.md`

## 2026-05-07 — MDEXT-003: Worker Go-Readability Extraction

Status:
- completed

Summary:
- Added a Worker-owned `MarkdownExtractor` contract and local go-readability extractor implementation.
- Added local HTML-to-Markdown conversion and typed local unreadable classification.

Changes:
- Added `src/worker/internal/markdown` with neutral extraction input/result types.
- Implemented local extraction with `codeberg.org/readeck/go-readability/v2`, `CheckDocument()`, and `github.com/JohannesKaufmann/html-to-markdown/v2`.
- Mapped local parse, extraction, conversion, and empty Markdown failures to `ARC-009`.
- Added tests for readable HTML, unreadable HTML, parse failure, and conversion failure.
- Added task-owned Worker dependencies required by the local extractor.

Decisions:
- Local unreadable results are non-terminal provider results so later Jina fallback can attach without changing the local extractor.
- Local extractor failures carry `ARC-009` and a diagnostic reason for logs, while the shared result model avoids provider library types.

Validation:
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.

Follow-ups:
- `MDEXT-004` should attach Jina behind the same `MarkdownExtractor` contract.
- `MDEXT-005` should integrate provider selection and artifact writes without importing provider-specific types.

Canonical Updates:
- `docs/specs/markdown-extraction/PLAN.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-003-worker-go-readability-extraction.md`

## 2026-05-07 — MDEXT-004: Worker Jina Reader Fallback

Status:
- completed

Summary:
- Implemented Jina Reader fallback extractor behind the `MarkdownExtractor` interface using a small internal HTTP adapter.
- Added `JINA_ENABLED` and `JINA_API_KEY` configuration fields.
- Wired `JinaExtractor` into the composition root.

Changes:
- Added `src/worker/internal/markdown/jina.go`: `JinaExtractor` implementing `MarkdownExtractor` with internal HTTP adapter for `https://r.jina.ai/<url>`.
- Added `src/worker/internal/markdown/jina_test.go`: tests for disabled fallback, successful extraction, API key passthrough, no-key path, general failure (ARC-010), transport failure (ARC-010), insufficient balance (ARC-011), and Accept header.
- Extended `src/worker/pkg/app/config/config.go` with `Jina` struct (`Enabled bool`, `APIKey string`) loaded from `APP_JINA_JINA__ENABLED` and `APP_JINA_JINA__API__KEY` env vars (configuro double-underscore convention for underscored tag names).
- Extended `src/worker/pkg/app/config/load_test.go` with Jina default and env var loading tests.
- Extended `src/worker/pkg/app/app.go` to construct and expose `JinaExtractor` in the composition root.
- Extended `src/worker/pkg/app/app_test.go` with tests verifying `JinaExtractor` is created and defaults to disabled.

Decisions:
- SDK selection: No official Jina Reader Go SDK exists as of 2026-05-07. The Jina AI GitHub organization provides TypeScript and Python implementations only. The `github.com/jina-ai/client-go` repo targets older Jina client semantics and is not a Reader API SDK. A small internal HTTP adapter was implemented using only the Go standard library (`net/http`, `io`), satisfying REQ-009 through REQ-011.
- The `JinaExtractor.baseURL` field is exported only within the package to allow test injection via `httptest.Server` without exposing it publicly.
- HTTP 402 (Payment Required) is mapped to ARC-011 (insufficient balance). All other non-200 responses and transport errors map to ARC-010.
- The `Accept: text/plain` header is sent to request Markdown output from Jina Reader.
- `Authorization: Bearer <key>` is only sent when `APIKey` is non-empty, preserving unauthenticated free-tier usage.
- The disabled path returns a `ResultStatusFailure` result with `ARC-010` rather than panicking or silently succeeding, so callers can observe disabled-extractor behavior consistently.

Validation:
- `cd src/worker && go build ./...` passed.
- `cd src/worker && go test ./...` passed (all 9 packages).
- `cd src/worker && go tool golangci-lint run ./internal/markdown/... ./pkg/app/...` passed (no issues in changed files).
- `cd src/worker && gofmt -l ...` passed (no formatting issues).
- Lefthook `build` step: Go and dotnet passed; npm failed with `tsc: command not found` (pre-existing UI toolchain issue, unrelated to this task).
- Lefthook `format` step: dotnet passed; biome failed with `command not found` (pre-existing UI toolchain issue); golangci errors reference `w1-integration` worktree file not present in this worktree (pre-existing, unrelated to this task).
- Lefthook `test` step: Go and dotnet passed; vitest failed with `command not found` (pre-existing UI toolchain issue).

Follow-ups:
- `MDEXT-005` should select between `GoReadabilityExtractor` and `JinaExtractor` at pipeline orchestration level without importing either concrete type directly (use `MarkdownExtractor` interface).

Canonical Updates:
- `docs/specs/markdown-extraction/PLAN.md`
- `docs/specs/markdown-extraction/tasks/MDEXT-004-worker-jina-reader-fallback.md`
- `docs/specs/markdown-extraction/plans/MDEXT-004-worker-jina-reader-fallback.execplan.md`

---

## 2026-05-09 — refactor: adopt imroc/req/v3

**Status:** completed

**Summary:** Switched `JinaExtractor` from a bare `*http.Client` to an injected `*req.Client` (`github.com/imroc/req/v3`). The HTTP client is now constructed once in the composition root (`pkg/app/NewApp`) and injected via the updated `NewJinaExtractor(client *req.Client, ...)` constructor. The convention mandating this approach for all Worker outbound HTTP has been promoted to `docs/conventions/WORKER.md` (new "HTTP client" section).

**Validation:** `go tool lefthook run build`, `format`, `lint`, `test` — all passed.

**Canonical Updates:**
- `docs/conventions/WORKER.md` (new "HTTP client" section; expanded "Composition Root and Poor Man's DI" section)

---

## 2026-05-09 — refactor: promote orchestration-owns-logging and error-helpers conventions

**Status:** completed

**Summary:** Promoted two cross-cutting decisions to canonical documents following the code review of MDEXT-003 and MDEXT-004:

1. **Orchestration-owns-logging**: `GoReadabilityExtractor` and `JinaExtractor` must not accept a `*slog.Logger` and must not emit `slog.Info` or `slog.Error` calls. Structured log entries for provider attempt, fallback reason, selected provider, ARC code, `article_id`, `job_id`, canonical URL, duration, and artifact write result are owned exclusively by MDEXT-005 pipeline orchestration. Adapters return result types (`ExtractResult`) with sufficient fields for orchestration to log everything.

2. **Error helpers in `errors.go`**: all error-building infrastructure for a package (ARC constants, error constructors, classification helpers) must live in `<package>/errors.go`. Applied to `src/worker/internal/markdown/errors.go`: error helpers from `jina.go` and `go_readability.go` moved there.

The REVIEW.md finding MDEXT-003-FIX-3 ("document logging deferral") is resolved by this diary entry and the canonical convention update. MDEXT-004-FIX-2 ("inject logger into Jina") is superseded — no logger injection is required.

**Canonical Updates:**
- `docs/conventions/WORKER.md` (new "Structured Logging" and "Error helpers" sections)
- `docs/conventions/ARTIFACTS.md` (artifact access layer logging responsibility)
- `docs/specs/markdown-extraction/tasks/MDEXT-003-worker-go-readability-extraction.md` (Notes)
- `docs/specs/markdown-extraction/tasks/MDEXT-004-worker-jina-reader-fallback.md` (Notes)
- `docs/specs/markdown-extraction/tasks/MDEXT-005-worker-markdown-pipeline-integration.md` (Notes)

## 2026-05-09 — MDEXT-003, MDEXT-004: Post-Review Corrective Fixes

Status:
- completed

Summary:
- Closed corrective test-coverage items for MDEXT-003 and MDEXT-004 identified during the 2026-05-08 code review. MDEXT-004-FIX-1 (HTTP timeout) was resolved by a prior refactor and required no further change.

Changes:
- MDEXT-003-FIX-1 (closed): Added `TestGoReadabilityExtractorRejectsInvalidCanonicalURL` to `src/worker/internal/markdown/go_readability_test.go`. Passes an unparseable URL string and asserts `ResultStatusFailure` with `ErrorCodeLocalExtractionFailed`, covering the `url.ParseRequestURI` failure branch at `go_readability.go:35-37`.
- MDEXT-003-FIX-2 (closed): Added `TestGoReadabilityExtractorRejectsEmptyMarkdown` to `src/worker/internal/markdown/go_readability_test.go`. Injects a `convert` function returning `("", nil)` and asserts `ResultStatusFailure` with `ErrorCodeLocalExtractionFailed`, covering the empty-Markdown guard at `go_readability.go:77-79`.
- MDEXT-004-FIX-1 (already resolved): HTTP timeout is satisfied by the shared `*req.Client` constructed with `SetTimeout(30s)` in `app.go` and injected into `JinaExtractor`. No change to `jina.go` was needed.
- MDEXT-004-FIX-3 (closed): Added `TestJinaExtractorEmptyResponseMapsToARC010` to `src/worker/internal/markdown/jina_test.go`. Uses `httptest.NewServer` returning 200 OK with an empty body; asserts `result.Status == ResultStatusFailure` and `result.ErrorCode == ErrorCodeJinaFailed`, covering the empty-body path at `jina.go:53-56`.

Decisions:
- HTTP timeout for `JinaExtractor` is governed by the shared `*req.Client` in the composition root, not by a per-extractor timeout field. This is consistent with the WORKER.md "HTTP client" convention: the client is constructed once in `NewApp` and injected into all adapters.

Validation:
- `go tool lefthook run lint && go tool lefthook run test` — passed.

Follow-ups:
- All open MDEXT-003 and MDEXT-004 corrective items are closed. MDEXT-005 pipeline integration can proceed.

Canonical Updates:
- None. No new durable decisions were made; existing conventions cover all changes.
