---
id: AUTHN-002
feature: authn
title: Password persistence and bootstrap
status: done
depends_on: [TELING-001]
blocks: [AUTHN-003]
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-002: Password persistence and bootstrap

## Objective

Implement the gateway persistence and application services required to store and verify the personal user's Argon2id password hash.

## Scope

This task includes:

- Ensuring a `users` table with `id`, nullable `telegram_user_id`, and nullable `password_hash`.
- Bootstrapping `password_hash` from `AUTH_BOOTSTRAP_PASSWORD` only when missing.
- Creating the personal user row with `id = 01ASB2XFCZJY7WHZ2FNRTMQJCT` when missing.
- Seeding the personal user's `telegram_user_id` to `1559957191` only when it is null.
- Argon2id PHC hashing and verification.
- Strict 2048-character printable ASCII password validation.


## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`

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

Scenario: Bootstrap seeds missing Telegram sender mapping
  Given the personal user has telegram_user_id null
  When the gateway initializes auth storage
  Then telegram_user_id is set to 1559957191

Scenario: Bootstrap preserves existing Telegram sender mapping
  Given the personal user has a non-null telegram_user_id
  When the gateway initializes auth storage
  Then telegram_user_id is unchanged
```

## Done When

- Bootstrap and verification tests pass.
- Bootstrap plaintext is not logged or persisted.
- Existing non-null Telegram sender mappings are preserved.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:
- `TELING-001`

Blocks:

- `AUTHN-003`

## ExecPlan

ExecPlan:

```text
null
```
