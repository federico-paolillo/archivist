---
id: AUTHN-002
feature: authn
title: Password persistence and bootstrap
status: done
depends_on: [AUTHN-001, TELING-001]
blocks: [AUTHN-003]
parallel: false
exec_plan: ../plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md
canonical: true
---

# AUTHN-002: Password persistence and bootstrap

## Objective

Implement the gateway persistence and application services required to store and verify the personal user's Argon2id password hash.

## Scope

This task includes:

- Ensuring a `users` table with `id`, nullable `telegram_user_id`, and nullable `password_hash`.
- Bootstrapping `password_hash` from `AUTH_BOOTSTRAP_PASSWORD` only when missing.
- Argon2id PHC hashing and verification.
- Strict 2048-character printable ASCII password validation.

## Out of Scope

This task does not include:

- Login HTTP endpoints.
- UI screens.
- Password rotation UI.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/plans/TELING-001-persistence-contracts.execplan.md`

## Acceptance Criteria

```gherkin
Scenario: Bootstrap missing password hash
  Given AUTH_BOOTSTRAP_PASSWORD is valid
  And the personal user has no password_hash
  When the gateway initializes auth storage
  Then the personal user's password_hash is an Argon2id PHC string

Scenario: Existing password hash is present
  Given the personal user already has password_hash
  When the gateway initializes auth storage
  Then AUTH_BOOTSTRAP_PASSWORD is not required
  And the existing hash is preserved
```

## Done When

- Bootstrap and verification tests pass.
- Bootstrap plaintext is not logged or persisted.
- The task's ExecPlan is accepted before implementation begins.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `AUTHN-001`
- `TELING-001`

Blocks:

- `AUTHN-003`

## ExecPlan

ExecPlan:

```text
../plans/AUTHN-002-password-persistence-and-bootstrap.execplan.md
```
