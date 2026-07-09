---
plan_id: 2026-07-09-14-25-00_refactor-android-rendering-and-fit-screen
title: Refactor Android Rendering and Fit Screen
summary: Clean up the custom Java-to-Go scale propagation by utilizing gobind's native Seq.setContext(), and allow the app to fill the screen in any orientation.
status: current
created_at: 2026-07-09-14-25-00
---

# Refactor Android Rendering and Fit Screen

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 5, `Android GUI APK Build Pipeline`.
- Product goal: Clean up temporary hacks, ensure the APK fits the full target screen, and establish a stable, standard Ebitengine initialization structure.

## Execution Plan

### 1. Register Context with Go Runtime
- [x] Initialize `go.Seq.setContext` in `MainActivity.java`.
  - Calling `go.Seq.setContext(getApplicationContext())` in `onCreate()` populates the Go side's global context variable.
  - This allows Ebitengine's native `displayInfo()` helper to query display metrics via JVM without custom APIs.

### 2. Clean Up Scale Hacks
- [x] Revert changes in `third_party/game/ebiten/internal/ui/ui_mobile.go`.
  - Removed the custom `SetDeviceScale` method.
- [x] Revert changes in `third_party/game/ebiten/mobile/ebitenmobileview/mobile.go`.
  - Removed the custom `SetDeviceScale` wrapper function.
- [x] Revert changes in `third_party/game/ebiten/internal/graphicsdriver/opengl/graphics.go`.
  - Restored `Begin()` to the original no-op; the framebuffer query hack is no longer needed.
- [x] Simplify patching in `scripts/android_wrapper.py`.
  - Removed the custom `setDeviceScale` call, `pxToDp` override, and all debug logging injections.
  - Only the template variable substitution and RGBA_8888 pixel format fix remain.
- [x] Clean up `internal/adapters/gui/ebiten_shell.go`.
  - Removed debug `slog` logging, `frameCount` instrumentation, and unnecessary 100,000 Layout guard.
- [x] Clean up `MainActivity.java`.
  - Removed transparent background hack, verbose lifecycle logging. Kept only essential lifecycle methods.

### 3. Support Full Screen Orientation and Android Permissions
- [x] Modify `android/apparat/AndroidManifest.xml`.
  - Removed `android:screenOrientation="portrait"` to allow the screen to rotate and fill landscape devices.
  - Added `<uses-permission android:name="android.permission.RECORD_AUDIO" />` for push-to-talk voice capture.

### 4. Add Touch Input Support
- [x] Update `internal/adapters/gui/ebiten_shell.go` to handle touch events alongside mouse events.
  - Used `inpututil.AppendJustPressedTouchIDs(nil)` to detect new touch-down events each frame.
  - For each new touch, calls `ebiten.TouchPosition(id)` to get coordinates.
  - Feeds those coordinates into the existing `tabIndexAt(x, y)` hit test.

### 5. Implement Gamepad Controller Parity
- [x] Update `internal/adapters/gui/ebiten_shell.go` to handle gamepad button inputs.
  - Added `updateGamepad()` method querying active gamepads via `ebiten.AppendGamepadIDs(nil)`.
  - Checks for `ebiten.IsStandardGamepadLayoutAvailable(id)`.
  - Maps `StandardGamepadButtonFrontTopRight` (R1) to next tab and `StandardGamepadButtonFrontTopLeft` (L1) to previous tab.
  - Maps `StandardGamepadButtonFrontBottomRight` (R2) to start/release voice capture.

### 6. Rebuild & Verify
- [x] Run the Android build pipeline to produce the latest APK.
- [x] Deploy the APK and verify touch-tap on tabs switches the active tab on device.
- [x] Capture a screenshot confirming touch interaction works (Projects tab active via touch, `input=select-tab` in diagnostics).
