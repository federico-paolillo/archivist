# DIARY.md

Append-only implementation notes for `worker-runtime-configuration`.

## 2026-05-16 — WCFG-001: Canonical Worker Config Loading

Status:
- done

Summary:
- Corrected Worker config loading from non-canonical legacy variables to canonical `ARCHIVIST_*` deployment variables.
- Reshaped Worker config structs to match configuro environment key nesting.
- Added strict load-time validation for required runtime values.
- Updated canonical docs so future rebuilds use `WCFG-001` instead of historical diary entries that referenced stale config names.

Validation:
- `cd src/worker && go test ./pkg/app/config ./pkg/app ./internal/runner ./internal/app` passed.
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.
- Stale config-name scan found only historical diary mentions and non-config `DataDir` identifiers; canonical docs and Worker config code no longer prescribe legacy application-prefixed Worker configuration.

## 2026-05-17 — WCFG-002: Non-Optional Worker Composition

Status:
- done

Summary:
- Tightened `pkg/app.NewApp` so valid Worker config always produces a complete composition root.
- Removed process-command validation for an impossible missing snapshot pipeline.
- Replaced partial-composition tests with invalid-config and full-composition tests.

Decisions:
- `NewApp` is now the boundary that rejects nil or invalid config before any app is returned.
- Lower-level nil checks in error types, optional provider fallback seams, and tests remain valid because they are not composition-root service guards.

Validation:
- `cd src/worker && go test ./pkg/app ./internal/app ./internal/runner` passed.
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.
- Targeted scans for redundant composition guards and stale canonical Worker config names passed.

## 2026-05-17 — WCFG-001/WCFG-002: Mandatory Jina Runtime Key

Status:
- done

Summary:
- Updated Worker runtime configuration so `JINA_API_KEY` is a required production value.
- Removed the Jina runtime toggle from the canonical Worker config surface.
- Updated composition tests and runner fixtures so valid Worker config includes a Jina key.

Decisions:
- `NewApp` remains the boundary that rejects invalid runtime config before returning an app.
- A composed Worker always has a Jina fallback adapter.

Validation:
- `cd src/worker && go test ./pkg/app/config ./pkg/app ./internal/markdown ./internal/pipeline` passed.
- `cd src/worker && go test ./internal/app ./internal/runner` passed after fixture updates.
- `cd src/worker && go tool lefthook run build` passed.
- `cd src/worker && go tool lefthook run format` passed.
- `cd src/worker && go tool lefthook run lint` passed.
- `cd src/worker && go tool lefthook run test` passed.
- Stale toggle scan passed for active canonical docs and Worker sources.

Follow-ups:
- None.

Canonical Updates:
- `docs/specs/worker-runtime-configuration/SPEC.md`
- `docs/specs/worker-runtime-configuration/PLAN.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-001-canonical-worker-config-loading.md`
- `docs/specs/worker-runtime-configuration/tasks/WCFG-002-non-optional-worker-composition.md`
- `docs/specs/worker-runtime-configuration/plans/WCFG-001-canonical-worker-config-loading.execplan.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
