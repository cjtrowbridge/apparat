---
plan_id: 2026-07-18-21-00-43_move-tasks-into-cluster-selector-panel
title: Move Tasks Into Cluster Selector Panel
summary: Replace the Tasks top-level tab with a Cluster selector-panel item, add semantic selector headings, and canonize selector/content-panel terminology.
status: current
created_at: 2026-07-18-21-00-43
---

# Move Tasks Into Cluster Selector Panel

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 Basic HUD Tabs And Content and its HUD tab-shell/Tasks content expectations.
- Product contract: preserve a controller-first, responsive selector/content-panel model. Cluster owns devices, routing, and task automation context; top-level navigation contains Comrades, Projects, Research, Cluster, and Settings.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Source inbox: `TODO.md` item `move the tasks tab into the cluster tab as a new left-item that selects its own right-pane with the contents`.

## Checklist

- [x] 1. Establish the governed implementation checkpoint.
  - [x] 1.1 Mark the exact source TODO item in progress without changing its wording.
  - [x] 1.2 Create this focused active plan and bind it to Phase 4.
  - [x] 1.3 Regenerate the active-plan index.

- [x] 2. Make selector headings a first-class HUD presentation concept.
  - [x] 2.1 Extend the HUD selector data model to distinguish selectable items from non-selectable, non-focusable headings.
  - [x] 2.2 Render headings as semantic text in the selector panel, preserving touch-target sizing only for selectable items.
  - [x] 2.3 Keep selection indexing, collapsed navigation, and controller/keyboard activation restricted to selectable items.
  - [x] 2.4 Add focused HUD/GUI regression coverage for heading display and selector behavior.

- [x] 3. Move Tasks from top-level navigation into Cluster.
  - [x] 3.1 Remove Tasks from top-level descriptors, default tabs, direct-tab bindings, and tab-order expectations.
  - [x] 3.2 Add a Tasks selector item under Cluster that renders the existing task schedules, webhooks, approvals, controls, and scenario data in the content panel.
  - [x] 3.3 Group Cluster selectors with non-interactive headings that distinguish device context from operations.
  - [x] 3.4 Preserve independent Cluster selector state and narrow-screen Back behavior.

- [x] 4. Canonize selector/content-panel terminology.
  - [x] 4.1 Replace affected master-detail, left/right-column, and left/right-pane terminology in HUD and GUI documentation with `selector panel` and `content panel`.
  - [x] 4.2 Update README, ROADMAP, input-binding hints, and tests to state five top-level tabs and `Alt+1` through `Alt+5`.
  - [x] 4.3 Update local code names/comments where they expose layout terminology without needlessly renaming stable implementation internals.

- [ ] 5. Verify and publish the checkpoint.
  - [x] 5.1 Run focused HUD and GUI tests, the GUI compile gate, repository tests, documentation/code-size checks, and `git diff --check`.
  - [ ] 5.2 Rebuild/install Android when available and capture desktop and Android screenshots showing Cluster headings, Tasks selection, and the five-tab strip.
  - [ ] 5.3 Mark the source TODO complete only after visual validation passes.
  - [x] 5.4 Update this plan, regenerate indexes, append the current-day journal evidence, confirm no Salvagecore files are staged, and review pending downtime reports.
  - [x] 5.5 Commit and push directly to `main` after the user-approved checkpoint summary; do not open a pull request.

## Scope Boundaries

- Headings are presentation-only: they do not change durable domain state, receive focus, or dispatch actions.
- Do not change the existing tab-strip scrolling/gesture behavior.
- Do not add persistence, scheduling, webhooks, or workflow execution; this keeps existing task presentation mock/future-facing.

## Execution Notes

- 2026-07-18: Implemented a `SelectorKind` model with non-selectable `SelectorHeading` text, moved Tasks into Cluster under an Operations heading, and reduced top-level navigation/direct selection to five tabs. Focused HUD/GUI tests, `make test`, GUI compile, documentation, code-size, diff, and plan-index checks passed. Linux and Android release artifacts rebuilt successfully. The connected Pixel 10 Pro XL was absent; a Pixel Tablet is connected but has not been updated without explicit device authorization, so on-device screenshots and TODO completion remain pending.

## Approved Plan Revision: 2026-07-18 Selector Heading Descriptions

- [?] 6. Add optional descriptive text beneath selector-heading titles.
  - [x] 6.1 Model a heading description as presentation-only data and populate concise descriptions for Cluster's Devices and Operations groups.
  - [x] 6.2 Render each description as small muted, wrapped text beneath its title without creating a focusable or clickable control.
  - [x] 6.3 Add focused HUD/GUI coverage that proves descriptions render and selector headings remain non-selectable.
  - [?] 6.4 Update local HUD/GUI documentation, run applicable checks, rebuild the APK, and update the current plan/journal evidence.

## Revision Execution Notes

- 2026-07-18: Focused HUD/GUI tests, `make test`, GUI compile, documentation, and code-size checks passed. `git diff --check` reports trailing whitespace only in concurrently added user-authored `TODO.md` lines; those immutable task lines were deliberately not changed. APK rebuild/install and visual evidence remain pending.
