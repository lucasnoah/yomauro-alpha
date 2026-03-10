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
