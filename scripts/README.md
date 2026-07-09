# Scripts

This directory contains repository-local automation used by developers and CI-like local checks.

Run scripts from the repository root unless a script explicitly says otherwise. Prefer Makefile targets when they exist because they provide the expected Go cache and tool settings.

## Inventory

- `build.py`: Builds canonical local release artifacts for `apparat`, `apparatd`, and the Android GUI APK.
- `check_code_file_lines.py`: Fails when included code files exceed the 400-line limit.
- `check_directory_docs.py`: Fails when source directories or scripts are missing required documentation.
- `regenerate_plan_indexes.py`: Validates plan files and regenerates plan indexes.
- `run_artifact.py`: Runs a built artifact while forwarding arguments after an optional `--`.

## Desktop Build Script

Canonical desktop build commands:

```bash
make build
python3 scripts/build.py
python3 scripts/build.py --target apparatd
python3 scripts/build.py --target apparat --print-path
```

`make build` is preferred because it applies repo-local Go cache settings. `python3 build.py` at the repository root is a compatibility wrapper that delegates to `python3 scripts/build.py`.

Desktop outputs:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

The `apparat` target is compiled with the `gui` build tag. On Linux, it requires native desktop development headers used by Ebitengine and GLFW, including `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libgl1-mesa-dev`, `libxxf86vm-dev`, and `libasound2-dev`.

The `apparatd` target avoids the GUI build tag and is the correct target for headless workers and display-free validation.

## Android Build Script

Canonical Android commands:

```bash
make check-android-build-env
make build-android
python3 scripts/build.py --check-android-env
python3 scripts/build.py --os android --arch arm64 --target apparat
python3 scripts/build.py --os android --arch arm64 --target apparat --print-path
```

Android output:

```text
releases/android/arm64/apparat/latest.apk
```

Phase 5 intentionally supports only `--os android --arch arm64 --target apparat`. Android `apparatd` builds fail with an explicit headless-out-of-scope message because Android headless needs a later Termux/service-worker strategy.

Android prerequisites:

- JDK 21, discovered through `JAVA_HOME` or the ignored repo-local `.tools/jdks/openjdk-21` path.
- Android SDK command-line tools under `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or ignored `.tools/android`.
- Android platform `android-35`.
- Android build-tools `35.0.0`.
- Android NDK `27.2.12479018`, discovered through `ANDROID_NDK_HOME` or the SDK `ndk/27.2.12479018` directory.
- Ebitengine `github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1` in the Go module cache.

`scripts/build.py --check-android-env` checks those prerequisites and prepares an ignored `.tools/bin/gomobile-apparat` helper if needed. That helper is a small local build-tool patch for the pinned Ebitengine gomobile scanner: the upstream tool checks for `github.com/ebitengine/gomobile/app` but its symbol-scanning regular expression only matches `golang.org/x` import paths. The patch broadens the scanner and does not fork application source.

Android build side effects:

- Creates or updates `.tools/bin/gomobile-apparat` when the helper is missing.
- Uses `.tmp/gomobile-apparat-src` as disposable patched-tool source.
- Writes the APK to `releases/android/arm64/apparat/latest.apk`.
- Does not read from or reference `third_party/salvagecore`.

## Verification Scripts

Run:

```bash
make check-code-size
make check-docs
make test-build
```

`make verify` runs all three checks along with Go formatting, Go tests, race tests, linting, and vulnerability scanning.

## Failure Modes

- Missing Go modules: rerun through `make build` or allow network access so Go can populate the module cache.
- Missing Linux GUI headers: install the native desktop development packages listed above, then rerun `make build`.
- Missing Android SDK/JDK/NDK: set `JAVA_HOME`, `ANDROID_HOME`/`ANDROID_SDK_ROOT`, and `ANDROID_NDK_HOME`, then rerun `make check-android-build-env`.
- Missing Android gomobile module: run `go mod download github.com/ebitengine/gomobile` with writable `GOMODCACHE`, then rerun the Android preflight.
- Missing `adb` or sandbox-blocked `adb`: APK builds can still pass, but install/launch validation must be performed from an environment where `adb devices`, `adb install`, and `adb logcat` work.
- Missing directory docs: add a `README.md` to the source directory reported by `check_directory_docs.py`.
- Missing script help: add argparse-style `--help` output before considering the script complete.
