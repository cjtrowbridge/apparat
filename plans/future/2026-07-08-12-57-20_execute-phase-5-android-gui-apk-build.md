---
plan_id: 2026-07-08-12-57-20_execute-phase-5-android-gui-apk-build
title: Execute Phase 5 Android GUI APK Build
summary: Build an Apparat-owned Android GUI APK pipeline that emits `releases/android/arm64/apparat/latest.apk` without depending on the temporary salvagecore checkout or producing an Android headless artifact.
status: future
created_at: 2026-07-08-12-57-20
---

# Execute Phase 5 Android GUI APK Build

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5, `Android GUI APK Build Pipeline`.
- Product goal: produce a working Android GUI APK artifact for the Phase 4 HUD in the canonical release directory.
- Artifact goal: `releases/android/arm64/apparat/latest.apk`.
- Target boundary: Android builds only the GUI `apparat` artifact during this phase.
- Headless boundary: do not create an Android `apparatd` artifact; future Android headless work requires a separate Termux/service-worker strategy.
- Salvagecore boundary: `third_party/salvagecore` is temporary ignored reference material only; the final APK pipeline must work after salvagecore is removed.
- Source ownership boundary: any Android source, scripts, manifests, wrapper code, or tooling required for the APK pipeline must live in Apparat-owned tracked paths or be documented as external prerequisites.
- Platform boundary: app-managed Android WireGuard/VPN-service integration is out of scope; use external WireGuard/local-network assumptions.

## Implementation Checklist

- [ ] 1. Audit Android reference inputs.
  - [ ] 1.1 Confirm `third_party/game/ebiten` is initialized and contains `cmd/ebitenmobile`.
  - [ ] 1.2 Inspect Ebitengine Android mobile package layout, examples, manifest expectations, and command options.
  - [ ] 1.3 Inspect salvagecore's ignored `third_party/cicd/mobile` checkout for `golang/mobile` conventions.
  - [ ] 1.4 Inspect salvagecore's Ebitengine Android examples and any APK/mobile build scripts.
  - [ ] 1.5 Write a concise adopted/rejected/deferred Android reference summary in Apparat-owned docs.
  - [ ] 1.6 Verify no source file under `third_party/salvagecore` is copied or staged.

- [ ] 2. Decide the durable Android build shape.
  - [ ] 2.1 Attempt the shortest path using admitted Ebitengine `ebitenmobile` tooling directly.
  - [ ] 2.2 Identify whether direct `ebitenmobile` can express package ID, label, manifest, orientation, permissions, signing, and output path needs.
  - [ ] 2.3 If direct `ebitenmobile` is sufficient, document why `golang/mobile` does not need to be admitted as a source checkout.
  - [ ] 2.4 If direct `ebitenmobile` is insufficient, define a host-owned Android wrapper or AAR project path and its responsibilities.
  - [ ] 2.5 If `golang/mobile` is required directly, create a separate source-admission task before implementation.
  - [ ] 2.6 Record the decision in `README.md`, `ROADMAP.md`, `scripts/README.md`, and the nearest Android build README.

- [ ] 3. Define Android toolchain prerequisites.
  - [ ] 3.1 Select and document the JDK version.
  - [ ] 3.2 Select and document Android SDK command-line tools version.
  - [ ] 3.3 Select and document Android platform/API level.
  - [ ] 3.4 Select and document Android build-tools version.
  - [ ] 3.5 Select and document Android NDK version if required.
  - [ ] 3.6 Define accepted environment variables: `ANDROID_HOME`, `ANDROID_SDK_ROOT`, `ANDROID_NDK_HOME`, and `JAVA_HOME`.
  - [ ] 3.7 Define where local tool caches may live without being tracked.
  - [ ] 3.8 Define failure messages for each missing prerequisite.

- [ ] 4. Add Android build preflight.
  - [ ] 4.1 Add a script or `scripts/build.py` mode for Android environment checks.
  - [ ] 4.2 Check Java availability and version.
  - [ ] 4.3 Check Android SDK root and command-line tools.
  - [ ] 4.4 Check Android platform and build-tools.
  - [ ] 4.5 Check Android NDK only if the selected path requires it.
  - [ ] 4.6 Check `adb` availability for optional device/emulator validation.
  - [ ] 4.7 Check Ebitengine `ebitenmobile` availability from the admitted submodule or generated tool path.
  - [ ] 4.8 Check unsupported target ABIs before build time.
  - [ ] 4.9 Check that no Android build script or config references `third_party/salvagecore`.
  - [ ] 4.10 Add `make check-android-build-env`.

- [ ] 5. Define canonical Android artifact behavior.
  - [ ] 5.1 Add Android APK artifact path logic for `releases/android/arm64/apparat/latest.apk`.
  - [ ] 5.2 Keep `--print-path --os android --arch arm64 --target apparat` non-building and deterministic.
  - [ ] 5.3 Reject `--os android --target apparatd` with a clear Android-headless-out-of-scope message.
  - [ ] 5.4 Decide whether `--target all --os android` builds only `apparat` or fails until explicitly supported.
  - [ ] 5.5 Keep generated APKs ignored by Git.
  - [ ] 5.6 Add `make build-android`.

- [ ] 6. Integrate APK build execution.
  - [ ] 6.1 Extend `scripts/build.py` to route Android GUI builds through the selected APK builder.
  - [ ] 6.2 Preserve existing desktop/Linux GUI and headless behavior.
  - [ ] 6.3 Preserve existing release paths for non-Android builds.
  - [ ] 6.4 Ensure failures produce actionable prerequisites and command hints.
  - [ ] 6.5 Ensure Android builds do not initialize or build `cmd/apparatd`.
  - [ ] 6.6 Ensure the APK output is copied or written exactly to `releases/android/arm64/apparat/latest.apk`.

- [ ] 7. Add Android app metadata and permissions.
  - [ ] 7.1 Define application ID/package name.
  - [ ] 7.2 Define app label and launcher metadata.
  - [ ] 7.3 Define version name and version code generation.
  - [ ] 7.4 Define minimum and target SDK values.
  - [ ] 7.5 Define landscape/portrait behavior for phone, tablet, and controller-first layouts.
  - [ ] 7.6 Add network permission for HTTPS over external WireGuard/local network.
  - [ ] 7.7 Add microphone permission only if the selected smoke path validates voice capture.
  - [ ] 7.8 Avoid broad storage permissions; use app-scoped storage.
  - [ ] 7.9 Defer Android VPN-service permission and app-managed WireGuard.

- [ ] 8. Adapt Android runtime behavior.
  - [ ] 8.1 Determine Android app-scoped runtime root.
  - [ ] 8.2 Ensure SQLite, logs, identity, cache, artifacts, backups, and recovery directories are created under Android-safe paths.
  - [ ] 8.3 Ensure `last_run.log` is deleted and recreated on every Android GUI launch.
  - [ ] 8.4 Surface Android runtime root and `last_run.log` path in Settings or diagnostics where feasible.
  - [ ] 8.5 Confirm structured logging redacts secrets, raw audio, private message bodies, and project file contents.
  - [ ] 8.6 Record Android path assumptions in `docs/platform-matrix.md`.

- [ ] 9. Validate Android GUI behavior.
  - [ ] 9.1 Build the debug APK locally.
  - [ ] 9.2 Install the APK on an emulator or physical Android device with `adb`.
  - [ ] 9.3 Launch the app and verify Ebitengine activity startup.
  - [ ] 9.4 Verify the seven-tab Phase 4 HUD renders.
  - [ ] 9.5 Verify touch/click tab selection.
  - [ ] 9.6 Verify keyboard/controller navigation where device support exists.
  - [ ] 9.7 Verify runtime directory creation on Android.
  - [ ] 9.8 Verify fresh `last_run.log` after launch.
  - [ ] 9.9 Capture `adb logcat` evidence for startup and failures.
  - [ ] 9.10 Record exact emulator/device model, Android version, ABI, and validation result.

- [ ] 10. Add tests.
  - [ ] 10.1 Unit-test Android APK artifact path selection.
  - [ ] 10.2 Unit-test Android supports only `apparat` during this phase.
  - [ ] 10.3 Unit-test Android `apparatd` rejection.
  - [ ] 10.4 Unit-test Android `--print-path` behavior.
  - [ ] 10.5 Unit-test preflight failures for missing Java, SDK, build-tools, NDK when required, `adb`, and `ebitenmobile`.
  - [ ] 10.6 Unit-test no Android pipeline path references `third_party/salvagecore`.
  - [ ] 10.7 Add optional integration-test command for build/install/launch when Android tools and a device are available.

- [ ] 11. Update documentation.
  - [ ] 11.1 Update root `README.md` with Android APK prerequisites, commands, artifact path, and scope boundaries.
  - [ ] 11.2 Update `scripts/README.md` with Android build modes, preflight, outputs, side effects, and common failures.
  - [ ] 11.3 Add or update Android wrapper/build directory README files.
  - [ ] 11.4 Update `docs/platform-matrix.md` with validated Android evidence and remaining caveats.
  - [ ] 11.5 Update `third_party/README.md` if any new source is admitted.
  - [ ] 11.6 Update `ROADMAP.md` Phase 5 checklist only after implementation evidence exists.
  - [ ] 11.7 Append each checkpoint to the journal.
  - [ ] 11.8 Regenerate plan indexes.

- [ ] 12. Verify and complete.
  - [ ] 12.1 Run `make fmt`.
  - [ ] 12.2 Run `make test`.
  - [ ] 12.3 Run `make test-build`.
  - [ ] 12.4 Run `make check-docs`.
  - [ ] 12.5 Run `make check-code-size`.
  - [ ] 12.6 Run Android preflight.
  - [ ] 12.7 Run `make build-android`.
  - [ ] 12.8 Confirm `releases/android/arm64/apparat/latest.apk` exists and is ignored by Git.
  - [ ] 12.9 Confirm `python3 scripts/build.py --os android --arch arm64 --target apparatd` fails clearly.
  - [ ] 12.10 Temporarily hide or move `third_party/salvagecore` and rerun the Android build or selected no-salvagecore guard.
  - [ ] 12.11 Run emulator/device install and launch validation if Android tooling and a target device are available.
  - [ ] 12.12 Confirm no files under `third_party/salvagecore` are staged.
  - [ ] 12.13 Review final diff and staged payload.
  - [ ] 12.14 Check pending downtime reports.
  - [ ] 12.15 Commit and push after approval.

## Open Decisions And Defaults

- Default Android ABI is `arm64`.
- Default Android output is GUI-only `apparat`.
- Android headless `apparatd` is intentionally unsupported in Phase 5.
- Use admitted Ebitengine `ebitenmobile` tooling first.
- Do not admit `golang/mobile` as a source checkout unless implementation proves it is required.
- Keep salvagecore as temporary reference only; the final pipeline must not reference it.
- External WireGuard/local-network assumptions remain in force.
- Android VPN-service/app-managed WireGuard remains future work.
