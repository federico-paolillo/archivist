---
id: MDEXT-001
feature: markdown-extraction
title: Create Feature Artifacts And Contracts
status: done
depends_on: []
blocks: [MDEXT-002, MDEXT-003, MDEXT-004]
parallel: false
exec_plan: null
canonical: true
---

# MDEXT-001: Create Feature Artifacts And Contracts

## Objective

Create the Markdown extraction feature ALM artifacts and promote the durable artifact, extraction, logging, configuration, and error-code contracts to canonical docs.

## Story / Context

As a rebuild agent, I need the Markdown extraction feature to be specified before Worker and Gateway implementation begins.

## Scope

This task includes:

- Feature `SPEC.md`, `PLAN.md`, `DIARY.md`, task files, and required ExecPlan.
- `docs/specs/INDEX.md` update.
- Artifact path convention under `docs/conventions/ARTIFACTS.md`.
- Error catalog extension for Markdown extraction failures.
- Architecture and design updates for Markdown-complete terminal success.
- Worker configuration and logging convention updates.

## Out of Scope

This task does not include:

- Production Worker implementation.
- Production Gateway implementation.
- Go dependency installation.
- Runtime validation.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- User-provided Markdown extraction plan.
- Existing `article-processing` feature.
- Existing `telegram-ingestion` dependency contracts.
- `docs/BOOKKEEPING.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Canonical feature artifacts under `docs/specs/markdown-extraction/`.
- Updated global canonical docs.

## Expected Affected Areas

```text
docs/specs/markdown-extraction/
docs/specs/INDEX.md
docs/ARCHITECTURE.md
docs/DESIGN.md
docs/conventions/
```

## Required Context

Read before execution:

- `AGENTS.md`
- `docs/REBUILD.md`
- `docs/BOOKKEEPING.md`
- `docs/specs/INDEX.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/article-processing/PLAN.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown extraction planning artifacts exist
  Given the repository uses Markdown ALM artifacts
  When this task is complete
  Then markdown-extraction has a SPEC.md
  And markdown-extraction has a PLAN.md
  And markdown-extraction has executable task files
  And markdown-extraction is listed in docs/specs/INDEX.md

Scenario: Durable contracts are promoted
  Given Markdown extraction changes artifact, error, logging, and terminal success behavior
  When this task is complete
  Then those contracts are documented in canonical docs
  And no required behavior exists only in DIARY.md
```

## Done When

- Feature folder and task artifacts exist.
- `MDEXT-005` has a linked ExecPlan.
- `docs/specs/INDEX.md` includes `markdown-extraction`.
- `docs/conventions/ARTIFACTS.md` defines article artifact paths.
- `docs/conventions/ERRORS.md` includes Markdown extraction ARC codes.
- Architecture, design, general, and worker conventions reflect the feature.
- `DIARY.md` records this planning task completion.

## Validation

Required checks:

```bash
rg -n "markdown-extraction|MDEXT|content.md|ARC-012|ARTIFACTS" docs
```

Manual validation, if any:

- Review created Markdown files for unresolved template placeholders.

## Dependencies

Depends on:

- None.

Blocks:

- `MDEXT-002`
- `MDEXT-003`
- `MDEXT-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This task creates planning and canonical documentation only.
