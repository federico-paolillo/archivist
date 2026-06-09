# Implementation Diary: UI Article Endpoints

This file is an append-only historical log. It records implementation outcomes, validation, decisions, and follow-ups.

It is not the canonical source of requirements. Durable decisions must be promoted to canonical documents.

---

## Log

## 2026-05-05 — UIEND-001: Create Canonical Artifacts

Status:
- completed

Summary:
- Created the canonical UI endpoint feature artifacts for authenticated article list, detail, and delete APIs.

Changes:
- Added `SPEC.md`, `PLAN.md`, task files, ExecPlans, and this diary.
- Added the feature to `docs/specs/INDEX.md`.
- Updated Gateway conventions for a narrow article-delete artifact cleanup operation.

Decisions:
- Article list uses fixed 25-item pages.
- Cursor pagination uses article ULID strings.
- Delete is a hard admin action and rejects running jobs.

Validation:
- Documentation structure inspected during implementation.

Follow-ups:
- Complete Gateway implementation and validation for `UIEND-002` and `UIEND-003`.

Canonical Updates:
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/ui-endpoints/PLAN.md`
- `docs/specs/ui-endpoints/tasks/*.md`
- `docs/specs/ui-endpoints/plans/*.execplan.md`
- `docs/specs/INDEX.md`
- `docs/conventions/GATEWAY.md`

## 2026-05-06 — DOCS-SANITY: Article API Contract Correction

Status:
- completed

Summary:
- Completed UI endpoint docs with explicit lower-camel DTOs and delete/worker race rules.

Changes:
- Added list/detail/error response shapes, cursor names, and lower-camel JSON requirements.
- Added SQLite write-transaction serialization rules for delete and worker claim.
- Reconciled `UIEND-002` and `UIEND-003` as non-parallel Gateway API tasks.

Decisions:
- JSON wire contracts use lower camel case.
- Delete rejects already-running jobs with `409`; delete and worker claim serialize through SQLite write transactions.

Validation:
- Structural docs check passed for task/PLAN drift, dependency drift, ExecPlan links, required context, Markdown links, and canonical TODOs.
- Targeted repository searches found no snake_case UI/UI endpoint wire field names.
- Production build/test validation was not required because this was a docs-only correction.

Follow-ups:
- Implement read/delete APIs against the explicit DTO contracts and race semantics.

Canonical Updates:
- `docs/specs/ui-endpoints/SPEC.md`
- `docs/specs/ui-endpoints/PLAN.md`
- `docs/specs/ui-endpoints/tasks/UIEND-002-gateway-article-read-api.md`
- `docs/specs/ui-endpoints/tasks/UIEND-003-gateway-article-delete-api.md`
- `docs/specs/ui-endpoints/plans/UIEND-002-gateway-article-read-api.execplan.md`
- `docs/specs/ui-endpoints/plans/UIEND-003-gateway-article-delete-api.execplan.md`

## 2026-05-12 — UIEND-003: Gateway Article Delete API

Status:
- completed

Summary:
- Implemented authenticated `DELETE /articles/{id}` for hard deletion of owned article state, associated jobs, associated notifications, and the deterministic artifact directory.

Changes:
- Added Gateway article route mapping with `RequireAuthorization` and same-origin unsafe-method filtering.
- Added an article delete application service using a SQLite `BEGIN IMMEDIATE` transaction, ownership recheck, running-job conflict detection, explicit associated row deletion, and commit after artifact cleanup succeeds.
- Added a delete-only artifact cleanup abstraction separate from read-only artifact access.
- Added integration tests for ready, failed, queued, running conflict, malformed ID, not found, missing artifact directory, cleanup failure with DB state intact, row/directory cleanup, cross-site rejection, and queued-job removal before later claim.

Decisions:
- Used the proposed ExecPlan as execution guidance under explicit user assignment override; marked it completed.
- `DATA_DIR` defaults to `/data` for Gateway article artifact deletion when not configured.
- Artifact cleanup failure returns `500` and rolls back the open SQLite transaction before database delete statements run.

Validation:
- `cd src/gateway && dotnet format`
- `cd src/gateway && dotnet build`
- `cd src/gateway && dotnet test`

Follow-ups:
- `UIEND-002` remains separate and was not implemented.

Canonical Updates:
- `docs/specs/ui-endpoints/PLAN.md`
- `docs/specs/ui-endpoints/tasks/UIEND-003-gateway-article-delete-api.md`
- `docs/specs/ui-endpoints/plans/UIEND-003-gateway-article-delete-api.execplan.md`

## 2026-05-15 — R-001: Worker claim skips orphan queued jobs

Task ID: R-001
Status: completed

Summary:
- Fixed `SQLiteRepository.ClaimQueued` so the atomic `UPDATE ... RETURNING` claim selects only queued jobs whose article row still exists.
- Added a Worker repository regression test that creates an orphan queued job through a controlled fixture and asserts `sql.ErrNoRows`.
- Moved the ui-endpoints feature status to `in_progress` because `UIEND-003` is done while `UIEND-002` remains blocked.

Decisions:
- Kept the claim in a single `UPDATE ... RETURNING` statement.
- Did not add a deterministic concurrent delete/claim test in this pass. The current Worker repository tests use a single-connection in-memory SQLite fixture, while Gateway delete tests use a separate EF/TestServer fixture. A safe cross-connection SQLite write-interleaving test requires a shared file-backed integration harness outside this targeted fix.

Validation:
- `go test ./pkg/jobs` — passed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — initial run failed on stale golangci cache entries for a deleted external worktree; after `go tool golangci-lint cache clean`, rerunning the required lint command passed.
- `cd src/worker && go tool lefthook run test` — passed.

Follow-ups:
- Add execution-level delete/claim serialization coverage if a shared file-backed Gateway/Worker SQLite integration harness is introduced.
- `UIEND-002` remains blocked on summary generation.

Canonical Updates:
- `docs/specs/ui-endpoints/SPEC.md` — status: in_progress.
- `docs/specs/ui-endpoints/PLAN.md` — status: in_progress.
- `docs/specs/ui-endpoints/tasks/UIEND-003-gateway-article-delete-api.md` — validation limitation recorded.
- `docs/specs/INDEX.md` — ui-endpoints status: in_progress.

## 2026-05-28 — UIEND-002: Gateway Article Read API

Status:
- completed

Summary:
- Implemented authenticated `GET /articles` and `GET /articles/{id}` for user-owned article metadata and detail.
- Completed the `ui-endpoints` feature because `UIEND-002` and `UIEND-003` are now done.

Changes:
- Added fixed 25-item article list pagination using ULID cursors with `after` for older rows and `before` for newer rows.
- Added authenticated user scoping through `ClaimTypes.NameIdentifier`.
- Added article detail loading with lower-camel JSON DTOs and read-only `summary.md` / `content.md` artifact access.
- Extended the existing read-only artifact reader to read `content.md` without adding write, create, rename, or delete operations.
- Added integration tests for authentication, cursor validation, pagination, malformed IDs, not found, ready artifact reads, ready missing artifacts, and queued/failed nullable artifacts.
- Re-ran the existing delete endpoint tests as part of the full Gateway test suite to guard the `DELETE /articles/{id}` behavior.

Decisions:
- Used the linked ExecPlan after promoting it from `proposed` to `in_progress`; marked it `completed` after implementation.
- Kept URLs as strings in application/API DTOs because the canonical wire contract and SQLite schema expose persisted URL text.
- Missing or unreadable artifacts are normalized inside the query service: ready articles return the documented `500`, while queued and failed articles return nullable Markdown fields.

Validation:
- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed.
- `cd src/gateway && dotnet test` — passed.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui-endpoints/SPEC.md` — status: done.
- `docs/specs/ui-endpoints/PLAN.md` — status: done, `UIEND-002` row done.
- `docs/specs/ui-endpoints/tasks/UIEND-002-gateway-article-read-api.md` — status: done.
- `docs/specs/ui-endpoints/plans/UIEND-002-gateway-article-read-api.execplan.md` — status: completed.
- `docs/specs/INDEX.md` — ui-endpoints status: done.

## 2026-05-30 — UIEND-003: Delete cleanup rollback ordering fix

Task ID: UIEND-003
Status: completed

Summary:
- Fixed Gateway hard delete ordering so article notifications, jobs, and the article row are deleted inside the open SQLite `BEGIN IMMEDIATE` transaction before artifact cleanup runs.
- The service now commits only after artifact cleanup succeeds; cleanup failure rolls back the database deletes and returns the existing `500` delete failure path.

Changes:
- Reordered `EfArticleDeleteService` to collect job IDs, delete notifications/jobs/article rows, invoke artifact cleanup, and commit only on cleanup success.
- Added regression coverage that records SQL delete command timing and proves artifact cleanup failure after database deletes leaves article, job, and notification rows intact after rollback.
- Updated the `UIEND-003` ExecPlan risk text to match the current database-delete-then-cleanup-before-commit contract.

Decisions:
- Preserved not-found, running-job conflict, and missing-artifact-directory behavior.
- Kept the delete-only artifact cleanup abstraction separate from read-only artifact access.

Validation:
- `cd src/gateway && dotnet test --filter ArticleDeleteEndpointTest` — passed, 14 tests.
- `cd src/gateway && dotnet format --verify-no-changes` — passed after applying formatter corrections.
- `cd src/gateway && dotnet build` — passed.
- `cd src/gateway && dotnet test` — passed, 148 tests.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/ui-endpoints/plans/UIEND-003-gateway-article-delete-api.execplan.md`

## 2026-05-31 — UIEND-003-REVIEW-P3: File-backed delete/claim coverage

Task ID: UIEND-003
Status: completed

Summary:
- Added file-backed SQLite cross-connection coverage for Gateway hard delete and Worker-equivalent job claim ordering.
- Covered both delete-first and claim-first outcomes without introducing Worker code dependencies into Gateway tests.

Changes:
- `ArticleDeleteEndpointTest` now exercises the real `EfArticleDeleteService` and filesystem artifact deletion service against a file-backed database.
- Added raw claim SQL that matches the canonical Worker claim shape: a SQLite write transaction, `UPDATE ... RETURNING`, and an `articles` join so deleted article jobs are not claimable.

Decisions:
- No canonical contract changes were required; `SPEC.md` already defines delete/claim serialization, delete-first no-claim behavior, and claim-first `409 Conflict` behavior.

Validation:
- `cd src/gateway && dotnet test Archivist.Gateway.Tests/Archivist.Gateway.Tests.csproj --filter "TelegramIngestionRepositoryTest|ArticleDeleteEndpointTest"` — passed, 21 tests.
- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed.
- `cd src/gateway && dotnet test` — passed, 165 tests.

Follow-ups:
- None.

Canonical Updates:
- None; historical diary entry only.

## 2026-06-06 — UIEND-004 done

- **Task:** UIEND-004 Delete review hardening
- **Status outcome:** done
- **Summary:** Normal delete now normalizes valid ULID route values before service calls. The known SQLite/filesystem delete atomicity limitation was documented as a v0 design decision and artifact contract note.
- **Decisions made:** v0 keeps the current transaction-before-artifact-cleanup ordering and documents the rare artifact-deleted/commit-failed repair case instead of adding repair queues, tombstones, or cleanup jobs.
- **Validation performed:** `git diff --check`; `cd src/gateway && dotnet format --verify-no-changes`; `cd src/gateway && dotnet build`; `cd src/gateway && dotnet test` passed with 197 tests.
- **Follow-ups:** None.
- **Canonical documents updated:** `SPEC.md`, `PLAN.md`, `tasks/UIEND-004-delete-review-hardening.md`, `docs/DESIGN.md`, `docs/ARTIFACTS.md`, and `docs/specs/job-recovery/SPEC.md`.
