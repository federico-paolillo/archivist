# ERRORS.md

Defines shared error-code conventions for persisted user-visible failures.

Error codes are stable contracts between Worker, Gateway, UI, and documentation. They are not log-only strings. If a persisted article or job error is intended for a user-facing surface, it must use a known code from this file.

---

## ARC Article Processing Errors

Article-processing errors use the `ARC-NNN` namespace.

Persisted public article errors must start with the code in square brackets, followed by a short user-safe message:

```text
[ARC-003] The URL was not found.
```

Detailed HTTP, filesystem, library, stack, or provider diagnostics belong in structured logs or job diagnostic context, not in `articles.error_message`.

## Initial Catalog

| Code | Meaning | Public Message Guidance |
|---|---|---|
| `ARC-001` | URL resolution failed | The URL could not be resolved. |
| `ARC-002` | URL access denied | The URL requires access Archivist does not have. |
| `ARC-003` | URL not found | The URL was not found. |
| `ARC-004` | URL fetch transient failure | The URL could not be fetched right now. |
| `ARC-005` | Response is not HTML | The URL did not return an HTML article page. |
| `ARC-006` | Response exceeds snapshot size limit | The HTML response is too large to archive. |
| `ARC-007` | Snapshot write failed | Archivist could not store the HTML snapshot. |
| `ARC-008` | go-readability found document unreadable | Archivist could not read this page locally. |
| `ARC-009` | go-readability extraction or Markdown conversion failed | Archivist could not extract this page locally. |
| `ARC-010` | Jina Reader fallback failed | Archivist could not extract this page with the fallback reader. |
| `ARC-011` | Jina Reader insufficient balance | Archivist could not use the fallback reader because the Jina account is out of credit. |
| `ARC-012` | Markdown artifact write failed | Archivist could not store the Markdown article. |
| `ARC-999` | Unknown processing failure | Archivist could not process the URL. |

## Rules

- Do not reuse a code for a different meaning.
- Do not delete a code once implementation or persisted data may reference it.
- Add new codes in canonical docs before implementation uses them.
- Keep public messages short and user-safe.
- Keep operational details in logs or job diagnostic context.
