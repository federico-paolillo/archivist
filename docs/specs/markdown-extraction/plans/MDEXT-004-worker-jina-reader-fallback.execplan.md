---
id: MDEXT-004-PLAN
task: ../tasks/MDEXT-004-worker-jina-reader-fallback.md
status: completed
canonical: true
---

# ExecPlan: MDEXT-004 Worker Jina Reader Fallback

## Objective

Implement the Jina Reader fallback as a Worker-owned `MarkdownExtractor` implementation while keeping Jina SDK or adapter details out of pipeline orchestration.

## Linked Task

- `../tasks/MDEXT-004-worker-jina-reader-fallback.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/MDEXT-004-worker-jina-reader-fallback.md`

Add only ExecPlan-specific context:

- `.agents/skills/archivist-worker/SKILL.md`
- `docs/ERRORS.md`
- Jina Reader API documentation
- Official Jina repositories or SDK documentation available at implementation time

## Assumptions

- `MDEXT-003` defines the shared extractor result model or the implementer coordinates the same contract across both tasks.
- Planning verification found the official Jina Reader API and `jina-ai/reader` repository, but no suitable official Go Reader SDK.
- Implementation must re-check current official Jina SDK availability before selecting custom adapter fallback.

## Non-Goals

- Do not implement local go-readability extraction.
- Do not write `content.md`.
- Do not update SQLite state.
- Do not use ReaderLM-v2 by default.
- Do not introduce untagged or low-adoption third-party wrappers.

## Implementation Sequence

1. Re-check official Jina documentation and repositories for a suitable Go SDK that supports Reader API.
2. If a suitable official SDK exists, use it inside the Jina extractor implementation.
3. If no suitable official SDK exists, implement a small internal HTTP adapter for Reader API.
4. Define or reuse the Worker-owned `MarkdownExtractor` contract and result types.
5. Implement Jina extraction against the article canonical URL.
6. Keep Jina callable whenever local extraction needs fallback; do not add a runtime toggle.
7. Load required `JINA_API_KEY` without logging it.
8. Map general Jina failures to `ARC-010`.
9. Map insufficient balance to `ARC-011` when exposed by status, code, or response body.
10. Keep SDK or adapter types private to the Jina package.
11. Add tests for success, provider failure, timeout/transport failure, insufficient balance, and abstraction boundaries.
12. Record the SDK selection outcome in `DIARY.md` when implementation is performed.

## Validation Plan

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual checks:

- Confirm current official Jina SDK availability from Jina documentation or repositories.
- Inspect imports from Worker pipeline code to confirm no Jina SDK or adapter package leaks into orchestration.

## Documentation Updates Required

- Update `../tasks/MDEXT-004-worker-jina-reader-fallback.md` status when complete.
- Update `../PLAN.md` task table when status changes.
- Append to `../DIARY.md` after implementation.
- Promote any new Jina configuration, error mapping, or provider behavior to canonical docs before relying on it.

## Risks

- Using an unofficial wrapper would add avoidable dependency risk and weaken provider-boundary control.
- Leaking adapter types into pipeline code would make future provider replacement harder.
- Failing to distinguish insufficient balance would hide the operator action required to restore extraction.

## Rollback / Recovery Notes

- Failed extraction remains terminal in v0; manual requeue is performed by sending the URL again.
- Failed Jina fallback remains terminal in v0; operators restore provider availability by correcting the required Jina configuration or account state and requeueing manually.

## Completion Criteria

- Jina fallback implements `MarkdownExtractor`.
- SDK selection is documented.
- ARC mappings are tested.
- Worker validation passes.
