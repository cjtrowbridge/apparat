---
plan_id: 2026-07-10-13-16-00_hud-body-layout-primitives
title: HUD Body Layout Primitives
summary: Add reusable responsive HUD body layout primitives for native control slots, block-level text/input wrapping, and scrollable master/detail panes.
status: past
created_at: 2026-07-10-13-16-00
---

# HUD Body Layout Primitives

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Source TODO Items

This plan pulls the first three items from `TODO.md` into executable planning scope:

- The Android update button is floating instead of aligning with the Settings `Updates` fieldset.
- Text fields and input fields need block-level layout behavior so text wraps or clips inside the visible area.
- Columns need touch-drag, click-drag, and mouse-wheel scrolling, with default bindings tracked in the configuration model and later Settings UI for rebinding.

- [x] 1. Define the shared HUD body layout contract.
  - [x] 1.1 Read `README.md`, `ROADMAP.md`, `internal/hud/README.md`, `internal/adapters/gui/README.md`, and `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
  - [x] 1.2 Confirm the plan remains aligned with Phase 4/5 HUD scope and does not introduce backend-dependent behavior.
  - [x] 1.3 Define a minimum body width and minimum touch target size for body rows, list rows, buttons, inputs, and native control slots.
  - [x] 1.4 Define the supported body patterns for this checkpoint: Settings fieldset stack and master/detail panes.
  - [x] 1.5 Define overflow behavior for this checkpoint: clipping, truncation, wrapping, and scroll container boundaries.
  - [x] 1.6 Audit current partial HUD body/native update-button implementation and decide which parts to keep, replace, or remove before adding new primitives.

- [x] 2. Add reusable body layout primitives.
  - [x] 2.1 Add or refine rectangle/layout helpers for body panes, fieldsets, rows, text blocks, inputs, buttons, and native control slots.
  - [x] 2.2 Ensure every primitive has a named minimum width or height where touch, readability, or responsive layout requires one.
  - [x] 2.3 Implement block-level text measurement and wrapping or truncation inside a bounded rectangle.
  - [x] 2.4 Implement input/text-field placeholder layout as block-level elements that cannot draw outside their owning fieldset or pane.
  - [x] 2.5 Add tests for text wrapping/truncation and input bounds at normal and narrow body widths.

- [x] 3. Fix the Settings `Updates` native control slot.
  - [x] 3.1 Model the `Updates` fieldset as HUD-owned visual content at the bottom of the Settings fieldset stack.
  - [x] 3.2 Add a stable native control slot id for the Android update button, such as `settings.updates.check_for_update`.
  - [x] 3.3 Calculate the native slot rectangle from the same layout data used to draw the `Updates` fieldset.
  - [x] 3.4 Expose the slot rectangle through the Android mobile bridge without hard-coded Android margins.
  - [x] 3.5 Convert HUD logical coordinates to Android view coordinates using the actual `EbitenView` size and rendered game layout scale.
  - [x] 3.6 Hide the native Android update button unless Settings is active and the slot is visible.
  - [x] 3.7 Add tests proving the native slot is inside the `Updates` fieldset and meets touch target size.
  - [x] 3.8 Rebuild the Android APK and verify the Java wrapper compiles against the bridge.

- [x] 4. Add scrollable master/detail panes.
  - [x] 4.1 Add scroll state for the left list pane and right detail pane.
  - [x] 4.2 Support mouse wheel scrolling for the pane under the pointer or the focused pane.
  - [x] 4.3 Support mouse click-drag scrolling in scrollable panes without breaking list-item selection.
  - [x] 4.4 Support touch-drag scrolling in scrollable panes without breaking tap selection.
  - [x] 4.5 Clamp scroll offsets so content cannot scroll past its start or end.
  - [-] 4.6 Keep the active/selected list item visible after keyboard/controller/touch selection changes.
  - [x] 4.7 Add tests for scroll clamping, drag threshold behavior, and tap-versus-drag selection behavior.

- [x] 5. Track default scroll and pane bindings in configuration.
  - [x] 5.1 Add default scroll-related HUD actions or binding entries to the configuration model.
  - [x] 5.2 Document mouse wheel, pointer drag, touch drag, and keyboard/controller scroll defaults.
  - [x] 5.3 Add later-phase roadmap or TODO-linked work for Settings UI that lets users reassign these bindings.
  - [x] 5.4 Add tests proving the default configuration exposes the new binding entries.

- [x] 6. Align current tab bodies with the new primitives.
  - [x] 6.1 Render Settings as a vertical stack of non-overlapping fieldsets.
  - [x] 6.2 Render Comrades as a master/detail layout with placeholder detail text in the right pane.
  - [x] 6.3 Render Projects as a master/detail layout with placeholder detail text in the right pane.
  - [x] 6.4 Render Cluster as a master/detail layout with placeholder detail text in the right pane.
  - [x] 6.5 Render Routing as a master/detail layout with placeholder detail text in the right pane.
  - [x] 6.6 Render Tasks as a master/detail layout with placeholder detail text in the right pane.
  - [x] 6.7 Keep Research aligned with the current placeholder pattern while documenting when it should become master/detail.
  - [x] 6.8 Confirm no tab body draws text, controls, or native views outside its owning element.

- [x] 7. Update documentation and governance.
  - [x] 7.1 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` with any refined slot, wrapping, or scrolling rules.
  - [x] 7.2 Update `internal/hud/README.md` for body-layout model and configuration expectations.
  - [x] 7.3 Update `internal/adapters/gui/README.md` for renderer behavior, native slot mapping, and scroll validation.
  - [x] 7.4 Update `android/apparat/README.md` if the Android update bridge behavior or assumptions change.
  - [x] 7.5 Update `README.md`, `ROADMAP.md`, or `docs/platform-matrix.md` if user-visible platform behavior changes.
  - [x] 7.6 Append the checkpoint to the daily journal and regenerate plan indexes.

- [?] 8. Verify.
  - [?] 8.1 Run focused unit tests for HUD configuration and GUI layout helpers.
  - [x] 8.2 Run `python3 -m unittest tests/build_pipeline_test.py`.
  - [x] 8.3 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [x] 8.4 Run `make check-docs`.
  - [x] 8.5 Run `python3 scripts/check_code_file_lines.py`.
  - [x] 8.6 Run `git diff --check`.
  - [x] 8.7 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - [x] 8.8 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build-android`.
  - [?] 8.9 Install and launch the rebuilt Android APK on the attached tablet when ADB approval is available.
  - [?] 8.10 Capture screenshot evidence showing the Settings `Updates` button aligned inside its fieldset and master/detail placeholder text inside right-hand panes.
  - [x] 8.11 Confirm TODO items 1-3 are checked off, reworded, or linked to this plan after implementation.

## Execution Notes

- 2026-07-10: Kept the existing partial fieldset/master-detail renderer, replaced its stop-drawing overflow behavior with clipped scroll panes, and anchored the Android update button to the `settings.updates.check_for_update` native slot derived from the `Updates` fieldset.
- 2026-07-10: Closed item 4.6 because the current mock rows do not yet expose a real selected item/focus model; future selectable panes should implement keep-visible behavior when that model lands.
- 2026-07-10: Focused `internal/hud` tests passed. `go test -tags gui ./internal/adapters/gui` remains blocked in this headless environment by Ebitengine/GLFW failing to open X display `:1`; compile coverage is provided by `make build` and `make build-android`.
- 2026-07-10: Corrected Android test feedback by moving the high-priority `Updates` fieldset to the top of Settings, removing fragile native-button view scaling, enforcing Android dp-based touch height, cache-busting GitHub APK downloads, and making the "already current" toast include the matching hash prefix.
- 2026-07-10: Archived as superseded by `plans/current/2026-07-10-18-58-38_recover-ebitenui-hud-settings-first.md`; useful intent should be reimplemented through EbitenUI rather than custom body layout primitives.
