---
plan_id: 2026-07-10-14-12-58_recover-hud-layout-with-visual-validation
title: Recover HUD Layout With Visual Validation
summary: Recover from the recent HUD layout regressions by reverting unsafe assumptions, rebuilding layout around verifiable rendering evidence, and requiring Android/Debian screenshots before release.
status: past
created_at: 2026-07-10-14-12-58
---

# Recover HUD Layout With Visual Validation

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Findings From Review

- Commit `79d821d Add HUD body layout primitives` changed too many layout concerns at once: clipping, scrolling, native Android button placement, text wrapping, touch target sizing, Android bridge functions, docs, and release artifacts.
- Commit `2c9d947 Fix HUD scroll layout regressions` patched coordinate symptoms without first producing visual reproduction evidence, so it did not prove the UI actually improved.
- The implementation treated unit geometry checks as enough for visual layout correctness. That was not enough because the regressions are visible rendering failures, not only pure math failures.
- The Ebitengine renderer now mixes parent-screen drawing, `SubImage` clipping, pane-local coordinates, and scroll offsets in fragile ways.
- The Android update button is a native view layered above the Ebitengine view. It cannot be treated as ordinary scrolled HUD content unless its position, clipping, and visibility are synchronized from live rendered state.
- Current Android update-button behavior and non-Settings text layout must be validated on actual rendered pixels before more inputs or controls are added.
- The safest recovery path is to get back to a small known-good body layout, then add one behavior at a time with screenshots and tests after each step.
- No submodule edits are allowed. `agents/` must remain clean and pinned to the parent-recorded commit.

## Recovery Strategy

This plan deliberately prioritizes evidence over speed. It should not begin with another speculative patch. It should first establish what the current UI actually renders, decide whether to revert parts of the last two commits, and then rebuild only the smallest layout changes needed to satisfy the original goals.

Original goals still worth preserving:

- Settings content is a vertical stack of non-overlapping fieldsets.
- Most non-Settings tabs use master/detail panes with list content on the left and placeholder/detail content on the right.
- List rows, fieldset rows, buttons, and future inputs are large enough for touch.
- Long tab/body content can scroll by mouse wheel, pointer drag, and touch drag.
- Android update checking remains available from Settings without occupying important global UI real estate.

## Proposed Direction

Preferred technical direction:

- First remove the Android native update button from scrolled body content. Put the update action in a stable Settings fieldset visually, but implement the actual touch target as Ebitengine-rendered HUD content if possible. If Android installation must remain native-triggered, make the native view a temporary fixed Settings header action, not a child of a scrolling fieldset.
- Revert or simplify the `SubImage` clipping approach until the app has screenshot-based layout tests. A simpler first recovery is to render body content directly in screen coordinates with explicit clipping deferred, restoring correct placement before reintroducing scroll clipping.
- Reintroduce scrolling per pane only after the static fieldset/master-detail layout is visually correct on Debian and Android.
- Add screenshot capture as a required validation gate for layout work. Code compile and unit tests are necessary but not sufficient for HUD layout changes.

- [?] 1. Establish evidence and rollback boundary.
  - [x] 1.1 Confirm `git status -sb` is clean before starting implementation.
  - [x] 1.2 Confirm `git submodule status agents` matches the parent-recorded commit and `agents/` has no local modifications.
  - [x] 1.3 Record current `main` commit and the two suspect commits: `79d821d` and `2c9d947`.
  - [?] 1.4 Capture current Android screenshots for Settings, Comrades, Projects, Cluster, Routing, and Tasks before making changes.
  - [?] 1.5 Capture current Debian screenshots for the same tabs before making changes, using an actual render path rather than pure unit tests.
  - [?] 1.6 Compare current screenshots against the last known acceptable screenshots/artifacts from before `79d821d`, if available.
  - [x] 1.7 Decide explicitly whether to revert `2c9d947`, revert both `2c9d947` and `79d821d`, or patch forward from current `main`.

- [ ] 2. Define the visual acceptance contract.
  - [ ] 2.1 Settings must show the `Updates` fieldset at the top without floating controls over unrelated content.
  - [ ] 2.2 Settings fieldsets must stack vertically with visible gaps and no overlap.
  - [ ] 2.3 Comrades, Projects, Cluster, Routing, and Tasks must show left-pane list content and right-pane placeholder/detail content in their own panes.
  - [ ] 2.4 Research must render according to its current placeholder pattern without text collapsing into the wrong origin.
  - [ ] 2.5 Text may truncate temporarily, but it must not draw over another element or disappear into the wrong pane.
  - [ ] 2.6 Android update checking must remain reachable from Settings, but it must not float over scrolled body content.
  - [ ] 2.7 Any scrollable area must visibly scroll only its own content, not unrelated tabs, diagnostics, or native overlays.

- [?] 3. Recover static body layout first.
  - [x] 3.1 Restore or implement a minimal static Settings fieldset stack using screen-space coordinates only.
  - [x] 3.2 Restore or implement minimal static master/detail rendering using screen-space coordinates only.
  - [x] 3.3 Remove ambiguous `SubImage` coordinate conversion from body content until static placement is visually correct.
  - [x] 3.4 Keep touch-sized rows and button-sized placeholders without adding new interactive inputs.
  - [?] 3.5 Verify static layout visually on Debian before reintroducing scroll behavior.
  - [?] 3.6 Verify static layout visually on Android before reintroducing scroll behavior.

- [?] 4. Redesign the Android update action safely.
  - [x] 4.1 Decide whether the temporary update action is Ebitengine-rendered or native Android-rendered.
  - [-] 4.2 If Ebitengine-rendered, add a HUD action hit target that calls a narrow Android bridge method to start update checking.
  - [x] 4.3 If native-rendered, place it in a fixed Settings header/action row outside the scrollable body and document that temporary compromise.
  - [x] 4.4 Remove native Android button placement from scrolled fieldset geometry.
  - [x] 4.5 Keep the existing download/cache-bust/hash/installer permission flow unless direct evidence shows another bug.
  - [?] 4.6 Verify the update action is visible, touch-sized, and not floating over unrelated fieldsets on Android.

- [?] 5. Reintroduce scrolling only after static layout is proven.
  - [-] 5.1 Add Settings scroll state only after the static Settings layout has screenshot evidence.
  - [-] 5.2 Add master/detail pane scroll state only after static master/detail layout has screenshot evidence.
  - [-] 5.3 Support mouse wheel on Debian for the pane under the pointer.
  - [-] 5.4 Support touch drag on Android with a threshold that does not steal taps.
  - [ ] 5.5 Do not auto-scroll selected top tabs while the user is manually dragging the tab strip.
  - [x] 5.6 Add the Android tab-strip jump issue from `TODO.md` to the scrolling validation set before closing this plan.

- [?] 6. Add meaningful automated coverage.
  - [x] 6.1 Add pure layout tests for screen-space pane and fieldset rectangles.
  - [x] 6.2 Add tests for scroll clamping independent of Ebitengine runtime.
  - [x] 6.3 Add tests for touch drag threshold and tap preservation.
  - [x] 6.4 Add tests that Android wrapper source does not position native controls inside scrolling body content unless a live visibility contract exists.
  - [?] 6.5 Add or document a screenshot smoke-test path for Debian GUI layout.
  - [?] 6.6 Add a manual Android screenshot checklist if full automated Android screenshots are not feasible yet.

- [?] 7. Validate with real visual evidence.
  - [x] 7.1 Run `python3 -m unittest tests/build_pipeline_test.py`.
  - [x] 7.2 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [x] 7.3 Run `make check-docs`.
  - [x] 7.4 Run `python3 scripts/check_code_file_lines.py`.
  - [x] 7.5 Run `git diff --check`.
  - [x] 7.6 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - [x] 7.7 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build-android`.
  - [?] 7.8 Install and launch the rebuilt APK on the attached tablet.
  - [?] 7.9 Capture Android screenshots for Settings before and after any Settings scroll.
  - [?] 7.10 Capture Android screenshots for at least one non-Settings master/detail tab.
  - [?] 7.11 Capture Debian screenshots for Settings and at least one non-Settings master/detail tab.
  - [?] 7.12 Do not mark layout work complete until screenshot evidence matches the visual acceptance contract.

- [x] 8. Update docs and planning state.
  - [x] 8.1 Update `internal/adapters/gui/README.md` with the final rendering and scrolling contract.
  - [x] 8.2 Update `android/apparat/README.md` with the final temporary update-action placement.
  - [x] 8.3 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` with the screenshot-validation requirement for layout changes.
  - [x] 8.4 Append findings and execution results to `journal/2026-07-10.md`.
  - [x] 8.5 Regenerate plan indexes.

- [?] 9. Publish only after evidence review.
  - [x] 9.1 Review `git status -sb` and confirm no submodule modifications are present.
  - [ ] 9.2 Review staged file list and confirm no `agents/` or `third_party/salvagecore/` files are staged.
  - [?] 9.3 Present screenshot evidence and residual risks before commit.
  - [ ] 9.4 Commit with a message that reflects the recovery path, not just the intended feature.
  - [ ] 9.5 Push only after the evidence-backed checkpoint is accepted.

## Execution Notes

- 2026-07-10: Attempted to capture current Android state before changes, but the first screenshot showed the launcher rather than Apparat. Added foreground-window verification to the validation flow.
- 2026-07-10: Captured `artifacts/apparat-recovery-settings.png` after installing the previous build; it showed the native update button overlapping the `Updates` fieldset title, confirming that native fieldset-child placement was still unsafe.
- 2026-07-10: Recovered static body layout by disabling body scroll update hooks and drawing Settings/master-detail bodies directly in screen coordinates. Scrolling is intentionally deferred.
- 2026-07-10: Moved the native Android update button to a stable Settings header slot rather than scrolled fieldset geometry.
- 2026-07-10: Rebuilt Android APK after the recovery changes, but attached-tablet install/screenshot validation was blocked by the approval usage limiter before the rebuilt APK could be installed.
- 2026-07-10: Archived as superseded by `plans/current/2026-07-10-18-58-38_recover-ebitenui-hud-settings-first.md`. The evidence-first lesson remains binding, but the new plan starts from the EbitenUI implementation and the current blank-tab failure.
