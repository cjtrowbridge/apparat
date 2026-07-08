#!/usr/bin/env python3
"""Fail when code files exceed the Apparat line-count limit."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


DEFAULT_LIMIT = 400
CODE_SUFFIXES = {".go", ".py", ".sh", ".yaml", ".yml", ".json"}
EXCLUDED_DIRS = {
    ".git",
    ".tools",
    "__pycache__",
    "third_party",
    "releases",
    "journal",
    "plans",
    "downtime",
}
EXCLUDED_FILES = {"go.sum"}


def is_code_file(path: Path) -> bool:
    return path.suffix in CODE_SUFFIXES and path.name not in EXCLUDED_FILES


def is_excluded(path: Path) -> bool:
    return any(part in EXCLUDED_DIRS for part in path.parts)


def iter_code_files(root: Path) -> list[Path]:
    files: list[Path] = []
    for path in root.rglob("*"):
        if path.is_dir() or is_excluded(path.relative_to(root)):
            continue
        if is_code_file(path):
            files.append(path)
    return sorted(files)


def line_count(path: Path) -> int:
    with path.open("r", encoding="utf-8", errors="replace") as handle:
        return sum(1 for _ in handle)


def violations(root: Path, limit: int) -> list[tuple[Path, int]]:
    result: list[tuple[Path, int]] = []
    for path in iter_code_files(root):
        count = line_count(path)
        if count > limit:
            result.append((path.relative_to(root), count))
    return result


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path.cwd(), help="repository root")
    parser.add_argument("--limit", type=int, default=DEFAULT_LIMIT, help="maximum allowed lines per code file")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv or sys.argv[1:])
    root = args.root.resolve()
    found = violations(root, args.limit)
    if found:
        print(f"code files over {args.limit} lines:")
        for path, count in found:
            print(f"- {path}: {count} lines")
        return 1
    print(f"code-size check passed: all code files <= {args.limit} lines")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
