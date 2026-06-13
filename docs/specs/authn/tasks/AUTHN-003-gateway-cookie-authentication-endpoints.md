---
id: AUTHN-003
feature: authn
title: Gateway opaque session cookie authentication
status: done
depends_on: [AUTHN-002]
blocks: [AUTHN-004]
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-003: Gateway Opaque Session Cookie Authentication

## Objective

Implement gateway opaque session cookie endpoints, session storage, and custom ASP.NET Core authentication handler for the browser UI/API surface.

## Scope

This task includes:

- Opaque `__Host-app-auth` session id cookies.
- `ISessionStore`, `SessionEntry`, and in-memory session storage.
- `AppCookieAuthenticationHandler`, `AppCookieDefaults`, and `AddAppCookie()`.
- `POST /login`, `POST /logout`, and `GET /auth/session`.
- In-memory login throttling.
- Password-only login across all password-bearing users, with session issuance only when exactly one stored hash matches.
- Trusted forwarded-header processing for the Docker reverse-proxy topology.
- Effective public HTTPS validation for `POST /login`.
- Same-origin rejection for unsafe methods using post-forwarding scheme, host, and effective port.
- `GATEWAY_PUBLIC_HOSTS` public host allowlisting.
- `401/403` API responses instead of redirects.


## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./AUTHN-002-password-persistence-and-bootstrap.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Login succeeds
  Given exactly one password-bearing user has a stored password_hash for the submitted password
  When the user posts the correct password to /login
  Then the gateway returns 204
  And emits a "__Host-app-auth" cookie with the required secure attributes
  And stores a fresh SessionEntry with absoluteExpiresAt 24 hours after issue
  And the session entry user id is the matching password owner's id

Scenario: Login fails
  Given no password-bearing user has a stored password_hash for the submitted password
  When the user posts the wrong password to /login
  Then the gateway returns 401
  And does not issue an auth cookie

Scenario: Duplicate password match fails closed
  Given two password-bearing users have password_hash values that verify the submitted password
  When the user posts the password to /login
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

Scenario: Login succeeds through trusted reverse proxy
  Given Gateway receives internal HTTP from Caddy
  And trusted forwarded headers set the effective public scheme to https
  And the Origin matches the effective public origin
  When the user posts valid credentials to /login
  Then the gateway returns 204
  And emits the required secure cookie attributes

Scenario: Login rejects ineffective public HTTPS context
  Given forwarded-header processing leaves the effective public scheme as http
  When the user posts valid credentials to /login
  Then the gateway returns 403
  And does not issue an auth cookie
```

## Done When

- Auth endpoint tests pass.
- Cookie attributes match `SPEC.md`.
- Session rotation, logout removal, unknown session, and expired session behavior are tested.
- Exact-one-match password ownership, duplicate-match failure, effective public HTTPS validation, forwarded public host constraints, and unsafe cross-site rejection are tested.

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
null
```
