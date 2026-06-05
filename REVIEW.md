# Archivist Quality Review

Date: 2026-06-05

Scope: multi-agent review of the runnable application modules:

- Worker (`src/worker`)
- Gateway (`src/gateway`)
- UI (`src/ui`)

Review criteria:

- Quality
- Testability
- Maintainability
- SOLID / GRASP
- Consistency and coherence
- Specification compliance
- Idiomatic framework use and established practices

## Executive Verdict

Archivist is **not AI slop** and it is **not bad-junior work**. The project shows real architecture, canonical ALM discipline, meaningful module separation, explicit contracts, and a surprisingly broad test culture for a small personal system.

However, it is **not yet a project I would call fully production-proud without reservations**. The review found release-blocking risks in Worker and UI validation confidence, plus several Gateway consistency/security-hardening issues. The codebase feels like a strong intermediate-to-senior assisted implementation that needs a focused hardening pass, not like throwaway generated code.

**Global quality rating: B-**

Interpretation:

- **A/S** would require stable full validation, no P1 security/testability findings, tighter framework alignment, and clearer cross-resource consistency decisions.
- **B-** means the foundation is coherent and worth investing in, but the current branch still has enough sharp edges that I would not present it as finished craft.
- This is much closer to “quality project with fixable hardening gaps” than “bad junior developer slop.”

## Cross-Cutting Themes

### What is good

- The repository uses canonical specs and rebuild docs as source of truth instead of letting code drift silently.
- Major module boundaries are understandable: ASP.NET Core Gateway, Go Worker, Preact/Vite UI, SQLite as source of truth, filesystem artifacts under deterministic paths.
- The code usually prefers small abstractions around real boundaries: HTTP clients, artifact stores, repositories, auth/session services, API clients, and handoff interfaces.
- Tests are present across all modules and often cover behavior rather than only implementation details.
- Security concerns are visible in the design: SSRF guard, same-origin unsafe-method checks, cookie auth, public ARC error messages, Markdown HTML disabled.

### What is not good enough yet

- Full validation is not clean in this environment: Gateway cannot be validated because `dotnet` is unavailable; Worker validation exposes a proxy/SSRF failure path; one UI agent reproduced an order-dependent UI test failure even though the parent run passed.
- Some contracts are implemented by custom or partially custom machinery that should be either hardened or made more framework-native.
- Cross-resource consistency between SQLite rows and filesystem artifacts needs a documented failure model.
- Accessibility semantics in the UI lag behind the visual behavior.
- Several comments/settings still suggest earlier implementation phases or make canonical constants look configurable.

---

# Worker Review

**Module rating: B-**

The Worker is generally coherent and idiomatic Go in its broad shape. It has a clear composition root, explicit provider boundaries, centralized ARC error rendering, traversal-resistant artifact operations, and linear pipeline orchestration that is easy to reason about. The main release blocker is the interaction between SSRF protections and ambient proxy configuration.

## Worker Findings

### Finding: P1 SSRF-guarded HTTP client can be affected by environment proxies

**File:** `src/worker/pkg/app/app.go:181`

**Issue:** The production Worker `req` client wires the SSRF request middleware, redirect policy, guarded dialer, timeout, and HTTP-version restrictions, but does not explicitly disable or control proxy use (`src/worker/pkg/app/app.go:181-190`). The dial guard validates the actual dial address, which is only the intended article host when the HTTP transport dials targets directly. If environment proxy variables are honored, the dial target becomes the proxy rather than the article URL.

**Why it matters:** The article-processing spec makes the Worker the SSRF boundary and requires HTTPS-only processing plus redirect targets passing the same SSRF policy (`docs/specs/article-processing/SPEC.md:68-75`, `docs/specs/article-processing/SPEC.md:237-247`). A proxy-aware transport can either break legitimate tests/fetches or weaken target validation if a proxy is allowed to fetch the final target.

**Observed validation impact:** The Worker sub-agent reported guarded fetcher test failures where requests attempted `proxyconnect tcp` to `proxy:8080`, causing the SSRF dial guard to reject the proxy address. The affected tests are in `src/worker/internal/fetcher/fetcher_test.go:259-305`, and the guarded test helper configures `req.NewClient()` without an explicit proxy policy at `src/worker/internal/fetcher/fetcher_test.go:363-383`.

**Required fix:** Explicitly configure article-fetching HTTP clients to ignore ambient proxy variables, or design and document an SSRF-safe proxy mode that validates the ultimate target independent of the proxy dial. Make the fetcher tests hermetic by disabling proxy use in their test client setup.

### Finding: P2 CLI behavior remains coupled to `urfave/cli` closures

**File:** `src/worker/internal/app/program.go:30`

**Issue:** `CliProgram` registers commands and embeds action closures that directly inspect `*cli.Command` values (`NArg`, `Args().First`, `Bool`, `Duration`) before calling behavior (`src/worker/internal/app/program.go:30-83`). This works, but it mixes CLI framework plumbing with command behavior and makes executable-surface behavior harder to test without constructing framework state.

**Required fix:** Keep `program.go` as the registration layer, but move command-specific argument extraction and behavior calls into small command-named wrappers that produce plain typed inputs for framework-free behavior functions.

### Finding: P3 Stale placeholder comment remains in the production pipeline package

**File:** `src/worker/internal/pipeline/snapshot.go:44`

**Issue:** `NoOpMarkdownHandoff` still says it is a placeholder that `MDEXT-005` replaces, even though production wiring now creates the real Markdown and Summary handoffs (`src/worker/internal/pipeline/snapshot.go:44-48`, `src/worker/pkg/app/app.go:81-109`).

**Required fix:** Reword the comment to say this is a test/isolated-stage helper, not a production placeholder.

## Worker Strengths

- `NewApp` centralizes DB, artifact store, SSRF guard, shared HTTP client, fetcher, Markdown extractors, summarizer, handoffs, and pipeline wiring (`src/worker/pkg/app/app.go:47-111`).
- Summary generation is behind a Worker-owned `SummarizerService`, and the Anthropic SDK types stay private to the adapter (`src/worker/internal/summary/contract.go:30-36`, `src/worker/internal/summary/anthropic.go:25-38`).
- Artifact access owns `os.Root`, validates article IDs, and performs atomic temp-file promotion (`src/worker/internal/artifacts/store.go:23-52`, `src/worker/internal/artifacts/store.go:174-238`).
- ARC public formatting is centralized rather than duplicated across packages (`src/worker/internal/arc/arc.go:53-72`, `docs/ERRORS.md:49-52`).

## Worker Approval Status

**Changes requested.** The Worker is promising, but the proxy/SSRF issue should be fixed before claiming production quality.

---

# Gateway Review

**Module rating: B**

The Gateway is the strongest module structurally. It has a coherent ASP.NET Core Minimal API organization, clear Application-layer services, broad tests, typed endpoint results, and clean separation for auth, articles, persistence, Telegram, and observability. The main issues are consistency/security hardening, not architectural collapse.

## Gateway Findings

### Finding: P2 Fixed auth cookie contract is exposed as bindable configuration

**File:** `src/gateway/Archivist.Gateway.Application/Auth/Settings/AppCookieSettings.cs:12`

**Issue:** `CookieName` and `SessionLifetime` are public bindable settings with defaults (`src/gateway/Archivist.Gateway.Application/Auth/Settings/AppCookieSettings.cs:12-16`). Login/logout/session code uses those options to rotate sessions, issue cookies, clear cookies, and calculate absolute expiration (`src/gateway/Archivist.Gateway.Api/Auth/Handlers.cs:35-40`, `src/gateway/Archivist.Gateway.Api/Auth/Handlers.cs:78-118`). The canonical auth contract fixes the cookie name as `__Host-app-auth` and the server-side session lifetime as 24 hours (`docs/specs/authn/SPEC.md:75-84`).

**Risk:** A deployment override can silently diverge from canonical behavior, weaken the `__Host-` cookie contract by changing the name, or alter session lifetime without a spec change.

**Required fix:** Make these constants, or add startup validation that rejects non-canonical values unless the auth spec and architecture docs promote configurability as durable behavior.

### Finding: P2 Forwarded-header trust is broader than the reverse-proxy deployment contract

**File:** `src/gateway/Archivist.Gateway.Api/Auth/ServiceCollectionExtensions.cs:27`

**Issue:** The Gateway enables `X-Forwarded-For`, `X-Forwarded-Proto`, and `X-Forwarded-Host`, then clears `KnownIPNetworks` and `KnownProxies` (`src/gateway/Archivist.Gateway.Api/Auth/ServiceCollectionExtensions.cs:27-35`). The architecture expects the Gateway to run privately behind the trusted reverse proxy and makes forwarded scheme/host part of the auth boundary (`docs/ARCHITECTURE.md:201-205`, `docs/ARCHITECTURE.md:290`, `docs/ARCHITECTURE.md:350-358`).

**Risk:** If Gateway is ever reachable directly, a client can spoof forwarded scheme/host values. Host allowlisting helps, but it does not prove the request came through Caddy. This matters because `/login` requires effective HTTPS before issuing the secure cookie (`src/gateway/Archivist.Gateway.Api/Auth/Handlers.cs:37-40`).

**Required fix:** Prefer explicit trusted proxy/network configuration, or document and enforce the invariant that Gateway is unreachable except via the internal trusted reverse proxy network.

### Finding: P2 Normal delete validates but does not normalize valid ULID route values

**File:** `src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:94`

**Issue:** `DELETE /articles/{id}` validates with `Ulid.TryParse`, then passes the original `id` to `DeleteAsync` (`src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:88-105`). `GET /articles/{id}` and `DELETE /articles/{id}/force` normalize through `TryNormalizeUlid` before querying/deleting (`src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:63-74`, `src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:126-137`, `src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:185-194`).

**Risk:** If `Ulid.TryParse` accepts lower/mixed-case values, normal delete can return `404` for the same semantically valid article ID that detail and force-delete routes would find.

**Required fix:** Use the same normalization helper in `DeleteArticle` and pass the normalized value to the delete service.

### Finding: P2 SQLite/filesystem delete consistency has no durable failure model

**File:** `src/gateway/Archivist.Gateway.Application/Articles/Defaults/EfArticleDeleteService.cs:121`

**Issue:** The delete service deletes database rows inside a transaction, deletes the artifact directory, then commits (`src/gateway/Archivist.Gateway.Application/Articles/Defaults/EfArticleDeleteService.cs:106-133`). If artifact deletion succeeds but DB commit fails, rollback cannot restore artifacts. The UI endpoint spec requires successful delete to remove rows and `{DATA_DIR}/articles/{article_id}`, while artifact cleanup failures should leave DB state intact (`docs/specs/ui-endpoints/SPEC.md:69-78`, `docs/specs/ui-endpoints/SPEC.md:179-185`).

**Risk:** This is a real cross-resource atomicity gap. It is not fully avoidable with plain SQLite plus filesystem operations, but the current design does not document the failure model or repair story.

**Required fix:** Add a durable design decision for delete consistency semantics. Options include a documented repair path, post-commit best-effort cleanup with compensating state, or an explicit acceptance of the current rare commit-failure risk.

## Gateway Non-Finding: null `started_at` force delete is spec-compliant

A sub-review initially flagged force-deleting `running` jobs with `started_at IS NULL`. On re-check, this is explicitly canonical: `REQ-009` says such jobs are stale for force-delete recovery (`docs/specs/job-recovery/SPEC.md:63-65`). Therefore this review does **not** count it as a defect.

## Gateway Strengths

- Minimal API modules are well organized around `Endpoints.cs`, `Handlers.cs`, and DTOs.
- Article handlers use typed result unions and `TypedResults`, which is idiomatic for Minimal APIs (`src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:20-51`, `src/gateway/Archivist.Gateway.Api/Articles/Handlers.cs:57-82`).
- Read paths use `AsNoTracking()` for list/detail and delete prechecks (`src/gateway/Archivist.Gateway.Application/Articles/Defaults/EfArticleQueryService.cs:33-35`, `src/gateway/Archivist.Gateway.Application/Articles/Defaults/EfArticleQueryService.cs:82-85`, `src/gateway/Archivist.Gateway.Application/Articles/Defaults/EfArticleDeleteService.cs:68-70`).
- Auth has focused abstractions for password validation, hashing, password storage, session storage, throttle, and cookie auth (`src/gateway/Archivist.Gateway.Application/Auth/Extensions/ServiceCollectionExtensions.cs:38-55`).
- Forwarded headers are applied before auth middleware, which is the correct order for the effective-scheme auth behavior (`src/gateway/Archivist.Gateway.Api/Program.cs:41-48`).
- Gateway tests are broad across auth, article APIs, delete behavior, Telegram, persistence, notification dispatch, and configuration.

## Gateway Approval Status

**Changes requested, but close.** The Gateway is not slop; it is a solid module with hardening and consistency work remaining.

---

# UI Review

**Module rating: B-**

The UI is coherent and mostly aligned with the repo-local Preact/Vite conventions: route composition is centralized, dependencies are injected, API credentials are included, Markdown HTML is disabled, and the force-delete control uses backend-computed eligibility. The module’s biggest issue is validation confidence: one full-suite run passed in the parent environment, while the UI sub-agent reproduced an order-dependent full-suite failure that passed in isolation.

## UI Findings

### Finding: P1 UI test isolation is suspect and can produce order-dependent failures

**File:** `src/ui/src/app.test.tsx:114`

**Issue:** `mountAt()` creates a new root, appends it to `document.body`, and renders Preact into that root (`src/ui/src/app.test.tsx:114-121`). `afterEach()` calls `render(null, document.body)`, replaces body children, resets history, and restores mocks, but does not unmount the actual per-test root (`src/ui/src/app.test.tsx:190-195`). The sub-agent reproduced a full-suite failure in `keeps the detail pane blank when a pending detail resolves after returning to /articles`, while the same test passed in isolation. That test uses a deferred detail request and route transition (`src/ui/src/app.test.tsx:367-395`), which is exactly the kind of case that leaks if effects or routers are not unmounted cleanly.

**Parent validation note:** The parent run of `cd src/ui && npm run test -- --run` passed 28 tests, but still emitted repeated `window.scrollTo` jsdom warnings. The disagreement between runs does not clear the risk; it strengthens the conclusion that the suite should be made hermetic and repeatable.

**Required fix:** Track the mounted root(s) and call `render(null, root)` in `afterEach()`, or switch consistently to Testing Library/Preact cleanup patterns. Resolve or cancel deferred work where applicable. The goal is repeated full-suite stability, not isolated-test success.

### Finding: P2 Selected article state is visual-only

**File:** `src/ui/src/pages/articles/components/article-master-list.tsx:31`

**Issue:** Article rows are rendered as buttons and selection is represented only by the CSS class `article-row-selected` (`src/ui/src/pages/articles/components/article-master-list.tsx:31-50`). Assistive technologies do not receive a programmatic selected/current state. The UI spec requires selected article navigation and selected-state behavior (`docs/specs/ui/SPEC.md:72-75`).

**Required fix:** Prefer route links with `aria-current="page"`, or implement a proper selected-state pattern such as `aria-pressed`, `aria-selected` in a listbox/grid, or equivalent accessible semantics.

### Finding: P2 User menu declares ARIA menu semantics without implementing the pattern

**File:** `src/ui/src/pages/articles/components/user-menu.tsx:20`

**Issue:** The popover has `role="menu"`, but the child button is a normal button without `role="menuitem"`, and there is no menu focus management or Escape behavior (`src/ui/src/pages/articles/components/user-menu.tsx:19-24`). The spec only requires a user icon menu containing `Logout`; it does not require the full ARIA menu pattern (`docs/specs/ui/SPEC.md:69-71`).

**Required fix:** Either remove `role="menu"` and use a simple disclosure/popover with native buttons, or fully implement the ARIA menu pattern.

### Finding: P3 jsdom `scrollTo` warnings are not handled in shared test setup

**File:** `src/ui/vite.config.ts:16`

**Issue:** Vitest is configured for jsdom but has no `setupFiles` hook for browser API shims (`src/ui/vite.config.ts:16-23`). Test runs emit repeated `Not implemented: Window's scrollTo() method` warnings.

**Required fix:** Add a small test setup file and wire it through `test.setupFiles` to stub `window.scrollTo`.

## UI Strengths

- Route composition is centralized in `app.tsx`, and page implementations live under route-oriented directories (`src/ui/src/app.tsx:11-28`).
- Dependency composition follows poor-man’s DI: `Deps` exposes the API client and `makeDeps()` builds it from `VITE_API_BASE_PATH` (`src/ui/src/deps.ts:3-10`).
- API base path normalization and credentialed fetch are implemented consistently (`src/ui/src/deps/api-client.ts:37-56`, `src/ui/src/deps/api-client.ts:60-83`, `src/ui/src/deps/api-client.ts:106-136`).
- Markdown rendering disables raw HTML and sets safe `target`/`rel` attributes for rendered links (`src/ui/src/pages/articles/components/markdown-content.tsx:9-13`, `src/ui/src/pages/articles/components/markdown-content.tsx:20-40`).
- Force delete is gated on backend-provided `canForceDelete`, which is the right side of the contract for job recovery (`src/ui/src/pages/articles/components/article-actions.tsx:18-33`, `docs/specs/job-recovery/SPEC.md:57-74`).

## UI Approval Status

**Changes requested.** The UI is close, but test isolation and accessibility semantics need to be cleaned up before it can be called polished.

---

# Architect Summary

## Quality Rating

**Global rating: B-**

| Module | Rating | Verdict |
|---|---:|---|
| Worker | B- | Strong core boundaries, but proxy/SSRF behavior is a real release blocker. |
| Gateway | B | Best-structured module; needs security/config hardening and delete consistency clarity. |
| UI | B- | Coherent and usable, but test isolation/accessibility issues prevent polish. |

## Is this a project to be proud of?

**Yes, conditionally.** You can be proud of the architecture and direction. The project has enough structure, tests, contracts, and security awareness that it does not resemble low-effort AI slop. It looks like serious work.

But you should **not** be proud of it as “finished quality” until the P1s are fixed and validation is consistently green. The difference is important:

- Proud of the foundation: **yes**.
- Proud to ship without caveats today: **not yet**.
- Worth throwing away as bad junior code: **absolutely not**.

## Highest-Priority Hardening Plan

1. Fix Worker proxy/SSRF behavior and make guarded fetcher tests hermetic.
2. Fix UI test cleanup so full-suite runs are repeatable, then add the `scrollTo` setup shim.
3. Normalize normal Gateway delete IDs exactly like detail and force-delete.
4. Decide and document Gateway SQLite/filesystem delete consistency semantics.
5. Lock or validate canonical auth cookie/session settings.
6. Replace the partial ARIA menu and visual-only selected state with accessible, framework-friendly semantics.

## Final Assessment

Archivist is a **real quality project in progress**, not a bad junior output. It is currently in the “good bones, needs hardening” stage. If the P1s and P2s above are addressed, it can become a project you can confidently show as thoughtful, maintainable, and idiomatic.
