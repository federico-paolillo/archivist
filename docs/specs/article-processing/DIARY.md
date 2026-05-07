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
