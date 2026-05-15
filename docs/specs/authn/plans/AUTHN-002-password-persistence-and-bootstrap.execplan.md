---
id: AUTHN-002-PLAN
task: ../tasks/AUTHN-002-password-persistence-and-bootstrap.md
status: completed
canonical: true
---

# ExecPlan: AUTHN-002 Password Persistence And Bootstrap

## Objective

Create the gateway application services and SQLite persistence needed for password bootstrap, Argon2id PHC storage, and password verification.

## Linked Task

- `../tasks/AUTHN-002-password-persistence-and-bootstrap.md`

## Required Context

- `../tasks/AUTHN-002-password-persistence-and-bootstrap.md`
- `../SPEC.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/telegram-ingestion/tasks/TELING-001-persistence-contracts.md`
- `docs/specs/telegram-ingestion/plans/TELING-001-persistence-contracts.execplan.md`

## Assumptions

- `SQLITE_PATH` points at the gateway database file.
- The personal user row may not exist before auth bootstrap.
- `telegram_user_id` is nullable at rest and filled by Telegram ingestion later.

## Non-Goals

- Do not implement password rotation UI.
- Do not implement auth session storage in this password persistence task.

## Implementation Sequence

1. Add SQLite and Argon2id dependencies to the gateway application layer.
2. Define auth settings for `SQLITE_PATH` and `AUTH_BOOTSTRAP_PASSWORD`; Gateway receives these logical keys as `ARCHIVIST_SQLITE_PATH` and `ARCHIVIST_AUTH_BOOTSTRAP_PASSWORD` in environment-based deployment.
3. Ensure the `users` table exists with `id`, nullable `telegram_user_id`, and nullable `password_hash`.
4. Insert the fixed personal user row if missing.
5. If `password_hash` is missing, validate and hash `AUTH_BOOTSTRAP_PASSWORD`; otherwise leave the stored hash untouched.
6. Implement Argon2id PHC encoding and verification with constant-time hash comparison.
7. Add tests for bootstrap, missing bootstrap, validation, and verification.

## Validation Plan

```bash
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual checks:

- Confirm bootstrap plaintext is not logged or stored.

## Documentation Updates Required

- `docs/specs/authn/SPEC.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/specs/telegram-ingestion/SPEC.md`

## Risks

- A 2048-character secret can become a CPU-amplification vector without request-size and throttling controls.
- Future Telegram ingestion work must preserve the auth-owned `password_hash` column.

## Rollback / Recovery Notes

- Removing `password_hash` disables UI login and requires a new bootstrap on next initialization.

## Completion Criteria

- Bootstrap and password verification tests pass.
