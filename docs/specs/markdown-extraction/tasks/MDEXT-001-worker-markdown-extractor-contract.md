---
id: MDEXT-001
feature: markdown-extraction
title: Worker Markdown Extractor Contract
depends_on: []
blocks: [MDEXT-003, MDEXT-004]
parallel: true
requires_exec_plan: false
canonical: true
---
# MDEXT-001: Worker Markdown Extractor Contract

## Objective

Define the Worker-owned `MarkdownExtractor` contract and result taxonomy shared by local extraction, Jina fallback, and Markdown pipeline orchestration.

## Story / Context

As the Worker, I need provider implementations to report success, local unreadable documents, and ARC-coded failures through a stable Archivist-owned contract so orchestration can apply fallback and terminal failure rules without importing provider-specific types.

## Scope

This task includes:

- Provider-neutral `MarkdownExtractor` contract semantics.
- Result taxonomy for Markdown success, local unreadable result, and ARC-coded provider or conversion failure.
- Provider identity fields sufficient for orchestration logging.
- Extracted title metadata field for best-effort `articles.title` persistence.
- Explicit rule that provider SDK/client/library types stay inside provider implementations.
- Contract boundary notes tying the shared extractor contract to provider and integration tasks.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Stable `MarkdownExtractor` contract and result taxonomy consumed by `MDEXT-003`, `MDEXT-004`, and `MDEXT-005`.
- Provider-neutral result fields for Markdown body, selected provider, optional title, local unreadable classification, and ARC-coded failure.

## Expected Affected Areas

```text
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Providers use a shared Markdown extraction contract
  Given local and Jina extraction providers exist
  When either provider is called by pipeline orchestration
  Then it returns a provider-neutral Markdown extraction result
  And provider SDK or library types do not leak into orchestration

Scenario: Local unreadable result is distinct from failure
  Given local readability rejects a document through CheckDocument
  When the local extractor returns
  Then orchestration can distinguish local unreadable from ARC-coded failure
  And Jina fallback can run

Scenario: Provider failures carry ARC-coded public errors
  Given a provider or conversion operation fails
  When the extractor returns failure
  Then the result carries the ARC-coded public error needed for terminal failure persistence
```

## Done When

- `MarkdownExtractor` semantics are stable before local and Jina implementations consume them.
- The result taxonomy distinguishes success, local unreadable, and ARC-coded failure.
- Provider identity and optional extracted title can be returned without leaking provider-specific types.
- `MDEXT-003`, `MDEXT-004`, and `MDEXT-005` can validate against this contract.

## Validation

Required checks:

```bash
git diff -- docs/specs/markdown-extraction
```

Manual validation, if any:

- None.

## Dependencies

Depends on:

- None.

Blocks:

- `MDEXT-003`
- `MDEXT-004`

## ExecPlan Requirement

Requires ExecPlan: false

## Open Questions

- None.

## Notes

- This task records the Markdown extractor contract boundary. It does not add runtime behavior beyond the Markdown extraction spec and provider/integration task contracts.
