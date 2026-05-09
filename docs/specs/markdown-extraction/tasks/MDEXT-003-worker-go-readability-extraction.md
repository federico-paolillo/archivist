---
id: MDEXT-003
feature: markdown-extraction
title: Worker Go-Readability Extraction
status: done
depends_on: [MDEXT-001]
blocks: [MDEXT-005]
parallel: true
exec_plan: null
canonical: true
---

# MDEXT-003: Worker Go-Readability Extraction

## Objective

Implement local Markdown extraction from HTML using go-readability v2 and HTML-to-Markdown conversion behind the Worker-owned `MarkdownExtractor` interface.

## Story / Context

As the Worker, I need a local low-cost extractor that can produce readable Markdown without using paid provider calls when the document is readable.

## Scope

This task includes:

- Adding `codeberg.org/readeck/go-readability/v2`.
- Adding `github.com/JohannesKaufmann/html-to-markdown/v2`.
- Implementing the shared Worker-owned `MarkdownExtractor` contract for the local extractor.
- Parsing saved HTML snapshot bytes with canonical URL context.
- Calling `CheckDocument()` before accepting local readability output.
- Returning a typed local unreadable result when `CheckDocument()` returns false.
- Converting readable HTML content to Markdown.
- Mapping local extraction and conversion failures to `ARC-009`.
- Tests for readable HTML, unreadable HTML, parse failure, and conversion failure.

## Out of Scope

This task does not include:

- Jina Reader fallback.
- Artifact writes.
- SQLite job state transitions.
- Gateway notification behavior.
- Extraction candidate scoring.
- LLM summarization.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `MDEXT-001`.
- `../SPEC.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker local `MarkdownExtractor` implementation.
- Tests for local extraction result classes.

## Expected Affected Areas

```text
src/worker/go.mod
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Readable HTML extracts locally
  Given saved HTML snapshot bytes
  And go-readability CheckDocument returns true
  When the local extractor runs
  Then it returns Markdown
  And it reports go-readability as the selected provider

Scenario: go-readability rejects a document
  Given saved HTML snapshot bytes
  And go-readability CheckDocument returns false
  When the local extractor runs
  Then it returns a local unreadable result
  And the caller can fallback to Jina Reader

Scenario: local conversion fails
  Given go-readability produces extracted HTML
  And HTML-to-Markdown conversion fails
  When the local extractor runs
  Then it returns an ARC-009 failure
```

## Done When

- Worker has a local `MarkdownExtractor` using go-readability v2.
- Local implementation does not leak go-readability or converter types into pipeline orchestration.
- `CheckDocument()` gates local success.
- Local extraction output is Markdown, not raw HTML.
- Local unreadable documents are distinguishable from unknown failures.
- Tests cover readable, unreadable, parse failure, and conversion failure behavior.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- None.

## Dependencies

Depends on:

- `MDEXT-001`

Blocks:

- `MDEXT-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- Do not add scoring or quality thresholds in this task.
- `GoReadabilityExtractor` does not accept a logger and must not emit structured log entries. Structured logging for provider attempt, selected provider, ARC code, duration, and artifact write result is owned by MDEXT-005 pipeline orchestration per `docs/conventions/WORKER.md`.
