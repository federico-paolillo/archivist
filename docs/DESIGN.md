# DESIGN.md

Records durable decisions that must survive rebuilds. This can be ADR-like but does not need heavy ceremony.

A decision discovered during implementation should be promoted here if it affects more than one task or must remain true across rebuilds.

---

## Decision Record Format

Each decision can be as lightweight as a heading + rationale paragraph, or as structured as a full ADR. Use the level of ceremony that matches the stakes.

Suggested minimal format:

```
### DSGN-NNN: <Title>

**Date:** YYYY-MM-DD
**Status:** accepted | superseded | under review

**Context:** Why this decision was needed.

**Decision:** What was decided.

**Consequences:** What changes as a result. What becomes easier or harder.
```

---

## Decisions