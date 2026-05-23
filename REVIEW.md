# Wave 5 Review

Date: 2026-05-20

## Scope

Reviewed Wave 5 implementation:

- `SUMGEN-002` Worker summary artifact access.
- `UI-002` UI routing, API base configuration, and auth shell.
- Wave 5 bookkeeping and `docs/MASTERPLAN.md` synchronization.

## Findings

### Resolved

- P2 UI: API base normalization allowed protocol-relative paths such as `//api`, which could produce off-origin requests. Fixed by collapsing multiple leading slashes and adding normalization coverage.
- P3 UI: UI tests used raw DOM event dispatch instead of the repository convention for Testing Library and `user-event`. Fixed by adding `@testing-library/react`, `@testing-library/user-event`, and rewriting route/auth tests around user-visible behavior.
- P2 Docs: `docs/MASTERPLAN.md` omitted `WCFG-001`/`WCFG-002` dependency edges into `SUMGEN-002` and `SUMGEN-004`. Fixed by adding worker-runtime-configuration DAG nodes, edges, classes, and clicks.
- P2 Docs: `SUMGEN-004` remained `blocked` even though its task dependencies are done. Fixed by explicitly documenting that it remains blocked because its ExecPlan is still `proposed`.
- P2 Docs: `summary-generation` and `ui` remained `draft` in `docs/specs/INDEX.md` and feature plan frontmatter despite completed tasks. Fixed by moving both feature statuses to `in_progress`.

### No Findings

- Worker review found no issues in `SUMGEN-002` artifact access. Coverage matches the task: deterministic `content.md` read behavior, atomic `summary.md` writes, failed-write cleanup, traversal and symlink rejection, no `summary.json`, and `ARC-016` wrapping for summary write failures.

## Validation

- `cd src/ui && npm run format` — passed.
- `cd src/ui && npm run lint` — passed.
- `cd src/ui && npm run build` — passed.
- `cd src/ui && npm run test` — passed.
- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed.

## Residual Risk

- `SUMGEN-004` is not ready until its proposed ExecPlan is accepted or updated.
- `UI-003` remains blocked on `UIEND-002`, so UI article workflows are intentionally incomplete after Wave 5.
