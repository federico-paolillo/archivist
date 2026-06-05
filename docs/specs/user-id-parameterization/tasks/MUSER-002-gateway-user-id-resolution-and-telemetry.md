---
id: MUSER-002
feature: user-id-parameterization
title: Gateway user-id resolution and telemetry
status: done
depends_on: [MUSER-001]
blocks: [MUSER-005]
parallel: true
exec_plan: ../plans/MUSER-002-gateway-user-id-resolution-and-telemetry.execplan.md
canonical: true
---

# MUSER-002: Gateway User-ID Resolution And Telemetry

## Objective

Remove runtime Gateway ownership hardcoding, support password-only login across multiple password-bearing users, resolve Telegram senders from persisted mappings, and attach `user_id` logs/spans when known.

## Scope

This task includes:

- Password store and login session ownership using all non-empty Argon2id PHC hashes and exact-one-match session issuance.
- Auth bootstrap seeding of the personal row's fixed Telegram sender id `1559957191` while preserving existing non-null `telegram_user_id`.
- Telegram sender-to-user lookup by `users.telegram_user_id`.
- Telegram ingestion article/job ownership from resolved `user_id`.
- Gateway article and Telegram telemetry using `user_id`.
- Gateway tests for auth, duplicate password match failure, bootstrap Telegram sender seeding, Telegram mapping, authorization, ownership, and telemetry.

## Out of Scope

This task does not include:

- Worker changes.
- UI changes unless required by a Gateway test helper.
- Registration or user-management endpoints.
- Deployment-configured personal Telegram sender ids such as `settings.PersonalTelegramUserId` or `Telegram:AllowedUserId`.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/MUSER-002-gateway-user-id-resolution-and-telemetry.execplan.md`
- `../../../ARCHITECTURE.md`
- `../../../DESIGN.md`
- `../../authn/SPEC.md`
- `../../telegram-ingestion/SPEC.md`
- `../../ui-endpoints/SPEC.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Gateway resolves runtime user ownership
  Given a login or Telegram request resolves a persisted user
  When Gateway creates a session, article, job, or authenticated article operation
  Then it uses the resolved user_id rather than the personal ULID constant

Scenario: Password-only login supports multiple password-bearing users
  Given multiple users have non-empty Argon2id PHC password_hash values
  And exactly one hash verifies the submitted password
  When Gateway handles POST /login
  Then the session is issued for the matching user's id

Scenario: Duplicate password matches fail closed
  Given multiple password_hash values verify the submitted password
  When Gateway handles POST /login
  Then no session is issued
  And the response is unauthorized

Scenario: Auth bootstrap hardcodes the personal Telegram sender id
  Given the personal user row has telegram_user_id null
  When Gateway auth bootstrap runs
  Then telegram_user_id is set to 1559957191
  And no Telegram allowed-user configuration is read for that seed

Scenario: Auth bootstrap preserves an existing Telegram sender mapping
  Given the personal user row has a non-null telegram_user_id
  When Gateway auth bootstrap runs
  Then telegram_user_id is unchanged
```

## Done When

- Gateway source no longer uses the personal user constant outside auth bootstrap.
- Password login verifies every non-empty Argon2id PHC hash and issues a session only when exactly one candidate matches.
- Multiple password-bearing rows are valid, and duplicate password matches fail closed.
- Auth bootstrap sets the personal row's `telegram_user_id` to `1559957191` only when null, preserves existing non-null values, and does not use `settings.PersonalTelegramUserId` or `Telegram:AllowedUserId`.
- Unknown Telegram senders create no rows and receive no reply.
- Gateway tests cover mapped and unmapped Telegram senders, row-derived login sessions, duplicate password matches, bootstrap Telegram sender seeding, and cross-user article isolation.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `MUSER-001`

Blocks:

- `MUSER-005`

## ExecPlan

ExecPlan:

```text
../plans/MUSER-002-gateway-user-id-resolution-and-telemetry.execplan.md
```

## Open Questions

- None.
