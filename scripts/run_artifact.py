#!/usr/bin/env python3
"""Run a built artifact with forwarded arguments."""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path


def main(argv: list[str]) -> int:
    if not argv:
        print("usage: run_artifact.py <artifact> [-- args...]", file=sys.stderr)
        return 2

    artifact = Path(argv[0])
    forwarded = argv[1:]
    if forwarded[:1] == ["--"]:
        forwarded = forwarded[1:]
    return subprocess.run([str(artifact), *forwarded], check=False).returncode


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
