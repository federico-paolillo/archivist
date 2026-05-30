---
id: WCFG-001-PLAN
task: ../tasks/WCFG-001-canonical-worker-config-loading.md
status: completed
canonical: true
---

# ExecPlan: WCFG-001 Canonical Worker Config Loading

## Objective

Correct Worker runtime configuration so documented deployment keys configure the Worker binary through configuro.

## Implementation Sequence

1. Change Worker configuro env prefix from `APP` to `ARCHIVIST`.
2. Reshape `config.Root` into configuro-bound nested structs for `SQLite.Path`, `Data.Dir`, `Jina.API.Key`, `LLM.Provider`, `LLM.API.Key`, and `LLM.Model`.
3. Add defaults for app name, debug, Anthropic provider, and the v0 Anthropic model.
4. Add load-time validation through configuro validation: require `SQLITE_PATH`, `DATA_DIR`, `JINA_API_KEY`, and `LLM_API_KEY` when provider is Anthropic; reject unsupported providers.
5. Remove manual Jina environment overrides and any Jina runtime toggle.
6. Update Worker consumers to read the new config shape.
7. Update tests for defaults, canonical env loading, required values, and unsupported providers.
8. Update canonical ALM, architecture, and convention docs.

## Validation Plan

```bash
cd src/worker && go test ./pkg/app/config ./pkg/app ./internal/runner ./internal/app
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
rg "legacy Worker config names" src/worker docs
```

## Documentation Updates Required

- `.agents/skills/archivist-worker/SKILL.md`
- `.agents/skills/archivist-general/SKILL.md`
- `docs/ARCHITECTURE.md`
- `docs/specs/INDEX.md`
- Affected completed task notes in article-processing, markdown-extraction, and summary-generation.

## Completion Criteria

- Canonical env names configure Worker runtime behavior.
- Missing required deployment config fails during `config.Load()`.
- Stale legacy Worker config names are not prescribed by canonical docs.
