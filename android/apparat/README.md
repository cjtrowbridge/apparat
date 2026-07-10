# Android Apparat Wrapper

This directory contains tracked Apparat-owned Android wrapper sources for the Phase 5 GUI APK.

The wrapper exists because direct `gomobile build` with `GoNativeActivity` can install and initialize Apparat runtime state, but it does not attach Ebitengine's Android `EbitenView`; the direct path only showed the Android splash/default icon. The wrapper uses Ebitengine's generated mobile view classes from `ebitenmobile bind` and a small `MainActivity` that sets that view as the activity content.

The build pipeline generates disposable intermediate files under `.tmp/android-apparat-wrapper`, then writes the canonical signed APK to:

```text
releases/android/arm64/apparat/latest.apk
```

The wrapper must preserve the existing Phase 5 gates: package ID `com.cjtrowbridge.apparat`, full-screen behavior without forced portrait orientation, `minSdkVersion=23`, `targetSdkVersion=30`, Android platform 35 packaging, `arm64-v8a`, v2/v3 signing, 16 KB native page alignment, app-private runtime storage, touch tab selection, and no dependency on `third_party/salvagecore`.

Phase 5 also includes a temporary Settings `Updates` fieldset rendered by the HUD, with a native Android `Check for update` button placed as a fixed Settings header action while Settings is active. The Java activity polls the mobile HUD bridge for the active tab and slot visibility, hides the native button outside Settings, places it from the stable HUD header slot rectangle, downloads the tracked GitHub `latest.apk`, compares its SHA-256 with the installed APK, opens Android's per-app unknown-source permission screen only when needed, and then launches the Android package installer through `UpdateApkProvider`. This is a dev-only bridge until a real Settings action, installed/latest version display, and signed update manifest replace it. Body scrolling is intentionally deferred during the HUD layout recovery checkpoint so the native button does not pretend to be a scrolled Ebitengine child.
