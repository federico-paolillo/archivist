---
id: ARTPROC-003
feature: article-processing
title: Worker Filesystem Artifact Access Layer
status: ready
depends_on: [ARTPROC-001]
blocks: [ARTPROC-004]
parallel: true
exec_plan: null
canonical: true
---

# ARTPROC-003: Worker Filesystem Artifact Access Layer

## Objective

Build a reusable Worker filesystem access layer for article artifacts under `DATA_DIR`, starting with atomic `snapshot.html` writes.

## Story / Context

As the Worker, I need one shared artifact access boundary so snapshotting, future extraction, future summarization, and any cleanup behavior use the same deterministic and traversal-resistant filesystem rules.

## Scope

This task includes:

- Article artifact root resolution from `DATA_DIR` and `article_id`.
- Deterministic `articles/{article_id}/snapshot.html` path behavior.
- Creation of the article artifact directory when needed.
- Atomic snapshot writes using a temporary file followed by rename.
- Traversal-resistant access using `os.Root` or `os.OpenInRoot` where functionally correct.
- Worker tests for deterministic paths, atomic writes, and traversal rejection.

## Out of Scope

This task does not include:

- HTTP fetching.
- SQLite state updates.
- Gateway artifact reads.
- Writing `content.md`, `summary.json`, `summary.md`, or `metadata.json`.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker artifact-store package or service.
- Tests proving artifact path and atomic write behavior.

## Expected Affected Areas

```text
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/ARTIFACTS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Snapshot is written atomically
  Given an article ID and HTML bytes
  When the Worker stores the snapshot
  Then the final file is DATA_DIR/articles/{article_id}/snapshot.html
  And the write uses a temporary file followed by rename
  And no partial temporary file is promoted on write failure

Scenario: Artifact access rejects traversal
  Given an invalid article artifact name attempts to escape DATA_DIR
  When the artifact layer resolves or opens it
  Then the operation fails
```

## Done When

- Worker has a reusable artifact access layer.
- `snapshot.html` writes are atomic.
- Artifact paths are deterministic and derived from `DATA_DIR` and `article_id`.
- Tests cover deterministic path, atomic write, and traversal rejection.
- No placeholder artifacts are created.
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

- Inspect a test artifact directory to confirm only `snapshot.html` is written.

## Dependencies

Depends on:

- `ARTPROC-001`

Blocks:

- `ARTPROC-004`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- This layer should be designed for reuse by later extraction and summarization features without implementing those features now.
