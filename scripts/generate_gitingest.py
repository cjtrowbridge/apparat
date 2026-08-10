#!/usr/bin/env python3
"""Generate the root GitIngest digest from this checkout and its submodules."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VENV = ROOT / ".tools" / "gitingest-venv"
OUTPUT = ROOT / "gitingest.txt"


class GitIngestError(RuntimeError):
    """A user-facing GitIngest setup or generation failure."""


def venv_python() -> Path:
    return VENV / ("Scripts/python.exe" if sys.platform == "win32" else "bin/python")


def gitingest_command() -> Path:
    return VENV / ("Scripts/gitingest.exe" if sys.platform == "win32" else "bin/gitingest")


def install_command() -> list[str]:
    return [str(venv_python()), "-m", "pip", "install", "--upgrade", "gitingest>=0.2.0"]


def generate_command() -> list[str]:
    return [str(gitingest_command()), str(ROOT), "--include-submodules", "--output", str(OUTPUT)]


def ensure_gitingest() -> None:
    if gitingest_command().is_file():
        return
    try:
        subprocess.run([sys.executable, "-m", "venv", str(VENV)], cwd=ROOT, check=True)
        subprocess.run(install_command(), cwd=ROOT, check=True)
    except (OSError, subprocess.CalledProcessError) as error:
        raise GitIngestError(
            "could not install GitIngest into .tools/gitingest-venv; "
            "ensure Python can create virtual environments and allow network access, then rerun the build"
        ) from error


def generate() -> Path:
    ensure_gitingest()
    try:
        subprocess.run(generate_command(), cwd=ROOT, check=True)
    except (OSError, subprocess.CalledProcessError) as error:
        raise GitIngestError(
            "could not generate gitingest.txt; ensure every required submodule is initialized with "
            "`git submodule update --init --recursive`, then rerun the build"
        ) from error
    if not OUTPUT.is_file():
        raise GitIngestError("GitIngest completed without producing gitingest.txt")
    return OUTPUT


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    parse_args(sys.argv[1:] if argv is None else argv)
    try:
        output = generate()
    except GitIngestError as error:
        print(error, file=sys.stderr)
        return 1
    print(f"GitIngest digest: {output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
