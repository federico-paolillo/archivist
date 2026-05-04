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
