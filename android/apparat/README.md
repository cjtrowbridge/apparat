# Android Apparat Wrapper

This directory contains tracked Apparat-owned Android wrapper sources for the Phase 5 GUI APK.

The wrapper exists because direct `gomobile build` with `GoNativeActivity` can install and initialize Apparat runtime state, but it does not attach Ebitengine's Android `EbitenView`; the direct path only showed the Android splash/default icon. The wrapper uses Ebitengine's generated mobile view classes from `ebitenmobile bind` and a small `MainActivity` that sets that view as the activity content.

The build pipeline generates disposable intermediate files under `.tmp/android-apparat-wrapper`, then writes the canonical signed APK to:

```text
releases/android/arm64/apparat/latest.apk
```

Android builds use OpenJDK 21 only (Eclipse Temurin preferred); Oracle JDK is prohibited by repository policy. The selected `JAVA_HOME` must expose `java`, `javac`, and `keytool`.

Windows builds use the repository-local Android SDK fallback and invoke `sdkmanager.bat`, `d8.bat`, and `apksigner.bat` alongside executable build tools. The wrapper creates its ignored development keystore before intermediate signing. A Windows-produced APK still requires separate authorized-device install and launch evidence.

The wrapper must preserve the existing Phase 5 gates: package ID `com.cjtrowbridge.apparat`, full-screen behavior without forced portrait orientation, `minSdkVersion=23`, `targetSdkVersion=30`, Android platform 35 packaging, `arm64-v8a`, v2/v3 signing, 16 KB native page alignment, app-private runtime storage, touch tab selection, and no dependency on `third_party/salvagecore`.

The launcher icon is generated from the tracked root `logo.svg` concept into `res/drawable/app_icon.xml` and referenced by `AndroidManifest.xml`. The build pipeline compiles tracked Android resources before linking the wrapper APK.

Phase 5 also includes a temporary Settings `Updates` fieldset rendered by the EbitenUI HUD. The `Check for update` button is now an EbitenUI control inside that fieldset, not a native Android overlay. The Go mobile package registers an `Updater` callback with `MainActivity`; tapping the HUD button calls that callback, and Java performs the platform-specific work: download the tracked GitHub `latest.apk`, compare its SHA-256 with the installed APK, open Android's per-app unknown-source permission screen only when needed, and launch the Android package installer through `UpdateApkProvider`. Java reports coarse status text back through the mobile binding so the HUD button can leave its `Checking...` state after the platform path reaches a result. `MainActivity` also runs the same check silently on startup and only surfaces it when an update path needs user action. This remains a dev-only bridge until a real Settings action, installed/latest version display, and signed update manifest replace it.
