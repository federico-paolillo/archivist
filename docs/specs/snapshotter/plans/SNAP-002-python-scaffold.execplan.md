---
id: SNAP-002-PLAN
task: ../tasks/SNAP-002-python-scaffold.md
status: completed
canonical: true
---

# ExecPlan: SNAP-002 Python Scaffold

## Objective

Create a locked Python 3.12 `uv` project with an importable package, console script, tests, and validation tooling.

## Linked Task

- `../tasks/SNAP-002-python-scaffold.md`

## Implementation Sequence

1. Create `src/snapshotter/pyproject.toml` with runtime dependency `boto3` and dev dependencies `pytest`, `ruff`, and `ty`.
2. Add `src/archivist_snapshotter/` package and executable entrypoint.
3. Add tests under `tests/`.
4. Generate `uv.lock`.
5. Run the task validation commands.

## Validation Plan

```bash
cd src/snapshotter && uv sync --locked --all-extras --dev
cd src/snapshotter && uv run ruff format --check .
cd src/snapshotter && uv run ruff check .
cd src/snapshotter && uv run ty check .
cd src/snapshotter && uv run pytest
```

## Risks

- This introduces the repository's first Python validation surface.

## Completion Criteria

- The Snapshotter Python project can be rebuilt from `pyproject.toml` and `uv.lock`.
