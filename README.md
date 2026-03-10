# yomauro

Go + Next.js starter template with session-based auth, login page, sidebar shell, and dashboard.

## Automated Pipeline (TaintFactory)

This repo uses [TaintFactory](https://taintfactory.itslucastime.com) to automatically implement GitHub issues.

### How it works

1. Create a GitHub issue describing the work
2. Add the `implementation` label to the issue
3. TaintFactory polls for labeled issues and picks them up automatically
4. Each issue goes through a multi-stage pipeline:
   - **implement** — a Claude agent writes the code in an isolated worktree
   - **review** — a second agent reviews the changes (code-only context)
   - **qa** — a third agent runs full QA (code + issue context)
   - **verify** — automated checks run (go vet, lint, tests)
   - **merge** — squash-merges the PR to main
5. If any stage fails checks, the pipeline loops back to `implement` for fixes

### Dashboard

Monitor pipeline progress at:

- **URL:** https://taintfactory.itslucastime.com
- **Username:** `yomauro`
- **Password:** `taintfactory2026`

### Labels

| Label | Purpose |
|-------|---------|
| `implementation` | Triggers the automated pipeline |
| `feature` | Feature request (human-written, not auto-processed) |

## Getting Started with Claude Code

This repo includes two Claude Code skills that turn natural language into fully planned, issue-tracked work. Before using them, set up your project constitution.

### First: Set Up Your Constitution

Run the following skill to define your project's core principles, constraints, and governance rules. The constitution is referenced during planning and adversarial review to catch violations early.

```
/speckit.constitution
```

This creates `.specify/memory/constitution.md` with your project-specific principles (e.g., test-first, simplicity, security requirements). Every plan generated after this will be checked against your constitution.

### /feature-request

End-to-end feature planning pipeline. Takes a natural language description and produces a fully specified, reviewed, and issue-tracked feature — without writing any code.

```
/feature-request I want to add user profiles with avatar uploads and a settings page
```

**What it does (6 phases):**

1. **Brainstorm** — Parallel explore agents scan the codebase, then a collaborative design session asks clarifying questions and proposes approaches. Produces an approved design doc.
2. **Specify** — Creates a formal feature spec (`specs/<feature>/spec.md`) with user stories, functional requirements, acceptance scenarios, and success criteria.
3. **Plan** — Generates implementation plan, research decisions, data model, API/type contracts, and test scenarios.
4. **Adversarial Review** — 3 parallel review agents attack the plan from different angles (spec coverage, security, data model consistency). Issues are fixed directly in the artifacts.
5. **Tasks** — Generates a dependency-ordered task list organized by user story.
6. **GitHub Issues** — Creates a parent feature issue and implementation sub-issues with precise data contracts, acceptance criteria, and dependency chains.

**Terminal state:** Feature issue (label: `feature`) + implementation issues (label: `implementation`) + all plan artifacts in `specs/<feature>/`. No code is written — implementation is a separate step.

### /bug

End-to-end bug-fix workflow. Same pipeline as `/feature-request` but adapted for diagnosing and fixing bugs.

```
/bug The login form shows a blank screen on Safari when submitting with autofill
```

**What it does:**

1. **Brainstorm** — Parallel investigation agents trace symptoms, check recent changes, and search for related issues.
2. **Specify + Plan** — Creates a bug spec and targeted fix plan.
3. **Adversarial Review** — Validates the fix won't introduce regressions.
4. **Tasks + Issues** — Produces implementation issues (typically a single slice for most bugs).

**Terminal state:** Bug issue (label: `bug`) + implementation issues (label: `implementation`) + plan artifacts.

### Other Useful Skills

| Skill | When to use |
|-------|-------------|
| `/brainstorming` | Lightweight design exploration without the full pipeline |
| `/speckit.implement` | Execute an existing implementation plan |
| `/speckit.clarify` | Ask targeted questions about underspecified areas in a spec |
| `/speckit.analyze` | Cross-artifact consistency check across spec, plan, and tasks |

## Development

```bash
make db            # Start PostgreSQL (docker-compose)
make migrate-up    # Run migrations
make dev-api       # Start Go API (port 8080)
cd web && npm run dev  # Start Next.js (port 3000)
make test          # Run all tests
```

## Stack

- **Backend:** Go 1.25+, chi v5, pgx v5, sqlc
- **Frontend:** Next.js 14 (App Router), Tailwind CSS 3.4, TypeScript 5
- **Database:** PostgreSQL
- **Testing:** go test (backend), vitest (frontend)
