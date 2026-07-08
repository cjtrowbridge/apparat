#!/usr/bin/env python3
"""Build Apparat binaries into canonical local release artifact paths.

Canonical invocation from the repository root:

    python3 scripts/build.py

The default build produces both `apparat` and `apparatd`. The GUI target
requires native desktop development headers on Linux because Ebitengine uses
GLFW/X11 there.
"""

from __future__ import annotations

import argparse
import os
import platform
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


@dataclass(frozen=True)
class Target:
    package: str
    tags: tuple[str, ...] = ()


TARGETS = {
    "apparat": Target("./cmd/apparat", ("gui",)),
    "apparatd": Target("./cmd/apparatd"),
}
ALL_TARGETS = tuple(TARGETS)
ANDROID_APK_HELP = (
    "Android APK builds are not integrated yet. The Ebitengine mobile source "
    "framework is present through the Ebitengine submodule, and salvagecore "
    "contains a golang/mobile reference checkout, but Apparat still needs a "
    "pinned Android SDK/NDK/JDK toolchain, host-owned Android wrapper or AAR "
    "pipeline, manifest/activity lifecycle code, signing configuration, "
    "permissions, and device/emulator validation before the pipeline can emit "
    "releases/android/<arch>/apparat/latest.apk."
)


def host_goos() -> str:
    system = platform.system().lower()
    if system == "darwin":
        return "darwin"
    if system == "linux":
        return "linux"
    if system == "windows":
        return "windows"
    if system.startswith("msys") or system.startswith("cygwin"):
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
    suffix = ".exe" if goos == "windows" else ""
    return ROOT / "releases" / goos / goarch / target / f"latest{suffix}"


def selected_targets(target: str) -> tuple[str, ...]:
    if target == "all":
        return ALL_TARGETS
    return (target,)


def build_command(go: str, target: str, output: Path) -> list[str]:
    spec = TARGETS[target]
    command = [go, "build", "-trimpath"]
    if spec.tags:
        command.extend(["-tags", ",".join(spec.tags)])
    command.extend(["-o", str(output), spec.package])
    return command


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Build Apparat GUI and headless binaries into canonical release paths.",
        epilog=(
            "Examples:\n"
            "  python3 scripts/build.py\n"
            "  python3 scripts/build.py --target apparatd\n"
            "  python3 scripts/build.py --target apparat --print-path\n\n"
            "Outputs:\n"
            "  releases/<goos>/<goarch>/apparat/latest[.exe]\n"
            "  releases/<goos>/<goarch>/apparatd/latest[.exe]\n\n"
            "Linux GUI prerequisite packages include libx11-dev, libxcursor-dev, "
            "libxrandr-dev, libxinerama-dev, libxi-dev, libgl1-mesa-dev, "
            "libxxf86vm-dev, and libasound2-dev. Prefer `make build` so the "
            "repo-local Go cache settings are applied.\n\n"
            f"{ANDROID_APK_HELP}"
        ),
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("--os", dest="goos", default=None, help="target GOOS; defaults to detected host")
    parser.add_argument("--arch", dest="goarch", default=None, help="target GOARCH; defaults to detected host")
    parser.add_argument("--target", choices=[*sorted(TARGETS), "all"], default="all", help="command target to build")
    parser.add_argument("--go", default=os.environ.get("GO", "go"), help="go executable")
    parser.add_argument("--print-path", action="store_true", help="print the expected artifact path without building")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    goos = args.goos or host_goos()
    goarch = args.goarch or host_goarch()
    if goos == "android":
        print(ANDROID_APK_HELP, file=sys.stderr)
        return 2
    targets = selected_targets(args.target)

    if args.print_path:
        for target in targets:
            print(artifact_path(goos, goarch, target))
        return 0

    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch

    for target in targets:
        output = artifact_path(goos, goarch, target)
        output.parent.mkdir(parents=True, exist_ok=True)
        command = build_command(args.go, target, output)
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
