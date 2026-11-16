#!/usr/bin/env python3
"""
Dispatch Apparat prerequisite checks to the selected bootstrap v1.
Forwards arguments and exit status without changing global machine configuration.
Only the selected host implementation owns installation and repair policy.
"""
import subprocess
import sys
from pathlib import Path


def main():
    command = [sys.executable, str(Path(__file__).with_name('bootstrap-v1.py'))]
    return subprocess.call(command + sys.argv[1:])


if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print('Bootstrap interrupted', file=sys.stderr)
        raise SystemExit(130)
