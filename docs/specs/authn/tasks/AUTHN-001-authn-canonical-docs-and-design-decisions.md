---
id: AUTHN-001
feature: authn
title: Authn canonical docs and design decisions
status: done
depends_on: []
blocks: [AUTHN-002]
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-001: Authn canonical docs and design decisions

## Objective

Create the authn feature artifacts and promote durable v0 UI/API authentication decisions into canonical architecture, design, and convention documents.

## Scope

This task includes:

- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/DIARY.md`
- Authn task files and ExecPlan links.
- Index, architecture, design, and convention updates.

## Out of Scope

This task does not include:

- Gateway code implementation.
- UI code implementation.

## Acceptance Criteria

- Authn feature artifacts exist under `docs/specs/authn`.
- `docs/specs/INDEX.md` lists `authn`.
- Durable cookie, password, bootstrap, and replica decisions are not hidden only in task files.

## Done When

- Canonical documents describe the accepted v0 auth design.
- No open planning questions block implementation.

## Validation

Required checks:

```bash
rg -n "authn|AUTHN|AUTH_BOOTSTRAP_PASSWORD|__Host-app-auth" docs
```

## Dependencies

Depends on:

- None.

Blocks:

- `AUTHN-002`
