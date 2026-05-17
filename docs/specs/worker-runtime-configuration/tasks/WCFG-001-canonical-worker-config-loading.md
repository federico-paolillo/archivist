---
id: WCFG-001
feature: worker-runtime-configuration
title: Canonical Worker Config Loading
status: done
depends_on: []
blocks: [SUMGEN-002, SUMGEN-004]
parallel: false
exec_plan: ../plans/WCFG-001-canonical-worker-config-loading.execplan.md
canonical: true
---

# WCFG-001: Canonical Worker Config Loading

## Objective

Make Worker configuration consume canonical `ARCHIVIST_`-prefixed runtime variables through configuro and reject missing required deployment values at load time.

## Scope

- Replace Worker configuro env prefix `APP` with `ARCHIVIST`.
- Reshape Worker config structs to match configuro underscore-to-nesting behavior.
- Remove direct environment reads outside configuro.
- Add strict config validation for `SQLITE_PATH`, `DATA_DIR`, supported `LLM_PROVIDER`, and Anthropic `LLM_API_KEY`.
- Update Worker config consumers and tests.
- Promote the config-loading rule to canonical architecture and convention docs.

## Out of Scope

- Gateway config implementation.
- Replacing configuro.
- Adding providers other than Anthropic.

## Acceptance Criteria

```gherkin
Scenario: ARCHIVIST-prefixed config loads
  Given canonical ARCHIVIST-prefixed Worker environment variables
  When config.Load runs
  Then SQLite, data, Jina, and LLM fields are populated from configuro-bound structs

Scenario: required values fail fast
  Given SQLITE_PATH, DATA_DIR, or required Anthropic LLM_API_KEY is absent
  When config.Load runs
  Then it returns a validation error

Scenario: no direct env overrides exist
  Given Worker production code outside pkg/app/config
  When the code is inspected
  Then it does not call os.LookupEnv or otherwise bypass configuro for runtime config
```

## Done When

- Legacy application-prefixed Worker config loading is removed.
- `ARCHIVIST_SQLITE_PATH`, `ARCHIVIST_DATA_DIR`, `ARCHIVIST_JINA_ENABLED`, `ARCHIVIST_JINA_API_KEY`, `ARCHIVIST_LLM_PROVIDER`, `ARCHIVIST_LLM_API_KEY`, and `ARCHIVIST_LLM_MODEL` are covered by tests.
- Required-value and unsupported-provider tests are present.
- Canonical docs identify `WCFG-001` as the corrective task.
- Worker validation passes.

## Validation

```bash
cd src/worker && go test ./pkg/app/config ./pkg/app ./internal/runner ./internal/app
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```
