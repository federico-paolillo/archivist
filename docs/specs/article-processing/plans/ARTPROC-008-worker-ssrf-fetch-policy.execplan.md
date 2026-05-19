---
task: ../tasks/ARTPROC-008-worker-ssrf-fetch-policy.md
status: completed
canonical: true
---

# ExecPlan: ARTPROC-008 Worker SSRF Fetch Policy

## Objective

Implement a reusable Worker SSRF guard and wire it into article fetching so arbitrary submitted URLs cannot reach local, internal, metadata, Docker-network, or special network targets.

## Context

Required context:

- `../tasks/ARTPROC-008-worker-ssrf-fetch-policy.md`
- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`
- `docs/conventions/GENERAL.md`
- `REVIEW.md`

## Implementation Plan

1. Promote durable behavior to canonical docs.
   - Add `ARC-017` to `docs/conventions/ERRORS.md` and `src/worker/internal/arc`.
   - Update article-processing requirements for HTTPS-only Worker fetches, one redirect, SSRF blocks, and DNS failure mapping.
   - Update Worker HTTP conventions with guarded-client and logging rules.

2. Add `src/worker/internal/ssrf`.
   - Implement a `Guard` with explicit constructor dependencies for logger, resolver, and dialer.
   - Implement URL validation for initial and redirect phases.
   - Normalize hostnames with IDNA lookup rules, lowercase, and trailing-dot removal.
   - Reject non-HTTPS, userinfo, empty host, non-443 explicit ports, IP literals, single-label names, localhost names, Docker names, metadata names, and invalid domain names.
   - Resolve DNS during dialing, reject any denied resolved IP, and dial only a validated IP address.
   - Return errors that unwrap to `arc.ErrSSRFDetected` for policy blocks and `arc.ErrURLResolutionFailed` for DNS failures.
   - Log allow decisions at debug and block decisions at warn using redacted URL fields and URL hash.

3. Wire Worker HTTP.
   - Create the SSRF guard in `pkg/app.NewApp`.
   - Configure the shared `*req.Client` with one redirect, SSRF redirect policy, guarded `SetDial`, 20 second timeout, and HTTP/3 disabled.
   - Keep fetcher focused on request execution, status mapping, content-type validation, and body limits.

4. Update tests.
   - Add focused `internal/ssrf` tests for allowed URLs, blocked URLs, DNS failures, mixed DNS answers, redirect policy, and logging.
   - Update fetcher tests to use HTTPS URLs without an explicit port and explicit `:443` through an injected resolver/dialer setup.
   - Add pipeline regression for `ARC-017` persistence.
   - Assert composition-root wiring for guarded dial and redirect policy.

5. Validate and record outcome.
   - Run Worker build, format, lint, and test commands.
   - Update task, ExecPlan, `PLAN.md`, `DIARY.md`, and `REVIEW.md`.

## Validation

```bash
cd src/worker && go tool lefthook run build
cd src/worker && go tool lefthook run format
cd src/worker && go tool lefthook run lint
cd src/worker && go tool lefthook run test
```

## Completion Notes

- Completed by ARTPROC-008.
- Validation passed: Worker build, format, lint, and test lefthook commands.
