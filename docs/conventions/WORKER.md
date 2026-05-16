# WORKER.md

Describes conventions and best-practices when working on the Worker module.

## In general

Worker code lives under `src/worker/` and targets Go 1.26

- Make lightweight interfaces to aid testing. Do not try to anticipate greater abstractions unless necessary.
- Do follow SOLID principles and GRASP principles but do not forget about KISS and YAGNI.
- Do not take shortcuts or make stub implementations. If you find something difficult to implement challenge the design.
- `CGO_ENABLED=0`. The resulting binary must be a single-executable file.
- We target Linux x64 first. If it's not too much of an hassle we support MacOS with Apple Silicon.
- We follow the [go-standards/project-layout](https://github.com/golang-standards/project-layout)
- Run worker verification from `src/worker/`:

```bash
go tool lefthook run build
go tool lefthook run format
go tool lefthook run lint
go tool lefthook run test
```

## Working with directories

Go 1.24 introduced new APIs in the os package to safely open a file in a location in a traversal-resistant fashion.

```golang
root, err := os.OpenRoot("/some/root/directory")
if err != nil {
  return err
}
defer root.Close()
```

`Root` provides methods to operate on files within the root. These methods all accept filenames relative to the root, and disallow any operations that would escape from the root either using relative path components ("..") or symlinks.

`f, err := root.Open("path/to/file")`

`Root` permits relative path components and symlinks that do not escape the root. For example, `root.Open("a/../b")` is permitted. Filenames are resolved using the semantics of the local platform: On Unix systems, this will follow any symlink in “a” (so long as that link does not escape the root); while on Windows systems this will open “b” (even if “a” does not exist).

In addition to the `Root` type, the new `os.OpenInRoot` function provides a simple way to open a potentially-untrusted filename within a specific directory:

`f, err := os.OpenInRoot("/some/root/directory", untrustedFilename)`

The `Root` type provides a simple, safe, portable API for operating with untrusted filenames.

When possible and functionally correct, use `Root` or `OpenInRoot` to handle filesystem interactions scoped to `DATA_DIR`, especially `/data/articles/{article_id}/` artifact paths.

## Composition root

See the **Composition Root and Poor Man's DI** section above for the full rationale and rules.

Anytime you add a new service to the composition root you must test that it is created correctly (or not, depending on the logic) in `app_test.go`.

`NewApp()` is the constructor of the composition root and can be optionally partitioned into multiple `createXxx()` factory functions that take care of complex initialization logic. These functions are not necessary if constructing a service is trivial.

## CLI commands

Worker CLI command code lives under `src/worker/internal/app/`.

`program.go` owns all `urfave/cli/v3` command configuration: command names, usage text, flags, default flag values, and extraction of typed flag values from `*cli.Command`.

Each CLI command action must forward from `program.go` to a function in its own file named after the command. For example, the `process` command forwards to `process(...)` in `process.go`, and the `version` command forwards to `version(...)` in `version.go`.

Command action functions must not accept or reference `urfave/cli/v3` types. They receive only `context.Context`, the Worker `App`, and already-parsed primitive or domain values needed to perform the command.

Any task that changes Worker production behavior through a CLI command must include executable-surface tests that invoke the command registration path. Tests that instantiate internal services directly are not sufficient for executable-boundary behavior.

## Configuration

Worker configuration is loaded from environment variables or equivalent deployment secret mechanisms. Worker settings include:

```text
DATA_DIR
SQLITE_PATH
LLM_PROVIDER
LLM_API_KEY
LLM_MODEL
JINA_ENABLED
JINA_API_KEY
```

`JINA_API_KEY` is optional unless a task or deployment requires authenticated Jina Reader requests. It is secret material and must not be logged.

Whenever you introduce a worker configuration key, document it in `docs/conventions/GENERAL.md`, `docs/ARCHITECTURE.md`, and any affected feature spec or task. Add sensible defaults only for non-secret values.

Always extend worker configuration tests to assert default values, required values, and environment variable loading.

## Composition Root and Poor Man's DI

The Worker uses *Pure DI* (also called *Poor Man's DI*), as described by Mark Seemann and Steven van Deursen in *Dependency Injection Principles, Practices, and Patterns*. This means: explicit constructor injection, no IoC container, no service locator, no globals.

### Composition Root

A **Composition Root** is the single location where the entire object graph is wired. It is the only place that knows how concrete types map to each other. For the Worker this is `pkg/app/NewApp`, called from `cmd/worker/main.go`. Nothing else in the codebase assembles services; adapters and services only receive collaborators through constructor parameters.

### Poor Man's DI (Pure DI)

Each `NewXxx(dep1, dep2, …)` constructor takes its collaborators as explicit arguments. `NewApp` calls these constructors in dependency order, passing already-constructed values down the call chain. Optional `createXxx` factory functions partition complex sub-graphs but remain internal to the composition root.

Rules:

- Every collaborator a service needs must be a constructor parameter — never read from a global, never resolved at call time.
- Long-lived singletons (DB, HTTP client, loggers, service adapters) live as fields of `App`. Per-request objects are created inside the call that needs them.
- The composition root (`pkg/app`) may import any internal package. No other package may import a sibling package solely to wire it.
- Every field added to `App` must be covered by a test in `app_test.go` (this restates the existing rule from the section below; it is listed here because it is a consequence of Pure DI: the root is the only wiring point and its tests are the only integration tests for wiring).

## Provider Boundaries

Worker pipeline orchestration must depend on Archivist-owned interfaces for external or replaceable processing providers.

Markdown extraction uses a Worker-owned `MarkdownExtractor` abstraction. The local go-readability implementation and the Jina Reader implementation must sit behind that abstraction. Pipeline code must not import or expose Jina SDK/client types. Use an official Jina-provided SDK when a suitable Reader API SDK exists for the implementation language. If no suitable official SDK exists, a small internal Reader adapter is acceptable; untagged or low-adoption third-party wrappers are not.

Summary generation uses a Worker-owned `SummarizerService` abstraction. Claude/Anthropic is the first provider implementation and must use official Anthropic SDKs when suitable SDKs exist. In Go, prefer `github.com/anthropics/anthropic-sdk-go`. Pipeline code must not import or expose Anthropic SDK request or response types.

Snapshot and Markdown stages are intermediate pipeline stages once summary generation is implemented. Final Worker success for v0 means `summary.md` has been atomically promoted and the article/job/notification terminal state has been committed.

## HTTP client

All outbound HTTP calls from Worker code must go through `github.com/imroc/req/v3`.

- Direct use of `net/http` for *outbound* requests is forbidden. `net/http` remains acceptable inside test files (`httptest.NewServer`, `http.HandlerFunc`, status code constants) and when an external SDK requires a plain `*http.Client` parameter — in that case call `reqClient.GetClient()` to obtain the underlying standard-library client.
- Never use the package-level global client (`req.C()`). Always construct an owned `*req.Client` with `req.NewClient()` inside the composition root and inject it into every adapter that needs it.
- The composition root owns and configures shared concerns — user-agent, timeout, and any future middleware. Adapters receive a pre-configured client; they must not mutate it or set global options.
- Tests construct their own isolated `*req.Client` instances (typically `req.NewClient()` pointing at an `httptest.Server`). Tests must not share or mutate a client owned by another test.

## Structured Logging

v0 logs to stdout using Go's `log/slog` package. Structured logs are required for all observable pipeline events.

### Logging ownership

**Pipeline orchestration owns all structured log entries** for article-processing stages (ARTPROC-005, MDEXT-005, SUMGEN-004). Provider adapters (`go_readability.go`, `jina.go`, `anthropic.go`, and any future adapter) must not accept a `*slog.Logger` and must not emit `slog.Info` or `slog.Error` calls. Adapters return data; orchestration logs decisions.

This rule keeps adapter constructors small, adapter tests free of logger mocks, and the log field schema in one place (the orchestration layer).

**Carve-out**: an adapter may emit `slog.Debug` for low-value diagnostics that orchestration cannot observe (e.g. internal retry counts). This is the only permitted adapter-level logging. API keys must never appear in any log entry, at any level.

### Adapter error-flow contract

Provider adapters must use idiomatic Go error flow. Successful calls return `(output, nil)`. Failed calls return a zero output plus an `error` that wraps an `internal/arc` sentinel when the failure maps to a persisted ARC failure.

Adapter contracts must not carry ARC codes in result DTO fields. Use typed diagnostic errors such as `markdown.ExtractionError` or `summary.ProviderError` when orchestration needs provider, status code, request id, or fallback-reason metadata. These typed errors must implement `Unwrap() error` so `errors.Is(err, arc.Err...)` and `arc.CodeOf(err)` work through wrapped errors.

Provider interfaces must expose a `Provider() Provider` method so orchestration can log the attempted provider without relying on failure DTOs.

### Orchestration log field set

Orchestration must log the following fields when available. Cross-module field names follow `docs/conventions/GENERAL.md`:

- `article_id`, `job_id`, `url` — article context, threaded via request types
- `provider` — which adapter was called
- `duration` — measured by orchestration around the adapter call with `time.Since`
- `arc_code` — ARC error code on failure
- `fallback_reason` — when a secondary provider is invoked
- `artifact_result` — success or failure of the artifact write
- model, provider request id — for LLM providers when available

Markdown extraction orchestration must additionally log: provider attempt, fallback reason, selected provider.

Logs must never include API keys or full article/summary content.

## Error helpers

All error-building infrastructure for a package must live in `<package>/errors.go`. This includes:

- package-specific diagnostic error type definitions
- Error type definitions
- Error constructor and classification helpers (e.g. `jinaFailure`, `classifyError`, `isBillingError`)

This keeps primary implementation files (`jina.go`, `anthropic.go`, `go_readability.go`, etc.) focused on adapter behavior rather than error plumbing.

When a package is modified, migrate any existing error helpers from implementation files into `errors.go` as part of that change.

ARC code constants, public messages, and ARC sentinel errors belong only in `src/worker/internal/arc`. Other packages must not expose package-local aliases for ARC sentinels unless the alias adds package-specific behavior. Prefer direct `arc.Err*` use plus typed diagnostic errors that preserve operation metadata and unwrap to the ARC sentinel.

## Error wrapping

Persisted user-facing article-processing failures must use ARC codes from `docs/conventions/ERRORS.md`. Store a short public message on `articles.error_message`, and keep detailed HTTP, filesystem, library, or stack diagnostics in logs or job diagnostic context.

Worker code implements this through `src/worker/internal/arc`. ARC failures may be wrapped with `%w` or typed diagnostic errors for logs and classification. Persistence must call `arc.PublicMessage(err)` or otherwise render from the ARC code; it must not persist `err.Error()` from a wrapped diagnostic error. Unknown non-ARC processing failures must be mapped to `arc.ErrUnknown` before terminal persistence.

When adding additional context to an error, either with fmt.Errorf or by implementing a custom type, you need to decide whether the new error should wrap the original. There is no single answer to this question; it depends on the context in which the new error is created. Wrap an error to expose it to callers. Do not wrap an error when doing so would expose implementation details.

As one example, imagine a Parse function which reads a complex data structure from an io.Reader. If an error occurs, we wish to report the line and column number at which it occurred. If the error occurs while reading from the io.Reader, we will want to wrap that error to allow inspection of the underlying problem. Since the caller provided the io.Reader to the function, it makes sense to expose the error produced by it.

In contrast, a function which makes several calls to a database probably should not return an error which unwraps to the result of one of those calls. If the database used by the function is an implementation detail, then exposing these errors is a violation of abstraction. For example, if the LookupUser function of your package pkg uses Go’s database/sql package, then it may encounter a sql.ErrNoRows error. If you return that error with fmt.Errorf("accessing DB: %v", err) then a caller cannot look inside to find the sql.ErrNoRows. But if the function instead returns fmt.Errorf("accessing DB: %w", err), then a caller could reasonably write

```golang
err := pkg.LookupUser(...)
if errors.Is(err, sql.ErrNoRows) …
```

At that point, the function must always return sql.ErrNoRows if you dont want to break your clients, even if you switch to a different database package. In other words, wrapping an error makes that error part of your API. If you don't want to commit to supporting that error as part of your API in the future, you shouldn't wrap the error.

It's important to remember that whether you wrap or not, the error text will be the same. A person trying to understand the error will have the same information either way; the choice to wrap is about whether to give programs additional information so they can make more informed decisions, or to withhold that information to preserve an abstraction layer.

## Error `As`

For most uses, prefer `AsType`. `As` is equivalent to AsType but sets its target argument rather than returning the matching error and doesn't require its target argument to implement error.

Go 1.26 introduces new `AsType` function is a generic version of As. It is type-safe, faster, and, in most cases, easier to use.

## Modern Go guidelines

> Follow these release notes excerpts when writing code to make sure it is modern

### Go 1.0+

- `time.Since`: `time.Since(start)` instead of `time.Now().Sub(start)`

### Go 1.8+

- `time.Until`: `time.Until(deadline)` instead of `deadline.Sub(time.Now())`

### Go 1.13+

- `errors.Is`: `errors.Is(err, target)` instead of `err == target` (works with wrapped errors)

### Go 1.18+

- `any`: Use `any` instead of `interface{}`
- `bytes.Cut`: `before, after, found := bytes.Cut(b, sep)` instead of Index+slice
- `strings.Cut`: `before, after, found := strings.Cut(s, sep)`

### Go 1.19+

- `fmt.Appendf`: `buf = fmt.Appendf(buf, "x=%d", x)` instead of `[]byte(fmt.Sprintf(...))`
- `atomic.Bool`/`atomic.Int64`/`atomic.Pointer[T]`: Type-safe atomics instead of `atomic.StoreInt32`

```go
var flag atomic.Bool
flag.Store(true)
if flag.Load() { ... }

var ptr atomic.Pointer[Config]
ptr.Store(cfg)
```

### Go 1.20+

- `strings.Clone`: `strings.Clone(s)` to copy string without sharing memory
- `bytes.Clone`: `bytes.Clone(b)` to copy byte slice
- `strings.CutPrefix/CutSuffix`: `if rest, ok := strings.CutPrefix(s, "pre:"); ok { ... }`
- `errors.Join`: `errors.Join(err1, err2)` to combine multiple errors
- `context.WithCancelCause`: `ctx, cancel := context.WithCancelCause(parent)` then `cancel(err)`
- `context.Cause`: `context.Cause(ctx)` to get the error that caused cancellation

### Go 1.21+

**Built-ins:**
- `min`/`max`: `max(a, b)` instead of if/else comparisons
- `clear`: `clear(m)` to delete all map entries, `clear(s)` to zero slice elements

**slices package:**
- `slices.Contains`: `slices.Contains(items, x)` instead of manual loops
- `slices.Index`: `slices.Index(items, x)` returns index (-1 if not found)
- `slices.IndexFunc`: `slices.IndexFunc(items, func(item T) bool { return item.ID == id })`
- `slices.SortFunc`: `slices.SortFunc(items, func(a, b T) int { return cmp.Compare(a.X, b.X) })`
- `slices.Sort`: `slices.Sort(items)` for ordered types
- `slices.Max`/`slices.Min`: `slices.Max(items)` instead of manual loop
- `slices.Reverse`: `slices.Reverse(items)` instead of manual swap loop
- `slices.Compact`: `slices.Compact(items)` removes consecutive duplicates in-place
- `slices.Clip`: `slices.Clip(s)` removes unused capacity
- `slices.Clone`: `slices.Clone(s)` creates a copy

**maps package:**
- `maps.Clone`: `maps.Clone(m)` instead of manual map iteration
- `maps.Copy`: `maps.Copy(dst, src)` copies entries from src to dst
- `maps.DeleteFunc`: `maps.DeleteFunc(m, func(k K, v V) bool { return condition })`

**sync package:**
- `sync.OnceFunc`: `f := sync.OnceFunc(func() { ... })` instead of `sync.Once` + wrapper
- `sync.OnceValue`: `getter := sync.OnceValue(func() T { return computeValue() })`

**context package:**
- `context.AfterFunc`: `stop := context.AfterFunc(ctx, cleanup)` runs cleanup on cancellation
- `context.WithTimeoutCause`: `ctx, cancel := context.WithTimeoutCause(parent, d, err)`
- `context.WithDeadlineCause`: Similar with deadline instead of duration

### Go 1.22+

**Loops:**
- `for i := range n`: `for i := range len(items)` instead of `for i := 0; i < len(items); i++`
- Loop variables are now safe to capture in goroutines (each iteration has its own copy)

**cmp package:**
- `cmp.Or`: `cmp.Or(flag, env, config, "default")` returns first non-zero value

```go
// Instead of:
name := os.Getenv("NAME")
if name == "" {
    name = "default"
}
// Use:
name := cmp.Or(os.Getenv("NAME"), "default")
```

**reflect package:**
- `reflect.TypeFor`: `reflect.TypeFor[T]()` instead of `reflect.TypeOf((*T)(nil)).Elem()`

**net/http:**
- Enhanced `http.ServeMux` patterns: `mux.HandleFunc("GET /api/{id}", handler)` with method and path params
- `r.PathValue("id")` to get path parameters

### Go 1.23+

- `maps.Keys(m)` / `maps.Values(m)` return iterators
- `slices.Collect(iter)` not manual loop to build slice from iterator
- `slices.Sorted(iter)` to collect and sort in one step

```go
keys := slices.Collect(maps.Keys(m))       // not: for k := range m { keys = append(keys, k) }
sortedKeys := slices.Sorted(maps.Keys(m))  // collect + sort
for k := range maps.Keys(m) { process(k) } // iterate directly
```

**time package**

- `time.Tick`: Use `time.Tick` freely — as of Go 1.23, the garbage collector can recover unreferenced tickers, even if they haven't been stopped. The Stop method is no longer necessary to help the garbage collector. There is no longer any reason to prefer NewTicker when Tick will do.

### Go 1.24+

- `t.Context()` not `context.WithCancel(context.Background())` in tests.
  ALWAYS use t.Context() when a test function needs a context.

Before:
```go
func TestFoo(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    result := doSomething(ctx)
}
```
After:
```go
func TestFoo(t *testing.T) {
    ctx := t.Context()
    result := doSomething(ctx)
}
```

- `omitzero` not `omitempty` in JSON struct tags.
  ALWAYS use omitzero for time.Duration, time.Time, structs, slices, maps.

Before:
```go
type Config struct {
    Timeout time.Duration `json:"timeout,omitempty"` // doesn't work for Duration!
}
```
After:
```go
type Config struct {
    Timeout time.Duration `json:"timeout,omitzero"`
}
```

- `b.Loop()` not `for i := 0; i < b.N; i++` in benchmarks.
  ALWAYS use b.Loop() for the main loop in benchmark functions.

Before:
```go
func BenchmarkFoo(b *testing.B) {
    for i := 0; i < b.N; i++ {
        doWork()
    }
}
```
After:
```go
func BenchmarkFoo(b *testing.B) {
    for b.Loop() {
        doWork()
    }
}
```

- `strings.SplitSeq` not `strings.Split` when iterating.
  ALWAYS use SplitSeq/FieldsSeq when iterating over split results in a for-range loop.

Before:
```go
for _, part := range strings.Split(s, ",") {
    process(part)
}
```
After:
```go
for part := range strings.SplitSeq(s, ",") {
    process(part)
}
```
Also: `strings.FieldsSeq`, `bytes.SplitSeq`, `bytes.FieldsSeq`.

### Go 1.25+

- `wg.Go(fn)` not `wg.Add(1)` + `go func() { defer wg.Done(); ... }()`.
  ALWAYS use wg.Go() when spawning goroutines with sync.WaitGroup.

Before:
```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func() {
        defer wg.Done()
        process(item)
    }()
}
wg.Wait()
```
After:
```go
var wg sync.WaitGroup
for _, item := range items {
    wg.Go(func() {
        process(item)
    })
}
wg.Wait()
```

### Go 1.26+

- `new(val)` not `x := val; &x` — returns pointer to any value.
  Go 1.26 extends new() to accept expressions, not just types.
  Type is inferred: new(0) → *int, new("s") → *string, new(T{}) → *T.
  DO NOT use `x := val; &x` pattern — always use new(val) directly.
  DO NOT use redundant casts like new(int(0)) — just write new(0).
  Common use case: struct fields with pointer types.

Before:
```go
timeout := 30
debug := true
cfg := Config{
    Timeout: &timeout,
    Debug:   &debug,
}
```
After:
```go
cfg := Config{
    Timeout: new(30),   // *int
    Debug:   new(true), // *bool
}
```

- `errors.AsType[T](err)` not `errors.As(err, &target)`.
  ALWAYS use errors.AsType when checking if error matches a specific type.

Before:
```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    handle(pathErr)
}
```
After:
```go
if pathErr, ok := errors.AsType[*os.PathError](err); ok {
    handle(pathErr)
}
```
