---
name: feature-request
description: End-to-end feature planning and issue creation. Starts with brainstorming via parallel explore agents, runs speckit.specify and speckit.plan, adversarial-reviews the plan via 3 parallel focused reviewers, generates tasks via speckit.tasks, compiles tasks into development slices with parallelization analysis, and creates GitHub implementation issues via background agent. Aggressively parallelizes with subagents. Use when the user wants to plan a new feature end-to-end or says "feature request".
---

# Feature Request

Orchestrates the full feature planning pipeline: brainstorm -> specify -> plan -> adversarial review -> tasks -> slices -> GitHub issues.

**Terminal state:** Feature issue has `feature` label, implementation issues have `implementation` label, plan artifacts exist in `specs/<feature>/`. This skill does NOT execute the implementation.

## Context Window Protection

**The main context is an orchestrator.** It dispatches, coordinates, and makes decisions. It does NOT:

- Read source files directly (dispatch Explore agents)
- Investigate architecture directly (dispatch exploration agents)
- Review plan artifacts alone (dispatch review agents)
- Create GitHub issues in bulk (dispatch background agents)

Every heavy operation goes to a subagent. The main context stays lean for decision-making and user interaction.

## Prerequisites

- Project has `.specify/` infrastructure (speckit templates and scripts)
- Project has `docs/implementation-issue-template.md`
- `gh` CLI authenticated with repo access
- `docs/prompts/adversarial-plan-review.md` exists (adversarial review prompt)

## Workflow

### Phase 0: Brainstorm (Agent-Driven Exploration)

**Always start here.** If the user has already completed brainstorming and has an approved design doc, skip to Phase 1.

Before invoking `/brainstorming`, dispatch **parallel Explore agents** to gather context that will inform the brainstorming session:

**Explore Agent 1 — Architecture scan (background):**
```
Explore the codebase architecture relevant to <feature area>.
Map: key packages involved (handler, analytics, repository, seed), entry points, data flow.
Check: boundary files (<package>.go) for existing exported types and interfaces.
Identify: extension points where new functionality would plug in.
Return: architecture summary, relevant files, patterns to follow.
```

**Explore Agent 2 — Similar features (background):**
```
Search the codebase for features similar to <proposed feature>.
Find: how existing features are structured across layers (migration -> repository -> analytics -> handler).
Check: sqlc query patterns in internal/repository/queries/.
Check: any prior attempts or related issues in git history.
Return: similar feature patterns, relevant precedents, lessons learned.
```

**Explore Agent 3 — Dependencies & constraints (background):**
```
Investigate the dependency landscape for <feature area>.
Check: existing tables and migrations that would be affected.
Check: API endpoint patterns in internal/handler/.
Check: any multi-tenancy implications (tenant_id scoping).
Identify: constraints that would affect implementation approach.
Return: dependency map, constraints, potential blockers.
```

Then invoke brainstorming (agents return context as they complete):

```
/brainstorming
```

When brainstorming completes and the user approves a design, capture the approved design doc path and proceed.

### Phase 1: Specify

Pass the approved brainstorming output to `speckit.specify` to create a formal feature specification.

```
/speckit.specify <approved design doc or feature description>
```

This produces `specs/<feature>/spec.md` — the canonical feature specification with requirements, user stories, and acceptance criteria.

**Wait for speckit.specify to complete before proceeding.**

### Phase 2: Plan

Invoke `/speckit.plan` to generate design artifacts from the spec.

```
/speckit.plan
```

This produces in `specs/<feature>/`:
- `plan.md` — technical context, structure, constitution check
- `research.md` — resolved unknowns, technology decisions
- `data-model.md` — entities, fields, relationships
- `contracts/` — interface contracts (API shapes, query signatures, types)
- `quickstart.md` — test scenarios

**Wait for speckit.plan to complete before proceeding.**

### Phase 3: Adversarial Review (3 Parallel Reviewers)

Launch **3 focused review agents in parallel** — each reviews from a different lens. All three read `docs/prompts/adversarial-plan-review.md` for the full review protocol but focus on their assigned domain.

**Agent 1 — Spec Coverage & Structure:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 1-2 and Step 6 (spec coverage + structure audit).

Review specs/<feature>/plan.md against specs/<feature>/spec.md.
Check: every spec requirement has a corresponding plan element.
Check: every user story has plan coverage.
Check: no plan elements exceed spec scope (scope creep).
Check: acceptance criteria are testable and complete.
Check: plan structure matches constitution conventions.
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Agent 2 — Security, Error Handling & Cross-Cutting:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 5, 7, 9 (security, quickstart validation, cross-cutting gaps).

Review all artifacts in specs/<feature>/.
Check: error handling for all failure modes.
Check: no security concerns (injection, auth bypass, data exposure).
Check: edge cases covered in test scenarios (quickstart.md).
Check: cross-cutting concerns (multi-tenancy scoping, timezone handling, closed days).
Check: input validation at API boundaries.
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Agent 3 — Data Model, Contracts & Research:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 3-4 and Step 8 (research, data model, contracts).

Review specs/<feature>/data-model.md and specs/<feature>/contracts/.
Check: data model changes follow expand-and-contract migration pattern.
Check: new tables include tenant_id column with proper default and scoping.
Check: sqlc query signatures match repository conventions.
Check: Go types match boundary file conventions (<package>.go).
Check: API contracts match handler patterns (Chi router, JSON responses).
Check: research decisions are sound and alternatives were considered.
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Wait for all three to complete.** Consolidate findings into an "## Adversarial Review Decisions" section appended to `specs/<feature>/plan.md` containing:
- Combined summary findings table (# | Severity | Issue)
- Decisions that needed human input (if any — noted but not blocking)
- Brief list of all files modified during review

### Phase 4: Tasks

Invoke `/speckit.tasks` to generate the task list from the reviewed plan.

```
/speckit.tasks
```

This produces `specs/<feature>/tasks.md` with:
- Setup phase (project initialization)
- Foundational phase (blocking prerequisites)
- Per-user-story phases (in priority order)
- Polish phase (cross-cutting concerns)
- Dependency graph and parallel execution opportunities

**Wait for speckit.tasks to complete before proceeding.**

### Phase 5: Compile Development Slices

Read the generated `tasks.md` and compile tasks into **development slices** — each slice becomes one GitHub implementation issue.

#### Slicing Rules

1. **A slice is the smallest unit that can be merged independently** without breaking the codebase.
2. **Target 1-3 hours of implementation work per slice.**
3. **Respect layer boundaries.** A slice should ideally touch one stack layer (see CLAUDE.md Stack Layers). Cross-layer slices are acceptable when the layers are tightly coupled (e.g., a new query + its handler).
4. **Preserve dependency order.** Slices must be orderable so each can be merged in sequence.
5. **Group by capability, not by file type.** "Add weather queries + handler" is a good slice. "Add all migrations" is a bad slice.

#### Parallelization Analysis

For each slice, determine:
- **Independent** — can be implemented in parallel with other slices (no shared files, no dependency)
- **Dependent** — must wait for a specific other slice to merge first

Present proposed slices with a **dependency graph** showing parallelization opportunities:

```
Proposed slices:
1. [Title] — T001-T003 — migrations/, repository/ — Blocked by: N/A
2. [Title] — T004-T007 — repository/, handler/ — Blocked by: #1
3. [Title] — T008-T012 — web/src/ — Blocked by: #2

Dependency graph:
  Slice 1 (DB + queries) --> Slice 2 (handler) --> Slice 3 (frontend)

Parallelization: Sequential chain — no parallel opportunities.
```

Or with parallel opportunities:

```
Proposed slices:
1. [Title] — T001-T003 — migrations/, repository/ — Blocked by: N/A
2. [Title] — T004-T006 — analytics/ — Blocked by: #1
3. [Title] — T007-T009 — handler/ — Blocked by: #1
4. [Title] — T010-T012 — web/src/ — Blocked by: #2, #3

Dependency graph:
  Slice 1 ──┬──> Slice 2 (analytics) ──┐
            └──> Slice 3 (handler)  ────┴──> Slice 4 (frontend)

Parallelization: Slices 2 and 3 can run in parallel after Slice 1 merges.
```

Ask: "Does this slicing look right? Want to merge or split any slices?"

**Wait for user approval before creating issues.**

### Phase 6: Create GitHub Issues (Background Agent)

Dispatch a **background agent** to create all issues:

```
Create GitHub issues for a feature implementation.
Use the implementation issue template at docs/implementation-issue-template.md.

1. Create parent feature issue (label: feature) using docs/feature-request-template.md
   if one doesn't exist yet.

2. For each slice, create an issue with `gh issue create`:
   - Label: implementation
   - Title: concise, action-oriented (e.g., "Add weather_daily table and repository queries")
   - Body: all 8 sections from the template

All 8 sections required:
1. Parent — feature issue number + reading instruction
2. Intent — what this slice builds + what it unblocks
3. Technical Scope — package/action/purpose table
4. Data Contracts — from contracts/, data-model.md (Schema, sqlc Query Signatures, Go Types, HTTP Shapes — include only applicable subsections)
5. Approach — prose rationale referencing plan decisions (no code blocks)
6. Acceptance Criteria — [ ] checkboxes, developer-verifiable
7. Dependencies — Blocked by / Blocks with issue numbers
8. Out of Scope — what this slice does NOT touch

After creating each issue, update Dependencies of previously created issues with `gh issue edit`.

Quality checks before publishing:
- Parent section includes full reading instruction
- Intent answers "what" and "what becomes possible"
- Technical Scope is package-level (not file-level)
- Data Contracts is a dedicated section (no schema buried in Approach)
- Data Contracts subsections match template structure
- Approach is prose rationale (no code blocks)
- All acceptance criteria are [ ] checkboxes
- Dependencies use exact "Blocked by / Blocks" format with issue numbers
- Out of Scope names excluded concerns explicitly
- Column names and types match the plan's contracts exactly

Parent feature description: <feature summary>

Slices to create:
<slice details with dependency info>

Return: summary table (# | Issue | Title | Blocked By | Blocks)
```

While the background agent works, inform the user that issues are being created. Present the summary table when the agent completes.

## Error Handling

- If `/brainstorming` is skipped (user has approved design): resume at Phase 1
- If `speckit.specify` fails: report the error, do not proceed
- If `speckit.plan` fails: report the error, do not proceed
- If any adversarial review agent fails: report findings from agents that completed, ask user whether to proceed with partial review
- If `speckit.tasks` fails: report the error, do not proceed
- If the user wants to re-slice: go back to Phase 5 with their feedback
- If `gh issue create` fails: retry once, then ask the user
- If no `.specify/` infrastructure exists: tell the user to run `/speckit.constitution` first
- If no implementation issue template exists: use the 8-section structure defined in Phase 6

## What This Skill Does NOT Do

- **Implementation** — creating issues is the terminal state; coding is a separate step
- **Deployment** — issues are created, not deployed

## Key Principles

- **Orchestrator pattern** — main context dispatches, never does heavy lifting
- **Parallel by default** — if two things don't depend on each other, run them simultaneously
- **Background for non-blocking** — use `run_in_background: true` when the main context doesn't need results immediately
- **Dependency graph visualization** — always show which slices can run in parallel
- **Hard gates are non-negotiable** — each phase must complete before the next begins
- **No implementation details in the feature issue** — file paths, function names, and package names go in the implementation issues only
