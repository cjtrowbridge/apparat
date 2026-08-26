---
plan_id: 2026-08-25-12-00-55_research-contribution-leaderboard
title: Research Contribution Leaderboard
summary: Replace the Research placeholder contribution row with an explicit personal contribution summary and sortable mock friend-contribution leaderboard.
status: current
created_at: 2026-08-25-12-00-55
---

# Research Contribution Leaderboard

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Recover a safe implementation checkout before modifying product code.
  - [x] 1.1 Preserve the current local documentation edits and any intentionally retained local commits.
  - [x] 1.2 Reclone or otherwise reconcile this checkout with the rewritten public history without reintroducing the removed `artifacts/` history.
  - [x] 1.3 Verify the recovered checkout is based on the rewritten public `main` before creating a working branch.

- [ ] 2. Define an explicit mock Research contribution leaderboard view model in `internal/hud`.
  - [ ] 2.1 Keep each project's personal contribution distinct from friend contributions and mark all values as mock until the later Research/BOINC integration.
  - [ ] 2.2 Move the current contribution row to the end of each enrolled Research-project detail and rename it `Your Contribution` while retaining its displayed contribution value and unit.
  - [ ] 2.3 Represent friend name and numeric contribution separately from formatted text so contribution sorting never parses display strings.
  - [ ] 2.4 Define table columns for friend and contribution, default order by contribution descending, and deterministic friend-name ascending tie-breaking.
  - [ ] 2.5 Support header-selected sort state for friend name ascending/descending and contribution ascending/descending; the first rendered state is contribution descending.

- [ ] 3. Render and interact with the mock leaderboard in the Research GUI detail.
  - [ ] 3.1 Add a bounded, touch/controller-accessible table layout with visible clickable column headers and an honest mock/placeholder description.
  - [ ] 3.2 Render `Your Contribution` after the project metadata and before the friend table, never as a friend-table entry.
  - [ ] 3.3 Route header activation through the existing GUI action/state boundary and rebuild only the affected Research detail projection.
  - [ ] 3.4 Preserve narrow-surface wrapping, scrolling, focus, and minimum touch-target behavior without changing other tab renderers.

- [ ] 4. Add focused evidence and documentation.
  - [ ] 4.1 Add HUD tests for personal-row placement/wording, default descending contribution order, both header sort directions, and stable tie behavior.
  - [ ] 4.2 Add GUI tests that locate both headers, activate them, and verify the visible order changes while the personal contribution remains outside the table.
  - [ ] 4.3 Update `internal/hud/README.md` and the closest user-facing Research documentation to describe the mock leaderboard and its non-authoritative status.
  - [ ] 4.4 Run focused HUD and GUI tests, the relevant full Go test/build checks, documentation validation, code-size validation, and `git diff --check`.
  - [x] 4.5 Correct pre-existing GUI tests that still assume the retired sixth Tasks tab, while preserving tab-strip interaction coverage and Settings-control assertions.
