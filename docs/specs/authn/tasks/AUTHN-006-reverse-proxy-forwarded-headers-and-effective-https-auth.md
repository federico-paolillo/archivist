---
id: AUTHN-006
feature: authn
title: Reverse-proxy forwarded headers and effective HTTPS auth
status: done
depends_on: [AUTHN-004]
blocks: [AUTHN-005]
parallel: false
exec_plan: ../plans/AUTHN-006-reverse-proxy-forwarded-headers.execplan.md
canonical: true
---

# AUTHN-006: Reverse-Proxy Forwarded Headers And Effective HTTPS Auth

## Objective

Update Gateway auth request interpretation for the primary Docker reverse-proxy topology, where upstream TLS termination has already established the public HTTPS context before Caddy forwards plaintext HTTP to Gateway.

## Scope

This task includes:

- Gateway forwarded-header startup configuration.
- `GATEWAY_PUBLIC_HOSTS` public host allowlisting.
- Effective public scheme validation for `POST /login`.
- Same-origin unsafe-method checks using post-forwarding scheme, host, and effective port.
- Gateway auth and routing tests for the `/api/*` public proxy contract.

## Out of Scope

This task does not include:

- Changing the `__Host-app-auth` cookie contract.
- Adding `GATEWAY_TRUSTED_PROXY_RANGES`.
- Exposing Gateway directly to the public Internet.
- Implementing deployment files when none exist.
- Changing UI browser routes or Gateway route paths.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `AMEND.md` as the source amendment note.
- `../SPEC.md` effective public HTTPS requirements.
- `docs/ARCHITECTURE.md` deployment topology.
- `.agents/skills/archivist-general/SKILL.md` configuration conventions.
- `.agents/skills/archivist-gateway/SKILL.md` Gateway middleware and auth conventions.

## Outputs

Expected outputs, files, behavior, or interfaces:

- Gateway processes trusted forwarded headers before auth middleware, authorization middleware, and endpoint mapping.
- `POST /login` returns `403 Forbidden` unless post-forwarding `Request.Scheme == "https"`.
- `SameOriginFilter` rejects missing, malformed, cross-scheme, cross-host, and cross-port origins or referers.
- `GATEWAY_PUBLIC_HOSTS` constrains accepted public host values.
- Login/logout cookie attributes remain unchanged.

## Expected Affected Areas

```text
src/gateway/
docs/specs/authn/
docs/ARCHITECTURE.md
.agents/skills/archivist-general/SKILL.md
.agents/skills/archivist-gateway/SKILL.md
docs/DESIGN.md
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `./AUTHN-003-gateway-cookie-authentication-endpoints.md`
- `./AUTHN-004-protect-ui-api-and-validate-auth-client-contract.md`
- `../plans/AUTHN-006-reverse-proxy-forwarded-headers.execplan.md`
- `docs/ARCHITECTURE.md`
- `docs/DESIGN.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-gateway/SKILL.md`

Do not load unrelated feature folders unless required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: Login succeeds through trusted reverse proxy
  Given Gateway receives internal HTTP from Caddy
  And trusted forwarded headers set the effective public scheme to https
  And the Origin matches the effective public origin
  When the browser posts valid credentials to /login
  Then the response status is 204
  And the exact secure cookie attributes are preserved

Scenario: Login rejects ineffective public HTTPS context
  Given forwarded-header processing leaves Request.Scheme as http
  When the browser posts valid credentials to /login
  Then the response status is 403
  And no auth cookie is issued

Scenario: Unsafe methods reject origin mismatches
  Given a valid auth cookie is present
  When POST /logout or DELETE /articles/{id} receives an Origin or Referer with mismatched effective public scheme, host, or port
  Then the response status is 403
```

## Done When

- Gateway forwarded-header behavior is implemented and tested.
- `GATEWAY_PUBLIC_HOSTS` is documented and enforced by startup/configuration tests.
- Login and same-origin tests cover effective public scheme, host, and port.
- Cookie issuance and clearing attributes remain unchanged.
- Required validation commands pass, or failures are recorded with cause.

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

Manual validation, if deployment files are added later:

- Verify public `/api/login`, `/api/logout`, and `/api/auth/session` are stripped to Gateway's unprefixed routes.
- Verify public root `/login`, `/articles`, and `/articles/{id}` remain UI routes.

## Dependencies

Depends on:

- `AUTHN-004`

Blocks:

- `AUTHN-005`

## ExecPlan

ExecPlan:

```text
../plans/AUTHN-006-reverse-proxy-forwarded-headers.execplan.md
```

## Open Questions

- None.

## Notes

- `AMEND.md` is an amendment/runbook note, not a canonical rebuild artifact.
- `AMEND.md`, `REVIEW.md`, `REFACTOR.md`, and `.claude/worktrees/` are local temporary review/worktree artifacts for this publication pass and must remain untracked and unstaged.
- The `src/gateway/.gitignore` `*.lscache` entry is intentional Gateway publication hygiene for local tooling cache files and is part of the publication set.
- The current harness validates Gateway's unprefixed route contract and forwarded-header behavior. It does not run Caddy, so public `/api/*` prefix stripping and root UI route ownership remain documented deployment assumptions.
