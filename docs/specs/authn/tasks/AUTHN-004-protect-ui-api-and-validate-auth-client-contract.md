---
id: AUTHN-004
feature: authn
title: Protect UI API and validate auth client contract
depends_on: [AUTHN-003]
blocks: [UIEND-002, UIEND-003, UI-002]
parallel: false
requires_exec_plan: false
canonical: true
---
# AUTHN-004: Protect UI API and validate auth client contract

## Objective

Validate the final Gateway auth contract consumed by UI clients and protected UI-facing APIs, without adding a production article or UI API surface in this auth feature.

## Scope

This task includes:

- Gateway auth endpoint behavior required by browser clients.
- `GET /auth/session` `204` and `401` behavior.
- `POST /logout` `204` and cookie clearing behavior.
- Protection-gate validation for UI-facing Gateway routes implemented by downstream features.
- A protected probe route used only by tests when needed to prove cookie enforcement, challenge behavior, and unsafe-method rejection before downstream UI endpoints exist.
- The protected probe route must be compiled, registered, or exposed only in the test host/test environment and must not be a production API surface, documented endpoint, or rebuild-visible route contract.
- Trusted reverse-proxy forwarded-header behavior for browser auth requests.
- Effective public HTTPS validation for `POST /login`.
- Same-origin unsafe-method rejection using post-forwarding scheme, host, and effective port.
- Authenticated user identity from session `ClaimTypes.NameIdentifier`.


## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `.agents/skills/archivist-gateway/SKILL.md`
- `docs/specs/ui/SPEC.md`
- `docs/specs/ui-endpoints/SPEC.md`

## Acceptance Criteria

```gherkin
Scenario: Protected endpoint rejects unauthenticated request
  Given the request has no valid app-cookie session
  When the browser calls a protected downstream UI-facing Gateway endpoint or test-only protected probe route
  Then the response status is 401

Scenario: Session endpoint confirms authenticated request
  Given the request has a valid app-cookie session
  When the browser calls GET /auth/session
  Then the response status is 204

Scenario: Login succeeds behind trusted reverse proxy
  Given trusted forwarded headers set the effective public scheme to https
  And the request Origin matches the effective public origin
  When the browser calls POST /login with valid credentials
  Then the response status is 204
  And the auth cookie attributes match the final auth contract

Scenario: Unsafe origin mismatch is rejected
  Given the request has a valid app-cookie session
  When the browser calls an unsafe downstream UI-facing Gateway endpoint or test-only protected probe route with a mismatched Origin
  Then the response status is 403
```

## Done When

- Gateway protected-route gate tests pass without adding a production auth-owned protected endpoint.
- Any protected probe route used for validation is test-only and unavailable in production route mapping.
- Auth endpoints retain the contracts consumed by `docs/specs/ui/SPEC.md` and `docs/specs/ui-endpoints/SPEC.md`.
- Reverse-proxy effective scheme, host, port, and same-origin behavior are covered by Gateway tests.

## Validation

Required checks:

```bash
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `AUTHN-003`

Blocks:

- `UIEND-002`
- `UIEND-003`
- `UI-002`
