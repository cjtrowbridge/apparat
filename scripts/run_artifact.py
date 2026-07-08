#!/usr/bin/env python3
"""Run a built artifact with forwarded arguments."""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Run a built Apparat artifact and forward any remaining arguments to it.",
        epilog=(
            "Example: python3 scripts/run_artifact.py releases/linux/amd64/apparatd/latest "
            "-- --smoke-test --runtime-dir /tmp/apparatd-smoke"
        ),
    )
    parser.add_argument("artifact", type=Path, help="path to the built binary artifact")
    parser.add_argument("forwarded", nargs=argparse.REMAINDER, help="arguments forwarded to the artifact after optional --")
    return parser.parse_args(argv)


def main(argv: list[str]) -> int:
    args = parse_args(argv)
    artifact = args.artifact
    forwarded = args.forwarded
    if forwarded[:1] == ["--"]:
        forwarded = forwarded[1:]
    return subprocess.run([str(artifact), *forwarded], check=False).returncode


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
