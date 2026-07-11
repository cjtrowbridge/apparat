---
plan_id: 2026-07-10-13-49-08_fix-hud-scroll-coordinate-and-native-slot-regressions
title: Fix HUD Scroll Coordinate And Native Slot Regressions
summary: Repair the HUD body layout regressions caused by mixed coordinate spaces and unsynchronized native Android controls in scrollable content.
status: past
created_at: 2026-07-10-13-49-08
---

# Fix HUD Scroll Coordinate And Native Slot Regressions

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Problem Statement

The last HUD body-layout checkpoint introduced two visible regressions:

- On Android, the native `Check for update` button floats over Settings content and does not move with the `Updates` fieldset when the Settings body scrolls.
- In non-Settings tabs, text in master/detail columns is smashed into the top-left area instead of drawing inside the intended list and detail panes.

The suspected root cause is not a single bad margin. The current implementation mixes absolute screen-space rectangles, subimage-local draw coordinates, scroll offsets, and Android native view overlay coordinates without one explicit conversion boundary.

## Non-Negotiable Fix Principles

- Ebitengine-rendered content and Android native controls must not pretend to share one rendering layer.
- Layout calculation must use one canonical coordinate model.
- Drawing helpers must receive rectangles in one coordinate space only, or their names/types must make conversion explicit.
- Native Android controls inside scrollable HUD content must be positioned from live game state, including scroll offset and clipping visibility.
- If a native control cannot correctly behave like scrolled content, it must be moved out of the scrolled content model instead of floating above it.
- No changes may be made inside submodules.

- [x] 1. Confirm scope and reproduce the regressions.
  - [x] 1.1 Read `README.md`, `ROADMAP.md`, `internal/hud/README.md`, `internal/adapters/gui/README.md`, `android/apparat/README.md`, and `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
  - [x] 1.2 Inspect `internal/adapters/gui/body_layout.go`, `internal/adapters/gui/body_scroll.go`, `internal/adapters/gui/ebiten_shell.go`, `cmd/apparatmobile/mobile.go`, and `android/apparat/src/com/cjtrowbridge/apparat/MainActivity.java`.
  - [x] 1.3 Record the exact current behavior from user screenshots: Settings update button floats over scrolled content, and non-Settings pane text renders at the wrong origin.
  - [x] 1.4 Confirm the active Apparat parent repo is clean before editing and confirm `agents` submodule has no local modifications.
  - [x] 1.5 Confirm no implementation work in this plan requires editing any file under `agents/` or another submodule.

- [x] 2. Define the corrected coordinate-space contract.
  - [x] 2.1 Define `screen rect` as the canonical output of layout calculation.
  - [x] 2.2 Define `pane-local rect` as a draw-only coordinate space produced explicitly at the clipping boundary.
  - [x] 2.3 Add helper names or types that make conversion explicit, such as `toLocalRect(parent, child)` or `drawInPane(pane, absoluteRect)`.
  - [x] 2.4 Document that scroll offsets are applied during layout, before clipping and drawing.
  - [x] 2.5 Document that Android native slots are screen-space overlay rectangles derived from the same scrolled layout data.

- [?] 3. Replace ad hoc body drawing with measured layout data.
  - [-] 3.1 Add a Settings layout function that returns body, content, fieldset, text, row, and native-slot rectangles in screen coordinates after applying `settingsScroll`.
  - [-] 3.2 Add a master/detail layout function that returns list-pane, detail-pane, list-item, detail-header, detail-summary, and detail-fieldset rectangles in screen coordinates after applying pane scroll offsets.
  - [-] 3.3 Ensure layout functions do not draw and do not depend on `*ebiten.Image`.
  - [x] 3.4 Ensure draw functions consume measured layout output and convert to pane-local coordinates only at the clipping boundary.
  - [x] 3.5 Remove or rewrite helper paths that accept ambiguous rectangles after clipping.

- [x] 4. Fix master/detail pane rendering.
  - [x] 4.1 Render the list pane using clipped content whose text coordinates are converted from screen-space item rectangles to list-pane-local coordinates.
  - [x] 4.2 Render the detail pane using clipped content whose text and fieldset coordinates are converted from screen-space detail rectangles to detail-pane-local coordinates.
  - [x] 4.3 Verify Comrades, Projects, Cluster, Routing, and Tasks place list text in the left pane and placeholder/detail text in the right pane.
  - [x] 4.4 Keep Research aligned with its documented placeholder pattern and verify it does not inherit the top-left smash behavior.
  - [x] 4.5 Preserve touch-sized list rows, fieldset rows, buttons, and input-like placeholders.

- [x] 5. Fix Settings scroll and native Android update button behavior.
  - [x] 5.1 Decide whether the temporary Android update button remains a native control inside scrolled Settings content or moves to a non-scrolled Settings toolbar.
  - [x] 5.2 If it remains inside scrolled content, compute its slot from live `game.bodyScroll.settings` rather than stateless width/height helpers.
  - [x] 5.3 Add a native slot result with `visible`, `x`, `y`, `w`, `h`, and owning clip rectangle semantics.
  - [x] 5.4 Hide the native Android button when the `Updates` fieldset or button slot is outside the visible Settings body clip.
  - [x] 5.5 Move the native Android button as Settings scrolls so it visually stays inside the `Updates` fieldset.
  - [x] 5.6 Keep Android `48dp` minimum native button height and sufficient width for the label.
  - [x] 5.7 Remove stale mobile bridge functions that expose stateless or unused view-scaled button geometry.

- [x] 6. Add regression tests for layout and native slots.
  - [x] 6.1 Add pure Go tests proving Settings fieldset and native slot rectangles move by exactly the Settings scroll offset.
  - [x] 6.2 Add tests proving the native update slot is invisible when scrolled out of the Settings body clip.
  - [x] 6.3 Add tests proving native slot rectangles remain inside their owning `Updates` fieldset when visible.
  - [x] 6.4 Add tests proving master/detail list text rectangles are inside the left pane.
  - [x] 6.5 Add tests proving master/detail detail text rectangles are inside the right pane.
  - [x] 6.6 Add tests proving clipped draw conversion maps screen rectangles to expected pane-local coordinates.
  - [x] 6.7 Update Android wrapper tests to check for the new native slot visibility bridge and removal of stale geometry bridge functions.

- [x] 7. Update Android wrapper integration.
  - [x] 7.1 Update `cmd/apparatmobile/mobile.go` to expose live native slot visibility and geometry from the active `game` instance.
  - [x] 7.2 Update `MainActivity.java` to hide the button unless the bridge reports the slot is visible.
  - [x] 7.3 Update `MainActivity.java` to place the button using live scrolled slot geometry, not only view width/height.
  - [x] 7.4 Keep the update download/cache-busting/hash comparison behavior unchanged unless a focused bug is found.
  - [x] 7.5 Rebuild the Android APK after bridge changes.

- [x] 8. Update documentation and plan state.
  - [x] 8.1 Update `internal/adapters/gui/README.md` to describe the screen-space layout and pane-local drawing boundary.
  - [x] 8.2 Update `android/apparat/README.md` to describe native slot visibility and scroll synchronization.
  - [-] 8.3 Update `internal/hud/README.md` if the Settings `Updates` placement or control model changes.
  - [x] 8.4 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to warn that native controls in scrollable panes require live slot geometry and clipping.
  - [x] 8.5 Append the checkpoint to `journal/2026-07-10.md`.
  - [x] 8.6 Regenerate plan indexes.

- [?] 9. Verify on Debian and Android.
  - [x] 9.1 Run focused HUD and GUI layout tests that do not require an X display.
  - [x] 9.2 Run `python3 -m unittest tests/build_pipeline_test.py`.
  - [x] 9.3 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [x] 9.4 Run `make check-docs`.
  - [x] 9.5 Run `python3 scripts/check_code_file_lines.py`.
  - [x] 9.6 Run `git diff --check`.
  - [x] 9.7 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - [x] 9.8 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build-android`.
  - [?] 9.9 Install and launch the rebuilt APK on the attached tablet when approval is available.
  - [?] 9.10 Capture Android screenshot evidence showing the update button inside the `Updates` fieldset before scrolling.
  - [?] 9.11 Capture Android screenshot evidence showing the button moving with the fieldset while partially scrolled.
  - [?] 9.12 Capture Android screenshot evidence showing the button hidden when the `Updates` slot scrolls out of view.
  - [?] 9.13 Capture Debian or screenshot-test evidence showing non-Settings tab text in the correct list/detail panes.

- [ ] 10. Review and publish.
  - [ ] 10.1 Review `git status -sb` and confirm no submodule modifications are present.
  - [ ] 10.2 Review staged diff and confirm no `agents/` or `third_party/salvagecore/` files are staged.
  - [ ] 10.3 Suggest commit message: `Fix HUD scroll layout regressions`.
  - [ ] 10.4 Commit and push only after user approval.

## Archive Note

- 2026-07-10: Archived as superseded by `plans/current/2026-07-10-18-58-38_recover-ebitenui-hud-settings-first.md`. The native-slot and custom-scroll coordinate strategy was replaced by an EbitenUI-first recovery direction.
