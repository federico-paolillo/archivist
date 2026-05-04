# GENERAL.md

Defines cross-module coding, testing, naming, configuration, error handling, logging, security, and repository conventions.

Agents must treat conventions as binding unless a task explicitly changes them.

---

## Language and Runtime

The accepted v0 architecture uses:

- ASP.NET Core Minimal API for the gateway/API process;
- Go for the background worker;
- Preact with Vite for the web UI;
- SQLite for metadata and queue state.

Exact language versions, package managers, linters, and formatters are specified in the relevant module convention files. Feature planning or implementation tasks that introduce project structure must update the appropriate convention file with concrete toolchain commands before marking the task done.

## Project Layout

Production source layout is module-specific and documented in the relevant file under `docs/conventions/`.

Canonical rebuild artifacts remain under `docs/` and `docs/specs/` according to `docs/REBUILD.md`. Feature planning tasks must create feature folders under `docs/specs/<feature-slug>/` and update `docs/specs/INDEX.md`.

## Naming

Language-specific naming conventions are defined in module convention files when needed.

Configuration keys use uppercase snake case, for example `DATA_DIR` and `TELEGRAM_BOT_TOKEN`.

## Error Handling

Processing failures must be persisted in SQLite with enough context for the UI to show a failure state, for Telegram completion replies to report the error, and for operators to diagnose the failed article/job.

User-facing persisted article-processing failures must use the shared ARC error-code catalog in `docs/conventions/ERRORS.md`. Public article error messages must not expose low-level HTTP, filesystem, library, stack, or provider details.

v0 does not automatically retry failed jobs or failed Telegram notifications. Manual requeue is performed by sending the URL again unless a future feature changes this convention.

## Logging and Observability

v0 logs to stdout. Structured logs are preferred.

Logs for article processing should include these fields when available:

- `article_id`
- `job_id`
- `url`
- `status`
- `duration`
- `error`

A dedicated observability stack is out of scope for v0.

## Testing

Module-specific test framework, naming, and command conventions are defined in the relevant module convention files.

Before any task is marked done, run the validation required by that task or its ExecPlan. If validation cannot be run, record the reason in the task and the relevant feature diary.

## Configuration

Runtime configuration is supplied through environment variables or equivalent deployment secret mechanisms.

Required v0 configuration keys are:

```text
DATA_DIR
SQLITE_PATH
TELEGRAM_BOT_TOKEN
TELEGRAM_ALLOWED_USER_ID
TELEGRAM_WEBHOOK_SECRET
AUTH_COOKIE_SECRET
LLM_PROVIDER
LLM_API_KEY
LLM_MODEL
JINA_ENABLED
```

Feature specs or tasks that add configuration keys must update this file or the relevant module convention file, plus any affected architecture or design decisions.

## Security

Secrets must not be committed to the repository.

The UI and UI-facing API must require cookie authentication. Telegram ingestion must validate both the webhook secret and allowed Telegram user ID.

## Dependencies

Keep dependencies minimal. Add external dependencies only when they replace non-trivial custom implementation or are required by an accepted architecture/design decision.

Feature planning or implementation tasks that add dependencies must document the reason in the relevant spec, task, or ExecPlan.

## File Writes

Artifact writes under `/data` must be atomic: write to a temporary path, then rename into place.

## Identifiers

Whenever an identifier is needed, use a ULID. Do not use GUIDs and do not delegate identifier generation to the database.
