# apparat

Host-owned source for android/apparat/src/com/cjtrowbridge/apparat.

## Source inventory

- `MainActivity.java`: Hosts the Android Ebitengine view inside modern system-bar/cutout safe bounds and provides bounded APK-update UI.
- `UpdateApkProvider.java`: Exposes the cached update APK to Android package installation via a content provider.

Android safe-area color comes from the GUI renderer through `BackgroundARGB` in the mobile bridge. The current fixed dark palette is authoritative; runtime theme switching is not implemented. The native parent and navigation background use that color, with light navigation icons and no automatic contrast scrim. Insets and touch bounds are unchanged. Verify matching background pixels and readable buttons in portrait/landscape on an authorized device; system/OEM styling on other devices requires separate evidence.

Updater rejection codes distinguish invalid package, mismatched signer and non-increasing version. The signer is checked before version messaging; none of these paths opens the installer or clears app data.
