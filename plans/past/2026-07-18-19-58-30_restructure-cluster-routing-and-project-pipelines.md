---
plan_id: 2026-07-18-19-58-30_restructure-cluster-routing-and-project-pipelines
title: Restructure Cluster Routing And Project Pipelines
summary: Move the mock Routing surface into Cluster detail content and add a mock Pipelines detail surface under Projects while preserving the left-selector/right-detail HUD pattern.
status: past
created_at: 2026-07-18-19-58-30
---

# Restructure Cluster Routing And Project Pipelines

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 Basic HUD Tabs And Content, specifically the Projects, Cluster, and Routing placeholder surfaces and the controller-first tab-shell contract.
- Product contract: reduce the top-level HUD to Comrades, Projects, Research, Cluster, Tasks, and Settings. Cluster owns the mock routing view as a left-selector item; Projects owns a left-selector `Pipelines` item. Both use the existing left-selector/right-detail pattern, with narrow layouts presenting the selector first and a Back button in detail.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Source inbox: the two final TODO.md items, which retain their exact text and are now in progress.

## Checklist

- [x] 1. Establish the governed HUD restructuring checkpoint.
  - [x] 1.1 Mark both selected TODO items `[-]` in progress without altering their text.
  - [x] 1.2 Record the approved left-selector/right-detail interpretation and roadmap/documentation impact.
  - [x] 1.3 Promote this plan to current and regenerate its status index before source edits.

- [x] 2. Move Routing into Cluster without losing its mock detail content.
  - [x] 2.1 Remove Routing from the top-level tab descriptors, default tab list, direct-tab expectations, and obsolete routing-only configuration fields.
  - [x] 2.2 Extend the HUD detail model so one selectable Cluster item can render the former Routing tab's grouped workload-class, queue, fallback, and scenario sections in its right detail pane.
  - [x] 2.3 Add a Cluster `Routing` selector item and retain the former Routing content's future/mock truthfulness markers.
  - [x] 2.4 Make expanded and collapsed master-detail views render only the currently selected left-pane item in the right detail pane; preserve touch targets, wrapping, bounded width, and vertical scrolling for the grouped routing detail.

- [x] 3. Add the Projects Pipelines detail surface.
  - [x] 3.1 Add a left-selector `Pipelines` item to Projects.
  - [x] 3.2 Model pipeline-building stages, triggers, inputs, typed steps, approvals, routing, and run records as clearly mock/future detail content.
  - [x] 3.3 Render the grouped pipeline content in the existing right detail pane and preserve narrow-screen Back navigation and vertical scrolling.

- [x] 4. Align contracts, tests, and visual validation.
  - [x] 4.1 Update root product/input documentation, roadmap language, controller map, and local HUD/GUI READMEs from seven top-level tabs/Alt+1..Alt+7 to six top-level tabs/Alt+1..Alt+6; retain Routing as Cluster detail content.
  - [x] 4.2 Update focused HUD and GUI tests for the new tab order, direct selection, nested/grouped detail content, Pipelines visibility, and master-detail behavior.
  - [x] 4.3 Run focused HUD/GUI tests, repository tests, documentation and code-size checks, and `git diff --check`.
  - [x] 4.4 Rebuild/install the Android APK and capture narrow-phone visual evidence for Cluster Routing and Projects Pipelines detail selection. On the connected Pixel 10 Pro XL, Cluster → Routing showed the grouped workload, queue, profile, compatibility, and queue-state detail; Projects → Pipelines showed the grouped builder, safety/routing, and run-history detail. The prior updater-signing incompatibility remained separate and no unknown-source permission was changed.

- [x] 5. Publish the verified checkpoint.
  - [x] 5.1 Update plan status/checklist evidence, regenerate indexes, append the journal, confirm no `third_party/salvagecore` files are staged, and review pending downtime reports.
  - [x] 5.2 Commit and push directly to `main` after the user-approved checkpoint summary; do not open a pull request.

## Approved Plan Revision

- 2026-07-18: Inspection found that expanded master-detail views currently render every section in the right pane, despite the documented selected-item contract. The user explicitly approved changing that shared behavior so the left selector controls the sole right detail pane at every viewport width. This is required for the requested Cluster Routing and Projects Pipelines selectors to be meaningful.
