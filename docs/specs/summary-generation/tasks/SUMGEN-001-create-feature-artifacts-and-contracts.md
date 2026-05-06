---
id: SUMGEN-001
feature: summary-generation
title: Create Feature Artifacts And Contracts
status: done
depends_on: []
blocks: [SUMGEN-002, SUMGEN-003]
parallel: false
exec_plan: null
canonical: true
---

# SUMGEN-001: Create Feature Artifacts And Contracts

## Objective

Create the Summary Generation feature ALM artifacts and promote the durable final-success, artifact, provider-boundary, configuration, logging, Gateway, and error-code contracts to canonical docs.

## Story / Context

As a rebuild agent, I need summary generation to be specified before Worker and Gateway implementation begins.

## Scope

This task includes:

- Feature `SPEC.md`, `PLAN.md`, `DIARY.md`, task files, and required ExecPlans.
- `docs/specs/INDEX.md` update.
- Artifact path convention update for `summary.md`.
- Error catalog extension for summary failures.
- Architecture and design updates for summary-complete final success.
- Worker provider-boundary and logging convention updates.
- Gateway read-only artifact convention updates.
- Amendments to `article-processing` and `markdown-extraction` terminal success wording.

## Out of Scope

This task does not include:

- Production Worker implementation.
- Production Gateway implementation.
- Go or .NET dependency installation.
- Runtime validation.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- User-approved summary generation plan.
- Existing `markdown-extraction` feature.
- Existing `article-processing` and `telegram-ingestion` dependency contracts.
- `docs/BOOKKEEPING.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Canonical feature artifacts under `docs/specs/summary-generation/`.
- Updated global canonical docs.

## Expected Affected Areas

```text
docs/specs/summary-generation/
docs/specs/markdown-extraction/
docs/specs/article-processing/
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
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/article-processing/SPEC.md`
- `docs/specs/markdown-extraction/SPEC.md`
- `docs/specs/telegram-ingestion/SPEC.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Summary generation planning artifacts exist
  Given the repository uses Markdown ALM artifacts
  When this task is complete
  Then summary-generation has a SPEC.md
  And summary-generation has a PLAN.md
  And summary-generation has executable task files
  And summary-generation is listed in docs/specs/INDEX.md

Scenario: Durable contracts are promoted
  Given summary generation changes artifact, error, provider, Gateway, and terminal success behavior
  When this task is complete
  Then those contracts are documented in canonical docs
  And no required behavior exists only in DIARY.md
```

## Done When

- Feature folder and task artifacts exist.
- `SUMGEN-003`, `SUMGEN-004`, and `SUMGEN-005` have linked ExecPlans.
- `docs/specs/INDEX.md` includes `summary-generation`.
- `docs/conventions/ARTIFACTS.md` defines `summary.md` as the v0 summary artifact.
- `docs/conventions/ERRORS.md` includes summary ARC codes.
- Architecture, design, Worker, and Gateway conventions reflect the feature.
- `DIARY.md` records this planning task completion.

## Validation

Required checks:

```bash
rg -n "summary-generation|SUMGEN|summary.md|ARC-016|SummarizerService|MarkdownExtractor" docs
```

Manual validation, if any:

- Review created Markdown files for unresolved template placeholders.

## Dependencies

Depends on:

- None.

Blocks:

- `SUMGEN-002`
- `SUMGEN-003`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This task creates planning and canonical documentation only.
