---
description: End-to-end bug-fix workflow. Brainstorm the bug, file a GitHub issue, engineer the fix, produce an implementation plan. Use when reporting, diagnosing, or fixing a bug.
---

# Bug Fix Workflow

## Intent

End-to-end bug-fix workflow: brainstorm the bug, file a GitHub issue, engineer the fix, produce an implementation plan. Use when reporting, diagnosing, or fixing a bug.

## Overview

Takes a bug from initial report through to a fully specified, pipeline-ready GitHub issue with an implementation plan. One skill invocation, four phases, no manual handoff.

**Terminal state:** Bug issue has `bug` + `implementation` labels, plan doc exists in `docs/plans/`. CI pipeline picks it up from here. This skill does NOT execute the fix.

## Entry Point

- If `$ARGUMENTS` is a number → treat as an existing GitHub issue number, fetch it with `gh issue view $ARGUMENTS --json title,body,number,labels`, skip to Phase 2
- If `$ARGUMENTS` is empty or text → user is describing a bug, start at Phase 1

---

## Phase 1 — Brainstorming (Bug-Adapted)

Full brainstorming rigor, adapted for diagnostic work.

> HARD GATE: Do NOT proceed to Phase 2 until a design doc is written and committed.

### Understanding the bug

- Explore project context first (files, docs, recent commits)
- If user describes symptoms: ask clarifying questions **one at a time** (reproduction steps, frequency, environment, when it started)
- If existing issue: read it, confirm understanding, ask what's missing
- **Root cause investigation** — read relevant code, trace the bug, confirm root cause with user
- Prefer multiple choice questions when possible
- Only one question per message

### Exploring fix approaches

- Propose 2-3 fix approaches with tradeoffs
- Lead with your recommendation and rationale

### Presenting the design

- Scale each section to its complexity
- Cover: root cause, fix approach, affected packages, risk assessment, testing strategy
- Ask after each section whether it looks right so far
- Be ready to go back and clarify

### Design doc

- Save to `docs/plans/YYYY-MM-DD-bug-<topic>-design.md`
- Commit before proceeding

---

## Phase 2 — Bug Issue

> HARD GATE: Do NOT proceed to Phase 3 until the issue is published with the `bug` label and you have its number.

Create a GitHub issue with the `bug` label. **All seven sections are required:**

### Bug Description

Narrative: who encounters it, what they're doing, what goes wrong, why it's painful. The reader should feel the friction. Write this as prose, not bullet points. This section grounds every subsequent decision.

### Reproduction Steps

Numbered steps to reproduce. Include preconditions (data state, config, environment).

### Expected vs Actual Behavior

Side-by-side: what should happen vs what does happen.

### Root Cause Analysis

What the brainstorming investigation found. Which code path, which logic error, which assumption failed.

### Affected Surfaces

Table of every touchpoint the bug impacts:

| Surface | Impact |
|---|---|
| [page, endpoint, table, job, etc.] | [what's broken about it] |

### Non-Requirements

What this fix explicitly does NOT address. Prevents scope creep.

### Open Questions

Decisions deliberately left for implementation.

### Rules

- **No implementation details** — no file paths, function names, package names, or API routes
- If a section doesn't apply, write "N/A" with a brief reason — do not omit it

### Publishing

```bash
gh issue create --label bug --title "<title>" --body "<body>"
```

Confirm the issue URL and number before proceeding.

---

## Phase 3 — Implementation Engineering

> HARD GATE: Do NOT proceed to Phase 4 until the issue has been edited with all 8 implementation template sections.

Same rigor as `/plan`. Operates on the bug issue created in Phase 2.

### Step 1 — Context

1. Read the bug issue in full
2. Explore codebase patterns relevant to the fix. At minimum read:
   - Existing migrations (`migrations/`) if the fix touches schema
   - Existing sqlc queries (`internal/repository/queries/`) for query patterns
   - Any package this fix will modify
3. Identify key architectural decisions (1-5). A decision IS key if it:
   - Shapes the database schema
   - Determines which packages are created or modified
   - Changes the API surface
   - Has multiple reasonable approaches with meaningfully different tradeoffs
   - Is difficult to reverse after implementation begins

### Step 2 — Decision Points

For each key decision, present this format exactly:

---

#### Decision N: [short name]

| Option | Approach | Tradeoffs |
|---|---|---|
| A | [description] | [pros and cons] |
| B | [description] | [pros and cons] |

**Recommendation:** Option [X] — [one sentence rationale].

---

**Present one decision at a time. Wait for the user to choose before presenting the next.**

After the user responds, confirm ("Locked: Option [X]") and present the next decision.

> HARD GATE: All decisions locked before Step 3.

### Step 3 — Contracts

Using the locked decisions, draw precise data contracts. For each data boundary this fix introduces or modifies:

**Schema** (new or modified tables)

| Column | Type | Constraints |
|---|---|---|

**sqlc Query Signatures**

```go
FunctionName(ctx context.Context, arg ParamType) (ReturnType, error)
```

**Go Types** (types that cross package boundaries)

```go
type TypeName struct {
    Field Type
}
```

**HTTP Shapes** (new or modified endpoints)

```
METHOD /api/v1/path
Request:  { "field": type }
Response: { "field": type }
```

Only include subsections that apply. Present to user and confirm before Step 4.

### Step 4 — Edit the bug issue

Append all 8 implementation template sections (per `docs/implementation-issue-template.md`) below the existing bug description. Use a horizontal rule (`---`) to separate the bug description from the implementation sections.

The 8 sections, in order:

1. **Parent** — "N/A — standalone bug fix" or reference a feature issue if related
2. **Intent** — what the fix accomplishes, what becomes possible after it lands
3. **Technical Scope** — package-level table (Package | Action | Purpose)
4. **Data Contracts** — from Step 3 (only subsections that apply)
5. **Approach** — prose rationale referencing locked decisions. No code blocks — those belong in Data Contracts
6. **Acceptance Criteria** — `- [ ]` checkboxes only, developer-verifiable
7. **Dependencies** — Blocked by / Blocks (or N/A)
8. **Out of Scope** — explicit exclusions

```bash
gh issue edit <number> --body "<original body + appended implementation sections>"
```

Do NOT add the `implementation` label yet.

---

## Phase 4 — Plan & Finalize

> HARD GATE: Do NOT add the `implementation` label until `writing-plans` completes and the plan doc is committed.

### Step 1 — Invoke `writing-plans`

Invoke the `writing-plans` skill with the bug issue number. It will:
- Read the issue (bug description + 8 implementation sections)
- Produce a step-by-step implementation plan
- Save to `docs/plans/YYYY-MM-DD-bug-<topic>.md`

### Step 2 — Add `implementation` label

Only after the plan doc is committed:

```bash
gh issue edit <number> --add-label implementation
```

### Terminal State

- Bug issue has `bug` + `implementation` labels
- Design doc exists in `docs/plans/` (from Phase 1)
- Implementation plan exists in `docs/plans/` (from Phase 4)
- CI pipeline picks up the issue
- This skill does NOT execute the fix

---

## Key Principles

- **One question at a time** — don't overwhelm during brainstorming
- **Multiple choice preferred** — easier to answer than open-ended
- **YAGNI ruthlessly** — fix the bug, don't refactor the neighborhood
- **Hard gates are non-negotiable** — each phase must complete before the next begins
- **No implementation details in the bug issue** — file paths, function names, and package names go in the implementation sections only
- **Single issue, evolving labels** — `bug` first, `implementation` added only at the end
