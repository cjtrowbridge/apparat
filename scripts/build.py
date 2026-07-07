#!/usr/bin/env python3
"""Build Apparat into the canonical local release artifact path."""

from __future__ import annotations

import argparse
import os
import platform
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGETS = {
    "apparat": "./cmd/apparat",
    "apparatd": "./cmd/apparatd",
}


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
    return ROOT / "releases" / goos / goarch / f"latest{suffix}"


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--os", dest="goos", default=None, help="target GOOS; defaults to detected host")
    parser.add_argument("--arch", dest="goarch", default=None, help="target GOARCH; defaults to detected host")
    parser.add_argument("--target", choices=sorted(TARGETS), default="apparat", help="command target to build")
    parser.add_argument("--go", default=os.environ.get("GO", "go"), help="go executable")
    parser.add_argument("--print-path", action="store_true", help="print the expected artifact path without building")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    goos = args.goos or host_goos()
    goarch = args.goarch or host_goarch()
    output = artifact_path(goos, goarch, args.target)

    if args.print_path:
        print(output)
        return 0

    output.parent.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch

    command = [args.go, "build", "-trimpath", "-o", str(output), TARGETS[args.target]]
    print(f"building target={args.target} goos={goos} goarch={goarch} output={output}")
    subprocess.run(command, cwd=ROOT, env=env, check=True)
    print(output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
