# Android Resources

This directory contains tracked Android resources used by the Apparat wrapper.

- `drawable/app_icon.xml`: launcher icon vector matching the root `logo.svg` path, stroke and transparent hub.

The wrapper build compiles this directory with `aapt2 compile --dir android/apparat/res` and links the compiled resources into `releases/android/arm64/apparat/latest.apk`.
