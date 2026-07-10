# Plan: Fix EbitenUI ScrollContainer Panic

## Problem Description

The Apparat application is currently launching to a black screen on both Android devices (phone and tablet) and Linux desktop. While this mirrors the symptom of a previous bug solved on July 9th, the root causes are completely different.

**History (July 9th):** The previous black screen on Android was caused by an Ebitengine native JNI context synchronization issue. The Android `deviceScaleFactor` evaluated to `0.0`, resulting in a division by zero in `pxToDp`, pushing layout values to `math.MaxInt64`, which caused the atlas allocator to fail. This was resolved by injecting the Android context earlier using `go.Seq.setContext()`.

**Current Issue:** The current black screen is a classic runtime panic in Go that instantly terminates the Ebitengine loop but leaves the Android Activity hanging on a black `SurfaceView` (or the desktop window blank). When migrating the settings pane to the EbitenUI framework, the `widget.NewScrollContainer()` constructor was used without providing the `widget.ScrollContainerOpts.Image()` property. The local Apparat version of the `ebitenui` library strictly requires this property to be explicitly defined during instantiation (`Validate()` enforces it) and does not automatically fall back to the central UI `Theme` (in fact, `widget.Theme` does not even have a field for `ScrollContainerParams`). If omitted, `ScrollContainer.Validate()` throws a fatal panic: `panic("ScrollContainer: Image is required.")`.

## Proposed Solution

1. **Define ScrollContainerImage:** We need to construct a valid `widget.ScrollContainerImage` inside `internal/adapters/gui/theme.go`. Since `widget.Theme` does not possess a slot for this, we will expose an additional builder function `createScrollContainerImage()` to vend this style.
2. **Inject ScrollContainerImage:** We will update `internal/adapters/gui/ui_builder.go` to call `createScrollContainerImage()` and inject it into the `widget.NewScrollContainer()` constructor using `widget.ScrollContainerOpts.Image()`.
3. **Build and Test:** Compile the application for Linux (`make verify`) and Android (`make build-android`). Deploy the new APK to the tablet (`3411105H803J8V`) using `adb` to verify the application loads successfully without panicking and the `Settings` pane renders correctly.
4. **Commit:** Commit and push the fix once visually validated.

## Execution Steps

- [ ] Update `internal/adapters/gui/theme.go` to construct and return a `widget.ScrollContainerImage`.
- [ ] Update `internal/adapters/gui/ui_builder.go` to provide the Image to `ScrollContainerOpts`.
- [ ] Run `make verify` and `make build-android` to ensure it compiles.
- [ ] Push the APK to the tablet via ADB.
- [ ] Commit and Push changes to the repository.
