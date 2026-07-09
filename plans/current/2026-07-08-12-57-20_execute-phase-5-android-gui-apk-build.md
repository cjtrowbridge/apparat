---
plan_id: 2026-07-08-12-57-20_execute-phase-5-android-gui-apk-build
title: Execute Phase 5 Android GUI APK Build
summary: Build an Apparat-owned Android GUI APK pipeline that emits `releases/android/arm64/apparat/latest.apk` without depending on the temporary salvagecore checkout or producing an Android headless artifact.
status: current
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

Phase 5 is partially executed: the APK build pipeline, preflight, artifact generation, package inspection, documentation, and unit tests are implemented. Device/emulator install and launch validation is blocked because `adb` daemon startup was rejected by the current sandbox/approval environment.

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
  - [x] 2.2 Confirm direct `gomobile build` can emit an installable-looking APK artifact for `cmd/apparat`.
  - [x] 2.3 Document why `golang/mobile` is not admitted as a source checkout.
  - [x] 2.4 Defer wrapper/AAR work until direct gomobile lacks required lifecycle, manifest, signing, SDK, or release behavior.
  - [x] 2.5 Avoid new third-party source admission in this checkpoint.
  - [x] 2.6 Record the decision in `README.md`, `ROADMAP.md`, `scripts/README.md`, `cmd/apparat/README.md`, and `docs/platform-matrix.md`.

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
  - [x] 8.4 Patch gomobile metadata synthesis to emit `minSdkVersion=23`, `targetSdkVersion=35`, and platform build version `35`.
  - [x] 8.5 Define landscape orientation for the controller-first HUD.
  - [x] 8.6 Add network permission for HTTPS over external WireGuard/local network.
  - [x] 8.7 Defer microphone permission until voice capture is enabled and tested.
  - [x] 8.8 Avoid broad storage permissions.
  - [x] 8.9 Defer Android VPN-service permission and app-managed WireGuard.

- [?] 9. Adapt Android runtime behavior.
  - [x] 9.1 Reuse the existing runtime startup path in the Android GUI entrypoint.
  - [?] 9.2 Verify actual Android app-scoped storage paths on device/emulator.
  - [?] 9.3 Verify `last_run.log` is deleted and recreated on every Android GUI launch.
  - [x] 9.4 Surface runtime root and `last_run.log` paths in existing Settings/diagnostics UI where runtime startup reaches the HUD.
  - [x] 9.5 Keep structured logging redaction rules documented.
  - [x] 9.6 Record Android path assumptions and validation gaps in `docs/platform-matrix.md`.

- [?] 10. Validate Android GUI behavior.
  - [x] 10.1 Build the debug APK locally.
  - [?] 10.2 Install the APK on an emulator or physical Android device with `adb`; blocked by sandbox/approval restrictions on `adb` daemon startup.
  - [?] 10.3 Launch the app and verify Ebitengine activity startup.
  - [?] 10.4 Verify the seven-tab Phase 4 HUD renders on Android.
  - [?] 10.5 Verify touch/click tab selection on Android.
  - [?] 10.6 Verify keyboard/controller navigation where device support exists.
  - [?] 10.7 Verify runtime directory creation on Android.
  - [?] 10.8 Verify fresh `last_run.log` after Android launch.
  - [?] 10.9 Capture `adb logcat` evidence for startup and failures.
  - [x] 10.10 Record exact build host, toolchain, ABI, package metadata, and validation blocker.

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

- [?] 13. Verify and complete.
  - [x] 13.1 Run `make fmt` equivalent through local Go path.
  - [x] 13.2 Run `make test` equivalent through local Go path.
  - [x] 13.3 Run `make test-build`.
  - [x] 13.4 Run `make check-docs`.
  - [x] 13.5 Run `make check-code-size`.
  - [x] 13.6 Run Android preflight.
  - [x] 13.7 Run Android build.
  - [x] 13.8 Confirm `releases/android/arm64/apparat/latest.apk` exists and is tracked by Git.
  - [x] 13.9 Confirm Android `apparatd` fails clearly.
  - [?] 13.10 Temporarily hide or move `third_party/salvagecore` and rerun Android build; not performed because the script/test path already proves no reference and moving local reference material is a separate destructive checkpoint.
  - [?] 13.11 Run emulator/device install and launch validation; Pixel install previously failed due obsolete SDK metadata, and rebuilt APK now declares min SDK 23 / target SDK 35, but final device validation still needs user-side retry or permitted `adb`.
  - [x] 13.12 Confirm no files under `third_party/salvagecore` are staged.
  - [x] 13.13 Review final diff and staged payload.
  - [?] 13.14 Check pending downtime reports before final summary.
  - [?] 13.15 Commit and push after approval.

## Open Follow-Up

- Retry install on Pixel or complete install/launch validation with a physical Android device or emulator where `adb devices`, `adb install`, `adb shell am start`, and `adb logcat` are permitted.
- Verify the HUD renders on Android and that tab touch/click input works.
- Verify Android app-scoped runtime root and `last_run.log` creation after launch.
- Decide whether direct `gomobile build` is sufficient for release hardening or whether a host-owned wrapper/AAR project is needed for SDK metadata, signing, icons, permissions, and store packaging.
