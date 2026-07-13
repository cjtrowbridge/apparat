---
plan_id: 2026-07-12-22-23-43_provision-windows-android-build-pipeline
title: Provision and Validate the Windows Android Build Pipeline
summary: Provision the pinned Android toolchain on Windows using OpenJDK only, build the GUI APK through the existing pipeline, and collect artifact and device evidence before claiming Windows Android-build support.
status: future
created_at: 2026-07-12-22-23-43
---

# Provision and Validate the Windows Android Build Pipeline

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5 Android GUI Build, Packaging, And Device Validation.
- Product contract: Phase 5 produces the GUI-only Android artifact at `releases/android/arm64/apparat/latest.apk`; Android `apparatd` remains intentionally unsupported.
- Current host evidence: the Windows build report can identify Android prerequisites but currently reports missing ADB, SDK command-line tools, platform `android-35`, build-tools `35.0.0`, and NDK `27.2.12479018`. Windows is not yet an evidence-producing Android build host.
- Toolchain policy: Oracle JDK is prohibited. Use OpenJDK 21 only, preferring Eclipse Temurin unless a documented compatibility reason selects another OpenJDK distribution.
- Related plans: `plans/current/2026-07-11-09-33-26_fix-mobile-overflow-and-tab-scroll.md`, `plans/current/2026-07-12-21-33-14_add-scrollable-hud-scenario-data.md`.

## Checklist

- [ ] 1. Establish an auditable Windows-host baseline.
  - [ ] 1.1 Inspect the current build report, environment overrides, available disk space, Go version/module cache, Java commands, Android paths, and `adb` visibility without changing machine state.
  - [ ] 1.2 Record the exact preflight failures and pinned versions in the journal and plan execution notes.
  - [ ] 1.3 Confirm `.tools/`, `build_environment.py`, debug keystores, patched helpers, and temporary source are ignored; do not stage downloaded tools or secrets.

- [ ] 2. Provision the permitted Java and Android dependencies.
  - [ ] 2.1 Install or configure an OpenJDK 21 distribution, preferably Eclipse Temurin, so `JAVA_HOME` provides `java`, `javac`, and `keytool`; reject Oracle JDK inputs.
  - [ ] 2.2 Install Android command-line tools, platform-tools (including ADB), platform `android-35`, build-tools `35.0.0`, and NDK `27.2.12479018` under ignored repo-local `.tools/android` or documented external paths.
  - [ ] 2.3 Accept required SDK licenses through the selected OpenJDK toolchain and configure `ANDROID_HOME`/`ANDROID_SDK_ROOT` and `ANDROID_NDK_HOME` using environment variables or ignored `build_environment.py`.
  - [ ] 2.4 Ensure the pinned Ebitengine `gomobile` module and its patched local helper can be generated without importing `third_party/salvagecore` or committing cache/tool artifacts.

- [ ] 3. Make the Windows invocation reliable and diagnosable.
  - [ ] 3.1 Run the documented no-flag build entry point and determine whether its Windows Python invocation can import Android wrapper support; fix only any confirmed invocation/package-resolution defect.
  - [ ] 3.2 Preserve a single canonical contributor command, accurate preflight reasons, pinned-version checks, and clear remediation for missing Java/SDK/NDK/ADB components.
  - [ ] 3.3 Update root build guidance, `scripts/README.md`, platform matrix, and Android wrapper documentation with the OpenJDK-only prerequisite, Windows setup, tool locations, large-download expectations, and known limits.

- [ ] 4. Produce and inspect the APK.
  - [ ] 4.1 Build `releases/android/arm64/apparat/latest.apk` on Windows through the pipeline; do not treat a successful Windows desktop build as Android evidence.
  - [ ] 4.2 Verify the APK exists at the canonical path, contains the expected package/activity/ABI, and passes `aapt2`, `apksigner`, and `zipalign` checks supplied by the pinned build-tools.
  - [ ] 4.3 Confirm signing uses only the ignored local debug keystore for this development checkpoint and does not expose passwords, certificate material, or raw device/project payloads in logs.

- [ ] 5. Gather device evidence before claiming Windows support.
  - [ ] 5.1 Connect an authorized Android device via ADB, install/replace the APK, launch the activity, and retain a redacted logcat/launch result.
  - [ ] 5.2 Capture screenshots demonstrating the tab strip and vertically scrollable content stay inside the body viewport, including the bottom diagnostics bar.
  - [ ] 5.3 Record device model/Android API level, package/activity result, process liveness, and any visual/input discrepancies in the plan and journal.
  - [ ] 5.4 Update `docs/platform-matrix.md` only with evidence actually produced on the Windows host; otherwise retain the existing Linux-only build-host claim and record Windows as pending.

- [ ] 6. Verify and publish the checkpoint.
  - [ ] 6.1 Run focused build/preflight tests, APK verification, relevant GUI tests, repository tests, code-size, documentation, and diff checks; distinguish pre-existing failures from plan-caused failures.
  - [ ] 6.2 Update this plan status/checklist and the current-day journal with commands, outcomes, paths, and remaining device-validation limits.
  - [ ] 6.3 Regenerate and validate plan indexes; confirm no files beneath `third_party/salvagecore/` are staged.
  - [ ] 6.4 Review pending downtime reports, then commit and push only after the user approves the verified checkpoint summary.

## Scope Boundaries

- This plan provisions and validates the existing GUI APK pipeline; it does not add Android headless support, release signing, store publishing, a new ABI, or broad Android feature work.
- Do not download, commit, or redistribute Oracle JDK. Do not commit SDK/NDK/JDK downloads, generated tool caches, debug keystores, ADB logs with sensitive content, or local environment configuration.
- A successful build is necessary but insufficient: Windows Android-build support is claimed only after artifact inspection and authorized device install/launch evidence.
- Toolchain acquisition changes local machine/repository ignored state and may require substantial downloads, network access, SDK-license acceptance, and a connected device; obtain runtime approval before beginning those steps.
