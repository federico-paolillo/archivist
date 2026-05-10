---
id: AUTHN-003-PLAN
task: ../tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md
status: completed
canonical: true
---

# ExecPlan: AUTHN-003 Gateway Opaque Session Cookie Authentication

## Objective

Implement the gateway HTTP authentication surface using opaque server-issued session id cookies, `ISessionStore`, and a custom ASP.NET Core authentication handler.

## Linked Task

- `../tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md`

## Required Context

- `../tasks/AUTHN-003-gateway-cookie-authentication-endpoints.md`
- `../SPEC.md`
- `docs/conventions/GATEWAY.md`

## Assumptions

- Gateway v0 is one process and may use in-memory throttling and in-memory session storage.
- UI and API share one trusted origin.
- Auth sessions expire absolutely on the server after 24 hours.

## Non-Goals

- Do not add encryption, signing, MACs, or embedded metadata to the cookie value.
- Do not add sliding expiry, refresh tokens, multiple concurrent sessions, or revocation lists beyond entry removal.
- Do not add an external throttle store in v0.
- Do not implement Redis session storage in v0.
- Do not redirect API clients to a login page.

## Implementation Sequence

1. Define `SessionEntry` and `ISessionStore` exactly as specified in `../SPEC.md`; `RemoveAsync` on missing keys is a no-op.
2. Implement the v0 in-memory store with `ConcurrentDictionary<string, SessionEntry>` and expired-entry cleanup on lookup and/or periodic sweep.
3. Define `AppCookieOptions` with `CookieName = "__Host-app-auth"` and `SessionLifetime = TimeSpan.FromHours(24)` defaults.
4. Implement `AddAppCookie(this AuthenticationBuilder builder, Action<AppCookieOptions>? configure = null)` with default scheme `"app-cookie"` and the custom handler.
5. Implement `AppCookieAuthenticationHandler` so missing cookie returns `NoResult`, unknown/expired sessions fail, expired sessions are removed, and valid sessions produce a one-claim `ClaimsPrincipal`.
6. Add auth endpoint route mapping for `POST /login`, `POST /logout`, and `GET /auth/session`.
7. Validate login request body size, password shape, TLS, and same-origin constraints before password verification.
8. Apply per-IP and global login throttling before Argon2id verification.
9. On successful login, remove any existing valid session from the request cookie, generate a fresh 32-byte CSPRNG base64url session id, store a new `SessionEntry`, set `__Host-app-auth` with the fixed attributes, and return `204`.
10. On failed login, return generic `401`, do not issue a cookie, and increment throttle counters.
11. On logout, remove the matching store entry when present, always emit the cookie-clearing header with `Max-Age=0`, and return `204`.
12. Ensure passwords, session ids, cookie values, and `Set-Cookie` headers are not logged.
13. Add tests for success, invalid login, throttling, oversized input, protected route `401`, same-origin rejection, session rotation, logout idempotency, unknown session, expired session cleanup, and handler principal shape.

## Validation Plan

```bash
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual checks:

- Inspect `Set-Cookie` for `__Host-app-auth`, `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, no `Expires`, and no `Max-Age` on login.
- Inspect logout `Set-Cookie` for `Max-Age=0`.

## Documentation Updates Required

- `docs/specs/authn/SPEC.md`
- `docs/conventions/GATEWAY.md`

## Risks

- Custom auth handlers can bypass normal `[Authorize]` behavior if the scheme name, ticket, or principal are wrong.
- SameSite is not sufficient by itself for all same-site request-forgery cases.
- Missing store cleanup can retain expired session entries until gateway restart.

## Rollback / Recovery Notes

- Disabling auth endpoint mapping should also remove `RequireAuthorization` from UI API groups to avoid unusable protected routes.
- Clearing the in-memory session store logs out all v0 browser sessions.

## Completion Criteria

- Auth endpoint, handler, session-store, and security-control tests pass.
