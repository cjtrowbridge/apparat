---
plan_id: 2026-07-08-08-49-21_execute-phase-4-basic-hud-tabs
title: Execute Phase 4 Basic HUD Tabs
summary: Implement the Phase 4 data-driven HUD tab system, temporary configuration manager, basic tab content, verification, and documentation before backend networking work.
status: past
created_at: 2026-07-08-08-49-21
---

# Execute Phase 4 Basic HUD Tabs

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4, `Basic HUD Tabs And Content`.
- Product goal: make the local seven-tab HUD usable before moving on to backend networking, queues, services, and transport work.
- Scope boundary: no real backend project mutation, no real comrade sharing, no BOINC execution, no task execution, no network API, and no persistent user-editable settings in SQLite during this phase.
- Persistence boundary: the HUD configuration manager uses hard-coded defaults now, with types and APIs shaped for future SQLite-backed user configuration tables.
- GUI framework boundary: use Ebitengine for the window/game/input/render loop and integrate EbitenUI for suitable widgets/layout primitives; document any custom widget built because EbitenUI does not fit controller-first behavior.

## Implementation Checklist

- [x] 1. Audit current HUD and GUI structure.
  - [x] 1.1 Inspect `internal/hud`, `internal/input`, `internal/adapters/gui`, and current tests.
  - [x] 1.2 Identify existing tab, snapshot, focus, and voice-state types that can be reused.
  - [x] 1.3 Identify any files approaching the 400-line limit before implementation.
  - [x] 1.4 Confirm applicable README files exist for every package that will be touched.

- [x] 2. Add the temporary HUD configuration manager.
  - [x] 2.1 Create or extend a HUD configuration package boundary under `internal/hud` or a clearly documented adjacent package.
  - [x] 2.2 Define a `ConfigManager` or equivalent interface that returns immutable/default HUD configuration values.
  - [x] 2.3 Define configuration value structs for tab view, bindings, display/accessibility, interaction, notifications, diagnostics, default views, and privacy/safety.
  - [x] 2.4 Implement a hard-coded default provider for Phase 4.
  - [x] 2.5 Keep field names and value types compatible with future SQLite serialization.
  - [x] 2.6 Add tests proving dark theme, top tab placement, canonical tab order, default scale, default font size, default bindings, and no-sharing defaults.
  - [x] 2.7 Document the temporary manager and future SQLite persistence boundary in the relevant README.

- [x] 3. Implement the data-driven tab-view model.
  - [x] 3.1 Define stable tab IDs for Comrades, Projects, Research, Cluster, Routing, Tasks, and Settings.
  - [x] 3.2 Define tab descriptors with ID, label, icon/glyph slot, accessibility label, visibility, enabled state, placement metadata, and badge/status placeholder fields.
  - [x] 3.3 Define tab placement values for top, left side rail, right side rail, and compact/sidebar-responsive future use.
  - [x] 3.4 Default the active placement to top.
  - [x] 3.5 Keep selected-tab state independent from visual placement.
  - [x] 3.6 Add tests for canonical order, stable IDs, default selected tab, visibility, labels, accessibility labels, and placement independence.

- [x] 4. Implement named input actions and default bindings.
  - [x] 4.1 Define named input actions for previous tab, next tab, direct tab selection, push-to-talk, cancel recording, focus movement, activation, back, context menu, command palette, scroll, and collection navigation.
  - [x] 4.2 Define default bindings for `L1`, `R1`, `R2`, right `Ctrl`, `Ctrl+PageUp`, `Ctrl+PageDown`, `Alt+1` through `Alt+7`, `Escape`, and existing focus/activation controls.
  - [x] 4.3 Route tab switching and push-to-talk through named actions rather than scattered literal key checks.
  - [x] 4.4 Preserve current Steam Deck/controller and Debian/Linux keyboard behavior.
  - [x] 4.5 Add deterministic tests for action lookup, tab switching, direct tab selection, push-to-talk, cancel recording, and unknown binding behavior.
  - [x] 4.6 Document the future user-editable key-binding contract.

- [x] 5. Build reusable HUD layout primitives.
  - [x] 5.1 Add view-model types for top-level shell, tab navigation, panels, lists, cards, empty states, status pills, action rows, and detail panes.
  - [x] 5.2 Implement top tab bar rendering from the tab-view model.
  - [x] 5.3 Keep the tab bar implementation structured so side-rail rendering can use the same tab descriptors later.
  - [x] 5.4 Integrate EbitenUI widgets for suitable panels, buttons, lists, forms, tab bars/rails, or focusable controls.
  - [x] 5.5 Add custom Ebitengine-rendered widgets only where EbitenUI does not satisfy controller-first/focus requirements.
  - [x] 5.6 Add visible focus styling suitable for Steam Deck scale.
  - [x] 5.7 Add loading, offline, warning, disabled, and future-placeholder states.
  - [x] 5.8 Keep rendering driven by HUD view models rather than direct database, network, or adapter calls.
  - [x] 5.9 Add tests for primitive view-model construction and disabled/future state flags.
  - [x] 5.10 Document which UI pieces use EbitenUI and which are Apparat custom widgets.

- [x] 6. Implement basic content for every tab.
  - [x] 6.1 Implement Comrades placeholder content.
    - [x] 6.1.1 Explain future real-friend chat.
    - [x] 6.1.2 Explain comrade queues for low-priority shared inference access.
    - [x] 6.1.3 Show placeholder sharing grants, queue access, quota, revocation, and audit concepts.
    - [x] 6.1.4 Mark backend-dependent controls disabled or future.
  - [x] 6.2 Implement Projects basic content.
    - [x] 6.2.1 Show project list and selected project summary.
    - [x] 6.2.2 Show chat preview, file tree placeholder, artifact list placeholder, and Git status placeholder.
    - [x] 6.2.3 Add mock/local-only selection and inspection actions.
    - [x] 6.2.4 Show offline draft and transaction concepts without real file changes.
  - [x] 6.3 Implement Research placeholder content.
    - [x] 6.3.1 Explain BOINC delegation as validated public-interest compute.
    - [x] 6.3.2 Show placeholder catalog, validation state, budget, schedule, contribution, and gameplay-validation concepts.
    - [x] 6.3.3 Keep BOINC execution controls disabled or future.
  - [x] 6.4 Implement Cluster basic content.
    - [x] 6.4.1 Show local device identity status, runtime mode, runtime root, database path, and `last_run.log` status.
    - [x] 6.4.2 Show mock device cards with roles, reachability, health, typed capabilities, and queue/service ownership.
    - [x] 6.4.3 Surface doctor status and recent diagnostics in a human-readable panel.
  - [x] 6.5 Implement Routing basic content.
    - [x] 6.5.1 Show workload classes for text generation, image generation, video generation, STT, TTS, and BOINC research compute.
    - [x] 6.5.2 Show mock queues, priorities, device assignments, compatibility filtering, fallback routes, and policy constraints.
    - [x] 6.5.3 Clearly state that BOINC is schedulable research compute, not model inference.
  - [x] 6.6 Implement Tasks basic content.
    - [x] 6.6.1 Show placeholder scheduled tasks, webhooks, event-driven tasks, Signal-driven tasks, manual approvals, and run history.
    - [x] 6.6.2 Disable create/edit controls until durable task storage and execution exist.
  - [x] 6.7 Implement Settings basic content.
    - [x] 6.7.1 Show runtime paths, build artifact paths, mode, identity status, documentation/check status, and developer diagnostics.
    - [x] 6.7.2 Show current temporary HUD configuration values including tab placement, theme, scale, font size, key-binding defaults, notifications, diagnostics, default views, and privacy/safety posture.
    - [x] 6.7.3 Label hard-coded Phase 4 settings as future SQLite-backed user settings.
    - [x] 6.7.4 Show command hints for `--doctor`, `--smoke-test`, `last_run.log`, `make verify`, and `make check-docs`.
    - [x] 6.7.5 Keep destructive identity/runtime operations disabled.
  - [x] 6.8 Add tests proving each tab has content, expected disabled/future labels, and no accidental backend action enablement.

- [x] 7. Wire the GUI adapter to render the Phase 4 shell.
  - [x] 7.1 Connect the Ebitengine `Game` to the HUD shell/view model and temporary configuration manager.
  - [x] 7.2 Render the tab bar and selected tab content instead of only filling the background.
  - [x] 7.3 Preserve `--smoke-test` as a non-window verification path.
  - [x] 7.4 Keep headless builds free of Ebitengine initialization.
  - [x] 7.5 Add GUI build-tag tests where possible without requiring a display server.
  - [x] 7.6 Document native GUI dependencies and any remaining host validation limits.
  - [x] 7.7 Tighten custom HUD shell layout after screenshot review with taller tab buttons, smaller outer margins, and less tab-to-body spacing.

- [x] 8. Update documentation.
  - [x] 8.1 Update `internal/hud/README.md` with Phase 4 tab behavior, tab-view model, configuration manager, and future backend boundaries.
  - [x] 8.2 Update `internal/adapters/gui/README.md` with Ebitengine/EbitenUI integration details and custom-widget decisions.
  - [x] 8.3 Update root `README.md` only if user-facing run/build behavior changes.
  - [x] 8.4 Update `ROADMAP.md` Phase 4 item statuses only after implementation evidence exists.
  - [x] 8.5 Add or update local README files for any new package directories created during Phase 4.
  - [x] 8.6 Append the implementation checkpoint to today's journal.
  - [x] 8.7 Regenerate plan indexes.

- [x] 9. Verify Phase 4.
  - [x] 9.1 Run `make fmt`.
  - [x] 9.2 Run `make test`.
  - [x] 9.3 Run `make test-race`.
  - [x] 9.4 Run `make test-build`.
  - [x] 9.5 Run `make check-code-size`.
  - [x] 9.6 Run `make check-docs`.
  - [x] 9.7 Run `make lint`.
  - [x] 9.8 Run `make audit`.
  - [x] 9.9 Run `make verify`.
  - [x] 9.10 Build `apparatd` and run the headless smoke test.
  - [x] 9.11 Attempt GUI build and GUI smoke/runtime validation if native desktop headers and display access are available.
  - [x] 9.12 If GUI validation is blocked by host dependencies, record the exact missing package/display evidence without marking GUI validation complete.
  - [x] 9.13 Confirm generated artifacts remain ignored.
  - [x] 9.14 Confirm no files under `third_party/salvagecore` are staged.
  - [x] 9.15 Confirm `last_run.log` exists for GUI/headless runtime roots and startup output reports the exact path.
  - [x] 9.16 Confirm Android APK builds are not falsely represented by plain `GOOS=android` binary output.

- [x] 10. Complete the checkpoint.
  - [x] 10.1 Review final diff and staged payload.
  - [x] 10.2 Check pending downtime reports.
  - [x] 10.3 Commit the completed, verified checkpoint after approval.
  - [x] 10.4 Push the checkpoint to `origin` after commit.

## Open Decisions And Defaults

- Default tab placement is top for Phase 4; side rails are designed but not necessarily exposed as editable UI yet.
- Default theme is dark.
- Default UI scale is `1.0`.
- Settings are hard-coded through the temporary configuration manager during Phase 4.
- SQLite persistence for user settings is explicitly future work.
- Real backend actions remain disabled or mocked in Phase 4.
- EbitenUI should be used for standard GUI widgets where it supports controller-first focus and layout needs; custom widgets are allowed when documented.
