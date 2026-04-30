# SPEC.md — Personal Article Archivist (v0)

## 1. Goal

A single-user system to archive web articles via Telegram.  
The system ingests URLs, extracts readable content, stores it as Markdown, generates a structured summary using an LLM, and exposes a minimal UI for review.

The primary objective is reliability and simplicity, not feature completeness.

---

## 2. Scope

### In Scope (v0)

- Telegram bot ingestion (single authorized user)
- URL → article ingestion pipeline
- HTML snapshot storage
- Readability-based extraction with fallback strategy
- Markdown generation
- LLM-based structured summarization
- SQLite-backed metadata and job queue
- Filesystem-backed artifact storage
- Minimal web UI (list + detail view)
- Cookie-based authentication (single user)
- Retry mechanism for failed jobs
- Basic admin actions (retry, delete)

### Out of Scope (v0)

- Playwright / headless browser rendering
- Full-text search or filtering
- Multi-user support
- Browser extensions
- Advanced tagging systems
- PWA/offline support
- Observability stack (beyond logs)

---

## 3. System Architecture

Telegram Bot
    ↓
ASP.NET Core Minimal API ("Gateway")
    ↓
SQLite (articles + jobs)
    ↓
Go Worker (single instance)
    - fetch URL
    - snapshot HTML
    - extract content
    - generate Markdown
    - score candidates
    - summarize via LLM
    ↓
SQLite + filesystem (/data)
    ↓
ASP.NET Core API
    ↓
Preact/Vite UI

---

## 4. Data Storage

### Directory Layout

/data/
  archive.db
  articles/
    {article_id}/
      snapshot.html
      content.md
      summary.json
      metadata.json

### Principles

- SQLite is the source of truth for state
- Filesystem stores large artifacts
- Writes are atomic (temp file → rename)
- Optional hashing for integrity/debugging

---

## 5. Data Model (High-Level)

### Article

id
original_url
canonical_url
title
domain
status (pending | processing | ready | failed)
selected_extractor
extractor_score
paths (snapshot, markdown, summary)
error_message
created_at
processed_at

### Job

id
article_id
type
status (queued | running | succeeded | failed | retrying | dead)
attempts
run_after
locked_at
locked_by
error_message
created_at

---

## 6. Ingestion Pipeline

### Steps

1. Receive URL from Telegram
2. Create article record (status: pending)
3. Enqueue job
4. Worker dequeues job
5. Fetch HTML (plain HTTP)
6. Attempt extraction via:
   - Jina Reader
   - readability/go-readability
7. Convert to Markdown
8. Score candidates
9. Select best candidate (threshold ≥ 0.6)
10. Store:
    - HTML snapshot
    - Markdown
11. Call LLM summarizer
12. Store summary
13. Mark article as ready

### Failure Handling

- Retry up to 3 times with backoff
- On repeated failure → mark as failed
- Errors persisted and visible in UI

---

## 7. Extraction Strategy

### Candidates

- Jina-based extraction
- Readability-based extraction

### Scoring Criteria (example)

- Title present
- Content length within bounds
- Paragraph count sufficient
- Low link density
- Low boilerplate ratio
- Sentence density
- Not an error page
- Canonical URL detected
- Language detected
- Heading structure present

### Selection

score = passed / total  
threshold = 0.6  
select highest scoring candidate  
fail if all below threshold  

---

## 8. LLM Summarization

### Input

- Extracted Markdown

### Output (strict JSON)

{
  "summary": "...",
  "key_points": ["..."],
  "tags": ["..."],
  "template_version": 1
}

### Constraints

- Provider-agnostic interface
- Schema validation required
- Retry on malformed output

---

## 9. Authentication

- Cookie-based authentication
- Single user
- Secret-based (long random key)
- Entire UI and API protected

---

## 10. UI

### Layout

- Left: article list (no search, no filtering)
- Right: article detail

### Article Detail

- Title
- Domain
- Summary
- Key points
- Tags
- Markdown content
- Link to original article
- Error (if failed)

### Actions

- Retry processing
- Delete article

---

## 11. Queue Design

- SQLite-backed job queue
- Single worker
- Atomic dequeue using UPDATE … RETURNING
- No external queue system

---

## 12. Configuration

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

---

## 13. Logging

- Stdout logging only
- Structured logs preferred
- Include:
  - article_id
  - job_id
  - url
  - status
  - duration
  - error

---

## 14. Deployment

- Single VPS
- All components deployed together
- Shared /data volume
- Backup via filesystem snapshot

---

## 15. Design Principles

- Favor simplicity over extensibility
- Deterministic processing where possible
- Clear failure states
- Minimal dependencies
- Avoid premature optimization
- Store raw + derived data

---

## 16. Non-Goals

- High scalability
- Multi-tenancy
- Real-time processing guarantees
- Perfect extraction accuracy

---

## 17. Future Extensions (Not in v0)

- Playwright fallback for JS-heavy sites
- Search / filtering
- Vector embeddings
- Tag management UI
- Browser extension
- Multi-user support
- Observability stack (OTel + collector)
