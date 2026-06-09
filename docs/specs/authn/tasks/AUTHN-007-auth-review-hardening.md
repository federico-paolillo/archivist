---
id: AUTHN-007
feature: authn
title: Auth review hardening
status: done
depends_on: [AUTHN-005]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# AUTHN-007: Auth review hardening

## Objective

Close review findings around canonical app-cookie settings and trusted forwarded-header deployment assumptions.

## Story / Context

As the Archivist operator, I want Gateway code to consume canonical auth cookie/session defaults directly so configuration cannot diverge from the security contract, while documenting that forwarded-header trust relies on Gateway remaining private behind ingress Caddy because the upstream load balancer has dynamic source IPs.

## Scope

This task includes:

- Remove the unnecessary app-cookie settings surface for canonical cookie name and session lifetime.
- Consume `AppCookieDefaults.CookieName` and `AppCookieDefaults.SessionLifetime` directly from auth handlers.
- Preserve the documented private reverse-proxy forwarded-header trust model.
- Add README and canonical design/architecture warnings for direct Gateway exposure.

## Out of Scope

This task does not include:

- Static trusted-proxy source IP configuration.
- Replacing the custom app-cookie authentication handler.
- Changing login/logout/session route contracts.

## Inputs

Required context:

- `../SPEC.md`
- `../PLAN.md`
- `docs/DESIGN.md`
- `docs/ARCHITECTURE.md`
- `.agents/skills/archivist-gateway/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Non-canonical cookie settings are ignored
  Given Gateway configuration attempts to change the auth cookie name or session lifetime
  When Gateway issues or authenticates app-cookie sessions
  Then Gateway still uses `AppCookieDefaults.CookieName` and `AppCookieDefaults.SessionLifetime`
```

```gherkin
Scenario: Forwarded-header trust remains deployment-scoped
  Given a rebuild of the production deployment docs
  Then Gateway remains private behind ingress Caddy
  And docs warn that direct Gateway exposure lets clients spoof forwarded scheme/host context
```

## Done When

- Gateway has no app-cookie settings object for canonical cookie name or session lifetime.
- Forwarded-header trust is documented as a private-network deployment invariant.
- Gateway validation passes or failures are recorded.
