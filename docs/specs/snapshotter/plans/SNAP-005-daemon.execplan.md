---
id: SNAP-005-PLAN
task: ../tasks/SNAP-005-daemon.md
status: completed
canonical: true
---

# ExecPlan: SNAP-005 Daemon

## Objective

Implement the service loop that sleeps before the first snapshot, handles failures by logging and continuing, and runs through the executable entrypoint.

## Linked Task

- `../tasks/SNAP-005-daemon.md`

## Implementation Sequence

1. Compose config, capture, and uploader services in the CLI entrypoint.
2. Implement an async interval loop with injectable sleep for tests.
3. Log startup, sleep, snapshot start, upload success, failure, and cleanup events.
4. Treat invalid startup configuration as fatal before entering the loop.
5. Treat per-attempt capture/upload failures as non-fatal.

## Validation Plan

```bash
cd src/snapshotter && uv run pytest
cd src/snapshotter && uv run archivist-snapshotter --help
```

## Risks

- A daemon loop can make tests slow unless sleep and attempt count are injectable.

## Completion Criteria

- Tests prove sleep-first and continue-after-failure behavior.
