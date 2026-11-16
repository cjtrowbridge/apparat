# Android Apparat Wrapper

This directory contains tracked Apparat-owned Android wrapper sources for the Phase 5 GUI APK.

The wrapper exists because direct `gomobile build` with `GoNativeActivity` can install and initialize Apparat runtime state, but it does not attach Ebitengine's Android `EbitenView`; the direct path only showed the Android splash/default icon. The wrapper uses Ebitengine's generated mobile view classes from `ebitenmobile bind` and a small `MainActivity` that sets that view as the activity content.

The build pipeline generates disposable intermediate files under `.tmp/android-apparat-wrapper`, then writes the canonical signed APK to:

```text
releases/android/arm64/apparat/latest.apk
```

Android builds use OpenJDK 21 only (Eclipse Temurin preferred); Oracle JDK is prohibited by repository policy. The selected `JAVA_HOME` must expose `java`, `javac`, and `keytool`.

Windows builds use the repository-local Android SDK fallback and invoke `sdkmanager.bat`, `d8.bat`, and `apksigner.bat` alongside executable build tools. The build verifies its existing ignored development keystore against the tracked certificate pin before signing. A Windows-produced APK still requires separate authorized-device install and launch evidence.

The wrapper must preserve the existing Phase 5 gates: package ID `com.cjtrowbridge.apparat`, full-screen behavior without forced portrait orientation, `minSdkVersion=23`, `targetSdkVersion=35`, Android platform 35 packaging, `arm64-v8a`, v2/v3 signing, 16 KB native page alignment, app-private runtime storage, touch tab selection, and no dependency on `third_party/salvagecore`.

The launcher icon is generated from the tracked root `logo.svg` concept into `res/drawable/app_icon.xml` and referenced by `AndroidManifest.xml`. The build pipeline compiles tracked Android resources before linking the wrapper APK.

Phase 5 also includes a temporary Settings `Updates` fieldset rendered by the EbitenUI HUD. The `Check for update` button is now an EbitenUI control inside that fieldset, not a native Android overlay. The Go mobile package registers an `Updater` callback with `MainActivity`; tapping the HUD button calls that callback, and Java performs the platform-specific work: download the tracked GitHub `latest.apk`, compare its SHA-256 with the installed APK, open Android's per-app unknown-source permission screen only when needed, and launch the Android package installer through `UpdateApkProvider`. Java reports coarse status text back through the mobile binding so the HUD button can leave its `Checking...` state after the platform path reaches a result. Update checks require an explicit action; startup does not download updates. This remains a dev-only bridge until a real Settings action, installed/latest version display, and signed update manifest replace it.

On Android 15 and newer, a native parent applies system-bar and display-cutout insets before sizing the engine view. This keeps rendering and touch coordinates in the same usable area during rotation and resume. Older Android versions retain decor-managed sizing. Without this parent, enforced edge-to-edge windows place the footer beneath navigation buttons and tabs beneath the camera. Verify with portrait/landscape screenshots, tab selection and pause/resume on an authorized device; API 23 remains the minimum.

Android safe-area color comes from the GUI renderer through `BackgroundARGB` in the mobile bridge. The current fixed dark palette is authoritative; runtime theme switching is not implemented. The native parent and navigation background use that color, with light navigation icons and no automatic contrast scrim. Insets and touch bounds are unchanged. Verify matching background pixels and readable buttons in portrait/landscape on an authorized device; system/OEM styling on other devices requires separate evidence.

Update continuity: `signing-certificate.sha256` pins the public certificate of the currently installed Pixel builds. Keep the existing ignored signing key; builds no longer create replacement keys. Version 0.1.1/code 2 supersedes code 1 and retains that signing identity. Future publications must increment the manifest code and pass `scripts/android_release.py` against the freshly downloaded GitHub APK. The prior GitHub APK used another certificate; new candidates follow the installed-device pin so those devices need no uninstall. Devices still using that old GitHub signer are a different signing lineage and cannot accept this new lineage in place. The updater distinguishes wrong package/signer and a non-newer version, while keeping all verification boundaries.
