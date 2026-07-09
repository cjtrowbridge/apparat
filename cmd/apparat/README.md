# `apparat`

`cmd/apparat` is the GUI console entrypoint.

It loads configuration with binary name `apparat`, creates the shared application runtime, supports `--smoke-test` and `--doctor`, and starts the GUI adapter during normal execution.

At startup it prints the selected runtime root and `last_run.log` path before entering the Ebitengine loop. By default on Linux, the GUI runtime root is `~/.local/share/apparat/apparat` unless `XDG_DATA_HOME`, `APPARAT_RUNTIME_DIR`, or `--runtime-dir` overrides it.

The desktop release pipeline builds this command with the `gui` build tag into:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
```

The Android release pipeline does not use this desktop command as the render entrypoint. It binds `cmd/apparatmobile` and the tracked `android/apparat` wrapper into:

```text
releases/android/arm64/apparat/latest.apk
```

Legacy direct-gomobile packaging files in this directory:

- `AndroidManifest.xml`: owns the package ID `com.cjtrowbridge.apparat`, app label `Apparat`, launcher `GoNativeActivity`, portrait phone orientation, debug metadata, and `INTERNET` permission for HTTPS over external WireGuard/local networks.
- `gomobile_app.go`: is Android-only and references `github.com/ebitengine/gomobile/app` so the Ebitengine gomobile builder recognizes the package as a mobile app.

Android broad storage, microphone, VPN-service, and app-managed WireGuard permissions are intentionally absent until those behaviors are implemented and validated. The build helper patches gomobile so the APK declares `minSdkVersion=23` and `targetSdkVersion=30`, preserves modern manifest attributes, signs with a generated debug keystore, and links the native library with 16 KB page alignment; release signing and store packaging are future work.

On Linux, normal GUI builds require Ebitengine native desktop dependencies such as X11, cursor, randr, xinerama, xi, OpenGL, xxf86vm, and ALSA development headers.
