---
name: bug
description: End-to-end bug-fix workflow. Starts with bug-adapted brainstorming via parallel explore agents, files a GitHub bug issue, runs speckit.specify and speckit.plan, adversarial-reviews the plan via 3 parallel focused reviewers, generates tasks via speckit.tasks, compiles tasks into development slices, and creates GitHub implementation issues using the project's implementation template. Aggressively parallelizes with subagents. Use when reporting, diagnosing, or fixing a bug.
---

# Bug Fix Workflow

Orchestrates the full bug-fix pipeline: brainstorm -> bug issue -> specify -> plan -> adversarial review -> tasks -> slices -> GitHub issues.

**Terminal state:** Bug issue has `bug` label, implementation issues have `implementation` label, plan artifacts exist in `specs/<bug>/`. This skill does NOT execute the fix.

## Context Window Protection

**The main context is an orchestrator.** It dispatches, coordinates, and makes decisions. It does NOT:

- Read source files directly (dispatch Explore agents)
- Investigate root causes directly (dispatch investigation agents)
- Review plan artifacts alone (dispatch review agents)
- Create GitHub issues in bulk (dispatch background agents)

Every heavy operation goes to a subagent. The main context stays lean for decision-making and user interaction.

## Prerequisites

- Project has `.specify/` infrastructure (speckit templates and scripts)
- Project has `docs/implementation-issue-template.md`
- `gh` CLI authenticated with repo access
- `docs/prompts/adversarial-plan-review.md` exists (adversarial review prompt)

## Entry Point

- If input is a number -> treat as an existing GitHub issue number, fetch it with `gh issue view <number> --json title,body,number,labels`, skip to Phase 2
- If input is empty or text -> user is describing a bug, start at Phase 0

## Workflow

### Phase 0: Brainstorm (Bug-Adapted, Agent-Driven)

**Always start here** (unless an existing issue number is provided). Full brainstorming rigor, adapted for diagnostic work.

> HARD GATE: Do NOT proceed to Phase 1 until a design doc is written and committed.

#### Understanding the bug — Parallel Exploration

Dispatch **2-3 parallel Explore agents** immediately to investigate while asking the user clarifying questions:

**Explore Agent 1 — Symptom trace (background):**
```
Explore the codebase to understand how <affected feature> works.
Trace the code path from <entry point> through to <where symptom appears>.
Map the call chain across packages (handler -> analytics -> repository -> SQL).
Identify where the behavior diverges from expectations.
Return: code path summary, key files involved, where the bug likely lives.
```

**Explore Agent 2 — Recent changes (background):**
```
Search git history for recent changes to files related to <affected area>.
Run: git log --oneline -20 -- <relevant paths>
Read the diffs of the most relevant commits.
Check if any recent migrations or query changes could have introduced this.
Return: summary of recent changes that could have introduced this bug.
```

**Explore Agent 3 (if applicable) — Related issues/tests (background):**
```
Search for existing tests covering <affected feature>.
Search GitHub issues for similar reports: gh issue list --search "<keywords>"
Check sqlc query definitions in internal/repository/queries/ for related queries.
Return: relevant test files, any related issues, what coverage exists.
```

All three run in the **background** while the main context asks clarifying questions:

- One question at a time
- Prefer multiple choice questions when possible
- Root cause investigation — synthesize agent findings as they return, confirm with user

#### Exploring fix approaches

- Propose 2-3 fix approaches with tradeoffs
- Lead with your recommendation and rationale

#### Presenting the design

- Scale each section to its complexity
- Cover: root cause, fix approach, affected packages, risk assessment, testing strategy
- Ask after each section whether it looks right so far
- Be ready to go back and clarify

#### Design doc

- Save to `docs/plans/YYYY-MM-DD-bug-<topic>-design.md`
- Commit before proceeding

### Phase 1 + Phase 2: Bug Issue + Specify (Parallel)

These two derive from the approved design doc and do not depend on each other. Run them simultaneously.

> HARD GATE: Do NOT proceed to Phase 3 until the bug issue is published with the `bug` label and you have its number, AND speckit.specify has completed.

**Background Agent — Bug Issue (label: `bug`):**
```
Create a GitHub issue with the `bug` label using `gh issue create`.
Use the design doc at <path> as the source of truth.

Include all seven required sections:
1. Bug Description — narrative prose: who encounters it, what goes wrong, why painful
2. Reproduction Steps — numbered steps with preconditions (data state, config, environment)
3. Expected vs Actual Behavior — side-by-side comparison
4. Root Cause Analysis — which code path, which logic error, which assumption failed
5. Affected Surfaces — table (Surface | Impact)
6. Non-Requirements — what this fix explicitly does NOT address
7. Open Questions — decisions left for implementation

Rules: NO implementation details — no file paths, function names, package names, or API routes.
If a section doesn't apply, write "N/A" with a brief reason — do not omit it.

Return the issue URL and number when created.
```

**Foreground — Specify:**
```
/speckit.specify <approved design doc or bug description>
```

This produces `specs/<bug>/spec.md`.

**Wait for both to complete. Capture the issue number from the background agent.**

### Phase 3: Plan

Invoke `/speckit.plan` to generate design artifacts from the spec.

```
/speckit.plan
```

This produces in `specs/<bug>/`:
- `plan.md` — technical context, structure, constitution check
- `research.md` — resolved unknowns, technology decisions
- `data-model.md` — entities, fields, relationships
- `contracts/` — interface contracts (API shapes, query signatures, types)
- `quickstart.md` — test scenarios

**Wait for speckit.plan to complete before proceeding.**

### Phase 4: Adversarial Review (3 Parallel Reviewers)

Launch **3 focused review agents in parallel** — each reviews from a different lens. All three read `docs/prompts/adversarial-plan-review.md` for the full review protocol but focus on their assigned domain.

**Agent 1 — Spec Coverage & Structure:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 1-2 and Step 6 (spec coverage + structure audit).

Review specs/<bug>/plan.md against specs/<bug>/spec.md.
Check: every spec requirement has a corresponding plan element.
Check: no plan elements exceed spec scope (scope creep).
Check: acceptance criteria are testable and complete.
Check: plan structure matches constitution conventions.
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Agent 2 — Security, Error Handling & Cross-Cutting:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 5, 9 (security, error handling, cross-cutting gaps).

Review all artifacts in specs/<bug>/.
Check: error handling for all failure modes.
Check: no security concerns (injection, auth bypass, data exposure).
Check: edge cases covered in test scenarios (quickstart.md).
Check: cross-cutting concerns (multi-tenancy scoping, timezone handling).
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Agent 3 — Data Model, Contracts & Research:**
```
Read docs/prompts/adversarial-plan-review.md for review protocol context.
Focus on Steps 3-4 and Step 8 (research, data model, contracts).

Review specs/<bug>/data-model.md and specs/<bug>/contracts/.
Check: data model changes follow expand-and-contract migration pattern.
Check: sqlc query signatures match repository conventions.
Check: Go types match boundary file conventions (<package>.go).
Check: API contracts match handler patterns.
Check: research decisions are sound and alternatives were considered.
If issues found: fix them directly in the plan artifacts.
Return: findings table (# | Severity | Issue | Resolution)
```

**Wait for all three to complete.** Consolidate findings into an "## Adversarial Review Decisions" section appended to `specs/<bug>/plan.md` containing:
- Combined summary findings table (# | Severity | Issue)
- Decisions that needed human input (if any — noted but not blocking)
- Brief list of all files modified during review

### Phase 5: Tasks

Invoke `/speckit.tasks` to generate the task list from the reviewed plan.

```
/speckit.tasks
```

This produces `specs/<bug>/tasks.md` with:
- Setup phase (project initialization)
- Foundational phase (blocking prerequisites)
- Fix phases (in priority order)
- Polish phase (cross-cutting concerns)
- Dependency graph and parallel execution opportunities

**Wait for speckit.tasks to complete before proceeding.**

### Phase 6: Compile Development Slices

Read the generated `tasks.md` and compile tasks into **development slices** — each slice becomes one GitHub implementation issue.

**Bug fixes should default to a single slice.** Most bugs are a focused fix across 1-2 layers. Only split into multiple slices if the fix genuinely requires independent, separately-mergeable changes.

#### When to split

- Fix requires a schema migration that other work depends on
- Fix spans backend + frontend with no shared merge path
- Fix exceeds ~3 hours of implementation work in a single slice

#### When NOT to split

- Fix touches multiple files in the same package — that's one slice
- Fix touches a query + its handler — that's one slice
- Fix modifies a migration + query + handler in a linear chain — still one slice if < 3 hours

#### Slice Compilation Process

1. Read `tasks.md` and the project's `docs/implementation-issue-template.md`
2. **Default: combine all tasks into a single slice.** Only split if the criteria above are met.
3. For each slice, identify:
   - Which tasks it includes (by task ID)
   - Which packages it touches (Technical Scope)
   - What data contracts it introduces (from `contracts/`, `data-model.md`)
   - What it's blocked by and what it blocks
   - Acceptance criteria (derived from task descriptions)
4. **Identify parallelization opportunities**: which slices are independent (can be implemented simultaneously) vs dependent (must be sequential)
5. Present the proposed slice(s) to the user with dependency graph:

```
Proposed slice:
1. [Title] — T001-T005 — repository/, handler/ — Blocked by: N/A
```

Or if multiple slices with parallelization analysis:

```
Proposed slices:
1. [Title] — T001-T002 — migrations/, repository/ — Blocked by: N/A
2. [Title] — T003-T005 — handler/, web/src/ — Blocked by: #1

Parallelization: Slices are sequential (slice 2 depends on slice 1).
```

Ask: "Does this slicing look right?"

**Wait for user approval before creating issues.**

### Phase 7: Create GitHub Issues (Background Agent)

Dispatch a **background agent** to create all implementation issues:

```
Create GitHub implementation issues for each approved slice.
The parent bug issue is #<number> (label: bug).
Use the implementation issue template at docs/implementation-issue-template.md.

For each slice, create an issue with `gh issue create`:
- Label: implementation
- Title: concise, action-oriented (e.g., "Fix revenue calculation for voided orders")
- Body: all 8 sections from the template

All 8 sections required:
1. Parent — bug issue number + reading instruction
2. Intent — what this slice fixes + what it unblocks
3. Technical Scope — package/action/purpose table
4. Data Contracts — from contracts/, data-model.md (Schema, sqlc Query Signatures, Go Types, HTTP Shapes — include only applicable subsections)
5. Approach — prose rationale referencing plan decisions (no code blocks)
6. Acceptance Criteria — [ ] checkboxes, developer-verifiable
7. Dependencies — Blocked by / Blocks with issue numbers
8. Out of Scope — what this slice does NOT touch

After creating each issue, update Dependencies of previously created issues with `gh issue edit`.

Quality checks before publishing:
- Parent section references the bug issue with full reading instruction
- Intent answers "what" and "what becomes possible"
- Technical Scope is package-level (not file-level)
- Data Contracts is a dedicated section (no schema buried in Approach)
- Data Contracts subsections match template structure
- Approach is prose rationale (no code blocks)
- All acceptance criteria are [ ] checkboxes
- Dependencies use exact "Blocked by / Blocks" format with issue numbers
- Out of Scope names excluded concerns explicitly
- Column names and types match the plan's contracts exactly

Slices to create:
<slice details>

Return: summary table (# | Issue | Title | Blocked By | Blocks)
```

While the background agent works, inform the user that issues are being created. Present the summary table when the agent completes.

## Error Handling

- If existing issue number provided: skip Phase 0-1, start at Phase 2
- If `speckit.specify` fails: report the error, do not proceed
- If `speckit.plan` fails: report the error, do not proceed
- If any adversarial review agent fails: report findings from agents that completed, ask user whether to proceed with partial review
- If `speckit.tasks` fails: report the error, do not proceed
- If the user wants to re-slice: go back to Phase 6 with their feedback
- If `gh issue create` fails: retry once, then ask the user
- If no `.specify/` infrastructure exists: tell the user to run `/speckit.constitution` first
- If no implementation issue template exists: use the 8-section structure defined in Phase 7

## What This Skill Does NOT Do

- **Implementation** — creating issues is the terminal state; coding is a separate step
- **Deployment** — issues are created, not deployed

## Key Principles

- **Orchestrator pattern** — main context dispatches, never does heavy lifting
- **Parallel by default** — if two things don't depend on each other, run them simultaneously
- **Background for non-blocking** — use `run_in_background: true` when the main context doesn't need results immediately
- **One question at a time** — don't overwhelm during brainstorming
- **Multiple choice preferred** — easier to answer than open-ended
- **YAGNI ruthlessly** — fix the bug, don't refactor the neighborhood
- **Single slice default** — most bugs are one focused fix
- **Hard gates are non-negotiable** — each phase must complete before the next begins
- **No implementation details in the bug issue** — file paths, function names, and package names go in the implementation issues only
