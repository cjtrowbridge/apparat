#!/usr/bin/env python3
"""Build Apparat artifacts into canonical local release paths.
Canonical invocation from the repository root:
    python3 scripts/build.py
The default desktop build produces both `apparat` and `apparatd`. Android
builds are GUI-only and produce `releases/android/arm64/apparat/latest.apk`.
"""
from __future__ import annotations
import os
import platform
import shutil
import stat
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
try:
    from scripts import android_wrapper
except ModuleNotFoundError:
    import android_wrapper
ROOT = Path(__file__).resolve().parents[1]
ANDROID_API = "35"
ANDROID_MIN_API = "23"
ANDROID_TARGET_API = "30"
ANDROID_PAGE_SIZE = "0x4000"
ANDROID_BUILD_TOOLS = "35.0.0"
ANDROID_NDK = "27.2.12479018"
GOMOBILE_VERSION = "v0.0.0-20250923094054-ea854a63cce1"
PATCHED_GOMOBILE = ROOT / ".tools" / "bin" / "gomobile-apparat"

@dataclass(frozen=True)
class Target:
    package: str
    tags: tuple[str, ...] = ()
@dataclass(frozen=True)
class AndroidToolchain:
    sdk_root: Path
    ndk_root: Path
    java_home: Path | None
    java: Path
    gomobile: Path
    adb: Path | None
TARGETS = {
    "apparat": Target("./cmd/apparat", ("gui",)),
    "apparatd": Target("./cmd/apparatd"),
}
ALL_TARGETS = tuple(TARGETS)
ANDROID_ARCHES = {"arm64": "android/arm64"}
ANDROID_NATIVE_CODES = {"arm64": "arm64-v8a"}
ANDROID_HEADLESS_HELP = "Android Phase 5 builds only the GUI `apparat` APK; use Linux/Termux for headless `apparatd`."
class BuildError(RuntimeError):
    """A build-time configuration error with user-facing guidance."""
def host_goos() -> str:
    system = platform.system().lower()
    if system == "darwin":
        return "darwin"
    if system == "linux":
        return "linux"
    if system == "windows" or system.startswith(("msys", "cygwin")):
        return "windows"
    raise ValueError(f"unsupported host OS: {platform.system()}")
def host_goarch() -> str:
    machine = platform.machine().lower()
    aliases = {
        "x86_64": "amd64",
        "amd64": "amd64",
        "aarch64": "arm64",
        "arm64": "arm64",
        "armv7l": "arm",
        "armv6l": "arm",
        "i386": "386",
        "i686": "386",
    }
    if machine not in aliases:
        raise ValueError(f"unsupported host architecture: {platform.machine()}")
    return aliases[machine]
def artifact_path(goos: str, goarch: str, target: str) -> Path:
    if goos == "android":
        return ROOT / "releases" / "android" / goarch / target / "latest.apk"
    suffix = ".exe" if goos == "windows" else ""
    return ROOT / "releases" / goos / goarch / target / f"latest{suffix}"
def selected_targets(target: str, goos: str | None = None) -> tuple[str, ...]:
    if target == "all" and goos == "android":
        return ("apparat",)
    if target == "all":
        return ALL_TARGETS
    return (target,)
def validate_target(goos: str, goarch: str, target: str) -> None:
    if goos != "android":
        return
    if goarch not in ANDROID_ARCHES:
        raise BuildError(f"unsupported Android arch {goarch!r}; supported: {', '.join(sorted(ANDROID_ARCHES))}")
    if target != "apparat":
        raise BuildError(ANDROID_HEADLESS_HELP)
def desktop_build_command(go: str, target: str, output: Path) -> list[str]:
    spec = TARGETS[target]
    command = [go, "build", "-trimpath"]
    if spec.tags:
        command.extend(["-tags", ",".join(spec.tags)])
    command.extend(["-o", str(output), spec.package])
    return command
def android_badging_expectations(goarch: str) -> tuple[str, ...]:
    return (
        f"minSdkVersion:'{ANDROID_MIN_API}'",
        f"targetSdkVersion:'{ANDROID_TARGET_API}'",
        f"native-code: '{ANDROID_NATIVE_CODES[goarch]}'",
    )
def default_sdk_root() -> Path:
    env = os.environ.get("ANDROID_HOME") or os.environ.get("ANDROID_SDK_ROOT")
    return Path(env).expanduser() if env else ROOT / ".tools" / "android"
def default_java_home() -> Path | None:
    if os.environ.get("JAVA_HOME"):
        return Path(os.environ["JAVA_HOME"]).expanduser()
    local = ROOT / ".tools" / "jdks" / "openjdk-21" / "usr" / "lib" / "jvm" / "java-21-openjdk-amd64"
    return local if local.exists() else None
def find_executable(name: str, extra_dirs: list[Path] | None = None) -> Path | None:
    for directory in extra_dirs or []:
        candidate = directory / name
        if candidate.is_file() and os.access(candidate, os.X_OK):
            return candidate
    found = shutil.which(name)
    return Path(found) if found else None
def android_tool_env(toolchain: AndroidToolchain) -> dict[str, str]:
    env = os.environ.copy()
    path_parts = [
        ROOT / ".tools" / "bin",
        toolchain.sdk_root / "platform-tools",
        toolchain.sdk_root / "build-tools" / ANDROID_BUILD_TOOLS,
        toolchain.sdk_root / "cmdline-tools" / "latest" / "bin",
    ]
    if toolchain.java_home:
        path_parts.insert(0, toolchain.java_home / "bin")
        env["JAVA_HOME"] = str(toolchain.java_home)
    env["ANDROID_HOME"] = str(toolchain.sdk_root)
    env["ANDROID_SDK_ROOT"] = str(toolchain.sdk_root)
    env["ANDROID_NDK_HOME"] = str(toolchain.ndk_root)
    env["PATH"] = os.pathsep.join([str(path) for path in path_parts] + [env.get("PATH", "")])
    return env
def check_path(path: Path, label: str, failures: list[str]) -> None:
    if not path.exists():
        failures.append(f"missing {label}: {path}")
def resolve_android_toolchain(go: str) -> tuple[AndroidToolchain | None, list[str], list[str]]:
    failures: list[str] = []
    warnings: list[str] = []
    sdk_root = default_sdk_root()
    ndk_root = Path(os.environ.get("ANDROID_NDK_HOME", sdk_root / "ndk" / ANDROID_NDK)).expanduser()
    java_home = default_java_home()
    java_dirs = [java_home / "bin"] if java_home else []
    java = find_executable("java", java_dirs)
    check_path(sdk_root / "cmdline-tools" / "latest" / "bin" / executable("sdkmanager"), "Android SDK command-line tools", failures)
    check_path(sdk_root / "platforms" / f"android-{ANDROID_API}", f"Android platform android-{ANDROID_API}", failures)
    check_path(sdk_root / "build-tools" / ANDROID_BUILD_TOOLS / executable("aapt2"), f"Android build-tools {ANDROID_BUILD_TOOLS}", failures)
    check_path(sdk_root / "build-tools" / ANDROID_BUILD_TOOLS / executable("apksigner"), f"Android apksigner {ANDROID_BUILD_TOOLS}", failures)
    check_path(sdk_root / "build-tools" / ANDROID_BUILD_TOOLS / executable("zipalign"), f"Android zipalign {ANDROID_BUILD_TOOLS}", failures)
    check_path(ndk_root, f"Android NDK {ANDROID_NDK}", failures)
    if not java:
        failures.append("missing Java runtime; set JAVA_HOME to JDK 21 or install Java on PATH")
    if platform.system().lower() != "linux":
        warnings.append("Android build host is unvalidated on this OS; Linux is the first evidence-producing host")
    adb = find_executable("adb", [sdk_root / "platform-tools"])
    if not adb:
        warnings.append("adb not found; APK build can run, but install/launch validation will be skipped")
    gomobile = ensure_patched_gomobile(go, failures)
    if failures or not java or not gomobile:
        return None, failures, warnings
    return AndroidToolchain(sdk_root, ndk_root, java_home, java, gomobile, adb), failures, warnings
def executable(name: str) -> str:
    return f"{name}.exe" if host_goos() == "windows" else name
def module_cache_dir(go: str) -> Path:
    result = subprocess.run([go, "env", "GOMODCACHE"], cwd=ROOT, check=True, capture_output=True, text=True)
    return Path(result.stdout.strip())
def ensure_patched_gomobile(go: str, failures: list[str]) -> Path | None:
    env_tool = os.environ.get("GOMOBILE")
    if env_tool:
        return Path(env_tool).expanduser()
    try:
        source = module_cache_dir(go) / f"github.com/ebitengine/gomobile@{GOMOBILE_VERSION}"
    except (subprocess.CalledProcessError, FileNotFoundError) as error:
        failures.append(f"cannot locate Go module cache for gomobile: {error}")
        return None
    if not source.exists():
        failures.append(f"missing Ebitengine gomobile module source: {source}")
        failures.append("run `go mod download github.com/ebitengine/gomobile` or allow a networked Android build once")
        return None
    try:
        build_patched_gomobile(go, source)
    except (OSError, subprocess.CalledProcessError, RuntimeError) as error:
        failures.append(f"failed to prepare patched Ebitengine gomobile tool: {error}")
        return None
    return PATCHED_GOMOBILE
def build_patched_gomobile(go: str, source: Path) -> None:
    temp = ROOT / ".tmp" / "gomobile-apparat-src"
    if temp.exists():
        make_writable(temp)
        shutil.rmtree(temp)
    shutil.copytree(source, temp)
    make_writable(temp)
    patch_file(temp / "cmd" / "gomobile" / "build.go", {
        r"`[0-9a-f]{8} t _?(?:.*/vendor/)?(golang.org/x.*/[^.]*)`":
        r"`[0-9a-f]{8} t _?(?:.*/vendor/)?((?:golang.org/x|github.com/ebitengine/gomobile).*/[^.]*)`",
    })
    patch_file(temp / "internal" / "binres" / "sdk.go", {
        "const MinSDK = 16": f"const MinSDK = {ANDROID_MIN_API}",
        "sdkpath.AndroidAPIPath(MinSDK)": f"sdkpath.AndroidAPIPath({ANDROID_API})",
    })
    patch_file(temp / "internal" / "binres" / "binres.go", {
        'Local: "platformBuildVersionCode",\n\t\t\t\t\t\t},\n\t\t\t\t\t\tValue: "16",':
        f'Local: "platformBuildVersionCode",\n\t\t\t\t\t\t}},\n\t\t\t\t\t\tValue: "{ANDROID_API}",',
        'Local: "platformBuildVersionName",\n\t\t\t\t\t\t},\n\t\t\t\t\t\tValue: "4.1.2-1425332",':
        f'Local: "platformBuildVersionName",\n\t\t\t\t\t\t}},\n\t\t\t\t\t\tValue: "{ANDROID_API}",',
        'Local: "minSdkVersion",\n\t\t\t\t\t\t\t\t},\n\t\t\t\t\t\t\t\tValue: fmt.Sprintf("%v", MinSDK),\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t},':
        f'Local: "minSdkVersion",\n\t\t\t\t\t\t\t\t}},\n\t\t\t\t\t\t\t\tValue: fmt.Sprintf("%v", MinSDK),\n\t\t\t\t\t\t\t}},\n\t\t\t\t\t\t\txml.Attr{{\n\t\t\t\t\t\t\t\tName: xml.Name{{\n\t\t\t\t\t\t\t\t\tSpace: androidSchema,\n\t\t\t\t\t\t\t\t\tLocal: "targetSdkVersion",\n\t\t\t\t\t\t\t\t}},\n\t\t\t\t\t\t\t\tValue: "{ANDROID_TARGET_API}",\n\t\t\t\t\t\t\t}},\n\t\t\t\t\t\t}},',
    })
    patch_file(temp / "cmd" / "gomobile" / "bind.go", {
        "if f == nil {\n\t\t\treturn nil\n\t\t}":
        f'if f == nil {{\\n\\t\\t\\t_, err := io.WriteString(w, `module gomobilebind\\n\\ngo 1.26.4\\n\\nrequire (\\n\\tgithub.com/cjtrowbridge/apparat v0.0.0\\n\\tgithub.com/ebitengine/gomobile {GOMOBILE_VERSION}\\n\\tgithub.com/hajimehoshi/ebiten/v2 v2.9.9\\n)\\n\\nreplace github.com/cjtrowbridge/apparat => {ROOT}\\nreplace github.com/ebitengine/gomobile => {temp}\\nreplace github.com/hajimehoshi/ebiten/v2 => {ROOT / "third_party" / "game" / "ebiten"}\\n`)\\n\\t\\t\\treturn err\\n\\t\\t}}'.replace("\\n", "\n").replace("\\t", "\t"),
    })
    PATCHED_GOMOBILE.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run([go, "build", "-o", str(PATCHED_GOMOBILE), "./cmd/gomobile"], cwd=temp, check=True)
    subprocess.run([go, "build", "-o", str(PATCHED_GOMOBILE.parent / executable("gobind")), "./cmd/gobind"], cwd=temp, check=True)

def patch_file(path: Path, replacements: dict[str, str]) -> None:
    text = path.read_text(encoding="utf-8")
    for old, new in replacements.items():
        if old not in text:
            raise RuntimeError(f"expected patch target not found in {path}")
        text = text.replace(old, new)
    path.write_text(text, encoding="utf-8")
def make_writable(path: Path) -> None:
    path.chmod(path.stat().st_mode | stat.S_IWUSR)
    for item in path.rglob("*"):
        try:
            item.chmod(item.stat().st_mode | stat.S_IWUSR)
        except OSError:
            pass
def main(argv: list[str] | None = None) -> int:
    try:
        from scripts import build_orchestrator
    except ModuleNotFoundError:
        import build_orchestrator
    return build_orchestrator.main(sys.argv[1:] if argv is None else argv)
def check_android_env(go: str) -> int:
    toolchain, failures, warnings = resolve_android_toolchain(go)
    for warning in warnings:
        print(f"warning: {warning}", file=sys.stderr)
    if failures or not toolchain:
        print("Android build environment check failed:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    print(f"Android SDK: {toolchain.sdk_root}")
    print(f"Android NDK: {toolchain.ndk_root}")
    print(f"Java: {toolchain.java}")
    print(f"gomobile: {toolchain.gomobile}")
    print(f"adb: {toolchain.adb or 'not found; install/launch validation unavailable'}")
    return 0
def build_android(go: str, goarch: str) -> int:
    toolchain, failures, warnings = resolve_android_toolchain(go)
    for warning in warnings:
        print(f"warning: {warning}", file=sys.stderr)
    if failures or not toolchain:
        print("Android build prerequisites are incomplete:", file=sys.stderr)
        for failure in failures:
            print(f"- {failure}", file=sys.stderr)
        return 1
    output = artifact_path("android", goarch, "apparat")
    output.parent.mkdir(parents=True, exist_ok=True)
    print(f"building target=apparat goos=android goarch={goarch} output={output}")
    try:
        android_wrapper.build_wrapper_apk(toolchain, go, goarch, output, {
            "api": ANDROID_API, "min_api": ANDROID_MIN_API,
            "target_api": ANDROID_TARGET_API, "build_tools": ANDROID_BUILD_TOOLS,
            "page_size": ANDROID_PAGE_SIZE,
        })
    except subprocess.CalledProcessError as error:
        print(f"Android APK build failed exit={error.returncode}", file=sys.stderr)
        print("Run `python3 scripts/build.py` to see prerequisite details for every target.", file=sys.stderr)
        return error.returncode
    try:
        sign_android_apk(toolchain, output)
        validate_android_apk(toolchain, output, goarch)
    except BuildError as error:
        print(error, file=sys.stderr)
        return 1
    print(output)
    return 0
def validate_android_apk(toolchain: AndroidToolchain, output: Path, goarch: str) -> None:
    aapt2 = toolchain.sdk_root / "build-tools" / ANDROID_BUILD_TOOLS / executable("aapt2")
    result = subprocess.run([str(aapt2), "dump", "badging", str(output)], check=True, capture_output=True, text=True)
    missing = [value for value in android_badging_expectations(goarch) if value not in result.stdout]
    if missing:
        raise BuildError(f"Android APK metadata validation failed; missing: {', '.join(missing)}")
    apksigner = toolchain.sdk_root / "build-tools" / ANDROID_BUILD_TOOLS / executable("apksigner")
    subprocess.run([str(apksigner), "verify", "--verbose", str(output)], check=True, capture_output=True, text=True)
def sign_android_apk(toolchain: AndroidToolchain, output: Path) -> None:
    build_tools = toolchain.sdk_root / "build-tools" / ANDROID_BUILD_TOOLS
    keystore = ROOT / ".tools" / "android" / "debug.keystore"
    unsigned = output.with_suffix(".unsigned.apk")
    aligned = output.with_suffix(".aligned.apk")
    ensure_debug_keystore(toolchain, keystore)
    if unsigned.exists():
        unsigned.unlink()
    if aligned.exists():
        aligned.unlink()
    output.replace(unsigned)
    subprocess.run([str(build_tools / executable("zipalign")), "-p", "-f", "4", str(unsigned), str(aligned)], check=True)
    subprocess.run([
        str(build_tools / executable("apksigner")), "sign", "--ks", str(keystore),
        "--ks-pass", "pass:android", "--key-pass", "pass:android", "--out", str(output), str(aligned),
    ], check=True)
    output.with_name(output.name + ".idsig").unlink(missing_ok=True)
    unsigned.unlink(missing_ok=True)
    aligned.unlink(missing_ok=True)
def ensure_debug_keystore(toolchain: AndroidToolchain, keystore: Path) -> None:
    if keystore.exists():
        return
    keytool = find_executable("keytool", [toolchain.java_home / "bin"] if toolchain.java_home else [])
    if not keytool:
        raise BuildError("missing JDK keytool for Android debug keystore generation")
    keystore.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run([
        str(keytool), "-genkeypair", "-v", "-keystore", str(keystore),
        "-storepass", "android", "-keypass", "android", "-alias", "apparat-debug",
        "-keyalg", "RSA", "-keysize", "2048", "-validity", "10000",
        "-dname", "CN=Apparat Debug,O=Apparat,C=US",
    ], check=True)
def build_desktop(go: str, goos: str, goarch: str, targets: tuple[str, ...]) -> int:
    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch
    for target in targets:
        output = artifact_path(goos, goarch, target)
        output.parent.mkdir(parents=True, exist_ok=True)
        command = desktop_build_command(go, target, output)
        print(f"building target={target} goos={goos} goarch={goarch} output={output}")
        try:
            subprocess.run(command, cwd=ROOT, env=env, check=True)
        except subprocess.CalledProcessError as error:
            print_build_failure_help(target, error)
            return error.returncode
        print(output)
    return 0
def print_build_failure_help(target: str, error: subprocess.CalledProcessError) -> None:
    print(f"build failed for target={target} exit={error.returncode}", file=sys.stderr)
    if target == "apparat":
        print(
            "The GUI target is compiled with `-tags gui`. On Linux it requires native "
            "desktop development headers. Install packages such as libx11-dev, "
            "libxcursor-dev, libxrandr-dev, libxinerama-dev, libxi-dev, "
            "libgl1-mesa-dev, libxxf86vm-dev, and libasound2-dev, then rerun "
            "`make build`.",
            file=sys.stderr,
        )
    print("For usage details run: python3 scripts/build.py --help", file=sys.stderr)
if __name__ == "__main__":
    raise SystemExit(main())
