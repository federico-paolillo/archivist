---
name: archivist-architecture-docs
description: Use when editing Archivist architecture or design documentation, including README architecture/design summaries and diagrams.
---

# Archivist Architecture Docs

Use this skill for changes to `docs/ARCHITECTURE.md`, `docs/DESIGN.md`, architecture/design summaries in `README.md`, or diagrams that describe system boundaries, durable decisions, runtime topology, storage, authentication, deployment, integrations, artifact contracts, or rebuild-relevant behavior.

## Canonical Boundary

- Repo-local skills are non-canonical development guidance.
- Rebuild-critical behavior belongs in canonical files listed by `docs/REBUILD.md`, not only in `.agents/skills/**`, README prose, comments, or diagrams.
- If a README architecture/design summary changes durable behavior, update the owning canonical document in the same change.
- Do not create ALM tasks from this skill. Use existing planning rules only when the user explicitly asks for ALM work.

## Required Context

Before changing durable architecture or design documentation, read:

1. `AGENTS.md`
2. `docs/REBUILD.md`
3. `docs/ARCHITECTURE.md`
4. `docs/DESIGN.md`
5. `docs/specs/INDEX.md`
6. Specific related feature specs, plans, tasks, or ExecPlans identified by `docs/specs/INDEX.md`, the affected canonical document, or the user request.

Do not load unrelated feature folders unless the change is explicitly cross-cutting or the affected specs point to them.

## Editing Guidance

- Keep `docs/ARCHITECTURE.md` focused on executables, service boundaries, data ownership, storage, runtime topology, integrations, authentication boundaries, deployment assumptions, and cross-feature contracts.
- Keep `docs/DESIGN.md` focused on accepted durable decisions, decision rationale, supersession history, and tradeoffs that must survive rebuilds.
- README architecture/design summaries should summarize canonical docs, not define new behavior.
- Preserve accepted decision status semantics: only accepted design decisions are binding when they conflict with superseded or under-review decisions.
- When changing public contracts, storage shape, authentication, deployment, artifacts, or observability behavior, identify the owning canonical document and affected feature specs before editing.

## Validation

For documentation-only changes, run:

```bash
git diff --check
```

Also run targeted link checks for changed Markdown references. When README Mermaid diagrams change, validate Mermaid syntax with an available Mermaid CLI or project-approved diagram check and report the exact command used. If a validation tool is unavailable, state that explicitly with the reason.

Do not run code formatters or module test suites unless source code changed or the task explicitly requires them.
