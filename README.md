# sharedtaint

Shared tooling, conventions, and patterns extracted from production repos.

## Dependencies

| Dependency | Required By | Install |
|------------|-------------|---------|
| [spec-kit](https://github.com/github/spec-kit) | `feature-request` skill | Follow spec-kit's setup instructions to create `.specify/` infrastructure |

## Contents

### Claude Code Skills

#### Commands (`.claude/commands/`)

| Skill | Description |
|-------|-------------|
| `bug` | End-to-end bug-fix workflow: brainstorm, file GitHub issue, engineer the fix, produce implementation plan |
| `brainstorming` | Collaborative design exploration before any creative/implementation work. Returns control to caller when design is complete — does NOT auto-chain into downstream skills. |
| `writing-plans` | Write bite-sized TDD implementation plans from specs or requirements |
| `plan` | Surface architectural decisions, draw contracts, create implementation sub-issues |

#### Skills (`.claude/skills/`)

| Skill | Description |
|-------|-------------|
| `feature-request` | Full pipeline: brainstorm -> specify -> plan -> adversarial review -> tasks -> slices -> GitHub issues. **Requires spec-kit.** |

### Templates (`docs/`)

| Template | Description |
|----------|-------------|
| `implementation-issue-template.md` | Standard format for implementation sub-issues with data contracts, acceptance criteria, and scope boundaries |
| `feature-request-template.md` | Standard format for feature issues: user intent, stories, requirements, affected surfaces |

### Prompts (`docs/prompts/`)

| Prompt | Description |
|--------|-------------|
| `adversarial-plan-review.md` | 10-step adversarial review protocol for plan artifacts — used by `feature-request` Phase 3 |

## Workflow Overview

```
brainstorming → design doc (returns control to caller)

feature-request (full pipeline, uses spec-kit)
  brainstorm → specify → plan → adversarial review → tasks → slices → issues

plan (lightweight decomposition)
  feature issue → decisions → contracts → implementation issues

bug (standalone pipeline)
  brainstorm → issue → implementation engineering → writing-plans → finalize
```
