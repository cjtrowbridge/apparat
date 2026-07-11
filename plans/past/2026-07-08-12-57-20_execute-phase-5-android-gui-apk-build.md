---
plan_id: 2026-07-08-12-57-20_execute-phase-5-android-gui-apk-build
title: Execute Phase 5 Android GUI APK Build
summary: Build an Apparat-owned Android GUI APK pipeline that emits `releases/android/arm64/apparat/latest.apk` without depending on the temporary salvagecore checkout or producing an Android headless artifact.
status: past
created_at: 2026-07-08-12-57-20
---

# Execute Phase 5 Android GUI APK Build

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5, `Android GUI APK Build Pipeline`.
- Product goal: produce an Android GUI APK artifact for the Phase 4 HUD in the canonical release directory.
- Artifact goal: `releases/android/arm64/apparat/latest.apk`.
- Target boundary: Android builds only the GUI `apparat` artifact during this phase.
- Headless boundary: do not create an Android `apparatd` artifact; future Android headless work requires a separate Termux/service-worker strategy.
- Salvagecore boundary: `third_party/salvagecore` is temporary ignored reference material only; the APK pipeline must work without referencing it.
- Source ownership boundary: Android source, scripts, manifests, and wrapper code must live in Apparat-owned tracked paths or be documented as external prerequisites.
- Build host boundary: Linux is the first evidence-producing build host; macOS and Windows Android build hosts remain unclaimed until validated.
- Platform boundary: app-managed Android WireGuard/VPN-service integration is out of scope.

## Execution Status

Phase 5 is partially executed: the APK build pipeline, preflight, artifact generation, package inspection, documentation, unit tests, Pixel install, process-liveness validation, app-private Android runtime storage, Android `last_run.log` creation, wrapper HUD rendering, full-screen Android view sizing, touch tab selection, and screenshot evidence are implemented. Direct `GoNativeActivity` paths initialized runtime state but stayed on the Android splash/default icon or failed with `internal/ui: Run is not implemented for GOOS=android`; the durable APK path now uses tracked Apparat wrapper sources plus Ebitengine generated mobile view classes. Remaining completion focuses on additional device testing, Android safe-area and density hardening, runtime-path validation depth, release-hardening deferrals, and resolving any local Ebitengine patch/submodule reproducibility work.

## Merged Checkpoint

The duplicate current plan `plans/current/2026-07-09-14-25-00_refactor-android-rendering-and-fit-screen.md` was merged into this Phase 5 plan. Its completed work is represented below in checklist items 8, 10, 13, and 14:

- Initialized `go.Seq.setContext(getApplicationContext())` in `MainActivity.java` so Ebitengine can query Android display metrics through gobind's native context path.
- Removed temporary Java-to-Go scale propagation, debug logging, transparent-background, framebuffer-query, and oversized-layout guard hacks.
- Removed the fixed portrait activity orientation so the Android wrapper can fill the screen in supported orientations.
- Added Android touch handling for tab selection and standard gamepad handling for `L1`, `R1`, and `R2`.
- Rebuilt and deployed the APK, verified touch tab selection on device, and captured screenshot evidence showing the Projects tab active with `input=select-tab`.

## Implementation Checklist

- [x] 1. Audit Android reference inputs.
  - [x] 1.1 Confirm `third_party/game/ebiten` contains `cmd/ebitenmobile` and mobile runtime packages.
  - [x] 1.2 Inspect Ebitengine mobile package layout and direct `gomobile build` behavior.
  - [x] 1.3 Inspect the ignored salvagecore Android/mobile references only as temporary context.
  - [x] 1.4 Reject durable Android pipeline dependence on salvagecore.
  - [x] 1.5 Record adopted/rejected/deferred Android reference conclusions in Apparat-owned docs.
  - [x] 1.6 Verify no source file under `third_party/salvagecore` is staged.

- [x] 2. Decide the durable Android build shape.
  - [x] 2.1 Attempt the shortest direct Ebitengine/gomobile path.
  - [x] 2.2 Confirm direct `gomobile build` can emit an installable APK artifact for `cmd/apparat`.
  - [x] 2.3 Document why `golang/mobile` is not admitted as a source checkout.
  - [x] 2.4 Confirm the Android-only `mobile.SetGame` path lacks required Android GUI lifecycle/render-surface behavior because Pixel startup remains on the Android splash/default icon.
  - [x] 2.5 Avoid new third-party source admission in this checkpoint.
  - [x] 2.6 Record the decision in `README.md`, `ROADMAP.md`, `scripts/README.md`, `cmd/apparat/README.md`, and `docs/platform-matrix.md`.
  - [x] 2.7 Test the shared `ebiten.RunGame` adapter and record that direct `GoNativeActivity` still fails on Android.
  - [x] 2.8 Promote the Apparat-owned wrapper/AAR-style path from deferred option to required Phase 5 work.

- [x] 3. Define Android toolchain prerequisites.
  - [x] 3.1 Select JDK 21.
  - [x] 3.2 Select Android SDK command-line tools as an external prerequisite.
  - [x] 3.3 Select Android platform/API `android-35`.
  - [x] 3.4 Select build-tools `35.0.0`.
  - [x] 3.5 Select NDK `27.2.12479018`.
  - [x] 3.6 Define `ANDROID_HOME`, `ANDROID_SDK_ROOT`, `ANDROID_NDK_HOME`, and `JAVA_HOME` discovery.
  - [x] 3.7 Define `.tools/` and `.tmp/` as ignored local tool/cache paths.
  - [x] 3.8 Define failure messages for missing prerequisites.

- [x] 4. Define Android build host support.
  - [x] 4.1 Treat Linux as the first evidence-producing APK build host.
  - [x] 4.2 Keep Python path and environment logic host-agnostic where practical.
  - [x] 4.3 Keep Linux-only convenience behavior outside core Python logic.
  - [x] 4.4 Add preflight messaging for unvalidated macOS/Windows hosts.
  - [x] 4.5 Avoid claiming macOS Android support before validation.
  - [x] 4.6 Avoid claiming Windows Android support before validation.
  - [x] 4.7 Allow device/emulator validation from any host with working `adb`.
  - [x] 4.8 Document device validation as a separate evidence step.

- [x] 5. Add Android build preflight.
  - [x] 5.1 Add `scripts/build.py --check-android-env`.
  - [x] 5.2 Check Java availability.
  - [x] 5.3 Check Android SDK command-line tools.
  - [x] 5.4 Check Android platform and build-tools.
  - [x] 5.5 Check Android NDK.
  - [x] 5.6 Check optional `adb` availability and report install/launch limitations.
  - [x] 5.7 Prepare/check the required Ebitengine gomobile helper.
  - [x] 5.8 Reject unsupported target ABIs before build time.
  - [x] 5.9 Unit-test that the Android build pipeline does not reference `third_party/salvagecore`.
  - [x] 5.10 Add `make check-android-build-env`.

- [x] 6. Define canonical Android artifact behavior.
  - [x] 6.1 Add Android APK artifact path logic for `releases/android/arm64/apparat/latest.apk`.
  - [x] 6.2 Keep `--print-path --os android --arch arm64 --target apparat` non-building and deterministic.
  - [x] 6.3 Reject `--os android --target apparatd` with a clear Android-headless-out-of-scope message.
  - [x] 6.4 Make `--target all --os android` build only `apparat`.
  - [x] 6.5 Track generated APKs in Git as the current latest Android build surface.
  - [x] 6.6 Add `make build-android`.

- [x] 7. Integrate APK build execution.
  - [x] 7.1 Extend `scripts/build.py` to route Android GUI builds through Ebitengine gomobile.
  - [x] 7.2 Preserve desktop/Linux GUI and headless behavior.
  - [x] 7.3 Preserve non-Android release paths.
  - [x] 7.4 Ensure failures include actionable prerequisite hints.
  - [x] 7.5 Ensure Android builds do not initialize or build `cmd/apparatd`.
  - [x] 7.6 Write the APK exactly to `releases/android/arm64/apparat/latest.apk`.

- [x] 8. Add Android app metadata and permissions.
  - [x] 8.1 Define package name `com.cjtrowbridge.apparat`.
  - [x] 8.2 Define app label `Apparat` and launcher activity metadata.
  - [x] 8.3 Define current version name/code as `0.1.0` and `1`.
  - [x] 8.4 Patch gomobile metadata synthesis to emit `minSdkVersion=23`, `targetSdkVersion=30`, and platform build version `35`.
  - [x] 8.5 Remove fixed portrait orientation from the wrapper activity so Android can fill the available screen in supported orientations.
  - [x] 8.6 Add network permission for HTTPS over external WireGuard/local network.
  - [x] 8.7 Add microphone permission for the existing push-to-talk state path while keeping real Android audio capture validation as future Phase 10 work.
  - [x] 8.8 Avoid broad storage permissions.
  - [x] 8.9 Defer Android VPN-service permission and app-managed WireGuard.
  - [x] 8.10 Add modern Pixel package/startup gates discovered during device testing: target SDK compatibility, v2/v3 signing, 16 KB native page alignment, and app-private runtime storage.

- [?] 9. Adapt Android runtime behavior.
  - [x] 9.1 Reuse the existing runtime startup path in the Android GUI entrypoint.
  - [?] 9.2 Verify actual Android app-scoped storage paths on device/emulator.
  - [?] 9.3 Verify `last_run.log` is deleted and recreated on every Android GUI launch.
  - [x] 9.4 Surface runtime root and `last_run.log` paths in existing Settings/diagnostics UI where runtime startup reaches the HUD.
  - [x] 9.5 Keep structured logging redaction rules documented.
  - [x] 9.6 Record Android path assumptions and validation gaps in `docs/platform-matrix.md`.

- [?] 10. Validate Android GUI behavior.
  - [x] 10.1 Build the debug APK locally.
  - [x] 10.2 Install the APK on a physical Pixel device with `adb`.
  - [x] 10.3 Launch the app and verify the Android process remains alive after startup.
  - [x] 10.4 Verify the seven-tab Phase 4 HUD renders on Android through the wrapper `EbitenView`.
  - [x] 10.5 Verify touch/click tab selection on Android.
  - [?] 10.6 Verify keyboard/controller navigation where device support exists.
  - [x] 10.7 Verify runtime directory creation on Android.
  - [x] 10.8 Verify fresh `last_run.log` after Android launch.
  - [x] 10.9 Capture `adb logcat` and `last_run.log` evidence for startup and failures.
  - [x] 10.10 Record exact build host, toolchain, ABI, package metadata, and validation blocker.
  - [x] 10.11 Record the Pixel install fixes required to reach app startup: modern SDK metadata, target SDK 30 compatibility workaround, debug APK signing, 16 KB page alignment, and app-private runtime root.
  - [x] 10.12 Verify Android opens to the Apparat HUD instead of remaining on the Android splash/default icon.
  - [-] 10.13 Verify portrait startup on a Pixel-class phone; closed because the wrapper no longer forces portrait orientation.
  - [x] 10.14 Capture visual evidence of the Android HUD matching the Debian HUD at the current Phase 4 feature level.

- [x] 11. Add tests.
  - [x] 11.1 Unit-test Android APK artifact path selection.
  - [x] 11.2 Unit-test Android supports only `apparat` during this phase.
  - [x] 11.3 Unit-test Android `apparatd` rejection.
  - [x] 11.4 Unit-test Android `--print-path` behavior.
  - [x] 11.5 Unit-test preflight failures for missing prerequisites.
  - [x] 11.6 Unit-test no Android pipeline path references `third_party/salvagecore`.
  - [x] 11.7 Unit-test unsupported Android ABI messaging.
  - [?] 11.8 Add optional integration-test command for build/install/launch when a device/emulator is available.

- [x] 12. Update documentation.
  - [x] 12.1 Update root `README.md` with Android APK prerequisites, commands, artifact path, host support policy, and scope boundaries.
  - [x] 12.2 Update `scripts/README.md` with Android build modes, preflight, outputs, side effects, and common failures.
  - [x] 12.3 Update `cmd/apparat/README.md` for Android manifest and gomobile files.
  - [x] 12.4 Update `docs/platform-matrix.md` with Android evidence and caveats.
  - [-] 12.5 Update `third_party/README.md`; no new source was admitted.
  - [x] 12.6 Update `ROADMAP.md` Phase 5 checklist with implementation evidence.
  - [x] 12.7 Append the checkpoint to the journal.
  - [x] 12.8 Regenerate plan indexes.

- [?] 13. Implement Android GUI parity.
  - [x] 13.1 Decide the minimal Apparat-owned Android integration path: tracked wrapper sources plus generated Ebitengine mobile view classes.
  - [x] 13.2 Add tracked Android wrapper sources/configuration so durable Android behavior lives outside `third_party/salvagecore`.
  - [x] 13.3 Preserve the existing package metadata, SDK metadata, signing, ABI, and release-artifact path gates.
  - [x] 13.4 Ensure Android startup enters the same `internal/app` runtime initialization path used by the Debian GUI.
  - [x] 13.5 Attach the actual Ebitengine render surface so the Apparat HUD is visible after launch.
  - [x] 13.6 Keep the seven canonical tabs, tab order, clickable tab behavior, disabled placeholders, runtime diagnostics, and Settings content aligned with Debian.
  - [ ] 13.7 Add Android safe-area/status-bar/navigation-bar layout handling.
  - [ ] 13.8 Add Android scale/density handling so tab buttons and body text remain readable on Pixel-class screens.
  - [x] 13.9 Remove fixed portrait orientation and allow the wrapper view to fill the Android screen in supported orientations.
  - [x] 13.10 Update `scripts/build.py`, tests, and documentation to build the wrapper/AAR-style APK path.
  - [x] 13.11 Verify the rebuilt APK installs, launches, renders the HUD, supports touch tab selection, writes `last_run.log`, and stays alive on the Pixel.
  - [ ] 13.12 Prove the final Android GUI path does not reference `third_party/salvagecore`.
  - [x] 13.13 Make the top tab strip responsive on small screens by sizing all tab buttons from the largest measured label plus balanced horizontal padding.
  - [x] 13.14 Make the top tab strip horizontally scrollable with mouse drag and touchscreen drag while preserving tap/click tab selection.
  - [x] 13.15 Keep the active tab visible when controller, keyboard, mouse, or touch input changes the selected tab.
  - [?] 13.16 Validate the responsive scrollable tab strip on Debian and the attached Android tablet; Android tablet build/install/launch/screenshot validation passed, Debian build/runtime validation passed, and interactive Debian window validation remains environment-limited.
  - [x] 13.17 Move the Android Ebitengine display-info guard out of the local dirty submodule checkout and into the Apparat-owned Android wrapper build pipeline.
  - [x] 13.18 Restore `third_party/game/ebiten` to a clean worktree after the Apparat-owned build-pipeline guard is verified.
  - [x] 13.19 Fix Android tab-strip touch handling so taps still select tabs and drag scrolling starts smoothly without a jump.
  - [?] 13.20 Add a temporary Android-native Settings `Updates` panel that downloads the tracked `latest.apk`, compares it with the installed APK, requests install-from-this-source permission only when needed, and launches the package installer with user consent; implemented and build/install validated, with final Settings-tab visual confirmation still pending because the attached tablet entered doze/lockscreen during screenshot capture.
  - [?] 13.21 Create HUD tab-content guidelines and align the current tab bodies so content is arranged in responsive, touch-sized, non-overlapping elements rather than floating overlays; implemented and build-validated, with final attached-tablet install/screenshot validation pending because ADB escalation was blocked by the approval usage limiter.

- [?] 14. Verify and complete.
  - [x] 14.1 Run `make fmt` equivalent through local Go path.
  - [x] 14.2 Run `make test` equivalent through local Go path.
  - [x] 14.3 Run `make test-build`.
  - [x] 14.4 Run `make check-docs`.
  - [x] 14.5 Run `make check-code-size`.
  - [x] 14.6 Run Android preflight.
  - [x] 14.7 Run Android build.
  - [x] 14.8 Confirm `releases/android/arm64/apparat/latest.apk` exists and is tracked by Git.
  - [x] 14.9 Confirm Android `apparatd` fails clearly.
  - [?] 14.10 Temporarily hide or move `third_party/salvagecore` and rerun Android build; not performed because the script/test path already proves no reference and moving local reference material is a separate destructive checkpoint.
  - [x] 14.11 Run Pixel install and launch validation with ADB; install succeeds, app process remains alive, and app-private `last_run.log` is created.
  - [x] 14.12 Run Pixel visual validation after Android GUI parity work; the HUD renders instead of the Android splash/default icon.
  - [-] 14.13 Validate Android phone portrait startup; closed because fixed portrait orientation was removed.
  - [x] 14.14 Validate Android touch/click tab selection.
  - [x] 14.15 Confirm no files under `third_party/salvagecore` are staged.
  - [x] 14.16 Review final diff and staged payload.
  - [?] 14.17 Check pending downtime reports before final summary.
  - [?] 14.18 Commit and push after approval.

## Open Follow-Up

- Direct `GoNativeActivity` is no longer the APK render path; the current APK uses the tracked Apparat wrapper activity and Ebitengine generated `EbitenView`.
- Android HUD rendering and touch tab selection have on-device screenshot evidence; additional device, controller, keyboard, safe-area, and density testing still needs to be added to this Phase 5 plan.
- The wrapper no longer forces portrait orientation; future testing should cover phone, tablet, portrait, landscape, keyboard, controller, and touch configurations explicitly.
- Preserve the already discovered Pixel gates: target SDK compatibility workaround, v2/v3 signing, 16 KB page alignment, app-private runtime storage, process liveness, and fresh `last_run.log`.

## Archive Note

- 2026-07-10: Archived as superseded by `plans/current/2026-07-10-18-58-38_recover-ebitenui-hud-settings-first.md`. Phase 5 remains the roadmap binding, but current execution now needs a Settings-first EbitenUI HUD recovery plan because the custom-coordinate UI plans no longer describe the active implementation paradigm.
