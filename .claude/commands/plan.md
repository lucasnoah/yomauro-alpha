---
description: Plan the implementation of a feature issue. Surfaces key architectural decisions with tradeoffs before drawing contracts and creating implementation sub-issues.
---

## User Input

Feature issue number: `$ARGUMENTS`

Fetch the issue: `gh issue view $ARGUMENTS --json title,body,number`

---

## Phase 1 — Context

1. Read the feature issue in full.
2. Explore codebase patterns relevant to this feature. At minimum read:
   - Existing migrations (`migrations/`) if the feature adds schema changes
   - Existing query patterns for upsert and query conventions
   - Any package this feature will extend
3. Identify **3–5 key architectural decisions** this feature requires.

   **A decision IS key if it:**
   - Shapes the database schema (column types, boolean vs enum, new table vs extending existing)
   - Determines which packages are created or modified
   - Changes the API surface
   - Has multiple reasonable approaches with meaningfully different tradeoffs
   - Is difficult to reverse after implementation begins

   **A decision is NOT key if:**
   - There is a clear existing convention in the codebase
   - It is an internal detail (variable names, log levels, timeout values)
   - It can be trivially changed later without a migration or API break

---

## Phase 2 — Decision Points

> **HARD GATE: Do NOT proceed to Phase 3 until every decision is approved.**

For each key decision, present this format exactly:

---

### Decision N: [short name]

| Option | Approach | Tradeoffs |
|---|---|---|
| A | [description] | [pros and cons] |
| B | [description] | [pros and cons] |
| C | [description] | [pros and cons] |

**Recommendation:** Option [X] — [one sentence rationale tied to this codebase's conventions and constraints].

---

**Present one decision at a time. Wait for the user to choose before presenting the next.**

After the user responds, confirm the locked decision ("Locked: Option [X]") and immediately present the next decision. Do not present multiple decisions in the same message.

---

## Phase 3 — Contracts

Using the locked decisions, write precise boundary artifacts. These feed directly into implementation issue Data Contracts sections.

For each data boundary this feature introduces or modifies:

**Schema** (new or modified tables)

| Column | Type | Constraints |
|---|---|---|

Note unique keys, indexes, and foreign keys below the table.

**sqlc Query Signatures** (every query the implementation will need)

```go
FunctionName(ctx context.Context, arg ParamType) (ReturnType, error)
```

**Go Types** (types that cross package boundaries — not sqlc-generated types)

```go
type TypeName struct {
    Field Type
}
```

**HTTP Shapes** (new or modified API endpoints only)

```
METHOD /api/v1/path
Request:  { "field": type }
Response: { "field": type }
```

Present the full contracts to the user and confirm before Phase 4.

---

## Phase 4 — Implementation Issues

Decompose the feature into logical implementation slices. A slice is the smallest unit that:
- Can be merged independently without breaking the codebase
- Has clear dependencies on other slices (if any)
- Can be fully specified from the contracts already drawn

For each slice, create a GitHub issue with `gh issue create`. Apply the `implementation` label.

> **HARD GATE: Every issue must contain all eight sections below in order. Do not publish any issue with a missing or malformed section.**

---

### Section 1 — Parent

```
**Feature:** #[N] — [title]

Read the parent before starting. It contains the user intent, user stories,
and external acceptance criteria this issue contributes to.
```

### Section 2 — Intent

One paragraph. Must answer both: (a) what this slice builds, and (b) what becomes possible after it lands that wasn't possible before.

### Section 3 — Technical Scope

Package-level only. No file names.

| Package | Action | Purpose |
|---|---|---|

### Section 4 — Data Contracts

Copy the relevant contracts from Phase 3 verbatim. Use only the subsections that apply (Schema, sqlc Query Signatures, Go Types, HTTP Shapes). Omit subsections that don't apply. If this slice introduces no data boundaries, write "N/A — this slice contains no data boundaries."

> Do NOT embed schema or Go type definitions inside Approach. Data contracts belong here, not there.

### Section 5 — Approach

Prose explaining the chosen strategy and WHY. Reference the Phase 2 decision that governs this slice. One to three paragraphs. No code blocks — those belong in Data Contracts.

### Section 6 — Acceptance Criteria

Checkboxes only. Developer-verifiable without a browser.

- [ ] ...

> Do NOT use prose bullets. Every criterion must be a `- [ ]` checkbox.

### Section 7 — Dependencies

- **Blocked by:** #NNN (reason), or N/A
- **Blocks:** #NNN (reason), or N/A

### Section 8 — Out of Scope

Explicit list of what this slice does NOT touch. If another slice handles the excluded concern, name it. Do not write "N/A" unless the feature is a single slice.

---

## Common mistakes — check before publishing each issue

- [ ] Parent section includes the full reading instruction (not just "Part of #N")
- [ ] Intent answers both "what" and "what becomes possible"
- [ ] Data Contracts is a dedicated section — no schema or types buried in Approach
- [ ] Approach is prose rationale — no code blocks
- [ ] Acceptance criteria are all `- [ ]` checkboxes
- [ ] Dependencies use the exact "Blocked by / Blocks" format
- [ ] Out of Scope is present and names excluded concerns explicitly
- [ ] Column names and package names match the Phase 3 contracts exactly
