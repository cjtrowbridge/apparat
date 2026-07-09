# Android Apparat Wrapper

This directory contains tracked Apparat-owned Android wrapper sources for the Phase 5 GUI APK.

The wrapper exists because direct `gomobile build` with `GoNativeActivity` can install and initialize Apparat runtime state, but it does not attach Ebitengine's Android `EbitenView`; the Pixel only shows the Android splash/default icon. The wrapper uses Ebitengine's generated mobile view classes from `ebitenmobile bind` and a small `MainActivity` that sets that view as the activity content.

The build pipeline generates disposable intermediate files under `.tmp/android-apparat-wrapper`, then writes the canonical signed APK to:

```text
releases/android/arm64/apparat/latest.apk
```

The wrapper must preserve the existing Phase 5 gates: package ID `com.cjtrowbridge.apparat`, portrait phone orientation, `minSdkVersion=23`, `targetSdkVersion=30`, Android platform 35 packaging, `arm64-v8a`, v2/v3 signing, 16 KB native page alignment, app-private runtime storage, and no dependency on `third_party/salvagecore`.
