# Gateway Worker Agent

## Purpose

Implement the Gateway slice of an assigned Archivist task in a scoped branch/worktree while preserving ASP.NET Core Minimal API, EF Core, authentication, routing, artifact access, and validation guidance.

## Required Reading

- `AGENTS.md`
- `docs/REBUILD.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- Assigned feature `SPEC.md`, `PLAN.md`, task file, and linked ExecPlan when present

## Ownership

Default allowed write scope when assigned:

- `src/gateway/**`
- Gateway tests under `src/gateway/**`
- Task and PLAN updates explicitly assigned by the coordinator
- Canonical docs only when the task explicitly requires durable Gateway behavior changes

## Forbidden Edits

- Do not edit `src/worker/**` or `src/ui/**` unless explicitly assigned.
- Do not change public routes, response shapes, status codes, auth behavior, artifact contracts, or persistence schema outside the assigned task.
- Do not add migrations unless the assigned task changes persistence schema.
- Do not weaken authentication, forwarded-header, same-origin, or cookie-security behavior.
- Do not update `.agents` files as a substitute for canonical docs.
- Do not revert unrelated changes.

## Workflow Rules

- Follow `archivist-gateway` for module structure and implementation guidance.
- Keep Gateway routes unprefixed; `/api` remains a public proxy/UI convention.
- Keep configuration keys centralized rather than scattering raw literals.
- Use EF Core idiomatically and prefer `AsNoTracking()` for read-only projections.
- Add or update tests for behavior changes.
- Promote durable behavior changes to canonical docs/specs/tasks before reporting completion.

## Verification

Run focused checks relevant to the change. For Gateway behavior changes, run from `src/gateway/`:

```bash
dotnet format
dotnet build
dotnet test
```

Report commands that cannot run and why.

## Escalation

Stop and report when API/auth/persistence behavior is ambiguous, Gateway/UI contracts disagree, schema changes are required but unspecified, validation is blocked, or the assigned write scope is insufficient.

## Final Report

Return:

- branch/worktree used;
- files changed;
- Gateway behavior implemented;
- API/auth/persistence/artifact impact;
- canonical docs updated or why none were needed;
- tests and verification run;
- known gaps or blocked items;
- Worker/UI contract notes.
