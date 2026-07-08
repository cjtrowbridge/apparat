#!/usr/bin/env python3
"""Compatibility wrapper for the canonical Apparat build script."""

from __future__ import annotations

import runpy
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
SCRIPT = ROOT / "scripts" / "build.py"


def main() -> int:
    print("Delegating to canonical build script: python3 scripts/build.py", file=sys.stderr)
    sys.argv[0] = str(SCRIPT)
    runpy.run_path(str(SCRIPT), run_name="__main__")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
