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
