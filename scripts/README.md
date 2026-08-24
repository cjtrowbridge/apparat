# Scripts

This directory contains repository-local automation used by developers and CI-like local checks.

Run scripts from the repository root unless a script explicitly says otherwise. Prefer Makefile targets when they exist because they provide the expected Go cache and tool settings.

## Inventory

- `build.py`: Builds canonical local release artifacts for `apparat`, `apparatd`, and the Android GUI APK.
- `build_orchestrator.py`: Implements the no-flag target detection/report/build loop used by `build.py`.
- `generate_gitingest.py`: Creates a project-focused GitIngest digest and provenance manifest for deep research.
- `android_wrapper.py`: Assembles the Android GUI wrapper APK from Ebitengine mobile binding output and tracked Apparat Android sources.
- `check_code_file_lines.py`: Fails when host-owned code files exceed the 400-line limit; Git submodules, third-party sources, generated releases, plans, and journals are excluded.
- `check_directory_docs.py`: Fails when source directories or scripts are missing required documentation.
- `regenerate_plan_indexes.py`: Validates plan files and regenerates plan indexes.
- `run_artifact.py`: Runs a built artifact while forwarding arguments after an optional `--`.
- `run_build_pipeline.ps1`: Host-owned Windows VS Code/local prerequisite boundary that verifies the selected Agentic Pipelines submodule and Python before delegating once to the canonical build script.

## Desktop Build Script

Canonical build commands:

```bash
make build
python3 scripts/build.py
```

`make build` is preferred because it applies repo-local Go cache settings. `python3 build.py` at the repository root is a compatibility wrapper that delegates to `python3 scripts/build.py`. The build script intentionally has one no-flag entry point: it detects the host, prints possible and impossible targets with reasons, and builds every possible target.

VS Code's local `Apparat: Build Pipeline` task and primary play action enter `run_build_pipeline.ps1`. The wrapper installs nothing and does not invoke inference: it checks that `agentic-pipelines/AGENTS.md` and Python are available, then runs `scripts/build.py` exactly once and preserves its exit status. Run it with `-Help` for usage. A missing submodule is repaired explicitly with `git submodule update --init --recursive agentic-pipelines`. Linux and macOS VS Code entrypoints remain unclaimed until they can be tested on those hosts; their existing direct `make build` and `python3 scripts/build.py` commands remain supported.

## GitIngest Digest

Generate a prompt-friendly digest separately from application builds:

```bash
python3 scripts/generate_gitingest.py
```

The command bootstraps the pinned GitIngest 0.3.1 release in the ignored `.tools/gitingest-venv` environment when needed. It atomically replaces two ignored root-level outputs:

- `gitingest.txt`: the default project-focused corpus, containing Apparat-owned code, tests, documentation, plans, governance, and build files. It excludes recursive third-party source, release artifacts, and tool/cache directories.
- `gitingest.manifest.json`: schema-versioned provenance containing the commit, branch, dirty state, sanitized remotes, recursive submodule revisions, `go.mod`/`go.sum` hashes, generator settings, and the digest SHA-256.

Use `python3 scripts/generate_gitingest.py --include-submodules` only for a targeted review that genuinely requires full third-party source bodies. The default does not recursively ingest submodule contents; their exact revisions remain in the manifest for later pinned-source retrieval. Use `--output` or `--manifest` to choose alternative artifact paths. Its first run needs network access; run `git submodule update --init --recursive` if generation reports a missing submodule.

Desktop outputs:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

The `apparat` target is compiled with the `gui` build tag. On Linux, it requires native desktop development headers used by Ebitengine and GLFW, including `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libgl1-mesa-dev`, `libxxf86vm-dev`, and `libasound2-dev`.

The `apparatd` target avoids the GUI build tag and is the correct target for headless workers and display-free validation.

## Android Build Script

Android is part of the same no-flag build pass:

```bash
make build
python3 scripts/build.py
```

Android output:

```text
releases/android/arm64/apparat/latest.apk
```

Phase 5 intentionally supports only the Android arm64 GUI APK. Android `apparatd` is reported as impossible with an explicit headless-out-of-scope message because Android headless needs a later Termux/service-worker strategy.

Android prerequisites:

- OpenJDK 21 (Eclipse Temurin preferred; Oracle JDK is prohibited), discovered through `JAVA_HOME` or the ignored repo-local `.tools/jdks/openjdk-21` path. The selected `JAVA_HOME` must provide `java`, `javac`, and `keytool`.
- Android SDK command-line tools under `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or ignored `.tools/android`.
- Android platform `android-35`.
- Android build-tools `35.0.0`.
- Android NDK `27.2.12479018`, discovered through `ANDROID_NDK_HOME` or the SDK `ndk/27.2.12479018` directory.
- Ebitengine `github.com/ebitengine/gomobile v0.0.0-20250923094054-ea854a63cce1` in the Go module cache.

`scripts/build.py` checks those prerequisites while producing the target report and prepares an ignored `.tools/bin/gomobile-apparat` helper if needed. That helper is a small local build-tool patch for the pinned Ebitengine gomobile scanner: the upstream tool checks for `github.com/ebitengine/gomobile/app` but its symbol-scanning regular expression only matches `golang.org/x` import paths. The patch broadens the scanner, supports local module replacement for wrapper binding, and synthesizes `minSdkVersion=23` plus `targetSdkVersion=30` while compiling/package-building against Android platform 35. The pipeline temporarily applies an Android Ebitengine display-metric guard during `gomobile bind`, restores the Ebitengine checkout afterward, binds `cmd/apparatmobile`, generates Ebitengine mobile view classes, compiles tracked `android/apparat` wrapper sources and resources, then zipaligns/signs the APK with a generated debug keystore. It does not fork application source.

The pipeline creates that ignored development keystore before wrapper assembly, because the wrapper signs its intermediate APK before final alignment, signing, and verification. It never writes the keystore or its password to tracked files.

On Windows, the pipeline resolves Android's actual launcher names: `sdkmanager.bat`, `apksigner.bat`, and `d8.bat`, alongside `aapt2.exe`, `zipalign.exe`, and `adb.exe`. It constructs child-process `PATH` values using the host path separator so the generated `gobind.exe` remains discoverable by Gomobile, and pins the patched helper plus generated binding module to Go 1.26.4 even when Gomobile's own module declares an older Go version. Both `python scripts/build.py` and `python -m scripts.build` work from the repository root; use the normal no-flag `make build`/`python3 scripts/build.py` workflow on platforms where those commands are available.

For local machine-specific paths, copy `build_environment.sample.py` to ignored `build_environment.py` and update environment values there. The build script loads that file opportunistically before target detection.

Android build side effects:

- Creates or updates `.tools/bin/gomobile-apparat` when the helper is missing.
- Uses `.tmp/gomobile-apparat-src` as disposable patched-tool source.
- Temporarily patches and restores `third_party/game/ebiten/internal/ui/ui_android.go` during Android binding so clean checkouts reproduce the Pixel display-metric guard without carrying a dirty submodule.
- Writes the APK to `releases/android/arm64/apparat/latest.apk`.
- Removes temporary unsigned/aligned APKs and optional signing sidecars after the final signed APK is verified.
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
- Missing GitIngest: allow network access for `python3 scripts/generate_gitingest.py` to create the pinned `.tools/gitingest-venv`; initialize submodules recursively before rerunning if provenance collection reports a missing submodule.
- Missing Linux GUI headers: install the native desktop development packages listed above, then rerun `make build`.
- Missing Android SDK/OpenJDK/NDK: set `JAVA_HOME`, `ANDROID_HOME`/`ANDROID_SDK_ROOT`, and `ANDROID_NDK_HOME`, or configure ignored `build_environment.py`, then rerun `make build`. Use an OpenJDK 21 distribution, never Oracle JDK.
- Missing Android gomobile module: run `go mod download github.com/ebitengine/gomobile` with writable `GOMODCACHE`, then rerun the Android preflight.
- Missing `adb` or sandbox-blocked `adb`: APK builds can still pass, but install/launch validation must be performed from an environment where `adb devices`, `adb install`, and `adb logcat` work.
- Missing directory docs: add a `README.md` to the source directory reported by `check_directory_docs.py`.
- Missing script help: add argparse-style `--help` output before considering the script complete.
