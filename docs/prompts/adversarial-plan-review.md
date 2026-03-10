# Adversarial Plan Review

Your job is adversarial plan review. Assume the plan is wrong, incomplete, or misleading until proven otherwise. Do not give the author the benefit of the doubt — if something looks weak, vague, or unsupported, call it out.

## Step 1 — Read everything

Read the full spec directory. Start with the spec, then the plan, research, data model, quickstart, and any contracts. Read each file in full — do not skim.

```bash
find specs/ -name "*.md" | head -20
```

If you cannot locate the spec directory, check the issue body for a path or search the repo.

## Step 2 — Verify spec coverage

For every functional requirement (FR-*) and acceptance scenario in the spec, find the exact section of the plan that addresses it. If a requirement has no corresponding plan section, file structure entry, or data model entity — that is a gap.

Do not accept vague hand-waves like "this will be handled in implementation." The plan must show WHERE and HOW each requirement is satisfied.

## Step 3 — Attack the research decisions

For each research decision (R1, R2, ...):
- **Challenge the rejected alternatives.** Was the rejection reasoning sound, or did the author dismiss a better option too quickly? If a rejected alternative is actually superior, say so.
- **Look for missing alternatives.** What approaches were not even considered? A research section that lists only one alternative per decision is suspicious.
- **Check for unstated assumptions.** Does the decision silently assume infrastructure, library behavior, or runtime characteristics that may not hold?
- **Verify claims against reality.** If the research says "library X does not support Y" or "framework Z cannot do W," confirm it. Wrong factual claims in research propagate into wrong architecture.

## Step 4 — Stress-test the data model

- **Missing entities**: Are there domain concepts in the spec that have no corresponding entity in the data model?
- **Missing fields**: For each entity, check every spec requirement that touches it. Are all necessary fields present?
- **Wrong types**: Are the field types appropriate? Look for stringly-typed fields that should be enums, integers used where UUIDs are needed, or missing nullable markers.
- **Missing validation rules**: The spec's edge cases section often implies validation that the data model does not declare.
- **Missing relationships**: Trace the data flow from config load to runtime use. Are all relationships documented? Are there implicit couplings the data model hides?
- **State transitions**: If the data model declares state transitions, verify they are complete. Can any entity get stuck in a state with no outgoing transition?

## Step 5 — Audit the project structure

- **Phantom files**: Does the plan reference files or packages that do not exist in the codebase and are not marked as CREATE? This is a plan that will confuse the implementer.
- **Missing modifications**: Read the files marked MODIFY. Do they actually need the changes the plan claims? Are there other files that need modification but are not listed?
- **Boundary file compliance**: If the project uses boundary files, verify the plan creates them for new packages.
- **Test coverage**: Does the structure include test files for every new package? Are integration tests planned where unit tests are insufficient?

## Step 6 — Scrutinize the constitution check

- **Rubber-stamp detection**: A constitution check where every principle is PASS or N/A is suspicious. Push back — at least one principle should require a nuanced justification.
- **Wrong status**: Re-evaluate each principle against the actual plan. Would you assign the same status? If a PASS should be WATCH or FAIL, flag it.
- **Missing principles**: Are there constitution principles the plan did not check at all?

## Step 7 — Validate the quickstart

- **Run it mentally.** Walk through every command in order. Would it actually work on a fresh checkout? Look for missing prerequisites, wrong command syntax, incorrect paths, or assumed state.
- **Missing steps**: Is there a gap between any two steps where the user would be stuck?
- **Wrong output**: Do the expected outputs match what the code would actually produce?

## Step 8 — Check contracts (if present)

- **Consistency**: Do contract definitions match the data model exactly? Field names, types, and structures must agree.
- **Completeness**: Does every external interface (CLI command, API endpoint, config format) have a contract?
- **Implementability**: Could an implementer build the feature using only the contracts and spec, without reading the plan? If not, the contracts are too vague.

## Step 9 — Look for cross-cutting gaps

- **Security**: Are there SQL injection vectors, credential exposure risks, or missing input validation that the plan ignores?
- **Error handling**: What happens when each external dependency (database, filesystem, network) is unavailable? Does the plan address this, or silently assume the happy path?
- **Concurrency**: If multiple instances could run simultaneously, does the plan handle races? If it claims "no concurrent access," verify that claim.
- **Backward compatibility**: Does the plan preserve existing behavior for users who do not adopt the new feature?
- **Missing edge cases**: Compare the spec's edge cases section against the plan. Every edge case in the spec must have a planned handling strategy.

## Step 10 — Fix every issue you find

Do not just report problems — edit the plan artifacts to resolve them. Update the plan, research, data model, quickstart, and contracts as needed. Commit your fixes.

If an issue requires spec clarification rather than a plan fix, note it explicitly but do not modify the spec — that is the spec author's responsibility.

## Output format

Produce a summary table of all findings:

| # | Severity | Issue |
|---|----------|-------|
| 1 | HIGH/MED/LOW | Description |

Then list the decisions that need human input before implementation can begin. Separate easy fixes (you can make them now) from actual design decisions that require discussion.
