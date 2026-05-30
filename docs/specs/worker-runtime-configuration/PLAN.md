---
feature: worker-runtime-configuration
status: done
canonical: true
---

# Feature Plan: Worker Runtime Configuration

## Purpose

Correct Worker runtime configuration binding so canonical deployment keys configure the executable and future rebuilds do not reintroduce legacy application-prefixed names.

## Task DAG

```text
WCFG-001 -> WCFG-002
```

## Tasks

| ID | Task | Status | Depends On | Blocks | Parallel | ExecPlan |
|---|---|---|---|---|---|---|
| `WCFG-001` | Canonical Worker config loading | done | - | `SUMGEN-002`, `SUMGEN-004` | no | `plans/WCFG-001-canonical-worker-config-loading.execplan.md` |
| `WCFG-002` | Non-Optional Worker Composition | done | `WCFG-001` | `SUMGEN-002`, `SUMGEN-004` | no | `plans/WCFG-002-non-optional-worker-composition.execplan.md` |

## Blocking Interfaces or Schemas

- Worker config package: `src/worker/pkg/app/config`.
- Worker composition root: `src/worker/pkg/app`.
- Canonical runtime keys from this feature spec, its task files, and `docs/ARCHITECTURE.md`.
- Required Worker runtime values include `SQLITE_PATH`, `DATA_DIR`, `JINA_API_KEY`, and `LLM_API_KEY` when `LLM_PROVIDER=anthropic`.

## Validation Sequence

```bash
cd src/worker && go test ./pkg/app/config ./pkg/app ./internal/runner ./internal/app
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

## Completion Criteria

- Worker config loads `ARCHIVIST_*` canonical deployment keys through configuro.
- Worker config validation rejects missing required production values.
- Existing Worker consumers use the config struct shape rather than reading environment variables directly.
- `NewApp` returns a fully wired Worker composition root or fails before returning an app.
- Worker commands do not revalidate service presence guaranteed by composition.
- Canonical docs point to `WCFG-001` and `WCFG-002` for the regression fixes.
