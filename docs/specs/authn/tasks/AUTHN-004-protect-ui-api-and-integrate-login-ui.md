---
id: AUTHN-004
feature: authn
title: Protect UI API and integrate login UI
status: blocked
depends_on: [AUTHN-003]
blocks: [AUTHN-005]
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-004: Protect UI API and integrate login UI

## Objective

Integrate the browser UI with gateway auth endpoints and ensure protected UI API behavior.

## Scope

This task includes:

- Password-only login form.
- Session check against `GET /auth/session`.
- Logout action using `POST /logout`.
- Client handling for `401` responses.
- A protected gateway route used to validate cookie enforcement.

## Out of Scope

This task does not include:

- Article list/detail UI.
- Password reset or rotation UI.
- PWA/offline behavior.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/UI.md`
- `docs/conventions/GATEWAY.md`

## Acceptance Criteria

```gherkin
Scenario: UI login flow
  Given the browser is unauthenticated
  When the user submits the correct 2048-character password
  Then the UI shows authenticated state
  And stores no password in local storage

Scenario: UI session expires
  Given the UI receives 401 from /auth/session
  Then it shows the login form
```

## Done When

- UI auth tests pass.
- Gateway protected route test passes.
- Login form accepts pasted 2048-character secrets.

## Validation

Required checks:

```bash
cd src/ui && npm run lint
cd src/ui && npm run build
cd src/ui && npm run test
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `AUTHN-003`

Blocks:

- `AUTHN-005`
