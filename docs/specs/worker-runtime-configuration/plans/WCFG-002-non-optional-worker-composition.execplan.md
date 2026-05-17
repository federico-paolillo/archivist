---
id: WCFG-002-PLAN
task: ../tasks/WCFG-002-non-optional-worker-composition.md
status: completed
canonical: true
---

# ExecPlan: WCFG-002 Non-Optional Worker Composition

## Objective

Make strict Worker config imply a fully wired Worker composition root.

## Implementation Sequence

1. Add nil-config and `cfg.Validate()` checks at the start of `pkg/app.NewApp`.
2. Create SQLite DB and artifact store unconditionally after validation.
3. Always create jobs repository, fetcher, Markdown extractors, Markdown handoff, and `SnapshotPipeline`.
4. Remove `process` command missing-pipeline guard.
5. Replace partial-composition tests with full-composition and invalid-config tests.
6. Update Worker conventions and this feature diary.

## Validation Plan

```bash
cd src/worker && go test ./pkg/app ./internal/app ./internal/runner
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Targeted scans:

```bash
rg -n "cfg\\.(SQLite\\.Path|Data\\.Dir).*!= \"\"|SnapshotPipeline == nil|Jobs != nil|ArtifactStore != nil" src/worker/pkg/app src/worker/internal/app
rg -n "legacy Worker config names" src/worker docs --glob '!docs/specs/*/DIARY.md'
```

## Completion Criteria

- No redundant composition guards remain in `pkg/app` or `internal/app`.
- Worker validation passes.
