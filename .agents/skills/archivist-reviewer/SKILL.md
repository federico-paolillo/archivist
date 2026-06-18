---
name: archivist-reviewer
description: Use when reviewing Archivist implementation branches for bugs, regressions, ALM compliance, canonical contract adherence, security, maintainability, test coverage, and unresolved integration risk.
---

# Archivist Reviewer

Use this skill for read-only review passes.

## Review Stance

Findings come first. Prioritize correctness, data integrity, security, canonical contract violations, behavioral regressions, missing validation, missing tests, and unresolved integration risk.

## Scope

Review only the assigned branch, task, diff, and worker report. Mention unrelated pre-existing issues only when they block the feature or create concrete integration risk.

Review against the relevant implementation skill:

- Gateway: `.agents/skills/archivist-gateway/SKILL.md`
- Worker: `.agents/skills/archivist-worker/SKILL.md`
- UI: `.agents/skills/archivist-ui/SKILL.md`
- Cross-module or documentation: `.agents/skills/archivist-general/SKILL.md`

## Required Checks

- Behavior is grounded in canonical ALM files, not just code or skills.
- Public APIs, auth behavior, persistence schema, artifact layout, and public ARC errors remain compatible unless the task explicitly changes them.
- Tests cover meaningful success and failure paths for new behavior.
- New abstractions are justified by external boundaries, meaningful tests, or real reuse.
- The branch avoids unrelated refactors and formatting churn.
- Canonical docs are updated only when durable behavior changed. Task completion state, validation evidence, diary entries, and coordinator notes remain non-canonical coordination state and must not be treated as rebuild contracts.
- Validation commands match the changed surface.

## Finding Format

```text
Finding: <severity> <title>
File: <path>:<line>
Issue: <specific problem>
Required fix: <concrete correction>
```

Severity:

- P0: breaks core behavior, data integrity, security, or deployment.
- P1: likely user-visible regression, missing required validation, or contract violation.
- P2: maintainability, idiomacy, accessibility, or moderate correctness risk.
- P3: minor cleanup only when worth fixing before merge.

## Verification

Run narrow relevant checks when feasible. Report commands that were not run and why.

## Output

End with approval status: `approved`, `approved with nits`, or `changes requested`.
