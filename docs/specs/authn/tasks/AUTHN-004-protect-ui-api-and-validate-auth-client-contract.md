---
id: AUTHN-004
feature: authn
title: Protect UI API and validate auth client contract
status: done
depends_on: [AUTHN-003]
blocks: [AUTHN-005, UI-002]
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-004: Protect UI API and validate auth client contract

## Objective

Validate Gateway auth endpoint behavior for UI clients and ensure protected UI API behavior.

## Scope

This task includes:

- Gateway auth endpoint behavior required by browser clients.
- `GET /auth/session` `204` and `401` behavior.
- `POST /logout` `204` and cookie clearing behavior.
- Protected UI-facing Gateway route behavior.
- A protected gateway route used to validate cookie enforcement.

## Out of Scope

This task does not include:

- Preact login form rendering.
- `/login/failed` browser route.
- Article list/detail UI.
- Password reset or rotation UI.
- PWA/offline behavior.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/GATEWAY.md`
- `docs/specs/ui/SPEC.md`

## Acceptance Criteria

```gherkin
Scenario: Protected endpoint rejects unauthenticated request
  Given the request has no valid app-cookie session
  When the browser calls a protected UI-facing Gateway endpoint
  Then the response status is 401

Scenario: Session endpoint confirms authenticated request
  Given the request has a valid app-cookie session
  When the browser calls GET /auth/session
  Then the response status is 204
```

## Done When

- Gateway protected route test passes.
- Auth endpoints retain the contracts consumed by `docs/specs/ui/SPEC.md`.

## Implementation Notes

- Status transitioned from `blocked` to `done` by explicit user assignment for Wave 4.
- No placeholder `/articles` route was added in this task. The concrete UI article endpoint contracts are owned by `ui-endpoints`; this task validates the existing authenticated Gateway route contract through `GET /auth/session` and preserves the auth client contract required by `docs/specs/ui/SPEC.md`.

## Validation

Required checks:

```bash
cd src/gateway && dotnet test
```

Result on 2026-05-12:

```text
Passed! - Failed: 0, Passed: 101, Skipped: 0, Total: 101
```

## Dependencies

Depends on:

- `AUTHN-003`

Blocks:

- `AUTHN-005`
