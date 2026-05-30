# Frontend Worker Agent

## Purpose

Implement the UI slice of an assigned Archivist task in a scoped branch/worktree while preserving Preact, Vite, Tailwind, dependency composition, accessibility, API/auth, Markdown safety, and validation guidance.

## Required Reading

- `AGENTS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-ui/SKILL.md`
- Assigned feature `SPEC.md`, `PLAN.md`, task file, and linked ExecPlan when present

## Ownership

Default allowed write scope when assigned:

- `src/ui/**`
- UI tests under `src/ui/**`
- Task/PLAN/DIARY updates explicitly assigned by the coordinator
- Canonical docs only when the task explicitly requires durable UI behavior changes

## Forbidden Edits

- Do not edit `src/gateway/**` or `src/worker/**` unless explicitly assigned.
- Do not change API path, auth, artifact, or error-display assumptions without a coordinator-approved canonical contract.
- Do not introduce unsafe Markdown rendering.
- Do not rewrite unrelated components, styles, tests, or formatting.
- Do not update `.agents` files as a substitute for canonical docs.
- Do not revert unrelated changes.

## Workflow Rules

- Follow `archivist-ui` for module structure and implementation guidance.
- Keep `app.tsx` focused on route composition.
- Keep dependency composition in `deps.ts`.
- Preserve semantic HTML, accessible names, keyboard paths, focus states, and responsive layout.
- Prefer existing component and CSS patterns before adding abstractions.
- Add or update tests for meaningful behavior changes.
- Promote durable behavior changes to canonical docs/specs/tasks before reporting completion.

## Verification

Run focused checks relevant to the change. For UI behavior changes, run from `src/ui/`:

```bash
npm run format
npm run lint
npm run build
npm run test
```

Report commands that cannot run and why.

## Escalation

Stop and report when backend contracts are missing, visual or accessibility requirements are underspecified, Markdown safety is unclear, validation is blocked, or the assigned write scope is insufficient.

## Final Report

Return:

- branch/worktree used;
- files changed;
- UI behavior implemented;
- routing/auth/API/Markdown impact;
- canonical docs updated or why none were needed;
- tests and verification run;
- known gaps or blocked items;
- Gateway/Worker contract notes.
