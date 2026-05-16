# Worker Code Review

Review scope: `src/worker`.

Review method: six read-only GPT-5.5 medium-effort subagent review slices, consolidated against canonical repository documents and local code evidence.

Validation:

- `cd src/worker && go build ./...`: passed
- `cd src/worker && go tool golangci-lint run`: passed
- `cd src/worker && go test -race -shuffle=on ./...`: passed outside the sandbox
- Initial sandbox test run failed only because `httptest` could not bind local ports.

## Overall

The worker does not fully meet completed-task expectations. The largest gaps are executable wiring, canonical configuration, and incomplete `SUMGEN-003` completion despite the task being marked done.

Core snapshot/Markdown non-terminal boundaries, active job terminal failure persistence, ARC public error persistence, `req/v3` use, and provider abstraction boundaries are mostly sound.

Missing `SUMGEN-002`, `SUMGEN-004`, and `SUMGEN-005` behavior is blocked todo, not a defect.

## Findings

### 1. Worker Binary Does Not Process Jobs

Severity: High

Area: Worker execution

References:

- `src/worker/cmd/app/main.go:11`
- `src/worker/internal/app/program.go:30`
- `src/worker/pkg/app/app.go:77`
- `docs/ARCHITECTURE.md:49`

What is wrong:

The binary runs only the CLI `version` command surface. No production path invokes `SnapshotPipeline.ProcessOne` or a processing loop.

Why:

Queued jobs will not be dequeued or processed by the worker executable, violating the worker's core architecture responsibility.

Recommended fix:

Add a default worker command or daemon path that builds `App`, validates `SnapshotPipeline != nil`, runs `ProcessOne` in a cancellable loop or explicit one-shot mode, and tests processing through the executable path.

### 2. Canonical Worker Config Is Not Loaded

Severity: High

Area: Worker configuration

References:

- `src/worker/pkg/app/config/config.go:4`
- `src/worker/pkg/app/config/load.go:12`
- `docs/conventions/WORKER.md:58`
- `docs/conventions/GENERAL.md:71`

What is wrong:

Code loads `APP_*` keys such as `APP_SQLITEPATH`, `APP_DATADIR`, and `APP_JINA_ENABLED`. Canonical worker keys are `DATA_DIR`, `SQLITE_PATH`, `JINA_ENABLED`, `JINA_API_KEY`, `LLM_PROVIDER`, `LLM_API_KEY`, and `LLM_MODEL`.

Why:

Documented deployment configuration will not configure SQLite, `/data`, Jina, or LLM settings correctly.

Recommended fix:

Bind canonical unprefixed environment variables directly. Add tests for defaults, required values, and environment loading.

### 3. `SUMGEN-003` Is Marked Done But Not Complete

Severity: High

Area: Summary generation

References:

- `docs/specs/summary-generation/tasks/SUMGEN-003-summarizer-provider-adapter.md:118`
- `src/worker/pkg/app/config/config.go:4`
- `src/worker/pkg/app/app.go:20`
- `src/worker/internal/summary/anthropic.go:34`

What is wrong:

`SummarizerService` and the Anthropic adapter exist, but config lacks `LLM_PROVIDER`, `LLM_API_KEY`, and `LLM_MODEL`, and `NewApp` does not construct or expose a summarizer.

Why:

The task explicitly requires provider/model/API-key configuration support. The adapter cannot be selected or used through Pure DI.

Recommended fix:

Add summary config with `LLM_PROVIDER=anthropic`, required `LLM_API_KEY`, default `LLM_MODEL=claude-3-5-haiku-20241022`; wire `summary.SummarizerService` in `App`; cover it in `app_test.go`.

### 4. Fetcher Has SSRF Exposure

Severity: High

Area: Fetching / security

References:

- `src/worker/internal/fetcher/fetcher.go:45`
- `src/worker/internal/fetcher/fetcher.go:88`
- `src/worker/pkg/app/app.go:41`

What is wrong:

URL validation allows any `http`/`https` URL and follows redirects without rejecting loopback, private, link-local, metadata, Docker-internal, or special IP ranges.

Why:

An article URL can make the worker fetch internal network resources.

Recommended fix:

Enforce SSRF policy in the fetch layer and dial path: validate initial and redirected targets, resolve DNS safely, block private/special ranges and localhost names, and add direct/private-redirect tests.

### 5. Runner Can Ignore Later Program Errors

Severity: High

Area: Runner reliability

References:

- `src/worker/internal/runner/execution.go:52`
- `src/worker/internal/runner/execution.go:78`

What is wrong:

`waitForTermination` consumes only the first program result. If the first result is `nil`, later errors can be ignored.

Why:

Multi-program worker mode can report `Ok` while a program failed.

Recommended fix:

Drain exactly `programsCount` results, aggregate errors with `errors.Join`, return `NotOk` on any error, and add a "success first, failure later" test.

### 6. Jina Fallback Runs Even When Disabled

Severity: Medium

Area: Markdown fallback behavior

References:

- `src/worker/internal/pipeline/markdown_handoff.go:145`
- `src/worker/internal/markdown/jina.go:40`
- `src/worker/pkg/app/app.go:47`
- `docs/specs/markdown-extraction/SPEC.md:69`

What is wrong:

Fallback is always wired. When `JINA_ENABLED=false`, local extraction failure becomes an `ARC-010` Jina failure.

Why:

The spec says call Jina only when enabled. Disabled fallback should preserve local `ARC-008` or `ARC-009`.

Recommended fix:

Wire fallback only when enabled, or make handoff return the local error when fallback is disabled.

### 7. Jina Response Handling Is Unbounded

Severity: Medium

Area: Jina response handling

References:

- `src/worker/internal/markdown/jina.go:71`
- `src/worker/internal/markdown/jina.go:88`

What is wrong:

Jina response body is read via `resp.String()` with no size limit and no response content-type validation.

Why:

An unexpected provider or proxy response can cause memory pressure or write non-Markdown content into `content.md`.

Recommended fix:

Read through a hard limit, validate accepted text content types, and add oversized/non-text tests.

### 8. Jina Balance Errors Are Underclassified

Severity: Medium

Area: Jina error mapping

References:

- `src/worker/internal/markdown/jina.go:76`
- `docs/specs/markdown-extraction/tasks/MDEXT-004-worker-jina-reader-fallback.md:35`

What is wrong:

Insufficient balance maps to `ARC-011` only for HTTP 402.

Why:

The task requires mapping by status, code, or response body when exposed.

Recommended fix:

Parse non-OK Jina error bodies/codes for known insufficient-balance markers before generic `ARC-010`.

### 9. Artifact Temp Creation Escapes Rooted API

Severity: Medium

Area: Artifact writes

References:

- `src/worker/internal/artifacts/store.go:114`
- `src/worker/internal/artifacts/store.go:119`
- `src/worker/internal/artifacts/store.go:148`

What is wrong:

`Store` creates directories via `os.Root`, then uses absolute `os.CreateTemp`.

Why:

This reintroduces a symlink/TOCTOU gap outside the rooted API.

Recommended fix:

Create temp files relative to a rooted article directory handle and rename within the same rooted scope; add symlink-escape regression tests.

### 10. Duplicate Persistence Package Encodes Stale Semantics

Severity: Medium

Area: Persistence design

References:

- `src/worker/pkg/app/persistence/repository.go:77`
- `src/worker/pkg/app/persistence/repository.go:195`
- `src/worker/pkg/app/persistence/sqlite.go:64`
- `src/worker/pkg/db/schema.go:46`

What is wrong:

`pkg/app/persistence` duplicates `pkg/db` and `pkg/jobs`, is unused by production code, and has divergent terminal/schema behavior.

Why:

It is a false persistence source of truth and can reintroduce pre-summary or non-Telegram notification behavior.

Recommended fix:

Delete it or migrate any useful tests into `pkg/db` or `pkg/jobs`.

### 11. Claim Does Not Filter Job Type

Severity: Medium

Area: Queue sequencing

References:

- `src/worker/pkg/jobs/repository.go:70`
- `src/worker/pkg/jobs/job.go:13`

What is wrong:

`ClaimQueued` claims any queued job with an article row, not only `article_processing`.

Why:

Future queued job types could be run through article fetch/snapshot/Markdown processing.

Recommended fix:

Add `type = 'article_processing'` to the claim subquery and test non-article jobs are skipped.

### 12. Anthropic Context-Overflow Mapping Is Incomplete

Severity: Medium

Area: Summary provider error mapping

References:

- `src/worker/internal/summary/errors.go:71`
- `docs/specs/summary-generation/SPEC.md:75`

What is wrong:

`ARC-014` is mapped only for HTTP 413.

Why:

Canonical behavior also requires context-window overflow or preflight size failures to map to `ARC-014`.

Recommended fix:

Recognize provider error types/messages for context/request-size overflow and add regression tests.

### 13. Raw URLs Leak Into Logs and Diagnostic Errors

Severity: Medium

Area: Logging / privacy

References:

- `src/worker/internal/fetcher/errors.go:31`
- `src/worker/internal/pipeline/snapshot.go:97`
- `src/worker/internal/pipeline/markdown_handoff.go:181`

What is wrong:

Logs and diagnostic errors include full URLs, including query strings, fragments, and userinfo.

Why:

Signed URLs and tokens can leak to stdout logging systems.

Recommended fix:

Add a URL redaction helper for logs/errors; strip query, fragment, and userinfo by default.

### 14. Config Example Is Stale

Severity: Medium

Area: Configuration documentation

References:

- `src/worker/config.example.yml:1`
- `docs/conventions/WORKER.md:58`

What is wrong:

Example config contains unrelated keys like `claude`, `harness`, and `loop`, and omits canonical worker keys.

Why:

It misleads deployment and contradicts docs.

Recommended fix:

Replace it with canonical worker config examples, omitting secret values.

### 15. Logging Field Names Drift

Severity: Low

Area: Structured logging consistency

References:

- `docs/conventions/GENERAL.md:44`
- `src/worker/internal/pipeline/snapshot.go:131`
- `src/worker/internal/runner/many.go:34`

What is wrong:

Code uses `err` while convention names `error`; some stage failure logs omit known `arc_code`.

Why:

This weakens log querying and consistency.

Recommended fix:

Standardize error field naming and include `arc_code` on stage-level ARC failures.

### 16. Minor Go 1.26 Runner Idiom Gaps

Severity: Low

Area: Modern Go idioms

References:

- `src/worker/internal/runner/execution.go:28`
- `src/worker/internal/runner/execution.go:37`

What is wrong:

Runner uses manual `WaitGroup.Add` plus `go`, and does not call `signal.Stop`.

Why:

This misses the modern `WaitGroup.Go` idiom and leaves signal ownership less explicit.

Recommended fix:

Use `wg.Go` and defer `signal.Stop(signalChan)` after `signal.Notify`.

## Passed Areas

No findings for:

- Active SQLite schema consistency in `pkg/db`.
- Active terminal failure transaction behavior in `pkg/jobs`.
- ARC public message persistence in pipeline failure handling.
- Snapshot and Markdown success staying non-terminal.
- Provider abstraction boundaries: pipeline does not import provider SDK types.
- Official Anthropic SDK use inside the adapter.
- Production outbound HTTP using `req/v3`.
- Fetcher HTML MIME and 10 MiB article body enforcement.

