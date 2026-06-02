# Reusable Prompt: Add Repo-Local Multi-Agent Workflow

Use this prompt in a repository that should adopt a local `.agents` multi-agent workflow while preserving that repository's existing source-of-truth, planning, rebuild, bookkeeping, or ALM system.

---

## Prompt

You are working inside a repository that needs repo-local multi-agent workflow tooling. Implement the workflow by adding or updating `.agents` agent profiles, skills, and workflow templates only. Do not change runtime product code unless I explicitly ask for product implementation.

The goal is to make it easy for a main coordinator agent to decompose a feature, dispatch module-specific worker agents, route work through matching reviewers, and have a merger/integrator consolidate approved work. This must be adapted to this repository's actual modules, languages, framework choices, validation commands, documentation structure, and source-of-truth system.

Treat any attached repositories, prompts, agent files, or skill files as examples and source material, not binding rules. Import only rules that fit this repository's actual language, framework, architecture, runtime, testing tools, deployment model, and current code shape.

## Hard Constraints

- Preserve the existing source-of-truth system. If this repository has ALM, specs, plans, tasks, ADRs, design docs, rebuild docs, bookkeeping docs, or equivalent governance files, treat them as canonical.
- Keep `.agents` non-canonical development tooling unless the repository already defines a different explicit rule.
- Do not create a new ALM system.
- Do not replace existing canonical docs with skills.
- Do not move durable contracts into `.agents`.
- Durable behavior, public interfaces, persistence, security, deployment, artifact layout, user-visible errors, and rebuild-relevant decisions must remain in canonical project documents.
- Skills may restate implementation guidance and checklists, but they must not become the only source for runtime/product contracts.
- Do not introduce references to source repositories, copied project names, or unrelated tooling from examples.
- Do not touch runtime code.

## Initial Inspection

First inspect the target repository and identify:

- repository name;
- root agent instructions such as `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, or equivalents;
- `README.md`, `docs/`, plans, diaries, architecture docs, design docs, and convention docs;
- package manifests and tool configuration files;
- application entrypoints;
- product/runtime modules;
- backend, frontend, worker, CLI, mobile, library, or generated-artifact directories;
- languages and frameworks per module;
- test, lint, format, and validation commands;
- existing `.agents` files, if any;
- existing canonical documentation and bookkeeping/source-of-truth files;
- existing implementation conventions, style guides, or development notes;
- existing dependency injection, routing, API, persistence, UI, generated-artifact, deployment, and authentication patterns;
- current git status;
- whether language-specific agent profiles already exist and should be merged into skills.

Read only the minimum context needed before editing. Prefer `rg` and `rg --files`.

## Target Structure

Create or update this repo-local structure, adapting names to the repository:

```text
.agents/
  agents/
    coordinator.md
    <module-a>-worker.md
    <module-a>-reviewer.md
    <module-b>-worker.md
    <module-b>-reviewer.md
    <frontend-or-ui>-worker.md
    <frontend-or-ui>-reviewer.md
    merger.md
  skills/
    <repo>-agentic-feature/
      SKILL.md
    <repo>-general/
      SKILL.md
    <repo>-<module-a>/
      SKILL.md
    <repo>-<module-b>/
      SKILL.md
    <repo>-<frontend-or-ui>/
      SKILL.md
    <repo>-reviewer/
      SKILL.md
    <repo>-integrator/
      SKILL.md
    planner-agent/
      SKILL.md
  workflows/
    templates/
      feature-state.md
      task-handoff.md
      review-report.md
      integration-report.md
```

Use the actual module names. If the repository has different modules, create one worker and one reviewer profile per independently assignable implementation surface.

Examples:

- `gateway-worker.md` / `gateway-reviewer.md` for an API/backend surface.
- `worker-worker.md` / `worker-reviewer.md` for an async worker or job processor.
- `frontend-worker.md` / `frontend-reviewer.md` for a web UI.
- `mobile-worker.md` / `mobile-reviewer.md` for a mobile app.
- `cli-worker.md` / `cli-reviewer.md` for a CLI module.

If the repository has only one implementation surface, still create a coordinator, one worker, one reviewer, merger, and the shared skills.

## Agent Profiles

Role profiles under `.agents/agents/` should be about ownership, workflow, allowed edits, forbidden edits, verification, escalation, and reporting. They should not carry large language/framework convention blocks. Put those conventions in skills.

Every role profile must contain these sections:

- `# <Role Name>`
- `## Purpose`
- `## Required Reading`
- `## Ownership`
- `## Forbidden Edits`
- `## Workflow Rules` or `## Review Rules`
- `## Verification`
- `## Escalation`
- `## Final Report`

### Coordinator

Create `.agents/agents/coordinator.md`.

The coordinator owns:

- initial source-of-truth and repo context inspection;
- feature decomposition;
- readiness checks;
- clarification gate before dispatch;
- selecting only the needed roles;
- assigning disjoint write scopes;
- branch and worktree naming guidance if the repo uses parallel worktrees;
- worker handoff creation;
- reviewer dispatch;
- routing findings back to workers;
- deciding when work is approved, waived, blocked, or ready to integrate;
- handing approved branches/scopes to merger;
- final validation summary;
- canonical doc/task/status updates only where the target repository's rules require them.

The coordinator must not invent product behavior outside canonical requirements.

### Workers

Create one worker role per implementation surface.

Each worker profile must:

- load the general skill and its module skill;
- declare its owned paths;
- declare tests and validation it may run;
- declare docs/task/spec files it may update when assigned;
- avoid editing other module surfaces unless explicitly assigned;
- record assumptions and blockers rather than invent durable behavior.

Example required reading:

```md
- `.agents/skills/<repo>-general/SKILL.md`
- `.agents/skills/<repo>-<module>/SKILL.md`
- canonical repository instructions and source-of-truth files identified during inspection
```

### Reviewers

Create one reviewer role per worker role.

Each reviewer must load at least the same implementation skills as its matching worker plus the shared reviewer skill.

Required parity rule:

- `<module>-worker.md` skills must be a subset of `<module>-reviewer.md` skills.
- `<module>-reviewer.md` must also load `.agents/skills/<repo>-reviewer/SKILL.md`.

Reviewers are read-only by default. They report findings first, ordered by severity, with file/line references when applicable. They must check:

- source-of-truth compliance;
- acceptance criteria;
- public contract changes;
- security boundaries;
- persistence or schema effects;
- error and artifact behavior;
- concurrency and integration risk;
- test coverage and validation gaps;
- module-specific conventions from the same skills workers use.

### Merger

Create `.agents/agents/merger.md`.

The merger owns integration of approved or explicitly waived work only. It must:

- read the integrator skill;
- merge or reconcile worker branches/scopes narrowly;
- resolve conflicts without adding new product behavior;
- run final validation;
- update coordination templates;
- report integrated branches, conflicts, validations, unresolved risks, and cleanup.

## Skills

Every skill must be a `SKILL.md` with frontmatter:

```md
---
name: <skill-name>
description: <when to use this skill>
---
```

Use lowercase hyphenated names.

### `<repo>-general`

Create or update `.agents/skills/<repo>-general/SKILL.md`.

Include cross-module development practice:

- source-of-truth discipline;
- dependency restraint;
- configuration hygiene;
- naming guidance;
- validation expectations;
- security reminders;
- testing standards;
- non-contract coding rules;
- how to handle missing durable decisions.

Keep this generic across modules.

### Module Skills

Create one module skill per independently assignable implementation surface.

Each module skill should include:

- owned paths;
- language and framework conventions;
- project-local architecture patterns;
- dependency and injection patterns;
- persistence or API conventions if implementation-level;
- test locations and validation commands;
- common pitfalls;
- references to canonical docs for durable contracts.

Do not make module skills the only home for runtime behavior or public contracts.

### Language-Specific Guidance

If the repository has standalone language-specific agent profiles, merge their code-specific guidance into the relevant module skill and delete the standalone role profile.

Examples:

- Move Go guidance into a worker/backend module skill.
- Move TypeScript/React guidance into a UI module skill.
- Move C#/.NET guidance into an API/gateway module skill.

Role profiles should stay focused on ownership and workflow. Skills should hold implementation practice.

### Reviewer Skill

Create `.agents/skills/<repo>-reviewer/SKILL.md`.

Include review procedure:

- review stance;
- severity ordering;
- source-of-truth checks;
- module skill parity requirement;
- security and public contract checks;
- testing and validation expectations;
- report format;
- when to request worker changes;
- when a finding can be waived.

### Integrator Skill

Create `.agents/skills/<repo>-integrator/SKILL.md`.

Include integration procedure:

- inputs expected from coordinator;
- approved/waived work only;
- branch/worktree handling if applicable;
- conflict resolution policy;
- final validation;
- final report;
- cleanup notes.

### Agentic Feature Skill

Create `.agents/skills/<repo>-agentic-feature/SKILL.md`.

Define the full coordinator loop:

1. inspect git status and source-of-truth context;
2. identify feature, modules, and durable surfaces;
3. run a pre-worker clarification gate;
4. select only the needed roles;
5. assign disjoint write scopes;
6. create worker handoffs;
7. spawn workers;
8. collect worker reports;
9. spawn matching reviewers;
10. route findings back to workers;
11. merge only approved or explicitly waived work;
12. run final validation;
13. update source-of-truth/task/status files if required by the repository;
14. produce a final report and cleanup notes.

### Planner Skill

If the repository already has planning or ALM conventions, update or create `.agents/skills/planner-agent/SKILL.md` to support them.

The planner skill must say:

- durable decisions go to canonical repo docs, not `.agents`;
- task/status updates follow the repository's existing process;
- `.agents/workflows/templates` are scratch coordination artifacts, not canonical product truth.

If the repository has no formal ALM system, the planner skill should remain lightweight and point to the repository's existing planning documents, issues, or README conventions.

## Workflow Templates

Create non-canonical templates under `.agents/workflows/templates/`:

- `feature-state.md`
- `task-handoff.md`
- `review-report.md`
- `integration-report.md`

Templates should record:

- feature/task identifier;
- source-of-truth context;
- branches/worktrees if used;
- role assignments;
- write scopes;
- read-only scopes;
- validation commands;
- worker status;
- reviewer findings;
- waived findings;
- integration state;
- open decisions;
- cleanup notes.

Each template must state that it is coordination scratch/state and does not replace canonical repository docs.

## Root Instruction Update

If the repository has a root agent instruction file such as `AGENTS.md`, `CLAUDE.md`, `.cursorrules`, or equivalent, add a short pointer:

- significant parallel feature work may use `.agents/skills/<repo>-agentic-feature/SKILL.md`;
- role profiles live in `.agents/agents/`;
- workflow templates live in `.agents/workflows/templates/`;
- `.agents` is non-canonical workflow tooling;
- durable behavior remains in canonical repository docs/source-of-truth files.

Do not substantially rewrite the root instruction file unless needed to remove contradictions.

## Convention Filtering Rules

When importing from example documentation, prompts, role profiles, or skills, keep only guidance that matches the target repository.

Keep:

- toolchain-specific rules that match this repository;
- existing architecture and module-boundary rules;
- existing dependency injection or composition-root rules;
- API contract, persistence, UI, worker, CLI, or generated-artifact rules that are already relevant;
- testing and verification rules that match available commands;
- security rules already relevant to the current architecture;
- review and integration rules that improve correctness without changing product behavior.

Discard:

- rules for languages, frameworks, package managers, CLIs, or deployment tools not used in this repository;
- domain concepts from other products;
- authentication, persistence, queue, worker, artifact, gateway, UI, or deployment assumptions not already true for this repository;
- rules that require broad architecture changes;
- rules that introduce speculative abstractions;
- rules that duplicate or conflict with the repository's canonical source-of-truth files.

When uncertain, do not import the questionable rule. Add a short `Incomplete Conventions` note inside the relevant skill explaining what still needs human or future-agent review.

## Reference Cleanup

Search for stale or copied references after editing:

- old repository names from examples;
- unrelated toolchains from examples;
- stale language-agent references after migrating guidance into skills;
- references that treat `.agents` as canonical product truth;
- references to non-existent roles or skills.

If stale references appear in canonical docs, update only the smallest necessary text to preserve the repository's authority model.

## Validation

Run structural and hygiene validation before final response.

Recommended checks:

```sh
find .agents/agents -maxdepth 1 -type f | sort
find .agents/skills -maxdepth 2 -name SKILL.md | sort
find .agents/workflows/templates -maxdepth 1 -type f | sort
rg -n '<old-repo-name>|<example-tooling>|<stale-workflow-doc-path>' .agents AGENTS.md 2>/dev/null || true
rg -n 'go-developer-agent|developer-agent.md' .agents 2>/dev/null || true
git diff --check
```

Adapt the stale-reference patterns to the actual example content being migrated.

Also verify manually:

- all expected role profiles exist;
- every role profile has the required sections;
- every `SKILL.md` has `name` and `description` frontmatter;
- reviewer roles include all matching worker skills plus the reviewer skill;
- module skills reference canonical docs for durable contracts;
- no `.agents` file is described as canonical rebuild/product truth;
- root instructions still preserve the repository's existing source-of-truth model.

## Final Response

Report:

- files created;
- files updated;
- files deleted;
- reviewer skill parity result;
- stale-reference result;
- validation commands run;
- runtime tests skipped or run, with reason;
- any unresolved assumptions.

Keep the final response concise.
