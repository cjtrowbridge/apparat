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
- Additional Windows evidence: Windows `amd64` produced the canonical signed APK through `python -m scripts.build` with Eclipse Temurin OpenJDK 21 and the pinned repo-local SDK/NDK. A connected Pixel Tablet (Android 16, API 36) accepted a fresh install after removal of a differently signed prior build; `MainActivity` resumed and remained alive. A touch swipe advanced body content while leaving the top tab strip stationary. This is one-device GUI evidence, not full Windows Android runtime support.
- Build command: `python3 scripts/build.py`.
- Preflight command: `python3 scripts/build.py` target report.
- Artifact: `releases/android/arm64/apparat/latest.apk`.
- Package ID: `com.cjtrowbridge.apparat`.
- App label: `Apparat`.
- Native ABI: `arm64-v8a`.
- Permissions: `android.permission.INTERNET` for HTTPS over external WireGuard/local networks, `android.permission.RECORD_AUDIO` for the push-to-talk state path, and temporary Phase 5 `android.permission.REQUEST_INSTALL_PACKAGES` for the dev Updates fieldset button.
- Activity: `com.cjtrowbridge.apparat.MainActivity` with Ebitengine `EbitenView`; the wrapper does not force portrait orientation.
- Toolchain: OpenJDK 21 (Eclipse Temurin preferred; Oracle JDK prohibited), Android platform `android-35`, build-tools `35.0.0`, NDK `27.2.12479018`, and pinned Ebitengine gomobile.
- SDK metadata: `minSdkVersion=23`, `targetSdkVersion=30`, and `platformBuildVersionCode=35`.
- Signing: debug keystore generated under ignored `.tools/android/debug.keystore`; APK verifies with v1, v2, and v3 signature schemes.
- Native page alignment: `lib/arm64-v8a/libapparat.so` LOAD segments align to `0x4000` for 16 KB page-size devices.
- Device validation: installed successfully on connected Pixel device `58051FDCQ002T9`, Android release `16`, SDK `36`; process remained alive after launch, wrote `last_run.log` under app-private storage, rendered the Apparat HUD, and accepted touch tab selection.

Known caveats:

- Additional Android device coverage is still pending for safe-area handling, density/readability hardening, keyboard/controller input, portrait and landscape behavior across form factors, and deeper runtime-path validation.
- The current APK uses the wrapper/AAR-style path. Future release-hardening still needs signing, icons, store packaging, additional ABI decisions, and broader Android device validation.
- Android `apparatd` is intentionally unsupported; headless Android work requires a later Termux/service-worker strategy.
- App-managed WireGuard/VPN-service, real microphone capture, broad storage, background execution, release signing, store packaging, signed update manifests, and additional Android ABIs are future work.
