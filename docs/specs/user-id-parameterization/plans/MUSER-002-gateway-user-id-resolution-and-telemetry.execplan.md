---
id: MUSER-002-PLAN
task: ../tasks/MUSER-002-gateway-user-id-resolution-and-telemetry.md
status: completed
canonical: true
---

# ExecPlan: MUSER-002 Gateway User-ID Resolution And Telemetry

## Objective

Remove runtime personal-user hardcoding from Gateway auth and Telegram ingestion while preserving existing HTTP behavior, support password-only login across multiple password-bearing users, and seed the personal Telegram sender mapping from the accepted fixed id.

## Linked Task

- `../tasks/MUSER-002-gateway-user-id-resolution-and-telemetry.md`

## Required Context

- `../SPEC.md`
- `../PLAN.md`
- `../tasks/MUSER-002-gateway-user-id-resolution-and-telemetry.md`
- `../../authn/SPEC.md`
- `../../telegram-ingestion/SPEC.md`
- `../../ui-endpoints/SPEC.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Assumptions

- Bootstrap remains the only Gateway production code allowed to contain the personal ULID.
- Bootstrap may hardcode the personal Telegram sender id `1559957191`.
- `Telegram:AllowedUserId` is not used as a runtime webhook authorization gate.

## Non-Goals

- New Gateway routes.
- User registration.
- UI changes.
- Deployment-configured personal Telegram sender ids such as `settings.PersonalTelegramUserId` or `Telegram:AllowedUserId`.

## Implementation Sequence

1. Change password-store abstractions to load every user row with a non-empty Argon2id PHC `password_hash`, returning each candidate's `user_id` and hash.
2. Make `POST /login` verify the submitted password against every loaded candidate.
3. Make `POST /login` create `SessionEntry` with the matching `user_id` only when exactly one candidate hash verifies; zero matches and duplicate matches return the existing generic unauthorized response.
4. Keep multiple password-bearing rows valid. Do not fail merely because more than one row has a non-empty `password_hash`.
5. Change auth bootstrap to set the personal row's `telegram_user_id` to `1559957191` when it is null, preserve an existing non-null value, and remove `settings.PersonalTelegramUserId` / `Telegram:AllowedUserId` as bootstrap inputs.
6. Add a Telegram user resolver that maps sender Telegram id to `users.id`.
7. Make Telegram webhook processing reject unmapped senders with no side effects and no reply.
8. Pass resolved `user_id` into Telegram ingestion persistence.
9. Remove user-row upsert and personal-id assignment from Telegram ingestion repository.
10. Add `user_id` to Gateway telemetry where session or Telegram mapping resolved a user.
11. Update Gateway tests for auth, duplicate password matches, bootstrap Telegram sender seeding, Telegram mapping, article isolation, and telemetry.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Documentation Updates Required

- Update existing Gateway-related specs only when implementation discovers a missing rebuild contract.

## Risks

- Duplicate password matches must fail closed, not silently choose one.
- More than one password-bearing row is valid when the submitted password matches exactly one row.
- Bootstrap must not overwrite an existing non-null `telegram_user_id`.
- Unknown Telegram users must remain silent to preserve current unauthorized behavior.

## Rollback / Recovery Notes

Revert Gateway auth/Telegram changes and restore the previous password-store contract if validation fails before integration.

## Completion Criteria

- Gateway tests prove exact-one-match row-derived login sessions, duplicate password match rejection, bootstrap Telegram sender seeding, and database-mapped Telegram ownership.
- Runtime Gateway source has no personal-user constant usage outside auth bootstrap.
