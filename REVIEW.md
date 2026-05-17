# Worker Code Review

Review scope: `src/worker` plus canonical worker-relevant ALM documents.

Review method: six read-only subagent review slices plus coordinator verification. Prior findings in the previous report were rechecked; resolved, stale, and blocked-scope findings were removed from the active list.

Validation:

- `cd src/worker && go tool lefthook run build`: passed
- `cd src/worker && go tool lefthook run format`: passed
- `cd src/worker && go tool lefthook run lint`: passed
- `cd src/worker && go tool lefthook run test`: passed

## Overall

The worker has materially improved since the prior review: executable processing, canonical configuration, summary adapter wiring, Jina response bounding, Jina balance classification, runner error aggregation, and stale duplicate persistence code are resolved.

The remaining issues are concentrated in security hardening, queue specificity, artifact write robustness, summary error classification, logging consistency, Pure DI consistency, and ALM task/dependency metadata. Missing `SUMGEN-002`, `SUMGEN-004`, and `SUMGEN-005` implementation remains blocked scope and is not reported as a code defect.

## Findings

### 1. Fetcher Has SSRF Exposure

Severity: High

Area: Fetching / security

References:

- `src/worker/internal/fetcher/fetcher.go:45`
- `src/worker/internal/fetcher/fetcher.go:51`
- `src/worker/internal/fetcher/fetcher.go:88`
- `src/worker/pkg/app/app.go:61`

What is wrong:

The fetcher validates only that the input URL uses `http` or `https`, then lets the shared client follow redirects. It does not reject loopback, private, link-local, metadata-service, Docker-internal, or other special address ranges for either the initial URL or redirected targets.

Why:

An article URL can make the worker fetch internal network resources. This is a standard SSRF class risk for any server-side URL fetcher.

Recommended fix:

Enforce SSRF policy in the fetch layer and dial path: validate initial and redirected targets, resolve DNS safely, block private/special ranges and localhost names, and add direct/private-redirect tests.

### 2. Artifact Temp Creation Escapes Rooted API

Severity: Medium

Area: Artifact writes

References:

- `src/worker/internal/artifacts/store.go:36`
- `src/worker/internal/artifacts/store.go:114`
- `src/worker/internal/artifacts/store.go:119`
- `src/worker/internal/artifacts/store.go:121`
- `docs/conventions/WORKER.md:48`

What is wrong:

`Store` opens `DATA_DIR` with `os.OpenRoot` and uses rooted APIs for directory creation and final rename, but temp files are created with absolute `os.CreateTemp(absDir, ...)`.

Why:

This reintroduces a symlink/TOCTOU gap outside the rooted API for the temp write step.

Recommended fix:

Create temp files relative to a rooted article-directory handle and rename within the same rooted scope. Add symlink-escape regression tests for artifact writes.

### 3. Article Processing Claim Does Not Filter Job Type

Severity: Medium - Deferred. We are not going to fix this. In v0 we don't have any other job types.

Area: Queue sequencing

References:

- `src/worker/pkg/jobs/repository.go:34`
- `src/worker/pkg/jobs/repository.go:70`
- `src/worker/pkg/jobs/repository.go:76`
- `src/worker/pkg/jobs/job.go:13`
- `docs/specs/article-processing/SPEC.md:67`

What is wrong:

`ClaimQueued` selects any queued job with an article row. It does not constrain the claimed job to `article_processing`.

Why:

Future queued job types could be claimed and run through the article fetch/snapshot/Markdown pipeline.

Recommended fix:

Make claiming type-specific, for example `ClaimQueuedArticleProcessing` or `ClaimQueued(ctx, jobs.TypeArticleProcessing)`. Add `AND type = ?` to the claim query and test that a queued non-article-processing job remains queued.

### 4. Anthropic Context-Overflow Mapping Is Incomplete

Severity: Medium

Area: Summary provider error mapping

References:

- `src/worker/internal/summary/errors.go:71`
- `src/worker/internal/summary/errors.go:88`
- `docs/specs/summary-generation/tasks/SUMGEN-003-summarizer-provider-adapter.md:34`
- `docs/specs/summary-generation/plans/SUMGEN-003-summarizer-provider-adapter.execplan.md:56`

What is wrong:

`ARC-014` is mapped only for HTTP 413. Other Anthropic context-window overflow or request-size signals fall through to `ARC-013`.

Why:

`SUMGEN-003` requires both request-too-large and context-window overflow failures to map to `ARC-014`.

Recommended fix:

Classify Anthropic context-window overflow provider signals as `ARC-014` in addition to HTTP 413. Add a focused adapter regression test for that provider error shape.

### 5. Raw URLs Leak Into Logs and Diagnostic Errors

Severity: Medium - Deferred. We are accepting this risk for now. I am the only v0 user anyway.

Area: Logging / privacy

References:

- `src/worker/internal/fetcher/errors.go:31`
- `src/worker/internal/pipeline/errors.go:38`
- `src/worker/internal/pipeline/snapshot.go:98`
- `src/worker/internal/pipeline/snapshot.go:141`
- `src/worker/internal/pipeline/markdown_handoff.go:178`

What is wrong:

Pipeline logs and diagnostic error strings include full article URLs, including any query strings, fragments, or userinfo.

Why:

Signed URLs, tokens, and credentials embedded in URLs can leak to stdout logging systems and diagnostic error text.

Recommended fix:

Add a URL redaction helper for logs/errors. Strip query, fragment, and userinfo by default while keeping enough host/path context for diagnosis.

### 6. Worker Logging Field Names Drift From Canonical Convention

Severity: Medium

Area: Structured logging consistency

References:

- `docs/conventions/GENERAL.md:44`
- `docs/conventions/WORKER.md:159`
- `src/worker/internal/pipeline/snapshot.go:145`
- `src/worker/internal/pipeline/markdown_handoff.go:87`
- `src/worker/internal/runner/many.go:36`

What is wrong:

The canonical article-processing log field set uses `error`, but worker code emits `err` in observable failure logs. Some stage failure logs also omit known `arc_code`, such as snapshot write failure.

Why:

The drift weakens log querying and makes rebuild agents choose between inconsistent conventions and implementation precedent.

Recommended fix:

Standardize worker structured logs on `slog.Any("error", err)` for observable failures and include `arc_code` where the ARC code is known, or update the canonical convention if `err` is intentionally the project-wide field name.

### 7. Shared HTTP Client Is Not Exposed As An App Singleton

Severity: Medium

Area: Composition root / Pure DI

References:

- `src/worker/pkg/app/app.go:24`
- `src/worker/pkg/app/app.go:61`
- `src/worker/pkg/app/app.go:66`
- `docs/conventions/WORKER.md:114`
- `docs/conventions/WORKER.md:116`

What is wrong:

`NewApp` creates the shared `*req.Client` as a local variable and injects it into services, but `App` has no field for the long-lived HTTP client.

Why:

Worker conventions say long-lived singletons, including HTTP clients, live as fields of `App`, and every `App` field must be covered in `app_test.go`. The current shape hides a shared singleton from the composition root contract.

Recommended fix:

Add `HTTPClient *req.Client` to `App`, assign it in `NewApp`, pass it from the `App` graph into fetcher/Jina/Anthropic, and assert it in `app_test.go`.

### 8. Job Repository Uses Global ULID Generation State

Severity: Medium

Area: Composition root / Pure DI

References:

- `src/worker/pkg/jobs/repository.go:14`
- `src/worker/pkg/jobs/repository.go:16`
- `src/worker/pkg/jobs/repository.go:214`
- `src/worker/pkg/jobs/repository.go:239`
- `docs/conventions/WORKER.md:101`
- `docs/conventions/WORKER.md:113`

What is wrong:

Notification ID generation uses package-level mutable ULID entropy and a mutex. `SQLiteRepository` resolves that hidden dependency internally instead of receiving an ID generator collaborator.

Why:

This conflicts with the Worker Pure DI rule: collaborators should be explicit constructor parameters, not hidden globals.

Recommended fix:

Introduce a small ID generator dependency for `SQLiteRepository`, wire the production ULID generator in `pkg/app.NewApp`, and inject deterministic generators in repository tests.

### 9. `SUMGEN-002` Is Blocked Despite Satisfied Dependencies

Severity: Medium

Area: ALM consistency

References:

- `docs/specs/summary-generation/PLAN.md:54`
- `docs/specs/summary-generation/tasks/SUMGEN-002-worker-summary-artifact-access.md:5`
- `docs/specs/summary-generation/tasks/SUMGEN-002-worker-summary-artifact-access.md:6`
- `docs/specs/markdown-extraction/PLAN.md:58`
- `docs/specs/worker-runtime-configuration/PLAN.md:23`

What is wrong:

`SUMGEN-002` is marked `blocked`, but its listed dependencies are already done in canonical plans.

Why:

Agents may execute only `ready` tasks. A rebuild or implementation agent would incorrectly skip the next executable worker summary task.

Recommended fix:

If no real blocker remains, mark `SUMGEN-002` ready in both task frontmatter and `summary-generation/PLAN.md`. If a blocker still exists, document the blocker explicitly in the canonical task or spec.

### 10. Summary Dependency Metadata Omits `WCFG-002`

Severity: Medium

Area: ALM consistency

References:

- `docs/specs/worker-runtime-configuration/PLAN.md:23`
- `docs/specs/worker-runtime-configuration/PLAN.md:24`
- `docs/specs/summary-generation/PLAN.md:54`
- `docs/specs/summary-generation/PLAN.md:56`
- `docs/specs/summary-generation/tasks/SUMGEN-002-worker-summary-artifact-access.md:6`

What is wrong:

`worker-runtime-configuration` says `WCFG-001` and `WCFG-002` both block `SUMGEN-002` and `SUMGEN-004`, but the summary feature records only `WCFG-001` in its PLAN rows, and task frontmatter omits worker config dependencies.

Why:

This creates contradictory cross-feature dependency data for rebuild ordering and task selection.

Recommended fix:

Update `summary-generation` PLAN rows and task frontmatter to include the canonical worker-runtime dependencies, including `WCFG-002`, or remove the reverse block from worker-runtime docs if it is no longer required.

### 11. `ARTPROC-005` Reverse Block Metadata Omits `ARTPROC-007`

Severity: Low

Area: ALM consistency

References:

- `docs/specs/article-processing/PLAN.md:24`
- `docs/specs/article-processing/PLAN.md:66`
- `docs/specs/article-processing/tasks/ARTPROC-005-worker-snapshot-pipeline-orchestration.md:7`
- `docs/specs/article-processing/tasks/ARTPROC-007-worker-executable-processing-command.md:6`

What is wrong:

The article-processing DAG and task table say `ARTPROC-005` blocks `ARTPROC-007`, and `ARTPROC-007` depends on `ARTPROC-005`. The `ARTPROC-005` task frontmatter omits `ARTPROC-007` from `blocks`.

Why:

This stale reverse-dependency metadata can mislead agents that use task frontmatter for dependency or concurrency checks.

Recommended fix:

Add `ARTPROC-007` to `ARTPROC-005` frontmatter `blocks`.

## Passed Areas

No active findings for:

- Worker executable `process` command wiring through the CLI registration path.
- Canonical Worker config loading with `ARCHIVIST_` environment variables.
- Summary adapter construction and exposure from `pkg/app.NewApp`.
- Runner aggregation of later program errors.
- Active SQLite schema consistency in `pkg/db`.
- Active terminal failure transaction behavior in `pkg/jobs`.
- ARC public message persistence in pipeline failure handling.
- Snapshot and Markdown success staying non-terminal while summary tasks remain blocked.
- Provider abstraction boundaries: pipeline does not import provider SDK types.
- Official Anthropic SDK use inside the adapter.
- Production outbound HTTP using `req/v3`.
- Fetcher HTML MIME and 10 MiB article body enforcement.
- Jina response size/content-type enforcement and insufficient-balance classification.
