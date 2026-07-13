---
plan_id: 2026-07-12-21-33-14_add-scrollable-hud-scenario-data
title: Add Scrollable HUD Scenario Data
summary: Populate each HUD tab with realistic multi-page scenario data and make every body, master-list, and detail viewport scroll only within the area between the tab strip and diagnostics bar.
status: current
created_at: 2026-07-12-21-33-14
---

# Add Scrollable HUD Scenario Data

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 Basic HUD Tabs And Content and Phase 5 Android GUI parity/safe-area validation.
- Product contract: the canonical seven tabs remain navigable placeholders backed by truthful mock data; content must scroll inside the body viewport, not under the top tab list or bottom diagnostics/status area.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Related active plans: `plans/current/2026-07-11-09-33-26_fix-mobile-overflow-and-tab-scroll.md` and `plans/current/2026-07-12-21-07-37_resolve-tab-drag-selection-race.md`.
- Confirmed diagnosis: EbitenUI `ScrollContainer` owns clipping and renders according to `ScrollTop`, but does not autonomously change `ScrollTop` for wheel, pointer-drag, or touch-drag input. Apparat presently customizes only horizontal tab-strip `ScrollLeft`; Settings, master lists, and details never receive vertical scroll input.

## Checklist

- [x] 1. Define realistic, clearly mock scenario data that exceeds each viewport.
  - [x] 1.1 Extend every canonical tab with enough labeled sections and rows to produce several phone-height pages without claiming that placeholder state is durable or live.
  - [x] 1.2 Model Comrades as conversations, trust posture, queue grants, quotas, and audit events; Projects as workspaces, chats, files, artifacts, Git activity, and drafts.
  - [x] 1.3 Model Research as candidate/validated projects, budgets, schedules, work-unit progress, and validation evidence; Cluster as device inventory, capabilities, health, and recent activity.
  - [x] 1.4 Model Routing as workload profiles, compatible destinations, queue state, fallbacks, and health; Tasks as schedules, webhooks, approvals, runs, and failures; Settings as configuration, bindings, diagnostics, logs, storage, and recovery.
  - [x] 1.5 Keep every generated label/detail visibly mock, planned, disabled, or future where no real backend exists.

- [x] 2. Give every vertical HUD viewport explicit scrolling behavior.
  - [x] 2.1 Retain references to the Settings body, collapsed master list, expanded master list, collapsed detail, and expanded detail scroll containers as they are rebuilt.
  - [x] 2.2 Route mouse wheel input over the innermost visible vertical viewport to bounded `ScrollTop` changes.
  - [x] 2.3 Route mouse/touch vertical drag gestures to the same viewport while preserving button taps, tab-strip horizontal drag behavior, divider dragging, and floating PTT behavior.
  - [x] 2.4 Apply a drag threshold and gesture ownership so vertical scrolling does not activate rows/buttons on release.
  - [x] 2.5 Keep horizontal tab-strip scrolling separate from vertical body scrolling.

- [x] 3. Enforce the body viewport geometry contract.
  - [x] 3.1 Preserve the top tab-strip row as a fixed-height shell child and the bottom diagnostics bar as root anchor clearance.
  - [x] 3.2 Verify Settings and every master-detail list/detail scroll container receives only the remaining body rectangle after tab height, body gap, margins, and diagnostics clearance.
  - [x] 3.3 Ensure content is clipped to its scroll-container viewport and cannot render under the tabs or status bar.
  - [x] 3.4 Preserve phone-width preferred-width bounds and wrapped body text while adding taller mock content.

- [x] 4. Add regression coverage.
  - [x] 4.1 Test that every tab contains enough scenario data to exceed a narrow body viewport.
  - [x] 4.2 Test vertical wheel and drag calculations change `ScrollTop` only for the targeted body/list/detail viewport.
  - [x] 4.3 Test drag cancellation prevents a row/button activation after a vertical scroll gesture.
  - [x] 4.4 Test tab-strip horizontal scrolling remains independent of body vertical scrolling.
  - [x] 4.5 Test each body viewport rectangle remains below the tab strip and above diagnostics at desktop and phone-sized dimensions.

- [?] 5. Document and validate the behavior.
  - [x] 5.1 Update `internal/hud/README.md` with the scope and truthfulness rules for scenario data.
  - [x] 5.2 Update `internal/adapters/gui/README.md` and the HUD playbook with vertical scroll ownership and viewport-boundary rules.
  - [?] 5.3 Run focused GUI compile/tests, repository tests, code-size, docs, and diff checks.
  - [?] 5.4 Rebuild/install Android when its toolchain and ADB are available.
  - [ ] 5.5 Capture desktop and phone screenshots showing several pages of content scrolling inside the body rectangle without overlapping tabs or diagnostics.

- [ ] 6. Publish the verified checkpoint.
  - [ ] 6.1 Update plan status and related-plan validation items to match verified behavior.
  - [ ] 6.2 Regenerate and validate plan indexes.
  - [ ] 6.3 Append implementation/validation evidence to the current-day journal.
  - [ ] 6.4 Confirm no files beneath `third_party/salvagecore/` are staged.
  - [ ] 6.5 Review pending downtime reports before final summary.
  - [ ] 6.6 Commit and push after the user-approved checkpoint summary.

## Scope Boundaries

- Scenario data is a presentation/scrolling test fixture, not a persistence, network, queue, or automation implementation.
- Do not expand the body viewport to the full screen; the top tab strip, window margins, body gap, and diagnostics clearance remain reserved.
- Do not alter the user-authored `TODO.md` lines until visual validation proves the corresponding behavior.

## Execution Notes

- 2026-07-12: Added six fictional scenario sections to every canonical tab, keeping all entries visibly mock/future. The resulting tab models contain at least eight sections and 24 rows each for multi-page scrolling evidence.
- 2026-07-12: Registered Settings, master-list, and detail scroll containers; added wheel, mouse-drag, and touch-drag `ScrollTop` routing with release-target suppression; preserved the tab strip's separate horizontal path; and verified body viewport rectangles after a real EbitenUI layout pass at 360px and 1280px widths.
- 2026-07-12: Focused GUI tests, ordinary repository tests, code-size, and diff checks passed. `check_directory_docs.py` remains blocked by pre-existing missing `--help` output in `scripts/build.py` and `scripts/build_orchestrator.py`. Android rebuild/install and phone screenshots remain blocked because this host lacks the configured SDK/NDK tools and ADB.
