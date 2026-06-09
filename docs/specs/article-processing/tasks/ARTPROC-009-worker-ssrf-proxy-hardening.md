---
id: ARTPROC-009
feature: article-processing
title: Worker SSRF proxy hardening
status: done
depends_on: [ARTPROC-008]
blocks: []
parallel: false
exec_plan: null
canonical: true
---

# ARTPROC-009: Worker SSRF proxy hardening

## Objective

Close the review finding that Worker article HTTP clients could be affected by ambient proxy environment variables.

## Story / Context

As the Archivist operator, I want the Worker SSRF guard to validate direct article fetch targets rather than host-level proxy dials so article fetching remains safe and tests remain hermetic on proxy-configured machines.

## Scope

This task includes:

- Configure Worker article-fetching/shared HTTP clients to ignore ambient proxy environment variables.
- Make guarded fetcher and process tests hermetic under proxy-populated environments.
- Reword stale snapshot pipeline handoff comments when they refer to superseded placeholder work.
- Update canonical SSRF requirements to make proxy-independent fetch behavior durable.

## Out of Scope

This task does not include:

- Configurable outbound proxy support.
- SSRF-safe proxy design.
- Gateway-side SSRF filtering.
- New ARC codes.

## Inputs

Required context:

- `../SPEC.md`
- `../PLAN.md`
- `./ARTPROC-008-worker-ssrf-fetch-policy.md`
- `.agents/skills/archivist-worker/SKILL.md`

## Acceptance Criteria

```gherkin
Scenario: Worker article fetch ignores ambient proxies
  Given the Worker process environment contains HTTP proxy variables
  When the Worker fetches an article URL
  Then the article HTTP client does not route the request through the ambient proxy
  And SSRF validation still applies to the original URL, redirects, DNS results, and direct dial target
```

## Done When

- Worker HTTP client construction disables ambient proxy use.
- Guarded fetcher/process tests are hermetic under proxy variables.
- Worker validation passes or failures are recorded.
- Canonical docs record proxy-independent SSRF behavior.
