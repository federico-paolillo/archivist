---
id: ARTPROC-004
feature: article-processing
title: Worker URL Resolver And HTML Fetcher
status: blocked
depends_on: [ARTPROC-002, ARTPROC-003]
blocks: [ARTPROC-005]
parallel: false
exec_plan: null
canonical: true
---

# ARTPROC-004: Worker URL Resolver And HTML Fetcher

## Objective

Implement Worker URL resolution and HTML fetching with `github.com/imroc/req/v3`, conservative limits, and ARC-coded failure mapping.

## Story / Context

As the Worker, I need to turn an article's original URL into bounded HTML bytes and a final resolved URL without introducing browser rendering or leaking low-level HTTP details into persisted article state.

## Scope

This task includes:

- Adding `github.com/imroc/req/v3` to the Worker module.
- Accepting only `http` and `https` URLs.
- Following at most 10 redirects.
- Applying a 20 second total timeout.
- Enforcing a 10 MiB maximum response body.
- Accepting only `text/html` and `application/xhtml+xml`.
- Returning the final redirected URL.
- Mapping resolution, HTTP status, content type, size, timeout, and unknown failures to ARC codes.
- Worker tests with local HTTP test servers.

## Out of Scope

This task does not include:

- Snapshot file writes.
- SQLite job state transitions.
- Extraction or summarization.
- Browser rendering.
- Automatic retries.

## Inputs

Required inputs, existing files, interfaces, or decisions:

- `../SPEC.md`
- Completed `ARTPROC-002`
- Completed `ARTPROC-003`
- `docs/conventions/ERRORS.md`
- `docs/conventions/WORKER.md`

## Outputs

Expected outputs, files, behavior, or interfaces:

- Worker fetcher/resolver service.
- ARC-coded public error mapping.
- Tests for success and failure classes.

## Expected Affected Areas

```text
src/worker/go.mod
src/worker/
```

## Required Context

Read before execution:

- `../SPEC.md`
- `../PLAN.md`
- `docs/conventions/ERRORS.md`
- `docs/conventions/GENERAL.md`
- `docs/conventions/WORKER.md`
- `./ARTPROC-002-define-shared-arc-error-code-convention.md`
- `./ARTPROC-003-worker-filesystem-artifact-access-layer.md`

Do not load unrelated feature folders unless listed here or required by dependencies.

## Acceptance Criteria

```gherkin
Scenario: HTML URL resolves successfully
  Given a URL redirects to a 200 text/html response
  When the Worker fetcher requests the URL
  Then it returns the final redirected URL
  And it returns the HTML bytes

Scenario: URL returns forbidden
  Given a URL returns 401 or 403
  When the Worker fetcher requests the URL
  Then it returns an ARC-002 public error

Scenario: URL returns not found
  Given a URL returns 404
  When the Worker fetcher requests the URL
  Then it returns an ARC-003 public error

Scenario: Response is not HTML
  Given a URL returns 200 application/pdf
  When the Worker fetcher requests the URL
  Then it returns an ARC-005 public error

Scenario: Response exceeds size limit
  Given a URL returns more than 10 MiB
  When the Worker fetcher requests the URL
  Then it returns an ARC-006 public error
```

## Done When

- Worker uses `github.com/imroc/req/v3` for article HTTP requests.
- URL scheme, redirect, timeout, body size, and content-type rules are enforced.
- Failure classes map to ARC-coded public errors.
- Tests cover redirects, 401/403, 404, timeout/5xx, non-HTML, and max body size.
- Task status and `PLAN.md` are updated if the task is completed.
- `DIARY.md` has an entry if implementation is performed.

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

## Dependencies

Depends on:

- `ARTPROC-002`
- `ARTPROC-003`

Blocks:

- `ARTPROC-005`

## ExecPlan

ExecPlan:

```text
null
```

## Open Questions

- None.

## Notes

- JavaScript-heavy empty app shells are accepted as raw HTML in this task. Detecting or rendering them is out of scope.
