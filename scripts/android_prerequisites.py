#!/usr/bin/env python3
"""
Resolve one OpenJDK 21 installation for Android setup and packaging.
Checks runtime, compiler and sibling tools without changing host configuration.
Rejects Oracle JDK and incompatible compiler majors before mobile binding.
"""
import argparse
import os
from pathlib import Path
import re
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[1]


def java_home():
    if os.environ.get('JAVA_HOME'):
        return Path(os.environ['JAVA_HOME']).expanduser()
    compiler = shutil.which('javac')
    if compiler:
        return Path(compiler).resolve().parent.parent
    local = ROOT / '.tools/jdk-21'
    if local.exists():
        return local
    return None


def validate_java(home):
    if home is None:
        return ['Missing OpenJDK 21; run bootstrap --install-android or set JAVA_HOME']
    failures = []
    suffix = '.exe' if os.name == 'nt' else ''
    for name in ('java', 'javac', 'jar', 'keytool'):
        executable = home / 'bin' / (name + suffix)
        if not executable.is_file():
            failures.append(f'Missing JDK tool: {executable}')
    if failures:
        return failures
    for name in ('java', 'javac'):
        try:
            result = subprocess.run([str(home / 'bin' / (name + suffix)), '-version'], capture_output=True, text=True, timeout=15)
            output = result.stdout + result.stderr
            pattern = r'openjdk version "21(?:\.|\")' if name == 'java' else r'javac 21(?:\.|\s|$)'
            if result.returncode or not re.search(pattern, output) or 'Oracle' in output:
                failures.append(f'{name} must be OpenJDK 21: {output.strip()}')
        except (OSError, subprocess.SubprocessError) as error:
            failures.append(f'Cannot check {name}: {error}')
    return failures


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.parse_args()
    errors = validate_java(java_home())
    print('\n'.join(errors) if errors else 'OpenJDK 21 tools available')
    raise SystemExit(bool(errors))
