# Implementation Diary: Article Processing

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Entry Template

```md
## YYYY-MM-DD - TASK-ID: Task Title

Status:
- completed / partially completed / blocked / skipped

Summary:
- Brief outcome.

Changes:
- Files, schemas, or behavior changed.

Decisions:
- Decisions made during implementation.

Validation:
- Commands and manual checks run.

Follow-ups:
- Remaining work, if any.

Canonical Updates:
- Specs, plans, standards, architecture docs, or design decisions updated.
```

---

## Log

## 2026-05-03 - ARTPROC-001: Create Feature Spec And Plan Artifacts

Status:
- completed

Summary:
- Created the `article-processing` feature folder, feature spec, feature plan, task files, and orchestration ExecPlan.

Changes:
- Added canonical article-processing ALM artifacts under `docs/specs/article-processing/`.
- Updated `docs/specs/INDEX.md` with the new feature dependency.

Decisions:
- Snapshot success is an interim completion point until the v0 extraction/summarization feature supersedes it.

Validation:
- Inspected generated Markdown structure and cross-links.

Follow-ups:
- Implement Worker and Gateway tasks according to `PLAN.md`.

Canonical Updates:
- `SPEC.md`, `PLAN.md`, task files, ExecPlan, and feature index.

## 2026-05-03 - ARTPROC-002: Define Shared ARC Error-Code Convention

Status:
- completed

Summary:
- Created a shared article-processing error-code catalog.

Changes:
- Added `docs/conventions/ERRORS.md`.
- Updated general and worker conventions to reference ARC-coded persisted errors.

Decisions:
- User-facing article processing failures use stable `ARC-NNN` codes and do not expose low-level HTTP or filesystem details.

Validation:
- Inspected generated Markdown structure and references.

Follow-ups:
- Worker implementation must map resolver, fetch, and snapshot failures to the catalog.

Canonical Updates:
- `docs/conventions/ERRORS.md`, `docs/conventions/GENERAL.md`, and `docs/conventions/WORKER.md`.

## 2026-05-04 - ARTPROC-006: Gateway Snapshot Success Notification Bridge

Status:
- skipped

Summary:
- Skipped the snapshot-only Gateway success bridge because `markdown-extraction` supersedes snapshot-only terminal success before the bridge is implemented.

Changes:
- Updated the task status and article-processing plan to point the next terminal success notification work at Markdown completion.

Decisions:
- Snapshot-only success remains an interim concept in the original article-processing spec, but it is no longer the next executable notification bridge once Markdown extraction is planned.

Validation:
- Reviewed the updated task, plan, and Markdown extraction dependency references.

Follow-ups:
- Implement Markdown-complete notification behavior through `MDEXT-006`.

Canonical Updates:
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/article-processing/tasks/ARTPROC-006-gateway-snapshot-success-notification-bridge.md`

## 2026-05-04 - ARTPROC-DOC: Summary-Complete Final Success Alignment

Status:
- completed

Summary:
- Aligned article-processing docs with final v0 done semantics from summary generation.

Changes:
- Updated the feature spec, plan, `ARTPROC-005`, `ARTPROC-006`, and `ARTPROC-005` ExecPlan.

Decisions:
- Snapshot success is not final v0 success once downstream Markdown extraction and summary generation are implemented.
- Final success notification work is owned by `SUMGEN-005`.

Validation:
- Documentation consistency checked by repository search and review.

Follow-ups:
- Implement snapshot pipeline handoff behavior when downstream pipeline dependencies are available.

Canonical Updates:
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `docs/specs/article-processing/tasks/ARTPROC-006-gateway-snapshot-success-notification-bridge.md`
- `docs/specs/article-processing/plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md`

## 2026-05-06 - DOCS-SANITY: Snapshot Handoff Correction

Status:
- completed

Summary:
- Corrected article-processing docs so snapshot success is only a handoff to Markdown extraction in final v0.

Changes:
- Updated `SPEC.md`, `PLAN.md`, `ARTPROC-003`, `ARTPROC-005`, `ARTPROC-006`, and the `ARTPROC-005` ExecPlan.
- Added the artifact convention to required context where snapshot artifact behavior depends on it.

Decisions:
- Snapshot success writes `snapshot.html`, updates the canonical URL when known, and keeps the job running.
- Snapshot success does not mark the article ready, mark the job succeeded, or create success notifications in final v0.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no stale terminal-success wording at the snapshot boundary.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement `ARTPROC-005` as a Markdown handoff once upstream dependencies are complete.

Canonical Updates:
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/tasks/ARTPROC-003-worker-filesystem-artifact-access-layer.md`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md`
- `docs/specs/article-processing/tasks/ARTPROC-006-gateway-snapshot-success-notification-bridge.md`
- `docs/specs/article-processing/plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md`

## 2026-05-07 - ARTPROC-003: Worker Filesystem Artifact Access Layer

Status:
- completed

Summary:
- Implemented a reusable Worker artifact store for deterministic `snapshot.html` writes under `DATA_DIR`.

Changes:
- Added `src/worker/internal/artifacts.Store`.
- Added tests for deterministic snapshot paths, atomic temp-file promotion, cleanup on failed promotion, empty `DATA_DIR`, and traversal-style invalid article IDs.

Decisions:
- Article IDs are validated as strict uppercase ULID-style identifiers before any artifact path is derived.
- Snapshot writes use `os.Root` scoped to `DATA_DIR`, create `articles/{article_id}/` on demand, write a `.snapshot.html.*.tmp` file, and rename it to `snapshot.html`.

Validation:
- `cd src/worker && go tool lefthook run build`
- `cd src/worker && go tool lefthook run format`
- `cd src/worker && go tool lefthook run lint`
- `cd src/worker && go tool lefthook run test`

Follow-ups:
- `ARTPROC-004` can consume the artifact store after fetch succeeds.

Canonical Updates:
- `docs/specs/article-processing/PLAN.md`
- `docs/specs/article-processing/tasks/ARTPROC-003-worker-filesystem-artifact-access-layer.md`

## 2026-05-09 - ARTPROC-003: Worker Artifact Store — Operation-First Refactor

Status:
- completed (refactor of completed task)

Summary:
- Refactored `internal/artifacts.Store` to an operation-first interface.
  Callers now name the artifact they want to open or write; the Store owns
  all path construction, atomicity, and traversal resistance internally.
  This change also wires the Store into the App composition root and
  establishes the pattern for future artifact kinds.

Changes:
- Rewrote `src/worker/internal/artifacts/store.go`:
  - Removed `SnapshotPath` (path-returning API).
  - Changed `WriteSnapshot(articleID string, html []byte)` to
    `WriteSnapshot(articleID string, html io.Reader)`.
  - Added `OpenSnapshot(articleID string) (io.ReadCloser, error)`.
  - Extracted private `writeArtifact` and `openArtifact` helpers as the
    pattern all future artifact kinds will follow.
  - Renamed temp-file helper to `createTempFile` with prefix/suffix params.
- Added `src/worker/internal/artifacts/article_id.go`: moved filename
  constants (`ArticlesDirectoryName`, `SnapshotHTMLFilename`, etc.),
  `ValidateArticleID`, and `ErrInvalidArticleID` from
  `pkg/app/artifacts/paths.go` into `internal/artifacts`.
- Deleted `src/worker/pkg/app/artifacts/paths.go` and `paths_test.go`:
  `ArticlePaths` type removed; no remaining public callers (Gateway is C#
  and has its own `ArticleArtifactPaths.cs`).
- Updated `src/worker/pkg/app/app.go`:
  - Replaced `ArtifactPaths *artifacts.ArticlePaths` with
    `Artifacts *artifacts.Store`.
  - Wired `artifacts.NewStore(cfg.Artifacts.DataDir)` into `NewApp`.
  - Extended `App.Close()` to close the Store alongside SQLite.
- Updated test files accordingly; added round-trip, not-found, and
  mid-stream-error-cleanup test cases.
- Set `cfg.Artifacts.DataDir = t.TempDir()` where tests previously relied
  on the default `/data` path that is unavailable in CI.

Decisions:
- `io.Reader` in / `io.ReadCloser` out is the artifact I/O idiom.
  Streams directly to the temp file without buffering; composes with
  `http.Response.Body`, `strings.NewReader`, `bytes.Buffer`. Future
  artifact kinds follow the same two-method pattern over the shared private
  helpers.
- Filename constants and `ValidateArticleID` live in `internal/artifacts`,
  not in a public package, because no other Go consumer needs them.
- Atomic write (temp file + rename + cleanup on failure) remains a private
  implementation detail; callers never coordinate temp files or renames.

Validation:
- `cd src/worker && go tool lefthook run build`
- `cd src/worker && go tool lefthook run format`
- `cd src/worker && go tool lefthook run lint`
- `cd src/worker && go tool lefthook run test`

Follow-ups:
- ARTPROC-004 can consume `App.Artifacts.WriteSnapshot` after fetch
  succeeds.
- Future tasks adding `content.md`, `summary.md`, `summary.json`, and
  `metadata.json` each add two one-liner public methods over
  `openArtifact` / `writeArtifact`.

Canonical Updates:
- `docs/conventions/ARTIFACTS.md` — added "Access Interface" section
  establishing the operation-first contract as a binding convention.

---

## 2026-05-09 — refactor: promote orchestration-owns-logging convention

**Status:** completed

**Summary:** The orchestration-owns-logging decision (see markdown-extraction and summary-generation diary entries of the same date) applies equally to ARTPROC-005. Any provider adapter introduced under article processing must follow the same rules: no logger parameter, no `slog.Info`/`slog.Error` calls, sufficient result fields for orchestration to log. The artifact access layer must not log write results; it must return enough data for the pipeline stage to emit an artifact write result log entry.

**Canonical Updates:**
- `docs/conventions/WORKER.md` (new "Structured Logging" and "Error helpers" sections)
- `docs/conventions/ARTIFACTS.md` (artifact access layer logging responsibility)

## 2026-05-09 — ARTPROC-003: Post-Review Corrective Fixes

Status:
- completed

Summary:
- Closed all corrective items from the 2026-05-08 code review of ARTPROC-003. No outstanding corrective items remain.

Changes:
- FIX-1 (closed): `artifacts.Store` is now wired into the composition root. `App` exposes an `Artifacts *artifacts.Store` field; `NewApp` initialises it from `cfg.Artifacts.DataDir`; `App.Close()` closes the store alongside SQLite. See the 2026-05-09 operation-first refactor entry above for full detail.
- FIX-2 (closed): `app_test.go:34` asserts that `application.Artifacts` is non-nil after `NewApp`, satisfying the WORKER.md requirement for a composition-root test for every new service.
- FIX-3 (N/A): `pkg/app/artifacts/paths_test.go` was deleted as part of the composition-root refactor — `ArticlePaths` and its tests were removed entirely when `ValidateArticleID` and filename constants moved to `internal/artifacts/article_id.go`. The parallel-test fix is therefore moot; the file no longer exists.

Decisions:
- No new decisions. All fixes follow the operation-first refactor design already recorded in the 2026-05-09 diary entry above.

Validation:
- `go tool lefthook run lint && go tool lefthook run test` — passed.

Follow-ups:
- No outstanding ARTPROC-003 corrective items remain.
- ARTPROC-004 can consume `App.Artifacts.WriteSnapshot` after fetch succeeds.

Canonical Updates:
- None beyond the operation-first refactor entry above.

---

## 2026-05-09 - ARTPROC-004: Worker URL Resolver And HTML Fetcher

Status:
- completed

Summary:
- Implemented the Worker fetcher package at `src/worker/internal/fetcher/`.
- Added `github.com/imroc/req/v3 v3.57.0` to `src/worker/go.mod`.
- All 13 tests pass, covering: success with redirect, 401, 403, 404, non-HTML content type, oversized body, timeout, ftp:// scheme, file:// scheme, empty scheme, and 5xx responses.

Changes:
- `src/worker/internal/fetcher/fetcher.go`: fetcher service with `New()` and `Fetch()`, ARC-coded sentinel errors, scheme validation, redirect following, content-type and body-size limits.
- `src/worker/internal/fetcher/fetcher_test.go`: httptest-based tests for all required failure classes.
- `src/worker/go.mod` and `src/worker/go.sum`: added `github.com/imroc/req/v3` and transitive dependencies.

Decisions:
- ARC-coded sentinel errors are exported package-level vars so callers can use `errors.Is` without depending on internal types.
- Error message strings end with a period to satisfy `docs/conventions/ERRORS.md` message format; `//nolint:staticcheck` comments are applied per-var to suppress the staticcheck linter that normally disallows trailing punctuation in Go error strings.
- `resp.Response.Request.URL` (the embedded `*http.Response.Request`) is used to obtain the final URL after redirect chain, because it holds the last `*http.Request` that produced the response.
- 4xx errors other than 401/403/404 (e.g. 410 Gone, 429 Too Many Requests) map to `ARC-004` (transient) as the most reasonable fallback; this decision is local and reversible.
- Used `errors.AsType[*url.Error]` (Go 1.26 generic form) as required by `docs/conventions/WORKER.md`.

Validation:
- `cd src/worker && go tool lefthook run build` — gobuild and dotnet pass.
- `cd src/worker && go tool lefthook run format` — golangci and dotnet pass.
- `cd src/worker && go tool lefthook run lint` — golangci and dotnet pass.
- `cd src/worker && go tool lefthook run test` — gotest and dotnet pass; all 13 `internal/fetcher` tests pass.

Follow-ups:
- `ARTPROC-005` can now unblock once TELING-001 and TELING-003 are confirmed done (they are).

Canonical Updates:
- `docs/specs/article-processing/tasks/ARTPROC-004-worker-url-resolver-and-html-fetcher.md` status: blocked → done.
- `docs/specs/article-processing/PLAN.md` ARTPROC-004 row: blocked → done.
