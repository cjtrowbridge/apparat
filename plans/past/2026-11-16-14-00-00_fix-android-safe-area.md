---
plan_id: 2026-11-16-14-00-00_fix-android-safe-area
title: Keep Android HUD inside the usable screen
summary: Respect modern Android system bars and camera cutouts and validate on Pixel 10.
status: past
created_at: 2026-11-16-14-00-00
---

# Keep Android HUD inside the usable screen

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

User's Pixel 10 overflow report continues the authorized implementation/install/testing task. Bind to Phase 8 GUI/input and Phase 17 continuous Android platform evidence. Keep HUD and touch coordinates inside the usable window in portrait and landscape; preserve runtime data, API 23 minimum, older-device layout, engine revision and build pipeline. No commit, release publication or data clearing.

## Work and acceptance

- [x] 1. Fit the Android 15+ engine view to system-bar and cutout insets through its native parent; document behavior.
- [x] 2. Build through the canonical pipeline, reinstall preserving data, check portrait/landscape/resume and touch alignment on the authorized Pixel 10, review and close.

## Diagnosis and evidence

Baseline: Pixel 10 Pro XL Android 17, 1080x2404, density 390. Screenshot `.tmp/pixel10-insets/before.png` shows the diagnostics footer beneath the three-button navigation bar, and tabs in the camera/rounded top region. Window dump reports EDGE_TO_EDGE_ENFORCED, full display bounds and zero content insets. MainActivity attaches EbitenView directly with FLAG_FULLSCREEN.

Hypotheses: (1) unhandled system insets: confirmed by window dump and screenshot; (2) wrong engine scaling: screenshot elements scale consistently at device density and engine onLayout reports actual child bounds, no evidence of independent scaling error; (3) widget minimum width overflow: does not explain navigation-bar overlap, defer unrelated mock selector truncation to existing Phase 17 notes. Minimal fix: FrameLayout parent padding on Android 15+ using systemBars/displayCutout union. EbitenView ignores its own padding, so inset the parent to resize rendering and local touch coordinates together. Older OS keeps decor-managed sizing. No new dependency or engine patch.

Acceptance: visible footer/tabs outside system UI and cutout, aligned touch targets, rotation/resume correct; build/source/docs checks pass. Preserve before/after images and tested APK in ignored `.tmp/pixel10-insets/`; restore tracked release artifacts after validation. Stop when these checks pass; do not claim all Android versions tested.

Implementation review: parent padding resizes both surface and local input space; no Go layout or engine source change. Insets are assigned absolutely on each dispatch, so rotation cannot accumulate padding. Return the original insets so the engine's IME reporting remains available. New APIs are guarded by SDK >=35; pre-35 parent has zero padding. First build exposed the direct javac bootclasspath's lack of lambda metafactory; replaced the listener lambda with the existing anonymous-class convention. Subsequent full build passed, as did source/400-line/docs checks and diff whitespace checks. MainActivity initially remained 368 lines. Installed APK SHA-256 `7a2cbe8800e9b5205958d78c70ddb2be3d3bd9777cfedad72be2db924eee2bd1`, install -r succeeded without uninstall or data clear. Tracked release files restored to pre-task bytes. Initial device acceptance confirmed portrait content between y=161 and y=2287 (161 px cutout, 117 px navigation), and landscape between x=161 and x=2287. Projects tap selected correctly; background/resume kept process 30719. The exposed default white parent was corrected to black with light navigation icons; final revalidation follows.

Final acceptance: retained fix. Final APK SHA-256 `4cfa7a25d14337d1d8373121b5b741f7269b175fd4316b7287a13139ba45f483`; installed using `adb install -r`, preserving data. Final portrait Projects and landscape Cluster taps selected their intended targets. Screenshots `portrait-final.png`, `landscape-final.png` and `resume-final.png` under `.tmp/pixel10-insets/` show controls and full footer inside the usable bounds, black cutout padding and visible system navigation controls. Background/resume returns to Cluster; process 31321 alive. Original accelerometer_rotation=0/user_rotation=0 restored. Full build/source/size checks pass; MainActivity 372 lines. Review found no remaining defect in this delta. Existing mock selector label truncation, other Android versions, gesture navigation, keyboard and accessibility are separate platform coverage gaps. No engine changes, dependencies, commits or publication.
