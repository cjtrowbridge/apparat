---
plan_id: 2026-11-16-11-13-16_repair-android-display-build
title: Repair Android display patch and build compatibility
summary: Remove the obsolete Android patch failure and validate the wrapper APK without mutating tracked engine source.
status: past
created_at: 2026-11-16-11-13-16
---

# Repair Android display patch and build compatibility

## Authorization and protected scope

The user requested three separate plans for the three submodule-review findings. The user explicitly authorized execution of all three plans and Pixel 4a testing. Implementation is active; initial bootstrap v1 is selected. Work on the existing branch and preserve the uncommitted submodule updates and inventory edits. This plan does not authorize commit, push, publication, new branches, destructive cleanup or machine-wide installation.

Use the requested source refresh as the starting candidate. Do not silently downgrade a submodule, fetch a different moving head, upgrade unrelated SQLite/EbitenUI application dependencies, activate reference-only trees, or redesign the product. Keep the host's deterministic no-flag build orchestration, target eligibility, artifact semantics and 400-physical-line limit. Do not add a second build engine or inference/remote execution concerns.

## Roadmap binding and dependencies

Finding 1 in the [original review](../past/2026-11-16-11-00-00_update-and-review-submodules.md) is reproduced: `scripts/android_wrapper.py` expects `scale := float64(cScale)`, which the new engine no longer contains.

Bind to ROADMAP Phase 5's admitted mobile source, host wrapper/AAR path, Android tests and device evidence, plus Phase 8's item to stop mutating tracked/submodule source during builds. Phase 5 checked commands are historical evidence, not proof that the updated checkout works. Add a current status/link and keep operational instructions aligned with the no-flag pipeline; preserve historical journal evidence.

Read-only diagnosis can start immediately. Implementation must use the engine choice and compatible mobile-tool outputs from [Reconcile Ebitengine and mobile tool dependencies](../past/2026-11-16-11-13-18_reconcile-engine-and-mobile-dependencies.md). Actual software builds require the bootstrap and source gate from [Adopt framework source and build governance](../past/2026-11-16-11-13-17_adopt-framework-source-build-governance.md). This plan owns the final integrated Android APK/device evidence consumed by the dependency plan; that plan's Linux/module/tool checks need not wait for this APK, avoiding a circular implementation dependency.

## Intended approach

Trace the old guard's purpose and selected upstream handling. Remove the workaround if upstream handles it. Otherwise use a narrow validated patch in an ignored source copy/cache, keyed to exact source and patch identity; never mutate the tracked engine checkout. Retain useful failure diagnostics and behavior rather than matching an obsolete source-text needle.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Repair and verify the Android display/build boundary (finding 1).
  - [x] 1.1 Trace the previous zero/invalid display-metric guard and selected upstream display handling. Determine whether the workaround is obsolete or still necessary; record the behavior that must be preserved before changing it.
  - [x] 1.2 Remove the obsolete tracked-tree patch when safe. If still required, apply a narrow validated patch to an ignored copy/cache tied to the exact engine revision and patch identity, wire binding to that input, and fail clearly on unexpected source. Never alter the tracked submodule during success, failure or interruption.
  - [x] 1.3 Add regression coverage for supported upstream handling, invalid/zero metrics behavior where testable, clear unsupported-source failure when applicable, and clean source trees after both successful and failed build preparation. Replace brittle needle-only expectations with checks of the supported behavior.
  - [x] 1.4 Build the android/arm64 wrapper APK through the existing canonical pipeline with OpenJDK 21 and declared SDK/NDK tools. Verify package/signing metadata and existing ABI/SDK/page-size requirements; record output hash and complete dependency/tool provenance.
  - [x] 1.5 On an available test device/emulator, install and launch, inspect process/last_run.log, and check HUD rendering, touch tab selection, orientation and display scaling. Record device/OS and evidence; unavailable hardware remains an explicit pending validation, not a passed test.

- [x] 2. Complete Android acceptance and documentation.
  - [x] 2.1 Verify repeat preparation/builds leave gitlinks, manifests and submodule worktrees unchanged on success and failure; demonstrate no salvagecore input using an isolated fixture/copy, not destructive local moves.
  - [x] 2.2 Record APK hash, source/tool provenance, metadata and device rendering/input evidence; link results to the dependency plan. Unavailable device testing remains pending or explicitly deferred, never passed.
  - [x] 2.3 Update operational/platform documentation and close finding 1 only with the formerly failing boundary resolved and accurate target limitations.

## Expected files and ownership

`scripts/android_wrapper.py`, focused build tests, necessary binding integration in `scripts/build.py`, `scripts/README.md`, root README/ROADMAP current-status notes and `docs/platform-matrix.md`. This plan does not own module/tool version selection or governance migration.

## Evidence, rollback and closure

Capture task-owned diffs and exact source/tool revisions before implementation; revert only this plan's changes if needed, preserving other work and diagnostic evidence. Do not disable gates, discard user state, uninstall host tools, or publish generated artifacts as a fallback. Preserve tracked release artifacts during validation. Any engine fallback belongs to the engine-version decision in the dependency plan.

Run relevant regressions, `make check-docs`, code-size/source checks for changed source, plan-index validation and diff checks. Record results and unavailable checks truthfully. The earlier 41 Python tests do not establish Go compilation or device compatibility; Go was unavailable during the original review. Require OpenJDK 21 for Android, preferably Temurin, and prohibit Oracle JDK.

Update the closest README inventories and script help/prerequisite/output/failure documentation as behavior changes. Record a concise append-only journal checkpoint and findings disposition. Archive only after all work is complete or explicitly deferred with a named follow-up and accurate support limitations. Update cross-plan links on lifecycle moves. Do not claim Phase 5 or Phase 8 completion from this maintenance work.

## Pixel installation finding

The built APK reached the Pixel 4a (Android 13), but Play Protect blocked installation with the explicit explanation that the app targeted an older Android version and lacked current privacy protections. The inherited targetSdkVersion=30 was the cause shown by the phone. Raise target SDK to 35 (matching compile SDK), retain minimum 23, update metadata regression expectations, rebuild and retest without disabling verification or bypassing the warning. This is a bounded packaging correction discovered during the authorized Pixel acceptance test.

## Completion evidence — 2026-11-16

The old Android guard rejected nonpositive native width/height/scale. Stable 2.10.6 moved these values into an atomic cached display record and only guarded nil. `scripts/mobile_sources.py` now exports the selected commit, retains the nonpositive guard in the ignored copy and records commit/patch/hash identity. Malformed source fails before binding; fixture tests verify tracked input preservation and regeneration. Java view placeholders were aligned with the selected templates.

Canonical builds succeeded, including a repeat and the final target-35 rebuild (`.tmp/remediation/build-api35.txt`). Metadata confirms package com.cjtrowbridge.apparat, arm64, min 23, target/compile 35; v1/v2/v3 signatures and 16 KB ZIP alignment verify. APK SHA-256: `76f1656aa5f271bcdfc25530d627614c5296d46fee564d89e81ae290572a7b6b`. Tool provenance: Go 1.26.4, Temurin 21.0.13+9, Android platform 35/build-tools 35.0.0, NDK 27.2.12479018, gomobile v0.0.0-20260820040257-d11f821a26a6, Ebitengine 5f363d17271b029d43e17d17dc00da34f43e2a86. The patched display file SHA-256 is 554a007026b521e343487f833c8d97f1af2a90b6d7b60f10aa22a317cd80a060 (display-values-v1).

Pixel 4a / Android 13 installed and launched this APK. All five tabs rendered and accepted held-touch selection; the tab strip scrolled, portrait/landscape rendered, and background/resume preserved a responsive process (PID 8788). App-private last_run.log recorded runtime readiness. Rotation preferences were restored. Screenshots/logs live in `.tmp/remediation/pixel-*`; the platform matrix records clipped labels/footer, short synthetic taps sometimes missed, rotation surface warnings without observed fatal exceptions, and untested older/Android 15 devices. The updater permission screen was dismissed without enabling installation permission. No broad device-support claim is made.

Repeat builds preserved module hashes and the engine checkout stayed clean. An isolated cmd/internal/module fixture with only the engine source reference resolved GUI dependencies without any salvagecore directory; source-preparation fixtures also exclude salvagecore. Validated binaries are retained under `.tmp/remediation/validated-artifacts`; tracked release artifacts were restored from pre-test backups. Finding 1 is resolved at the build/display boundary; broader Android form-factor and safe-area work remains the existing Phase 5 follow-up.
