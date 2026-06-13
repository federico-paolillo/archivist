---
name: archivist-agentic-feature
description: Use when coordinating significant Archivist feature work with repo-local multi-agent roles, ALM task readiness, worker/reviewer handoffs, branches, worktrees, integration, verification, cleanup, and optional non-canonical coordination notes.
---

# Archivist Agentic Feature

Use this skill for significant features, cross-module work, multi-day changes, or work that benefits from Gateway, Worker, and UI parallelism.

## Authority Model

- `.agents` files are non-canonical tooling.
- Durable behavior lives in canonical ALM files: `AGENTS.md`, `docs/REBUILD.md`, `docs/BOOKKEEPING.md`, `docs/ARCHITECTURE.md`, `docs/DESIGN.md`, `docs/ERRORS.md`, `docs/ARTIFACTS.md`, and `docs/specs/**`.
- Workflow templates under `.agents/workflows/templates/` are optional non-canonical coordination scratch/state. They do not replace task files, accepted/in-progress ExecPlans, feature `PLAN.md`, or other canonical ALM files.

## Trigger Protocol

When the user asks to use the multi-agent workflow, treat it as a request to coordinate through this workflow, not as automatic permission to skip ALM gates.

If the task is small, single-module, or not parallel-safe, state that the single-agent path is cheaper and continue only if the user still wants the full workflow.

## Required Inputs

- User goal and acceptance criteria.
- Current git status.
- Orientation bundle: `AGENTS.md`, `docs/REBUILD.md`, `docs/specs/INDEX.md`.
- Relevant feature `SPEC.md`, `PLAN.md`, task files, and linked ExecPlans.
- Relevant module skills and role profiles.

## Pre-Worker Clarification Gate

Do not spawn workers until this gate passes.

Ask the user or update canonical ALM artifacts before workers when:

- acceptance criteria are weak, missing, or not verifiable;
- the task is not `ready` and the user has not explicitly assigned it;
- task dependencies are incomplete and the user has not explicitly overridden them;
- Gateway, Worker, UI, persistence, artifact, ARC, API, auth, deployment, or data-flow contracts are ambiguous;
- requested behavior conflicts with canonical docs or accepted design decisions;
- the scope is too large, too vague, or mixes unrelated features;
- the requested approach implies unsafe data behavior, overengineering, architecture drift, or weakened security;
- worker write scopes cannot be made disjoint.

State contradictions directly and ask for the smallest decision that unblocks safe decomposition.

## Role Selection

Use only the roles needed for the feature:

- Gateway implementation: `.agents/agents/gateway-worker.md`
- Worker implementation: `.agents/agents/worker-worker.md`
- UI implementation: `.agents/agents/frontend-worker.md`
- Gateway review: `.agents/agents/gateway-reviewer.md`
- Worker review: `.agents/agents/worker-reviewer.md`
- UI review: `.agents/agents/frontend-reviewer.md`
- Integration: `.agents/agents/merger.md`

The coordinator owns decomposition and final acceptance.

## Execution Loop

1. Inspect repo state, ALM context, relevant code, and relevant skills.
2. Identify affected modules, durable surfaces, and readiness constraints.
3. Run the pre-worker clarification gate.
4. Create or update non-canonical feature state from `.agents/workflows/templates/feature-state.md` when coordination state would otherwise be lost.
5. Decompose work into tasks with explicit objectives, ownership, forbidden paths, validation, and final report requirements.
6. Dispatch workers only with disjoint write scopes.
7. Dispatch reviewers after worker completion.
8. Route reviewer findings back to the owning worker.
9. Repeat review/fix until approved, approved with nits accepted by the coordinator, or explicitly waived.
10. Dispatch the merger only after required approvals or waivers are recorded.
11. Run final verification from the coordinator context.
12. Ensure completed implementation work updates task status and feature `PLAN.md` according to `AGENTS.md`. Diary or coordinator notes may be updated for coordination only, but they are never completion gates.

## Branch And Worktree Defaults

- Feature branch: `codex/<feature-id>`
- Worker branch: `codex/<feature-id>/<module-or-slice>`
- Worktree: `../archivist.worktrees/<feature-id>-<module-or-slice>`

Use collision-free variants when needed.

## Documentation-Only Changes

- Run `git diff --check`.
- Do not run formatters that rewrite source code.
- Full module verification is optional unless root workflow instructions changed in a way that needs extra confidence.
- If optional app verification fails without source edits, report it as existing repository state unless the user assigns a fix.

## Stop Conditions

Stop and report when a destructive operation is required, worker scopes overlap, product intent is ambiguous, a requested approach is a bad practice, canonical docs contradict each other, or validation is blocked by missing external state.

## Output

Coordinator output must include:

- active branches/worktrees;
- roles used and outcomes;
- task state and ALM updates;
- review state;
- integration result;
- validation result;
- cleanup status.
