# Platform Matrix

Support is claimed only after target-specific build and behavior evidence exists.

## Steam Deck And Linux GUI

Requires controller-first HUD, readable 1280x800 sizing, keyboard/mouse fallback, audio capture, local storage paths, external WireGuard compatibility, runtime-root `last_run.log`, and generated release artifact under `releases/linux/<arch>/apparat/latest`.

Phase 3 keeps `--smoke-test` non-window for CI and build environments. The Ebitengine run loop is wired behind the `gui` build tag so native GUI validation can be performed on systems with the required desktop libraries.

## Linux Headless

Requires no Ebitengine initialization, CLI/API/service-manager control, health checks, graceful `SIGINT`/`SIGTERM`, durable storage paths, runtime-root `last_run.log`, and generated release artifact under `releases/linux/<arch>/apparatd/latest`.

Phase 3 headless startup initializes the same config, directory, logging, SQLite, identity-status, cluster-directory, and messaging primitives without importing Ebitengine.

## Windows

Requires `releases/windows/<arch>/apparat/latest.exe` and `releases/windows/<arch>/apparatd/latest.exe` artifacts, external WireGuard validation, service or tray decisions, certificate store handling, audio and controller validation, and path migration tests.

## macOS

Requires signing/notarization decisions, external WireGuard validation, keychain/certificate handling, app lifecycle, audio permissions, and controller validation.

## Android

Phase 5 adds the first Android build artifact but does not yet claim full Android runtime support.

Current evidence:

- Build host: Linux `amd64` development environment.
- Build command: `python3 scripts/build.py --os android --arch arm64 --target apparat`.
- Preflight command: `python3 scripts/build.py --check-android-env`.
- Artifact: `releases/android/arm64/apparat/latest.apk`.
- Package ID: `com.cjtrowbridge.apparat`.
- App label: `Apparat`.
- Native ABI: `arm64-v8a`.
- Permission: `android.permission.INTERNET` for HTTPS over external WireGuard/local networks.
- Activity: `org.golang.app.GoNativeActivity` with landscape orientation.
- Toolchain: JDK 21, Android platform `android-35`, build-tools `35.0.0`, NDK `27.2.12479018`, and pinned Ebitengine gomobile.

Known caveats:

- The APK has been built and inspected with Android build-tools, but install/launch validation is still pending because `adb` daemon startup was blocked by the current sandbox/approval environment.
- The Android GUI has not yet been observed on a physical device or emulator in this checkpoint.
- `last_run.log` creation inside Android app-scoped storage still needs device/emulator launch evidence.
- `gomobile` emits its default `minSdkVersion` metadata; explicit min/target SDK control is deferred to a future wrapper, signing, or release-hardening phase if direct `gomobile build` remains insufficient.
- Android `apparatd` is intentionally unsupported; headless Android work requires a later Termux/service-worker strategy.
- App-managed WireGuard/VPN-service, microphone capture, broad storage, background execution, release signing, store packaging, and additional Android ABIs are future work.
