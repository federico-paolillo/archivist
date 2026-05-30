---
name: archivist-integrator
description: Use when merging reviewed Archivist worker branches, resolving conflicts, preserving approved behavior, running integration verification, recording integration results, and preparing cleanup.
---

# Archivist Integrator

Use this skill after required worker branches have completed and reviewer findings are resolved or explicitly waived.

## Preconditions

- Feature branch exists or coordinator has assigned an integration base.
- Worker branches are identified.
- Review status is recorded.
- No unresolved blocking findings remain.
- The working tree is clean except for coordinator-approved integration state.

## Integration Rules

- Start from the feature branch.
- Confirm git status before each merge.
- Merge one reviewed branch at a time.
- Preserve reviewed behavior from each branch.
- Resolve conflicts narrowly and according to canonical ALM contracts.
- Do not introduce new product behavior during integration.
- Do not drop worker changes without recording why.
- Stop when conflict resolution requires a product decision.

## Generated And Derived Artifacts

Regenerate derived artifacts only when the repository already defines a regeneration command for that artifact. Record the command and result. Do not hand-edit generated artifacts unless the repository explicitly uses checked-in generated files and no generator is available.

## Verification

Run validation required by the merged task set. For full cross-module integration, use the relevant subset:

```bash
cd src/gateway && dotnet build && dotnet test
cd src/worker && go tool lefthook run build && go tool lefthook run lint && go tool lefthook run test
cd src/ui && npm run lint && npm run build && npm run test
```

Run formatters only when integration changed source files and the coordinator expects formatting.

## Output

Report:

- feature branch used;
- branches merged;
- conflicts and resolutions;
- generated or derived artifacts updated;
- verification commands and results;
- cleanup targets safe to remove;
- residual risks.
