---
id: MDEXT-002
feature: markdown-extraction
title: Worker Markdown Artifact Access
status: done
depends_on: [MDEXT-001, ARTPROC-003]
blocks: [MDEXT-005]
parallel: true
exec_plan: null
canonical: true
---

# MDEXT-002: Worker Markdown Artifact Access

## Objective

Extend the Worker artifact access layer with deterministic, traversal-resistant, atomic Markdown writes for `content.md`.

## Story / Context

As the Worker, I need the same artifact boundary used for HTML snapshots to persist extracted Markdown before terminal success is committed.

## Scope

This task includes:

- Deterministic `{DATA_DIR}/articles/{article_id}/content.md` path behavior.
- Creation of the article artifact directory when needed.
- Atomic Markdown writes using a temporary file followed by rename.
- Traversal-resistant access using `os.Root` or `os.OpenInRoot` where functionally correct.
- Tests for deterministic path behavior, atomic writes, failed-write cleanup, and traversal rejection.

## Out of Scope

This task does not include:

- HTML fetching.
- Markdown extraction.
- Jina Reader fallback.
- SQLite state updates.
- Gateway artifact reads.
- Writing `summary.json`, `summary.md`, or `metadata.json`.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- Completed `MDEXT-001`.
- Completed `ARTPROC-003`.
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker artifact-store behavior for `content.md`.
- Tests proving path and atomic write behavior.

## Expected Affected Areas

```text
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `docs/specs/article-processing/tasks/ARTPROC-003-worker-filesystem-artifact-access-layer.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Markdown is written atomically
  Given an article ID and Markdown bytes
  When the Worker stores Markdown content
  Then the final file is DATA_DIR/articles/{article_id}/content.md
  And the write uses a temporary file followed by rename
  And no partial temporary file is promoted on write failure

Scenario: Artifact access rejects traversal
  Given an invalid article artifact name attempts to escape DATA_DIR
  When the artifact layer resolves or opens it
  Then the operation fails
```

## Done When

- Worker artifact access supports atomic `content.md` writes.
- Artifact paths match `docs/conventions/ARTIFACTS.md`.
- Tests cover deterministic path, atomic write, failed-write cleanup, and traversal rejection.
- No placeholder future artifacts are created.
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

- Inspect a test artifact directory to confirm only expected artifacts are written.

## Dependencies

Depends on:

- `MDEXT-001`
- `ARTPROC-003`

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

- This task extends, rather than replaces, the artifact access layer from article processing.
