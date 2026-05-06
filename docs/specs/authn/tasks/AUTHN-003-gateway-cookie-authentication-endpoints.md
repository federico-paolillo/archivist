---
id: AUTHN-003
feature: authn
title: Gateway opaque session cookie authentication
status: blocked
depends_on: [AUTHN-002]
blocks: [AUTHN-004, UIEND-002, UIEND-003]
parallel: false
exec_plan: ../plans/AUTHN-003-gateway-cookie-authentication.execplan.md
canonical: true
---

# AUTHN-003: Gateway Opaque Session Cookie Authentication

## Objective

Implement gateway opaque session cookie endpoints, session storage, and custom ASP.NET Core authentication handler for the v0 browser UI/API surface.

## Scope

This task includes:

- Opaque `__Host-app-auth` session id cookies.
- `ISessionStore`, `SessionEntry`, and v0 in-memory session storage.
- `AppCookieAuthenticationHandler`, `AppCookieOptions`, and `AddAppCookie()`.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- In-memory login throttling.
- Same-origin rejection for unsafe methods.
- `401/403` API responses instead of redirects.

## Out of Scope

This task does not include:

- UI implementation.
- Multi-replica auth support.
- Redis session store implementation.
- Sliding expiry, refresh tokens, multiple concurrent sessions, or cryptographic transforms over the cookie value.

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./AUTHN-002-password-persistence-and-bootstrap.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`

## Acceptance Criteria

```gherkin
Scenario: Login succeeds
  Given a valid stored password_hash
  When the user posts the correct password to /login
  Then the gateway returns 204
  And emits a "__Host-app-auth" cookie with the required secure attributes
  And stores a fresh SessionEntry with absoluteExpiresAt 24 hours after issue

Scenario: Login fails
  Given a valid stored password_hash
  When the user posts the wrong password to /login
  Then the gateway returns 401
  And does not issue an auth cookie

Scenario: Existing session is rotated on login
  Given the request includes a valid "__Host-app-auth" cookie
  When login succeeds
  Then the previous session entry is removed
  And a new session id is issued

Scenario: Logout is idempotent
  Given the request may or may not include a valid "__Host-app-auth" cookie
  When the user posts to /logout
  Then the gateway removes any matching session entry
  And always clears the cookie with Max-Age=0

Scenario: Handler authenticates valid session
  Given the request includes a valid non-expired session id
  When the app-cookie handler authenticates the request
  Then HttpContext.User has authentication type "app-cookie"
  And contains only a NameIdentifier claim
```

## Done When

- Auth endpoint tests pass.
- Cookie attributes match `SPEC.md`.
- Session rotation, logout removal, unknown session, and expired session behavior are tested.
- Unsafe cross-site requests are rejected.

## Validation

Required checks:

```bash
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

## Dependencies

Depends on:

- `AUTHN-002`

Blocks:

- `AUTHN-004`

## ExecPlan

ExecPlan:

```text
../plans/AUTHN-003-gateway-cookie-authentication.execplan.md
```
