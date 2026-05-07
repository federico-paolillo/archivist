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
