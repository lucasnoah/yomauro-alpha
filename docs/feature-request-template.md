# Feature Request Issue Template

There are two issue templates in this project. Use the right one:

| Template | Label | Use when |
|---|---|---|
| **This file** — Feature Request | `feature` | Defining a new user-facing capability. No implementation details. |
| `implementation-issue-template.md` | `implementation` | Breaking a feature into shippable technical slices. Created by the planning agent. |

Feature issues are the source of truth for *what* the system should do. Implementation issues
are the source of truth for *how* a specific slice gets built. The two are linked via the
parent reference in each implementation issue.

---

When planning a new feature, the output is a GitHub issue that serves as the source of truth
for implementation, QA validation, and future reference. Every feature issue MUST follow this
template and carry the `feature` GitHub label. Do not skip sections — if a section doesn't
apply, write "N/A" with a brief reason.

## Principles

- **No code, only requirements.** The issue describes what the system should do, not how to
  build it. No file paths, function names, API routes, or implementation details.
- **Testable from the outside.** User stories must be verifiable by someone who has never
  seen the codebase — just a browser and the testmail inbox.
- **User intent first.** Every feature exists because a real person has a real problem. Start
  there. If you can't articulate the pain, you don't understand the feature yet.
- **Explicit scope boundaries.** What you're NOT building is as important as what you are.
  Unstated non-requirements become scope creep.

## Template

### User Intent

Describe the real-world situation this feature addresses. Who is the user? What are they
doing when they encounter this problem? What does their environment look like? What are
they juggling? Why is the current state painful?

Write this as a narrative, not bullet points. The reader should feel the friction. This
section grounds every subsequent decision — if a requirement doesn't trace back to the
intent, question whether it belongs.

### User Stories

Testable scenarios written from the user's perspective. Each story should be independently
verifiable through browser automation and maps directly to a workflow in
`docs/validation/registry.json`.

For each story, include:

- **Narrative**: "As a [role], I [action] and [expected outcome]." One sentence.
- **Preconditions**: What must exist before the story can be exercised (test data, system
  state). These become the `preconditions` array in the validation registry.
- **Assertions**: Observable outcomes the tester checks — what appears on screen, what
  emails arrive, what changes. These become the `assertions` array in the registry.

Cover the happy path first, then error/edge cases. Number the stories sequentially.

### Requirements

The functional spec. Organize into logical sub-sections (e.g., by component or capability).
Use tables for structured data like error messages. Be specific enough to implement from
but don't prescribe architecture.

### Affected Surfaces

A table of every user-facing touchpoint that changes. This tells the implementer the blast
radius and tells QA where to look.

| Surface | Change |
|---|---|
| [page, email, table, nav element, etc.] | [what changes about it] |

Include new pages, modified pages, email templates, database tables, and navigation changes.

### Non-Requirements

Explicitly state what is out of scope. This prevents scope creep during implementation and
sets expectations for reviewers. Be specific — "no X" is better than "keep it simple."

### Open Questions

Decisions deliberately left for the implementer or for a future conversation. These are
things the feature owner considered but chose not to lock down yet. Each question should
be answerable without revisiting the entire design.
