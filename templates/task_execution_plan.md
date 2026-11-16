---
plan_id: YYYY-MM-DD-HH-mm-ss_slug
title: Plan Title
summary: One-sentence summary of what this plan accomplishes.
status: future
created_at: YYYY-MM-DD-HH-mm-ss
---

# Plan Title

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

State the observable result, governing roadmap items, protected behavior, execution authority and excluded scope. Describe the simplest path through existing code and material uncertainty. A future plan remains queued until implementation begins under the applicable authorization.

## Work and acceptance

- [ ] 1. Deliver a coherent outcome through the existing owners and boundaries.
- [ ] 2. Verify the outcome, relevant failure/recovery behavior and any selected improvement under comparable conditions.

Add substeps only for necessary dependencies, uncertainty or independently verifiable outcomes. Reference expected files/owners and concrete verification commands or methods; checklist count does not prescribe architecture.

## Evidence and stopping condition

Link reusable evidence; distinguish existing completion from pending revalidation. For selected improvements record baseline, expected benefit, protected behavior, comparison method and retain/revert disposition. State acceptance, relevant budgets, known gaps and when to stop. Keep the roadmap and plan consistent; preserve append-only journal history and refresh plan indexes.

## Authoring Rules

- Use filename format: `YYYY-MM-DD-HH-mm-ss_slug.md`.
- Use lowercase, hyphenated slugs (letters, numbers, hyphens only).
- Include required front matter keys only: `plan_id`, `title`, `summary`, `status`, `created_at`.
- Do not add modification-time fields to YAML front matter.
- Decompose only until the next implementation and verification step is clear; do not require recursive task trees or interfaces per checklist item.
- Use `[-]` for intentionally de-scoped/closed items that are not completed implementation.
- Keep plan location aligned to status:
  - `plans/future/`: queued work
  - `plans/current/`: actively executed work
  - `plans/past/`: archived work with no planned follow-up
