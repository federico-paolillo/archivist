---
id: WCFG-002
feature: worker-runtime-configuration
title: Non-Optional Worker Composition
status: done
depends_on: [WCFG-001]
blocks: [SUMGEN-002, SUMGEN-004]
parallel: false
exec_plan: ../plans/WCFG-002-non-optional-worker-composition.execplan.md
canonical: true
---

# WCFG-002: Non-Optional Worker Composition

## Objective

Remove partial Worker composition after strict runtime configuration made required values non-optional.

## Scope

- Validate config at the start of `pkg/app.NewApp`.
- Return an error for nil or invalid config.
- Build DB, jobs repository, artifact store, HTTP fetcher, Markdown providers, Markdown handoff, and snapshot pipeline unconditionally when config is valid.
- Remove process-command service-presence validation for the snapshot pipeline.
- Update tests to cover invalid config and full composition.

## Out of Scope

- New configuration keys.
- Gateway configuration.
- Summary pipeline integration.
- Lower-level optional test seams such as nil fallback extractors.

## Acceptance Criteria

```gherkin
Scenario: valid config builds a full Worker graph
  Given Worker config contains required SQLite, data, and Anthropic API key values
  When NewApp runs
  Then DB, Jobs, ArtifactStore, Fetcher, Markdown providers, and SnapshotPipeline are present

Scenario: invalid config fails before composition
  Given Worker config is nil or missing a required value
  When NewApp runs
  Then it returns an error and no app

Scenario: process command trusts composition
  Given process is called with a composed Worker app
  When the command runs
  Then it invokes SnapshotPipeline without a redundant nil-pipeline check
```

## Done When

- `NewApp` no longer has conditional construction based on empty config values.
- `process` no longer returns a missing-pipeline configuration error.
- Tests cover full composition and invalid config.
- Worker validation passes.

## Validation

```bash
cd src/worker && go test ./pkg/app ./internal/app ./internal/runner
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```
