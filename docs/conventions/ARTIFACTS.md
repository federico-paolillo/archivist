# ARTIFACTS.md

Defines deterministic filesystem artifact conventions for article processing.

Artifact paths are stable contracts between Worker, Gateway, UI, backups, and rebuild documentation. SQLite stores article and job state; it does not store artifact path columns in v0.

---

## Article Artifact Root

Article artifacts live under:

```text
{DATA_DIR}/articles/{article_id}/
```

`article_id` is the SQLite article ULID. Implementations must treat it as an identifier, not as an arbitrary path segment.

Filesystem access scoped to `DATA_DIR` should use traversal-resistant APIs such as Go `os.Root` or `os.OpenInRoot` where functionally correct.

## Article Artifacts

| Artifact | Path | Producer | Required For |
|---|---|---|---|
| HTML snapshot | `{DATA_DIR}/articles/{article_id}/snapshot.html` | Worker HTML snapshot stage | Markdown extraction and future reprocessing |
| Markdown content | `{DATA_DIR}/articles/{article_id}/content.md` | Worker Markdown extraction stage | Summary generation input |
| Summary Markdown | `{DATA_DIR}/articles/{article_id}/summary.md` | Worker summary generation stage | Final v0 article success, Telegram completion replies, and UI summary display |
| Summary JSON | `{DATA_DIR}/articles/{article_id}/summary.json` | Future structured summary stage | Future structured summary display |
| Metadata | `{DATA_DIR}/articles/{article_id}/metadata.json` | Future metadata stage | Future diagnostics or enrichment |

## Access Interface

The artifact access layer is operation-first. It exposes per-artifact `Open<Artifact>` / `Write<Artifact>` operations (e.g. `OpenSnapshot`, `WriteSnapshot`). Callers identify an artifact by article ID and artifact kind; they do not receive or pass filesystem paths.

- `Open<Artifact>` returns a streaming reader (e.g. `io.ReadCloser` in Go). The caller is responsible for closing it. A non-existent artifact must surface as a not-found error (e.g. `fs.ErrNotExist`), not as an empty result.
- `Write<Artifact>` consumes a streaming source (e.g. `io.Reader` in Go) and persists it to the deterministic artifact path. Atomicity is an implementation detail of the access layer; callers do not coordinate temp files or renames.
- Filesystem paths and the rooted FS handle are private to the access layer. No public API returns absolute or relative artifact paths.

## Write Rules

- Artifact writes must be atomic: write to a temporary file in the article artifact directory, then rename into place.
- A failed write must not promote a partial final artifact.
- Temporary files should be named so cleanup can identify them as internal artifacts.
- Do not create placeholder artifacts for future stages.
- The artifact access layer must not log write results itself. It must surface sufficient information (success, failure, bytes written or equivalent) in its return value so the calling pipeline stage can emit a structured log entry for the artifact write result.
- `summary.md` is the v0 summary artifact. Do not write `summary.json` unless a future canonical spec reintroduces structured summary output.
- Do not add artifact path columns to SQLite unless a future canonical spec changes this convention.
- Delete or cleanup behavior that removes article state must remove the article artifact directory through the same deterministic root.

## Rebuild Notes

- Rebuilds must derive artifact paths from `DATA_DIR`, `articles/`, `article_id`, and the stable artifact filenames above.
- Implementations must not infer required behavior from files that happen to exist under an artifact directory unless a canonical spec declares that artifact as produced by an implemented feature.
