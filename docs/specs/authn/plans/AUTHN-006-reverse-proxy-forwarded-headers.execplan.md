---
id: AUTHN-006-PLAN
task: ../tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md
status: completed
canonical: true
---

# ExecPlan: AUTHN-006 Reverse-Proxy Forwarded Headers

## Objective

Implement Gateway auth behavior for the primary reverse-proxy deployment topology while preserving the existing opaque session cookie contract.

## Linked Task

- `../tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md`

## Required Context

Start from the linked task's `Required Context` and the linked task file:

- `../tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md`

ExecPlan-specific context:

- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `AMEND.md`

## Assumptions

- Docker internal networking is the v0 trust boundary for forwarded headers.
- Caddy is the only component allowed to reach Gateway.
- The VPS/cloud-provider layer terminates TLS before traffic reaches Caddy.
- Caddy overwrites forwarded headers instead of passing client-supplied values through.
- Gateway route contracts remain unprefixed.

## Non-Goals

- Do not change the auth cookie name, value format, attributes, or session expiry.
- Do not add `GATEWAY_TRUSTED_PROXY_RANGES`.
- Do not expose Gateway directly to the public Internet while trusting forwarded headers.
- Do not make Caddy own or present the public TLS certificate in the primary topology.
- Do not add explicit root-level public Gateway routes for `/login`, `/logout`, or `/auth/session`.

## Implementation Sequence

1. Add Gateway configuration binding for `GATEWAY_PUBLIC_HOSTS` as a required production-style public host allowlist for forwarded host validation.
2. Configure ASP.NET Core forwarded headers to process `X-Forwarded-Proto` and `X-Forwarded-For`.
3. Set forwarded header `ForwardLimit = 1`.
4. Apply allowed public hosts from `GATEWAY_PUBLIC_HOSTS`; fail fast in production-style configuration when forwarded headers are enabled and no public hosts are configured.
5. Call `UseForwardedHeaders()` before authentication middleware, authorization middleware, and endpoint mapping.
6. Keep login cookie issuance and logout cookie clearing unchanged: `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/`, no `Domain`, and login without `Expires` or `Max-Age`.
7. Change login effective-scheme validation so `POST /login` returns `403 Forbidden` unless post-forwarding `Request.Scheme == "https"`.
8. Update `SameOriginFilter` to compare `Request.Scheme`, `Request.Host`, and effective port after forwarded-header processing.
9. Reject missing, malformed, cross-scheme, cross-host, and cross-port origins or referers for unsafe methods.
10. Preserve public routing assumptions: `/api/*` is stripped by Caddy before Gateway receives requests; Gateway keeps unprefixed route definitions.
11. Add or update Gateway tests for forwarded `https` login success, effective `http` login `403`, origin scheme mismatch, origin host mismatch, origin port mismatch, same-origin `POST /logout`, same-origin `DELETE /articles/{id}`, cookie attributes, `GATEWAY_PUBLIC_HOSTS`, and `/api/*` proxy routing assumptions.

## Validation Plan

```bash
cd src/gateway && dotnet format
cd src/gateway && dotnet build
cd src/gateway && dotnet test
```

Manual checks:

- Confirm `Set-Cookie` for login still contains `__Host-app-auth`, `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/`, no `Domain`, no `Expires`, and no `Max-Age`.
- Confirm logout clearing still contains `Max-Age=0`.
- Confirm a Caddy deployment uses `http://:443`, sets `X-Forwarded-Proto` to literal `https`, overwrites forwarded headers, and strips `/api` before forwarding to Gateway.

## Documentation Updates Required

- `docs/specs/authn/SPEC.md`
- `docs/specs/authn/PLAN.md`
- `docs/specs/authn/tasks/AUTHN-006-reverse-proxy-forwarded-headers-and-effective-https-auth.md`
- `docs/ARCHITECTURE.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/GATEWAY.md`
- `docs/DESIGN.md`

## Risks

- Trusting forwarded headers while Gateway is publicly exposed lets clients spoof the effective public scheme and origin.
- Running forwarded-header middleware after authentication or endpoint mapping leaves auth decisions based on internal HTTP context.
- Using Caddy `:443` instead of `http://:443` makes Caddy expect TLS and contradicts the primary upstream-terminated topology.
- Passing through client-supplied forwarded headers can bypass effective scheme, host, or port checks.

## Rollback / Recovery Notes

- Reverting this task restores direct request-scheme TLS checks, which will reject the primary upstream-terminated deployment.
- If forwarded-header validation blocks production traffic, first verify Caddy overwrites `X-Forwarded-Proto: https`, forwards the public `Host`, and that `GATEWAY_PUBLIC_HOSTS` matches that host.

## Completion Criteria

- Gateway tests cover the forwarded-header and effective-origin behavior listed in this ExecPlan.
- Required validation commands pass, or failures are recorded with cause.
- No required behavior remains only in `AMEND.md`.
