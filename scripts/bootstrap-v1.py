#!/usr/bin/env python3
"""
Inspect and repair declared Apparat prerequisites in ignored project-local paths.
Check mode never writes; normal mode installs only explicitly selected inputs.
Reports independent probes, preserves failure status, and delegates hook setup.
"""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
GO_VERSION = 'go1.26.11'


def probe(command):
    try:
        result = subprocess.run(command, capture_output=True, text=True, timeout=30)
        return result.returncode == 0, (result.stdout + result.stderr).strip()
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        return False, str(error)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='Probe without writing or installing')
    parser.add_argument('--go-source', type=Path, help='Copy an existing Go 1.26.11 distribution into .tools')
    parser.add_argument('--native-source', type=Path, help='Copy an existing native dependency sysroot into .tools')
    parser.add_argument('--android', action='store_true', help='Require Android SDK, NDK, OpenJDK 21 and system ADB')
    parser.add_argument('--install-android', action='store_true', help='Install verified Linux amd64 SDK/NDK and Temurin 21 locally')
    parser.add_argument("--install-go", action="store_true", help="Install checksum-verified Go 1.26.11 locally on Linux amd64")
    args = parser.parse_args()
    try:
        from scripts.checkout_identity import check
    except ImportError:
        from checkout_identity import check
    check(ROOT)
    if args.install_go and not args.check:
        from bootstrap_downloads import install_go
        install_go(GO_VERSION)
    if args.install_android and not args.check:
        from bootstrap_downloads import install_android
        install_android()
    results = []
    for source, destination in ((args.go_source, ROOT / '.tools/go1.26.11'), (args.native_source, ROOT / '.tools/native')):
        if source and not destination.exists() and not args.check:
            print(f'Installing local prerequisite: {destination.name}', flush=True)
            shutil.copytree(source, destination, symlinks=True)
    go = ROOT / '.tools/go1.26.11/bin/go'
    if not go.exists():
        go = shutil.which('go') or 'go'
    ok, detail = probe([str(go), 'version'])
    results.append(('Go 1.26.11', ok and GO_VERSION in detail, detail))
    results.append(('Framework', (ROOT / 'agentic-pipelines/scripts/check_source_governance.py').is_file(), 'Initialize agentic-pipelines if absent'))
    hook = ROOT / 'agentic-pipelines/scripts/install_source_governance_hook.py'
    command = [sys.executable, str(hook), '--repo-root', str(ROOT)]
    if args.check:
        command.append('--check')
    ok, detail = probe(command)
    results.append(('Source hook', ok, detail))
    if args.android or args.install_android:
        sdk = Path(os.environ.get('ANDROID_HOME', ROOT / '.tools/android'))
        for name in ('platforms/android-35/android.jar', 'build-tools/35.0.0/aapt2', 'ndk/27.2.12479018/source.properties'):
            results.append((name, (sdk / name).is_file(), str(sdk / name)))
        from android_prerequisites import java_home, validate_java
        failures = validate_java(java_home())
        results.append(('OpenJDK 21 tools', not failures, '; '.join(failures) or 'runtime/compiler/jar/keytool available'))
        results.append(('System ADB', shutil.which('adb') is not None, 'Install Debian adb and android-sdk-platform-tools-common'))
    for name, ok, detail in results:
        print(f'{"PASS" if ok else "FAIL"}: {name}: {detail}', flush=True)
    failures = sum(not ok for _, ok, _ in results)
    print(f'Bootstrap: {len(results) - failures} passed, {failures} failed', flush=True)
    return int(failures != 0)


if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except KeyboardInterrupt:
        print('Bootstrap interrupted', file=sys.stderr)
        raise SystemExit(130)
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        print(f'Bootstrap failed: {error}', file=sys.stderr)
        raise SystemExit(1)
