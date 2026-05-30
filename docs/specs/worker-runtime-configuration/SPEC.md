---
slug: worker-runtime-configuration
title: Worker Runtime Configuration
status: done
depends_on: [markdown-extraction]
canonical: true
---

# Worker Runtime Configuration

## Intent

Make Worker runtime configuration match the canonical deployment surface and prevent regressions where code loads non-canonical environment variable names.

## Scope

- Worker configuration loading through configuro.
- `ARCHIVIST_`-prefixed environment variables for canonical Worker runtime keys.
- Strict Worker load-time validation for required production configuration.
- ALM corrections for prior completed tasks that introduced or depended on non-canonical Worker config names.

## Out of Scope

- Gateway configuration implementation.
- UI build-time configuration.
- New runtime configuration keys.
- Replacing configuro.

## Requirements

- REQ-001: Worker configuration must load from `ARCHIVIST_`-prefixed environment variables.
- REQ-002: Worker code must expose configuration through `src/worker/pkg/app/config/config.go`; production code outside that package must not read environment variables directly.
- REQ-003: Worker config must use configuro-compatible structure for canonical keys:
  - `ARCHIVIST_SQLITE_PATH` -> `SQLite.Path`
  - `ARCHIVIST_DATA_DIR` -> `Data.Dir`
  - `ARCHIVIST_JINA_API_KEY` -> `Jina.API.Key`
  - `ARCHIVIST_LLM_PROVIDER` -> `LLM.Provider`
  - `ARCHIVIST_LLM_API_KEY` -> `LLM.API.Key`
  - `ARCHIVIST_LLM_MODEL` -> `LLM.Model`
- REQ-004: `SQLITE_PATH`, `DATA_DIR`, `JINA_API_KEY`, and `LLM_API_KEY` when `LLM_PROVIDER=anthropic` are required at `config.Load()` time.
- REQ-005: `JINA_API_KEY` is required and the old Jina runtime toggle is not a supported Worker configuration key.
- REQ-006: `LLM_PROVIDER=anthropic` is the only supported v0 provider value.
- REQ-007: Worker config tests must cover defaults, required values, environment loading, and unsupported provider validation.
- REQ-008: Worker composition must treat `config.Load()` and `NewApp` validation as the boundary for required runtime values, so returned `App` instances have DB, job repository, artifact store, provider adapters, and processing pipeline wired.
- REQ-009: Worker command handlers must not revalidate service presence already guaranteed by `NewApp`.

## Acceptance Criteria

```gherkin
Scenario: canonical environment variables configure the Worker
  Given ARCHIVIST_SQLITE_PATH, ARCHIVIST_DATA_DIR, ARCHIVIST_JINA_API_KEY, and ARCHIVIST_LLM_API_KEY are set
  And optional LLM values are set with ARCHIVIST-prefixed canonical names
  When Worker configuration loads
  Then configuro binds those values to the Worker config structs

Scenario: missing required values fail at load time
  Given one of SQLITE_PATH, DATA_DIR, JINA_API_KEY, or required Anthropic LLM_API_KEY is missing
  When Worker configuration loads
  Then loading fails with a configuration validation error

Scenario: stale application-prefixed variables are not used
  Given legacy application-prefixed Worker variables are set
  When Worker configuration loads
  Then those variables do not satisfy Worker runtime configuration

Scenario: strict config produces a fully wired Worker
  Given Worker configuration validates successfully
  When NewApp constructs the composition root
  Then all required Worker services are non-optional
  And the process command does not check for a missing processing pipeline
```

## Rebuild Notes

Rebuild agents must implement Worker runtime configuration from this spec, its task files, and `docs/ARCHITECTURE.md`, not from historical diary entries that mention legacy application-prefixed variables or flattened configuro-derived names. Implementation agents may use `.agents/skills/archivist-worker/SKILL.md` for Worker coding guidance.
