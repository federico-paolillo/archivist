# Personal Article Archivist Concept

This file is a non-canonical product seed for later feature planning. Rebuild-critical architecture, constraints, and durable decisions live in `docs/ARCHITECTURE.md`, `docs/DESIGN.md`, and `docs/CONVENTIONS.md`.

## Goal

Archivist is a single-user system for saving web articles through Telegram.

The system ingests URLs, extracts readable content, stores article artifacts, generates structured LLM summaries, and exposes a minimal private UI for review.

The primary objective for v0 is reliability and simplicity, not feature completeness.

## v0 Capability Seeds

- Telegram bot ingestion for one authorized user.
- URL-to-article processing pipeline.
- HTML snapshot storage.
- Article content extraction with fallback handling.
- Markdown generation.
- Structured LLM summarization.
- Persistent article metadata, job state, and artifact storage.
- Minimal web UI with article list and detail view.
- Private authentication for the UI and API.
- Retry mechanism for failed jobs.
- Basic admin actions: retry and delete.

## v0 Non-Goals

- High scalability.
- Multi-tenancy.
- Real-time processing guarantees.
- Perfect extraction accuracy.
- Playwright or headless browser rendering.
- Full-text search or filtering.
- Browser extensions.
- Advanced tagging systems.
- PWA or offline support.
- Dedicated observability stack beyond logs.

## Future Feature Seeds

- Quiz generation from archived articles.
- Playwright fallback for JavaScript-heavy sites.
- Search and filtering.
- Vector embeddings.
- Tag management UI.
- Browser extension.
- Multi-user support.
- Observability stack with OpenTelemetry and collector support.
