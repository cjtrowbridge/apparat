---
plan_id: 2026-07-10-23-59-48_complete-ui-and-build-backlog
title: Complete UI And Build Backlog
summary: Convert the remaining TODO inbox items into planned work for app icon integration, responsive HUD polish, Debug UI/PTT controls, and automatic build orchestration.
status: past
created_at: 2026-07-10-23-59-48
---

# Complete UI And Build Backlog

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5 Android GUI APK Build Pipeline, Phase 4 HUD tab shell expectations, and later release/build hardening items.
- Product contract: keep the seven canonical HUD tabs, maintain touch-first responsive behavior on Android and small screens, keep release artifacts tracked under `releases/`, and preserve the GUI/headless target split.
- Source inbox: `TODO.md` items captured by the user on 2026-07-10.
- Scope boundary: this plan supersedes the EbitenUI HUD recovery plan after the `844bf16` checkpoint. It covers the remaining TODO-derived UI polish and build-system changes.

## Checklist

- [x] 1. Integrate the Apparat app icon.
  - [x] 1.1 Decide whether `logo.svg` is the canonical source icon or whether it needs design refinement first.
  - [x] 1.2 Add tracked source icon assets in the appropriate product/assets location.
  - [x] 1.3 Generate and wire Linux desktop/window icon assets where supported by the current GUI stack.
  - [x] 1.4 Generate and wire Android launcher/icon resources through the tracked wrapper build pipeline.
  - [x] 1.5 Document the icon source, generation command, output paths, and verification path.
  - [x] 1.6 Rebuild Linux and Android artifacts and capture evidence that the icon is visible in at least Android launcher/app info or the supported desktop surface.

- [x] 2. Fix block-level responsive HUD layout overflow.
  - [x] 2.1 Reproduce the tab-button and tab-content overflow on a narrow viewport/device capture.
  - [x] 2.2 Ensure tab body content is strictly width-bounded to the visible screen.
  - [x] 2.3 Ensure vertical content panes can scroll as needed without widening their parent.
  - [x] 2.4 Ensure the top tab list scrolls horizontally without making the whole page overflow.
  - [x] 2.5 Fix the collapsed detail Back button alignment so its text is centered and the button does not stretch off-screen.
  - [x] 2.6 Add focused GUI construction or layout-state tests for narrow widths where practical.
  - [x] 2.7 Capture narrow-screen visual evidence after the overflow and Back-button fixes.

- [x] 3. Fix tab selection state during overflow scrolling.
  - [x] 3.1 Reproduce the partially off-screen tab selection case where the selected tab and following tab both appear highlighted.
  - [x] 3.2 Identify whether the cause is button toggle state, rebuild timing, pointer drag/click sequencing, or scroll-centering behavior.
  - [x] 3.3 Ensure exactly one tab button has the selected visual state after pointer, touch, keyboard, and controller tab changes.
  - [x] 3.4 Add regression coverage for the single-selected-tab invariant.
  - [x] 3.5 Capture visual evidence on a narrow tab-strip surface.

- [x] 4. Improve list/detail touch targets and chat-like alignment.
  - [x] 4.1 Increase Back buttons and right-column/list buttons to match the tab-button scale closely enough for touch comfort.
  - [x] 4.2 Keep all interactive controls at or above the 44px minimum touch target.
  - [x] 4.3 Left-align right-column button text like a threaded chat/list app.
  - [x] 4.4 Leave room in the row layout for later avatars for projects, cluster devices, research projects, comrades, and tasks.
  - [x] 4.5 Update GUI docs/playbook guidance for the row/button sizing and alignment standard.
  - [x] 4.6 Capture visual evidence on expanded and collapsed master-detail views.

- [x] 5. Polish Debug UI and global Push-To-Talk controls.
  - [x] 5.1 Change the Settings Debug UI control text from open to close while the overlay is open.
  - [x] 5.2 Preserve Debug UI overlay behavior across Settings rebuilds without desynchronizing the control state.
  - [x] 5.3 Add a floating circular PTT button near the bottom-right of the app.
  - [x] 5.4 Make the PTT button globally visible without covering core navigation, Settings update controls, or diagnostics by default.
  - [x] 5.5 Wire the PTT button to the existing push-to-talk state as a hold interaction when the app is ready for it; until then, keep it visibly disabled or diagnostic-only.
  - [x] 5.6 Validate the Debug UI label change and PTT placement on desktop-sized and phone-sized surfaces.

- [x] 6. Replace build flags with automatic build orchestration.
  - [x] 6.1 Inventory the current `scripts/build.py`, `Makefile`, Android wrapper, toolchain, and release artifact entry points.
  - [x] 6.2 Define a single no-flag `build.py` entry point that detects the host environment and reports every possible and impossible target with reasons.
  - [x] 6.3 Add `build_environment.sample.py` with local hook functions for machine-specific paths or environment preparation.
  - [x] 6.4 Ignore local `build_environment.py` in Git while loading it opportunistically from `build.py`.
  - [x] 6.5 Make the no-flag build run all possible targets for the current host and report per-target outputs/results.
  - [x] 6.6 Preserve a documented path for CI or scripts to query/report target feasibility without hiding failures.
  - [x] 6.7 Update `scripts/README.md`, root `README.md`, and build/release docs for the new build entry point.
  - [x] 6.8 Update tests for build-target detection and impossible-target reporting.
  - [x] 6.9 Run the new build entry point locally and confirm Linux and Android artifacts are rebuilt when both are possible.

- [x] 7. Verify and publish the backlog checkpoint.
  - [x] 7.1 Run focused tests for each changed subsystem.
  - [x] 7.2 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [x] 7.3 Run `make check-docs`.
  - [x] 7.4 Run `python3 scripts/check_code_file_lines.py`.
  - [x] 7.5 Run `git diff --check`.
  - [x] 7.6 Rebuild all possible targets through the new build entry point.
  - [x] 7.7 Validate Android changes on device when ADB/device access is available.
  - [x] 7.8 Confirm no files under `third_party/salvagecore/` are staged.
  - [x] 7.9 Review pending downtime reports before final summary.
  - [x] 7.10 Commit and push after the user-approved checkpoint summary.

## Notes

- `logo.svg` existed as an untracked file when this plan was created. Do not assume it should be staged until icon integration work explicitly selects the source and destination paths.
- The TODO inbox remains user-authored. Leave TODO lines intact unless the user explicitly asks to edit or check them off.
