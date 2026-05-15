---
id: AUTHN-005
feature: authn
title: Security validation pass
status: done
depends_on: [AUTHN-006]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-005: Security validation pass

## Objective

Validate the implemented v0 auth surface against the feature's accepted security requirements.

## Scope

This task includes:

- Cookie attribute verification.
- Oversized login rejection.
- Login throttling verification.
- Protected endpoint unauthenticated `401`.
- Same-origin unsafe request rejection.
- Forwarded-header effective public HTTPS verification.
- Gateway validation commands.

## Out of Scope

This task does not include:

- Penetration testing.
- Dedicated observability stack.
- Multi-replica auth validation.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`

## Acceptance Criteria

- Required validation commands pass, or failures are recorded with cause.
- Security-relevant behavior is covered by automated tests.
- Reverse-proxy forwarded-header behavior is covered by automated tests.
- Any durable decision discovered during implementation is promoted to canonical docs.

## Done When

- The feature is marked `done`.
- `DIARY.md` records validation.
- `docs/specs/INDEX.md` lists `authn` as `done`.

## Validation

Required checks:

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Validation completed on 2026-05-15:

- `cd src/gateway && dotnet format` — passed.
- `cd src/gateway && dotnet build` — passed with 0 warnings and 0 errors.
- `cd src/gateway && dotnet test` — passed: 124 tests, 0 failed.

## Dependencies

Depends on:

- `AUTHN-006`

Blocks:

- None.
