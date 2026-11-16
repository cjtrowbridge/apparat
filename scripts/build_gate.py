#!/usr/bin/env python3
"""
Prepare the host environment and enforce bootstrap/source checks before builds.
Uses the canonical framework validator and retains Apparat's stricter size limit.
Provides the same environment to direct builds, Make targets and GUI test commands.
"""
import argparse
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def environment():
    env = dict(os.environ)
    paths = [ROOT / '.tools/go1.26.11/bin', ROOT / '.tools/bin']
    native = ROOT / '.tools/native/usr'
    if native.exists():
        paths.append(native / 'bin')
        env.setdefault('CGO_CFLAGS', f'-I{native / "include"}')
        env.setdefault('CGO_LDFLAGS', f'-L{native / "lib/x86_64-linux-gnu"}')
        env.setdefault('LD_LIBRARY_PATH', str(native / 'lib/x86_64-linux-gnu'))
    jdk = ROOT / '.tools/jdk-21'
    if jdk.exists():
        env.setdefault('JAVA_HOME', str(jdk))
    if env.get('JAVA_HOME'):
        paths.append(Path(env['JAVA_HOME']) / 'bin')
    env['PATH'] = os.pathsep.join(map(str, paths)) + os.pathsep + env.get('PATH', '')
    env.setdefault('GOCACHE', str(ROOT / '.tools/cache/go-build'))
    env.setdefault('GOMODCACHE', str(ROOT / '.tools/cache/go-mod'))
    env.setdefault('GOTOOLCHAIN', 'local')
    return env


def prepare():
    env = environment()
    checks = (
        [sys.executable, str(ROOT / 'scripts/bootstrap.py'), '--check'],
        [sys.executable, str(ROOT / 'agentic-pipelines/scripts/check_source_governance.py'), '--repo-root', str(ROOT)],
        [sys.executable, str(ROOT / 'scripts/check_code_file_lines.py')],
    )
    for command in checks:
        print(f'Build prerequisite: {Path(command[1]).name}', flush=True)
        subprocess.run(command, cwd=ROOT, env=env, check=True)
    return env


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', nargs=argparse.REMAINDER, help='Command to execute after checks; separate with --')
    args = parser.parse_args()
    env = prepare()
    command = args.command
    if command[:1] == ['--']:
        command = command[1:]
    if not command:
        return 0
    return subprocess.call(command, cwd=ROOT, env=env)


if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as error:
        raise SystemExit(error.returncode)
    except KeyboardInterrupt:
        print('Governed command interrupted', file=sys.stderr)
        raise SystemExit(130)
