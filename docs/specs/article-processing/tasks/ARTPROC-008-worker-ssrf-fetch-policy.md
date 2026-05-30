---
id: ARTPROC-008
feature: article-processing
title: Worker SSRF Fetch Policy
status: done
depends_on: [ARTPROC-007]
blocks: []
parallel: false
exec_plan: ../plans/ARTPROC-008-worker-ssrf-fetch-policy.execplan.md
canonical: true
---

# ARTPROC-008: Worker SSRF Fetch Policy

## Objective

Harden Worker article fetching against SSRF while preserving Gateway's permissive URL ingestion behavior.

## Story / Context

As the Worker, I need to process arbitrary URLs submitted through Telegram without allowing those URLs or their redirects to reach internal, local, metadata, Docker-network, or otherwise special network targets.

## Scope

This task includes:

- Adding reusable Worker SSRF guard code under `src/worker/internal/ssrf`.
- Adding `ARC-017` for SSRF policy blocks.
- Requiring article fetch URLs and redirect targets to be absolute `https` URLs.
- Treating omitted HTTPS ports as port `443`, allowing explicit `:443`, and rejecting every other explicit port.
- Limiting redirects to one.
- Rejecting userinfo, empty hosts, invalid hostnames, all IP literals, single-label hostnames, localhost names, Docker-internal names, cloud metadata names, and private or special resolved IP ranges.
- Resolving DNS at dial time and dialing only a validated address.
- Logging SSRF allow decisions at debug level and block decisions at warn level with redacted URL fields.
- Disabling HTTP/3 for the guarded shared Worker HTTP client.
- Worker tests for policy, fetcher wiring, and pipeline persistence of `ARC-017`.

## Out of Scope

This task does not include:

- Gateway-side SSRF validation.
- Runtime configuration that weakens SSRF defaults.
- Docker firewall rules, `DOCKER-USER` rules, or an egress proxy implementation.
- Automatic banning of users who trigger SSRF policy.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- `../PLAN.md`
- `../plans/ARTPROC-008-worker-ssrf-fetch-policy.execplan.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `.agents/skills/archivist-general/SKILL.md`
- `REVIEW.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Reusable Worker SSRF guard package.
- Shared `*req.Client` configured with the guard in `pkg/app.NewApp`.
- `ARC-017` public message and sentinel.
- Fetcher and pipeline behavior that persists `[ARC-017] Archivist refused to process suspicious URL.` for SSRF policy blocks.
- Updated canonical docs and review finding status.

## Expected Affected Areas

```text
docs/ERRORS.md
.agents/skills/archivist-worker/SKILL.md
docs/specs/article-processing/
REVIEW.md
src/worker/internal/arc/
src/worker/internal/ssrf/
src/worker/internal/fetcher/
src/worker/internal/pipeline/
src/worker/pkg/app/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/BOOKKEEPING.md`
- `docs/ERRORS.md`
- `.agents/skills/archivist-general/SKILL.md`
- `.agents/skills/archivist-worker/SKILL.md`
- `REVIEW.md`
- `../plans/ARTPROC-008-worker-ssrf-fetch-policy.execplan.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: HTTPS URL without explicit port is allowed
  Given a submitted article URL is "https://example.com/path"
  And DNS resolves example.com to only public IPs
  When the Worker fetch policy evaluates the URL
  Then the URL is allowed
  And the effective port is "443"

Scenario: Explicit HTTPS port 443 is allowed
  Given a submitted article URL is "https://example.com:443/path"
  And DNS resolves example.com to only public IPs
  When the Worker fetch policy evaluates the URL
  Then the URL is allowed

Scenario: Non-HTTPS URL is blocked
  Given a submitted article URL is "http://example.com/path"
  When the Worker processes the job
  Then processing fails with ARC-017

Scenario: Suspicious target is blocked
  Given a submitted article URL uses a localhost, IP literal, single-label, Docker-internal, metadata, private, link-local, or other special target
  When the Worker processes the job
  Then processing fails with ARC-017

Scenario: DNS failure remains URL resolution failure
  Given a submitted article URL has a syntactically valid public-looking hostname
  And DNS resolution fails
  When the Worker processes the job
  Then processing fails with ARC-001

Scenario: Redirect target is policy checked
  Given a submitted article URL redirects
  When the Worker follows the redirect
  Then the redirect target must pass the same SSRF policy
  And a second redirect is rejected
```

## Done When

- Canonical docs define Worker SSRF behavior and `ARC-017`.
- `internal/ssrf` validates URLs, redirect targets, DNS answers, and dial targets.
- The shared Worker HTTP client uses the SSRF guard and disables HTTP/3.
- Fetcher/pipeline failures persist `ARC-017` for SSRF blocks and keep DNS failures as `ARC-001`.
- Tests cover policy allows, policy blocks, redirects, logging, and pipeline persistence.
- Task status and `PLAN.md` are updated.
- `DIARY.md` has an implementation entry with validation results.

## Validation

Required checks:

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

Manual validation, if any:

- None.

Validation performed:

- `cd src/worker && go tool lefthook run build` — passed.
- `cd src/worker && go tool lefthook run format` — passed.
- `cd src/worker && go tool lefthook run lint` — passed.
- `cd src/worker && go tool lefthook run test` — passed.

## Dependencies

Depends on:

- `ARTPROC-007`

Blocks:

- None.

## ExecPlan

ExecPlan:

```text
../plans/ARTPROC-008-worker-ssrf-fetch-policy.execplan.md
```

## Open Questions

- None.

## Notes

- Gateway intentionally remains permissive. Worker processing is the SSRF security boundary.
- Docker network hardening is documented as defense-in-depth guidance only in this task.
