# Current Plans Index

Format: `last_modified | path | title | summary`

2026-07-10-16-39-38 | plans/current/2026-07-10-16-33-53_fix-ebitenui-regressions.md | Fix EbitenUI Layout Regressions | Restores the canonical layout, master-detail UI, sizing constraints, and dynamic update logic missing after EbitenUI migration.
2026-07-10-15-50-46 | plans/current/2026-07-10-15-44-00_fix-hud-layout-and-update-button.md | Fix HUD Layout and Android Update Button | Clean up the Ebitengine layout code to prevent text clumping on narrow screens, and replace the native Android update button overlay with a fully Ebitengine-rendered control triggered via JNI.
2026-07-10-14-24-43 | plans/current/2026-07-10-14-12-58_recover-hud-layout-with-visual-validation.md | Recover HUD Layout With Visual Validation | Recover from the recent HUD layout regressions by reverting unsafe assumptions, rebuilding layout around verifiable rendering evidence, and requiring Android/Debian screenshots before release.
2026-07-10-13-55-54 | plans/current/2026-07-10-13-49-08_fix-hud-scroll-coordinate-and-native-slot-regressions.md | Fix HUD Scroll Coordinate And Native Slot Regressions | Repair the HUD body layout regressions caused by mixed coordinate spaces and unsynchronized native Android controls in scrollable content.
2026-07-10-13-37-08 | plans/current/2026-07-10-13-16-00_hud-body-layout-primitives.md | HUD Body Layout Primitives | Add reusable responsive HUD body layout primitives for native control slots, block-level text/input wrapping, and scrollable master/detail panes.
2026-07-10-09-16-32 | plans/current/2026-07-08-12-57-20_execute-phase-5-android-gui-apk-build.md | Execute Phase 5 Android GUI APK Build | Build an Apparat-owned Android GUI APK pipeline that emits `releases/android/arm64/apparat/latest.apk` without depending on the temporary salvagecore checkout or producing an Android headless artifact.
