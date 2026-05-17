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

---

## 2026-05-09 — ARTPROC-004: Post-Review Corrective Fixes (Phase C)

Status:
- completed (correction, not a new feature)

Summary:
- Addressed all 7 reviewer findings for ARTPROC-004 identified after Wave 2 integration. The fetcher package now conforms to the WORKER.md `*req.Client` injection convention, exposes a proper per-package `errors.go`, and has a clean test file.

Changes:

Finding 1 — MAJOR: `New()` global client removed. `fetcher.New()` now accepts `*req.Client` and stores it; all client configuration (`SetRedirectPolicy`, `SetTimeout`, `DisableForceHttpVersion`) moved to `app.go`.
- `src/worker/internal/fetcher/fetcher.go` line 39: `func New(client *req.Client) *Fetcher`

Finding 2 — MAJOR: Fetcher wired into composition root. `App` gains a `Fetcher *fetcher.Fetcher` field. `NewApp` constructs a single shared `*req.Client` and calls `fetcher.New(httpClient)` unconditionally (not gated on `SqlitePath`).
- `src/worker/pkg/app/app.go`: added `fetcher` and `req` imports, `Fetcher` field, `httpClient` construction, `fetcher.New` call.
- `src/worker/pkg/app/app_test.go`: added `require.NotNil(t, application.Fetcher)` assertion.

Finding 3 — MAJOR: Per-package `errors.go` created. All 6 `Err*` sentinel vars and the `classifyHTTPStatus` / `classifyRequestError` functions moved from `fetcher.go` to a new `src/worker/internal/fetcher/errors.go`.
- New file: `src/worker/internal/fetcher/errors.go`
- Removed from: `src/worker/internal/fetcher/fetcher.go`

Finding 4 — MINOR: Dead branches in `classifyRequestError` collapsed. All three branches returned `ErrTransientFailure`; the function now has a single unconditional return and uses `_ error` to signal the parameter is intentionally unused.
- `src/worker/internal/fetcher/errors.go`: `func classifyRequestError(_ error) error { return ErrTransientFailure }`

Finding 5 — MINOR: Superfluous `newFetcher()` test helper removed. All call sites replaced with `fetcher.New(req.NewClient())` inline; `github.com/imroc/req/v3` import added to the test file.
- `src/worker/internal/fetcher/fetcher_test.go`: removed helper, added import, inlined calls.

Finding 6 — MINOR: Misleading comment at `TestFetchOversizedBodyReturnsARC006` deleted. The comment ("We stream it directly so we do not allocate a full 10 MiB in the test process") was inaccurate; the test name is self-describing.
- `src/worker/internal/fetcher/fetcher_test.go`: comment removed.

Finding 7 — NIT: Package docstring trimmed to a single concise line per the project's "no comment unless WHY is non-obvious" policy.
- `src/worker/internal/fetcher/fetcher.go` line 1: `// Package fetcher fetches bounded HTML content and maps failures to ARC error codes.`

Decisions:
- No new durable decisions. All fixes apply existing WORKER.md conventions.

Validation:
- `cd src/worker && go tool lefthook run build` — gobuild ✔, dotnet ✔ (npm pre-existing failure unrelated).
- `cd src/worker && go tool lefthook run format` — golangci ✔, dotnet ✔ (biome pre-existing failure unrelated).
- `cd src/worker && go tool lefthook run lint` — golangci ✔, dotnet ✔.
- `cd src/worker && go tool lefthook run test` — gotest ✔ (all packages), dotnet ✔ (vitest pre-existing failure unrelated).

Follow-ups:
- No outstanding ARTPROC-004 corrective items remain.

Canonical Updates:
- None. All changes are corrections to existing implementation; no durable decisions changed.

---

## 2026-05-10 - ARTPROC-005: Worker Snapshot Pipeline Orchestration

Status:
- completed

Summary:
- Implemented the Worker snapshot pipeline that claims queued article-processing jobs,
  resolves URLs, fetches HTML, writes `snapshot.html` atomically, updates `canonical_url`,
  invokes the Markdown extraction handoff point, and commits ARC-coded terminal failure state
  transactionally when any stage fails. Snapshot success is non-terminal in final v0.

Changes:
- New `src/worker/internal/pipeline/snapshot.go`:
  - `SnapshotPipeline` struct with `ProcessOne(ctx)`.
  - Pipeline stages: fetch HTML → write snapshot → update canonical URL → markdown handoff → no-op rating slot.
  - `MarkdownHandoff` function type (the MDEXT-005 extension point) with `NoOpMarkdownHandoff` placeholder.
  - `persistFailure` helper for structured logging + transactional terminal failure via `jobs.Repository.CompleteTerminal`.
  - `isARCError` / `arcCode` helpers for ARC-coded error detection and log field extraction.
  - Snapshot success does not mark `articles.status = ready`, `jobs.status = succeeded`, or insert a success notification.
- New `src/worker/internal/pipeline/errors.go`:
  - `ErrSnapshotWrite` (`[ARC-007]`) and `ErrUnknown` (`[ARC-999]`) sentinel errors.
- Extended `src/worker/pkg/jobs/repository.go`:
  - Added `ArticleURL(ctx, articleID)` and `UpdateCanonicalURL(ctx, articleID, canonicalURL)` to the `Repository` interface and `SQLiteRepository` implementation.
- Updated `src/worker/pkg/app/app.go`:
  - Added `ArtifactStore *artifacts.Store` and `SnapshotPipeline *pipeline.SnapshotPipeline` fields.
  - Wired `artifacts.NewStore(cfg.DataDir)` and `pipeline.NewSnapshotPipeline(...)` into `NewApp`.
  - Updated `Close()` to close the artifact store before the database.
- Updated `src/worker/pkg/app/app_test.go`:
  - Added tests for `ArtifactStore` nil when `DataDir` is empty.
  - Added `TestNewAppWithSQLiteAndDataDirWiresSnapshotPipeline` asserting all three pipeline dependencies are non-nil together.
- New `src/worker/internal/pipeline/snapshot_test.go`:
  - `TestSnapshotSuccessWritesSnapshotAndUpdatesCanonicalURL` — happy path: snapshot exists, canonical_url set, no terminal success markers.
  - `TestSnapshotFetchFailureCommitsARCCodedFailureTransactionally` — ARC-003 (404) failure: article failed, job failed, one pending notification.
  - `TestSnapshotForbiddenFailureMapsToARC002` — ARC-002 (403) failure.
  - `TestSnapshotNonHTMLFailureMapsToARC005` — ARC-005 (non-HTML content) failure.
  - `TestSnapshotTransactionRollbackOnNotificationFailure` — pre-seeded duplicate notification causes UNIQUE violation; `ProcessOne` returns an error and the article status is not updated.
  - `TestSnapshotNoQueuedJobReturnsNil` — empty queue returns nil.
  - `TestMarkdownHandoffIsCalledOnSnapshotSuccess` — extension point is invoked after successful snapshot.
  - `TestMarkdownHandoffFailureCommitsTerminalFailure` — ARC-coded handoff error triggers terminal failure.

Decisions:
- The `MarkdownHandoff` function type is the explicit extension point for MDEXT-005. It is wired as `NoOpMarkdownHandoff` in `NewApp` until MDEXT-005 replaces it. See follow-up for the contract.
- ARC-coded fetcher errors are passed through as-is (with `//nolint:wrapcheck`) because wrapping changes the error text and the text IS the persisted user-facing message.
- Snapshot boundary is non-terminal in final v0: `ProcessOne` logs `status=snapshot_done` and returns nil on success without calling `CompleteTerminal`.
- Unknown (non-ARC) errors from fetch and markdown handoff are mapped to `ErrUnknown` (ARC-999) to prevent low-level details from appearing in `articles.error_message`.
- `App.Close()` now closes ArtifactStore before DB; both close even if the first fails.

Validation:
- `cd src/worker && go tool lefthook run build` — gobuild ✔, dotnet ✔ (npm pre-existing failure unrelated).
- `cd src/worker && go tool lefthook run format` — golangci ✔, dotnet ✔ (biome pre-existing failure unrelated).
- `cd src/worker && go tool lefthook run lint` — golangci ✔, dotnet ✔ (biome pre-existing failure unrelated).
- `cd src/worker && go tool lefthook run test` — gotest ✔ (all packages including internal/pipeline), dotnet ✔ (vitest pre-existing failure unrelated).
- All 8 pipeline tests pass.

Follow-ups:

### MDEXT-005 Handoff Interface Contract

MDEXT-005 must replace `pipeline.NoOpMarkdownHandoff` with a real implementation.

The function type signature is:

```go
type MarkdownHandoff func(ctx context.Context, job *jobs.Job, canonicalURL string) error
```

Wiring point: `pkg/app/app.go` → `pipeline.NewSnapshotPipeline(..., mdHandoff)`.
Currently: `pipeline.NoOpMarkdownHandoff`.
MDEXT-005 replaces with: its own function (or closure) that opens `snapshot.html` via `artifacts.Store.OpenSnapshot(job.ArticleID)`, runs Markdown extraction, writes `content.md`, and continues to SUMGEN-005.

Contract:
- Input: `ctx`, `job` (with `ArticleID` and `ID`), `canonicalURL` (the resolved final URL).
- On success: must return `nil`. Must not call `jobs.Repository.CompleteTerminal` — the snapshot pipeline calls it on failure only; SUMGEN-005 owns the terminal success call.
- On failure: must return an ARC-coded error (`errors.New("[ARC-NNN] ...")`). The snapshot pipeline calls `CompleteTerminal(failed, errorMessage)` automatically.
- Must not set `articles.status = ready`, `jobs.status = succeeded`, or insert a success notification.

Canonical Updates:
- `docs/specs/article-processing/PLAN.md` — ARTPROC-005 row: in_progress → done.
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md` — status: in_progress → done.
- `docs/specs/article-processing/plans/ARTPROC-005-worker-snapshot-pipeline-orchestration.execplan.md` — status: accepted → completed.

---

## 2026-05-16 - ARTPROC-007: Worker Executable Processing Command

Status:
- completed

Summary:
- Added the explicit Worker processing command so queued jobs are reachable from the deployed executable path. The previous implementation wired `SnapshotPipeline` into `App` but left the binary with only the diagnostic `version` command.

Changes:
- Added `src/worker/internal/app/process.go`:
  - `archivist-worker process`
  - `process --once`
  - `process --idle-sleep`
  - startup validation that `SnapshotPipeline` is configured
  - cancellable single-worker polling loop
- Updated `src/worker/internal/app/program.go` to own `urfave/cli/v3` command configuration and register the `process` command.
- Updated `src/worker/internal/pipeline/snapshot.go` so `ProcessOne(ctx)` returns `(processed bool, err error)`.
- Updated pipeline tests for the new `ProcessOne` contract.
- Added `src/worker/internal/app/process_test.go` with executable-surface coverage for `process --once`, missing pipeline configuration, and idle cancellation.

Decisions:
- The root command remains command-only; production processing is invoked with `archivist-worker process`.
- The loop is v0 single-worker polling, not a scheduler, retry system, or worker pool.
- `ProcessOne` distinguishes no work from processed work so daemon mode can sleep only when idle and drain queued work without unnecessary delay.
- Worker CLI configuration stays in `program.go`; each command forwards to a command-named function that does not accept or reference `urfave/cli/v3` types.

Validation:
- `cd src/worker && go run ./cmd/app process --help` — passed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed after loop helper refactor.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed.

Follow-ups:
- Worker configuration-key mismatch remains separate from this executable processing command task.

Canonical Updates:
- `docs/specs/article-processing/SPEC.md` — added executable processing requirement, acceptance scenario, interface, and rebuild note.
- `docs/specs/article-processing/PLAN.md` — added `ARTPROC-007` DAG edge and task row.
- `docs/specs/article-processing/tasks/ARTPROC-007-worker-executable-processing-command.md` — added completed corrective task.
- `docs/specs/article-processing/plans/ARTPROC-007-worker-executable-processing-command.execplan.md` — added completed ExecPlan.
- `docs/ARCHITECTURE.md` — documented `archivist-worker process` as the Worker production command.
- `docs/conventions/WORKER.md` — added CLI command boundary convention and executable-surface test requirement.
- `docs/BOOKKEEPING.md` — added executable/service-boundary quality gate.
- `docs/REBUILD.md` — added executable/service-boundary rebuild validation rule.

---

## 2026-05-17 - ARTPROC-005/ARTPROC-007: Worker Structured Log Field Correction

Status:
- done

Summary:
- Corrected Worker snapshot pipeline and runner observable failure logs to use the canonical structured field name `error` instead of `err`.
- Added missing `arc_code` fields to snapshot-stage failure logs where the ARC mapping is known.

Decisions:
- No canonical convention change was needed; `docs/conventions/GENERAL.md` and `docs/conventions/WORKER.md` already require the `error` field and `arc_code` when available.
- Completed task statuses were not reopened for this review correction.

Validation:
- Added snapshot write failure log assertions for `artifact_result`, `error`, `arc_code`, and absence of `err`.
- `rg 'slog\\.Any\\("err"' src/worker -g '*.go'` returned no production Worker matches.

Follow-ups:
- None.

Canonical Updates:
- None.
