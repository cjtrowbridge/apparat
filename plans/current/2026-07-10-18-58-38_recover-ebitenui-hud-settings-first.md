---
plan_id: 2026-07-10-18-58-38_recover-ebitenui-hud-settings-first
title: Recover EbitenUI HUD Settings First
summary: Rebuild the blank EbitenUI HUD from the Settings tab outward, restore the update button path, and sweep stale custom-layout code, tests, docs, and plans.
status: current
created_at: 2026-07-10-18-58-38
---

# Recover EbitenUI HUD Settings First

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5, `Android GUI APK Build Pipeline`.
- Product contract: keep the seven canonical HUD tabs from `README.md`, restore the Settings update path first, and keep Android GUI support bound to tracked Apparat wrapper sources.
- Recovery reason: the custom-coordinate HUD recovery plans no longer match the active EbitenUI implementation, and all tabs currently render blank.
- Scope boundary: this plan recovers the HUD shell and visible tab bodies; it does not add backend services, real project data, production update manifests, release signing, Android headless support, or app-managed WireGuard.
- Validation rule: geometry/unit tests alone are insufficient for HUD recovery. Settings and each restored tab pattern need rendered visual evidence before being marked complete.

## Superseded Current Plans

- `plans/past/2026-07-08-12-57-20_execute-phase-5-android-gui-apk-build.md`
- `plans/past/2026-07-10-13-16-00_hud-body-layout-primitives.md`
- `plans/past/2026-07-10-13-49-08_fix-hud-scroll-coordinate-and-native-slot-regressions.md`
- `plans/past/2026-07-10-14-12-58_recover-hud-layout-with-visual-validation.md`
- `plans/past/2026-07-10-15-44-00_fix-hud-layout-and-update-button.md`

## Recovery Checklist

- [x] 1. Reset plan governance.
  - [x] 1.1 Confirm the parent repo starts clean before plan lifecycle edits.
  - [x] 1.2 Archive all stale `plans/current/` execution plans that describe superseded custom-coordinate or incomplete EbitenUI migration work.
  - [x] 1.3 Create this single current recovery plan as the active execution authority.
  - [x] 1.4 Regenerate plan indexes after lifecycle moves.
  - [x] 1.5 Append the governance reset checkpoint to `journal/2026-07-10.md`.

- [ ] 2. Reproduce and characterize the blank-tab failure.
  - [ ] 2.1 Run the focused GUI test/build commands needed to distinguish compile failures from runtime rendering failures.
  - [ ] 2.2 Capture current Debian visual evidence for Settings and at least one non-Settings tab.
  - [ ] 2.3 Capture current Android visual evidence for Settings when ADB/device access is available.
  - [ ] 2.4 Inspect the current EbitenUI widget tree construction for missing stretch, min-size, theme, tab content attachment, or layout invalidation problems.
  - [ ] 2.5 Record the root cause before broad tab reconstruction begins.

- [ ] 3. Restore the Settings tab and update-button path first.
  - [ ] 3.1 Make one Settings fieldset visibly render through EbitenUI with nonzero bounds.
  - [ ] 3.2 Render all Settings sections from `internal/hud` data, preserving `Updates`, `HUD Configuration`, `Bindings`, and `Diagnostics`.
  - [ ] 3.3 Render a clearly visible EbitenUI `Check for update` button inside the `Updates` section.
  - [ ] 3.4 Wire the Settings update button to the existing Android updater callback path without reintroducing a native overlay button.
  - [ ] 3.5 Add focused tests proving Settings sections and the update button are present in the EbitenUI construction path.
  - [ ] 3.6 Validate the Settings tab visually on Debian.
  - [ ] 3.7 Validate the Settings update button visually and functionally on Android when device access is available.

- [ ] 4. Rebuild non-Settings tab bodies under the EbitenUI paradigm.
  - [ ] 4.1 Define the restored EbitenUI tab-body patterns for Settings fieldsets, Research placeholder/review content, and master-detail tabs.
  - [ ] 4.2 Restore Comrades from `internal/hud` sections with left-list and right-detail content.
  - [ ] 4.3 Restore Projects from `internal/hud` sections with left-list and right-detail content.
  - [ ] 4.4 Restore Research as a readable placeholder/review body.
  - [ ] 4.5 Restore Cluster from `internal/hud` sections with left-list and right-detail content.
  - [ ] 4.6 Restore Routing from `internal/hud` sections with left-list and right-detail content.
  - [ ] 4.7 Restore Tasks from `internal/hud` sections with left-list and right-detail content.
  - [ ] 4.8 Add visual validation evidence for every restored tab or explicitly mark any tab as still blocked.

- [ ] 5. Reintegrate or replace expected HUD behavior.
  - [ ] 5.1 Restore controller and keyboard tab switching through the EbitenUI `TabBook` without desynchronizing `hud.Shell`.
  - [ ] 5.2 Restore touch/click tab selection and verify it works on Android.
  - [ ] 5.3 Decide whether tab overflow is handled by EbitenUI or needs an approved EbitenUI-compatible wrapper.
  - [ ] 5.4 Define and implement the narrow-screen fallback for master-detail tabs.
  - [ ] 5.5 Preserve 44px minimum touch targets for tabs, list rows, and buttons.
  - [ ] 5.6 Keep diagnostics visible without covering primary Settings/update controls.

- [ ] 6. Sweep stale custom-layout code, tests, and docs.
  - [ ] 6.1 Remove or rewrite GUI tests that reference deleted custom-coordinate symbols.
  - [ ] 6.2 Add replacement GUI tests for EbitenUI theme creation, tab construction, Settings content, update callback wiring, and tab action synchronization.
  - [ ] 6.3 Remove dead custom layout helpers, obsolete bridge functions, and stale artifacts only when they are no longer referenced.
  - [ ] 6.4 Update `internal/adapters/gui/README.md` so it describes the current EbitenUI shell instead of the custom tab/body renderer.
  - [ ] 6.5 Update `android/apparat/README.md` so it describes the current update callback path instead of a native button overlay.
  - [ ] 6.6 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` if the practical EbitenUI recovery reveals better constraints.
  - [ ] 6.7 Update `ROADMAP.md` Phase 5 status only when implementation evidence changes.

- [ ] 7. Verify and publish the recovery checkpoint.
  - [ ] 7.1 Run focused GUI tests with the `gui` tag.
  - [ ] 7.2 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [ ] 7.3 Run `make check-docs`.
  - [ ] 7.4 Run `python3 scripts/check_code_file_lines.py`.
  - [ ] 7.5 Run `git diff --check`.
  - [ ] 7.6 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - [ ] 7.7 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build-android`.
  - [ ] 7.8 Confirm no files under `third_party/salvagecore/` are staged.
  - [ ] 7.9 Review pending downtime reports before final summary.
  - [ ] 7.10 Commit and push after the user-approved checkpoint summary.

## Execution Notes

- 2026-07-10: Governance reset created this plan because the active current plans described obsolete custom-coordinate recovery paths while the codebase had already pivoted to EbitenUI and all tabs were reported blank.
