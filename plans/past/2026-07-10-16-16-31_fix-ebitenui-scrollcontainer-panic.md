---
plan_id: 2026-07-10-16-16-31_fix-ebitenui-scrollcontainer-panic
title: Fix EbitenUI ScrollContainer Panic
summary: Solves the black screen panic by correctly assigning PrimaryTheme.
status: past
created_at: 2026-07-10-16-16-31
---

# Plan: Fix EbitenUI ScrollContainer Panic

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Problem Description

The Apparat application is currently launching to a black screen on both Android devices (phone and tablet) and Linux desktop. While this mirrors the symptom of a previous bug solved on July 9th, the root causes are completely different.

**History (July 9th):** The previous black screen on Android was caused by an Ebitengine native JNI context synchronization issue. The Android `deviceScaleFactor` evaluated to `0.0`, resulting in a division by zero in `pxToDp`, pushing layout values to `math.MaxInt64`, which caused the atlas allocator to fail. This was resolved by injecting the Android context earlier using `go.Seq.setContext()`.

**Current Issue:** My previous diagnosis of a missing `ScrollContainerOpts.Image` was incomplete, as fixing that did not resolve the issue. After analyzing the tablet's ADB logcat and running the app locally on Linux, I found the true root cause: when instantiating `ebitenui.UI{Container: root}` in `ui_builder.go`, I omitted the `PrimaryTheme` property. Because `PrimaryTheme` was `nil`, the entire EbitenUI widget tree had no theme applied. When the `TabBook` was instantiated, it initialized its theme parameters to empty structs because it could not find a parent theme. Since `TabBook.Validate()` strictly enforces that the TabButton has a valid font color (`TabButtonText Color.Idle is required`), it crashed immediately with a fatal panic on Android (due to touch lifecycle forcing a relayout/validation). On Linux, it just failed to render styled components, leaving a black screen.

## Proposed Solution

1. **Inject PrimaryTheme:** The single missing piece was simply assigning `PrimaryTheme: game.theme` when creating the `ebitenui.UI` struct in `ui_builder.go`.
2. **Validate Locally:** Run the application locally on Linux to ensure Ebitengine successfully draws the first frame without panicking.
3. **Build and Test:** Compile the application for Android (`make build-android`) and deploy it to the tablet (`3411105H803J8V`) using `adb` to confirm the application launches and renders perfectly.
4. **Commit:** Commit and push the fix.

## Execution Steps

- [ ] Update `internal/adapters/gui/theme.go` to construct and return a `widget.ScrollContainerImage`.
- [ ] Update `internal/adapters/gui/ui_builder.go` to provide the Image to `ScrollContainerOpts`.
- [ ] Run `make verify` and `make build-android` to ensure it compiles.
- [ ] Push the APK to the tablet via ADB.
- [ ] Commit and Push changes to the repository.
