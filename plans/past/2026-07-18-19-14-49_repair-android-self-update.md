---
plan_id: 2026-07-18-19-14-49_repair-android-self-update
title: Repair Android Self Update
summary: Diagnose the Package Installer rejection, report its actual result, and make the tracked Android latest APK a compatible update artifact.
status: past
created_at: 2026-07-18-19-14-49
---

# Repair Android Self Update

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5 Android GUI APK Build Pipeline, especially Android GUI smoke validation, release signing/version generation deferrals, and the later packaging/update path.
- Product contract: the temporary Android Settings updater must download the tracked GitHub `latest.apk`, preserve app data when Android accepts the update, and report Android's actual install result rather than implying that opening the installer succeeded.
- Governing playbook: `playbooks/debugging_changes_that_lead_to_errors.md`.
- Source inbox: `TODO.md` item `the apk artifact is broken and android refuses to install it`.

## Checklist

- [x] 1. Establish the governed repair checkpoint.
  - [x] 1.1 Mark the selected TODO item `[-]` in progress without altering its text.
  - [x] 1.2 Record the initial byte, certificate, installed-package, and prior-updater evidence.
  - [x] 1.3 Keep this plan current and its index entry regenerated before device or implementation work.

- [x] 2. Reproduce and classify the Android installer outcome on the connected device.
  - [x] 2.1 Preserve the currently installed package/data long enough to identify the installed signer and capture the incompatible-update failure; no tracked prior APK shares that signer.
  - [x] 2.2 When an explicitly user-approved migration is required, uninstall only the incompatible `com.cjtrowbridge.apparat` package, record the resulting local-data loss, and fresh-install the current laptop/GitHub APK.
  - [x] 2.3 Build a controlled older-code APK with the configured canonical signer, then trigger the in-app updater against the current `latest.apk` without clearing data.
  - [x] 2.4 Capture Package Installer and Apparat updater logs through the full permission, scan, confirmation, and completion path.
  - [x] 2.5 Verify package identity, initial startup, and updater-cache bytes after the signer-migration installation or controlled update.

- [x] 3. Make update outcomes observable and truthful.
  - [x] 3.1 Replace the fire-and-forget installer launch with a PackageInstaller result path that reports pending user action, success, or Android's exact failure code/message to the HUD and structured logs.
  - [x] 3.2 Keep downloaded APK access private, grant only the installer read access needed for the chosen install mechanism, and preserve Android unknown-source permission behavior.
  - [x] 3.3 Add focused tests or deterministic seams for package-status mapping and updater state transitions.

- [x] 4. Make the GitHub artifact a valid forward update target.
  - [x] 4.1 Define deterministic monotonically increasing Android version-code/version-name generation for tracked `latest.apk` builds.
  - [x] 4.2 Require a stable configured signing identity for artifacts intended to update an installed Apparat package; never track a private keystore or secret in Git.
  - [x] 4.3 Make build preflight fail clearly when a release/update artifact would be signed with an incompatible or unspecified identity.
  - [x] 4.4 Document key custody, build configuration, versioning, update limitations, and recovery from a prior differently signed development install.

- [x] 5. Verify and publish the repair.
  - [x] 5.1 Run focused Java/build tests, Android package/signature inspection, repository checks, and diff checks.
  - [x] 5.2 Build the new APK and perform a real in-place device update from an earlier compatible build without clearing app data.
  - [x] 5.3 Capture installer-result and post-update version/signature evidence.
  - [x] 5.4 Update this plan, regenerate indexes, append the journal evidence, confirm no `third_party/salvagecore` files are staged, and review pending downtime reports.
  - [x] 5.5 Commit and push after the user-approved checkpoint summary.

## Initial Evidence

- The updater cache, the tracked `releases/android/arm64/apparat/latest.apk`, and the connected device's installed base APK have SHA-256 `837f96e292a63f153728e7f63b5a9fa5f43434bd0d10dc63dd04d887500e6578`.
- The tracked APK verifies with v1, v2, and v3 signing. Its signer certificate SHA-256 is `db825efcf89ec77a5d14ee1381eb5ae5644d57c4a908672c6fd06fa953993015`, matching the installed APK.
- The connected Pixel Tablet package record shows `versionCode=1`, `versionName=0.1.0`, and an Apparat-originated Package Installer update at 2026-07-12 23:29. The prior journal records that a differently signed install had to be removed before this build was fresh-installed.
- The connected Pixel 10 Pro XL (`58051FDCQ002T9`) rejected an `adb install -r` of the laptop/GitHub APK without changing app data: `INSTALL_FAILED_UPDATE_INCOMPATIBLE: Existing package com.cjtrowbridge.apparat signatures do not match newer version`. Its installed package was last updated at 2026-07-12 21:02:56, confirming it belongs to the prior signing-key cohort.
- After the user's explicit approval, the connected Pixel 10 Pro XL package was removed and fresh-installed from `releases/android/arm64/apparat/latest.apk`. `adb install` returned `Success`; `MainActivity` became the focused foreground activity (PID `19435`), and the silent startup update check reported the matching `837f96e292a6` hash. The incompatible installation's Apparat-local data was necessarily removed.
- The updater currently treats `startActivity(Intent.ACTION_VIEW)` as success and cannot report the system install completion/failure. The manifest hard-codes `versionCode=1`, and the build uses an ignored generated debug keystore; these are not a durable release-update contract.

## Approved Plan Revision

- 2026-07-18: The available tracked older APK is signed by a different certificate, so it cannot support the planned no-data-wipe update reproduction. The user explicitly approved a one-time migration of connected Pixel 10 Pro XL `58051FDCQ002T9`: uninstall the incompatible Apparat package and fresh-install the laptop/GitHub APK. The plan now separates that data-clearing migration from the later same-signer controlled-update test.
