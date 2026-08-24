# Playbook: Daily Journal and TODO Snapshot

*Status: Stable*

## Objective

Discover today's journal and root TODO state, obtain approval before writes, preserve user-owned text, and capture an implementation checkpoint linked to the active plan.

## Prerequisites

- Read `README.md` and `AGENTS.md`.
- Load `templates/daily_journal_entry.md` only if today's journal is missing.

## Procedure

1. Resolve the local date and inspect `journal/YYYY-MM-DD.md`, root `TODO.md`, and `plans/current/index.md` without writing.
2. Report what exists, what is missing, and any proposed journal creation or update before editing.
3. If approved, create a missing journal from the template. Copy user-only intentions or reflections verbatim; otherwise leave `-`.
4. Record relevant TODO lines exactly. Change a TODO item only when the user directly requested its execution, following root `AGENTS.md` state rules.
5. Before non-trivial implementation, identify or create the approved active plan through `how_to_create_and_maintain_task_execution_plans.md`.
6. After work, record plan checklist deltas, verification evidence, changed files, and the exact next action in the agent-managed journal fields.
7. Review the resulting diff and use the applicable host commit playbook. Never push without explicit user approval.

## Verification

- User-owned journal fields contain verbatim user text or `-`.
- TODO text remains unchanged unless the user explicitly requested rewriting it.
- Non-trivial implementation links to an approved active plan.
- Journal evidence and plan checklist states match the repository diff.
