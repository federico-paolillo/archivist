---
name: archivist-worker
description: Use when implementing, reviewing, or planning Archivist Worker module changes under src/worker, including Go code, Worker CLI commands, configuration, provider adapters, artifact writes, pipeline orchestration, jobs, errors, logging, and Worker validation.
allowed-tools: Read Write Edit Grep Glob Bash
---

# Archivist Worker

## Purpose

Use this skill for repository-specific Worker work in Archivist. This skill does not replace canonical docs; it provides Worker implementation guidance, including Go language standards for Worker code.

## Required Context

Start with the repository orientation bundle:

```text
AGENTS.md
docs/REBUILD.md
docs/specs/INDEX.md
```

Load additional context by trigger:

- `docs/BOOKKEEPING.md`: changing task status, `PLAN.md`, specs, ExecPlans, or dependency/concurrency state.
- `docs/ARCHITECTURE.md`: changing executables, service boundaries, storage, integrations, deployment, runtime topology, or configuration semantics.
- `docs/DESIGN.md`: relying on or changing durable decisions or rebuild-relevant tradeoffs.
- `docs/ERRORS.md`: changing persisted article-processing failure behavior, ARC codes, public messages, or error classification.
- `docs/ARTIFACTS.md`: changing artifact paths, filenames, atomic writes, or article data layout.
- Relevant feature `SPEC.md`, `PLAN.md`, task files, and linked ExecPlans before implementing a task.

Do not load unrelated feature folders unless dependencies or canonical docs require them.

## Source-Of-Truth Rules

- Canonical Markdown controls required behavior. Existing Go code is implementation evidence, not source of truth.
- Work only on tasks marked `ready` or explicitly assigned by the user.
- Do not invent durable behavior. If the task lacks required behavior, update the relevant spec/task or mark the task blocked.
- Promote durable changes to canonical docs, not just skills, code comments, or coordinator notes.
- After material task implementation, update task frontmatter and the feature `PLAN.md` when AGENTS/BOOKKEEPING require it. Diary or coordinator notes are optional non-canonical coordination only.

## Worker Stack

Worker code lives under `src/worker/` and targets Go 1.26.

- Keep `CGO_ENABLED=0`; the Worker binary must remain a single executable.
- Target Linux x64 first. Support macOS Apple Silicon when practical.
- Follow the repository's existing Go project layout.
- Use lightweight interfaces to aid tests. Do not speculate with broad abstractions.
- Do not use stubs, placeholders, or fake production behavior. If implementation is blocked, challenge the design or record the missing decision.

## Modern Go Standards

- Prefer simple, explicit designs over premature abstraction.
- Make lightweight interfaces at consumer boundaries to aid testing. Do not introduce interfaces for unsupported variation.
- Apply SOLID and GRASP only where they reduce coupling or clarify ownership; KISS and YAGNI are stronger defaults.
- Prefer explicit constructor injection over globals, service locators, or hidden runtime resolution.
- Keep package boundaries narrow. Do not leak third-party SDK types across owned domain or orchestration interfaces unless that is the documented public contract.
- Treat error wrapping as an API decision. Wrap with `%w` only when callers should inspect the underlying error.
- Keep package-specific error types, constructors, and classification helpers in `errors.go` when a package has enough error logic to justify separation.
- Use structured logging for observable runtime behavior. Do not log secrets or large user/provider payloads.
- Add focused tests around behavior and boundaries. Prefer executable or public-surface tests when the task changes executable behavior.

For Go 1.26 Worker code, prefer:

- `any` instead of `interface{}`.
- `time.Since(start)` and `time.Until(deadline)`.
- `errors.Is` for sentinel matching through wrapped errors.
- `errors.AsType[T](err)` instead of `errors.As(err, &target)`.
- `bytes.Cut`, `strings.Cut`, `strings.CutPrefix`, and `strings.CutSuffix` instead of index-and-slice parsing.
- `slices` and `maps` helpers instead of open-coded collection utilities.
- `cmp.Or` for simple fallback selection.
- `for i := range n` when ranging over an integer count.
- `maps.Keys` and `maps.Values` iterators with `slices.Collect` or `slices.Sorted` where appropriate.
- `t.Context()` in tests instead of manually deriving from `context.Background()`.
- `json:",omitzero"` for zero-sensitive struct, time, duration, slice, and map fields when Go 1.24+ semantics are intended.
- `b.Loop()` for benchmark loops.
- `strings.SplitSeq`, `strings.FieldsSeq`, and byte equivalents when iterating over split parts.
- `sync.WaitGroup.Go` instead of manual `Add` plus goroutine plus `Done`.
- `new(value)` instead of temporary local variables solely to take an address.
- traversal-resistant filesystem APIs such as `os.Root` or `os.OpenInRoot` for potentially untrusted names inside trusted roots.

## Worker-Specific Checkpoints

- Composition root: `pkg/app.NewApp` owns Worker wiring. Use explicit constructor injection. Test new `App` fields or service creation logic in `pkg/app/app_test.go`.
- CLI commands: `internal/app/program.go` owns `urfave/cli/v3` registration and typed flag extraction. Command action functions live in command-named files and must not accept or reference CLI types. Production behavior changed through a CLI command requires executable-surface tests through command registration.
- Configuration: use `pkg/app/config` and configuro with the `ARCHIVIST_` prefix and canonical nested shape. Production Worker code outside config must not read environment variables directly.
- Provider boundaries: orchestration depends on Archivist-owned interfaces. Pipeline code must not import or expose external provider SDK request/response types.
- HTTP: outbound Worker HTTP goes through an injected `*req.Client`. Do not use `req.C()`. Direct outbound `net/http` is forbidden except for documented SDK bridge cases; tests may use `httptest`.
- Logging: pipeline orchestration owns structured logs for article-processing stages. Provider adapters return data and errors; they do not emit info/error logs. Secrets and full article/summary content must never be logged.
- Errors: package error infrastructure belongs in `errors.go`. ARC sentinels and public messages belong in `internal/arc`; persistence must use ARC public messages, not raw diagnostic errors.
- Filesystem/artifacts: use traversal-resistant APIs when operating under `DATA_DIR`, especially article artifact paths. Artifact writes under `/data` must be atomic.
- Configuration changes: update `docs/ARCHITECTURE.md` and affected feature specs/tasks when adding Worker configuration keys.

## Configuration

Worker configuration is loaded from environment variables or equivalent deployment secret mechanisms.

- Use configuro in `src/worker/pkg/app/config`.
- Load environment variables with the `ARCHIVIST_` prefix.
- Canonical deployment variables include `ARCHIVIST_SQLITE_PATH`, `ARCHIVIST_DATA_DIR`, `ARCHIVIST_JINA_API_KEY`, `ARCHIVIST_LLM_PROVIDER`, `ARCHIVIST_LLM_API_KEY`, and `ARCHIVIST_LLM_MODEL`.
- Because configuro maps underscores to nested keys, use the canonical nested shape documented by Worker architecture and current Worker feature specs: `SQLite.Path`, `Data.Dir`, `Jina.API.Key`, `LLM.Provider`, `LLM.API.Key`, and `LLM.Model`.
- `SQLITE_PATH`, `DATA_DIR`, `JINA_API_KEY`, and `LLM_API_KEY` when `LLM_PROVIDER=anthropic` are required at `config.Load()` time.
- `LLM_PROVIDER=anthropic` is the only currently supported provider value.
- `LLM_MODEL` defaults to `claude-3-5-haiku-20241022`.
- `pkg/app.NewApp` validates config before constructing services and returns either an error or a fully wired Worker composition root.
- Worker command handlers must not re-check for missing DB, artifact store, provider adapters, or processing pipeline services that `NewApp` guarantees.
- `JINA_API_KEY` is required for Jina Reader fallback and must not be logged.
- Extend worker configuration tests for default values, required values, and environment loading whenever configuration changes.

## Pure DI And Composition Root

- The Worker uses Pure DI: explicit constructor injection, no IoC container, no service locator, no globals.
- `pkg/app.NewApp`, called from `cmd/worker/main.go`, is the composition root.
- Every collaborator a service needs must be a constructor parameter.
- Long-lived singletons such as DB, HTTP clients, loggers, and adapters live as fields of `App`.
- Per-request objects are created inside the call that needs them.
- Optional `createXxx` factory functions may partition complex subgraphs but remain internal to the composition root.
- The composition root may import any internal package. Other packages must not import sibling packages solely to wire them.
- Every field added to `App` must be covered by a test in `pkg/app/app_test.go`.

## Provider Boundaries

- Markdown extraction uses a Worker-owned `MarkdownExtractor` abstraction. The local go-readability implementation and Jina Reader implementation sit behind that abstraction.
- Summary generation uses a Worker-owned `SummarizerService` abstraction. Claude/Anthropic is the first provider implementation and should use `github.com/anthropics/anthropic-sdk-go`.
- Pipeline code must not import or expose external provider SDK request/response types.
- Use official SDKs when suitable SDKs exist. If no suitable official SDK exists, a small internal adapter is acceptable.
- Snapshot and Markdown stages are intermediate pipeline stages. Final Worker success means `summary.md` has been atomically promoted and the article/job/notification terminal state has been committed.

## HTTP Client

- All outbound Worker HTTP calls go through `github.com/imroc/req/v3`.
- Direct outbound `net/http` is forbidden except in tests or SDK bridge cases that require a plain `*http.Client`; use `reqClient.GetClient()` for those bridge cases.
- Never use the package-level global client `req.C()`.
- Construct owned `*req.Client` instances in the composition root and inject them.
- The composition root owns shared concerns such as user-agent, timeout, and HTTP middleware.
- Adapters receive a preconfigured client and must not mutate it or set global options.
- Article fetching must use the composition-root SSRF-guarded client defined by canonical specs/tasks.
- Tests construct isolated `*req.Client` instances and must not share or mutate another test's client.

## Structured Logging

- The Worker logs to stdout using Go `log/slog`.
- Pipeline orchestration owns structured log entries for article-processing stages.
- Provider adapters must not accept a `*slog.Logger` and must not emit `slog.Info` or `slog.Error`.
- Adapter `slog.Debug` is allowed only for low-value diagnostics orchestration cannot observe.
- API keys and full article/summary content must never appear in logs.
- Use stable field names such as `article_id`, `job_id`, `url`, `provider`, `duration`, `arc_code`, `fallback_reason`, and `artifact_result`.

## Errors

- Provider adapters use idiomatic Go error flow: successful calls return `(output, nil)`; failures return a zero output plus an `error`.
- ARC failures wrap sentinels from `src/worker/internal/arc` when the failure maps to persisted public article/job failure.
- Adapter contracts must not carry ARC codes in result DTO fields.
- Use typed diagnostic errors when orchestration needs provider, status code, request id, fallback reason, or operation metadata.
- Typed diagnostic errors should implement `Unwrap() error` when callers need `errors.Is`, `errors.AsType`, or ARC classification.
- All error-building infrastructure for a package belongs in `<package>/errors.go`.
- Do not expose package-local aliases for ARC sentinels unless the alias adds package-specific behavior.
- Persistence must use `arc.PublicMessage(err)` or equivalent rendering from `docs/ERRORS.md`; it must not persist raw diagnostic `err.Error()` strings.

## Filesystem

- Use Go traversal-resistant APIs such as `os.Root` or `os.OpenInRoot` when operating under `DATA_DIR` and functionally correct.
- Article artifact paths and atomic write behavior must follow `docs/ARTIFACTS.md`.

## Validation

Run Worker validation from `src/worker/` unless the task or ExecPlan specifies a narrower or broader command set:

```bash
go tool lefthook run build
go tool lefthook run format
go tool lefthook run lint
go tool lefthook run test
```

Before marking a task done, record validation in the task according to repository bookkeeping rules. If validation cannot run, record the exact reason in the task.

## Output

When reporting Worker work, include:

- task ID when applicable;
- loaded context summary;
- changed Worker areas;
- canonical docs updated or why none were needed;
- validation commands and results;
- blockers or follow-ups.
