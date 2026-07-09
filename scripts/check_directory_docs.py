#!/usr/bin/env python3
"""Fail when tracked source directories or scripts are missing documentation."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

SOURCE_SUFFIXES = {".go", ".py", ".sh"}
EXCLUDED_DIR_PARTS = {".git", ".tools", ".tmp", "__pycache__", "third_party", "releases", "journal", "plans", "downtime"}
SCRIPT_DIR = Path("scripts")


def tracked_files(root: Path) -> list[Path]:
    result = subprocess.run(["git", "ls-files", "--cached", "--others", "--exclude-standard"], cwd=root, check=True, capture_output=True, text=True)
    return [Path(line) for line in result.stdout.splitlines() if line]


def is_excluded(path: Path) -> bool:
    return any(part in EXCLUDED_DIR_PARTS for part in path.parts)


def source_directories(files: list[Path]) -> set[Path]:
    dirs: set[Path] = set()
    for path in files:
        if is_excluded(path) or path.suffix not in SOURCE_SUFFIXES:
            continue
        if path.name == "__init__.py":
            continue
        dirs.add(path.parent)
    return dirs


def missing_directory_readmes(root: Path, files: list[Path]) -> list[Path]:
    missing: list[Path] = []
    for directory in sorted(source_directories(files)):
        if not (root / directory / "README.md").is_file():
            missing.append(directory)
    return missing


def script_files(files: list[Path]) -> list[Path]:
    scripts: list[Path] = []
    for path in files:
        if path.parent == SCRIPT_DIR and path.suffix == ".py" and path.name != "__init__.py":
            scripts.append(path)
    return sorted(scripts)


def undocumented_scripts(root: Path, files: list[Path]) -> list[Path]:
    readme = root / SCRIPT_DIR / "README.md"
    if not readme.is_file():
        return script_files(files)
    text = readme.read_text(encoding="utf-8")
    return [path for path in script_files(files) if path.name not in text and path.as_posix() not in text]


def scripts_without_help(root: Path, files: list[Path]) -> list[Path]:
    missing: list[Path] = []
    for path in script_files(files):
        result = subprocess.run([sys.executable, str(root / path), "--help"], cwd=root, capture_output=True, text=True)
        output = f"{result.stdout}\n{result.stderr}".lower()
        if result.returncode != 0 or "usage:" not in output:
            missing.append(path)
    return missing


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path.cwd(), help="repository root")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    root = args.root.resolve()
    files = tracked_files(root)
    failures: list[str] = []

    for directory in missing_directory_readmes(root, files):
        failures.append(f"missing README.md for source directory: {directory}")
    for path in undocumented_scripts(root, files):
        failures.append(f"script missing from scripts/README.md: {path}")
    for path in scripts_without_help(root, files):
        failures.append(f"script does not provide --help usage: {path}")

    if failures:
        print("documentation check failed:")
        for failure in failures:
            print(f"- {failure}")
        return 1
    print("documentation check passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
