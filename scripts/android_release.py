#!/usr/bin/env python3
"""
Verify Android signing continuity and a strictly newer publication candidate.
Reads APK metadata and public certificates; never creates keys or publishes files.
Run with --candidate APK --previous DOWNLOADED_APK through the build prerequisite gate.
"""
import argparse
import hashlib
import os
from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parents[1]
PACKAGE = 'com.cjtrowbridge.apparat'


def expected_signer():
    value = (ROOT / 'android/apparat/signing-certificate.sha256').read_text().strip()
    if not re.fullmatch('[0-9a-f]{64}', value):
        raise ValueError('Invalid Android signing certificate pin')
    return value


def validate_keystore(java_home, keystore):
    if not keystore.is_file():
        raise ValueError('Android signing key missing: restore the existing .tools/android/debug.keystore; do not generate a replacement')
    suffix = '.exe' if os.name == 'nt' else ''
    certificate = subprocess.check_output([
        str(java_home / 'bin' / ('keytool' + suffix)), '-exportcert',
        '-keystore', str(keystore), '-storepass', 'android', '-alias', 'apparat-debug',
    ], stderr=subprocess.PIPE)
    if hashlib.sha256(certificate).hexdigest() != expected_signer():
        raise ValueError('Android signing key does not match the installed-device certificate pin')


def apk_metadata(build_tools, apk):
    suffix = '.bat' if os.name == 'nt' else ''
    verified = subprocess.check_output([
        str(build_tools / ('apksigner' + suffix)), 'verify', '--print-certs', str(apk),
    ], text=True, stderr=subprocess.PIPE)
    suffix = '.exe' if os.name == 'nt' else ''
    badging = subprocess.check_output([
        str(build_tools / ('aapt2' + suffix)), 'dump', 'badging', str(apk),
    ], text=True, stderr=subprocess.PIPE)
    match = re.search(r"package: name='([^']+)' versionCode='([0-9]+)'", badging)
    if not match:
        raise ValueError('Cannot read APK package/version metadata')
    signers = re.findall(r'Signer #\d+ certificate SHA-256 digest: ([0-9a-f]+)', verified)
    return {'package': match[1], 'version': int(match[2]), 'signers': signers}


def validate_candidate(candidate, previous, signer):
    if candidate['package'] != PACKAGE or previous['package'] != PACKAGE:
        raise ValueError('Wrong package in update candidate or previous APK')
    if candidate['signers'] != [signer]:
        raise ValueError('Candidate signer differs from the installed-device certificate pin')
    if candidate['version'] <= previous['version']:
        raise ValueError('Increment AndroidManifest.xml versionCode above the published APK before publication')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--candidate', type=Path, required=True, help='Locally built signed APK to publish')
    parser.add_argument('--previous', type=Path, required=True, help='Freshly downloaded current GitHub APK (read only)')
    args = parser.parse_args()
    from build_gate import prepare
    os.environ.update(prepare())
    import build
    toolchain, failures, _ = build.resolve_android_toolchain('go')
    if failures or not toolchain:
        raise ValueError('Android prerequisites unavailable: ' + '; '.join(failures))
    print('Checking candidate and published APK signatures and versions', flush=True)
    directory = toolchain.sdk_root / 'build-tools' / build.ANDROID_BUILD_TOOLS
    candidate = apk_metadata(directory, args.candidate)
    previous = apk_metadata(directory, args.previous)
    validate_candidate(candidate, previous, expected_signer())
    print(f"Release verified: {previous['version']} -> {candidate['version']}; installed-device signer matches")
    print('Candidate SHA-256: ' + hashlib.sha256(args.candidate.read_bytes()).hexdigest())


if __name__ == '__main__':
    try:
        main()
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        raise SystemExit(str(error))
    except KeyboardInterrupt:
        raise SystemExit(130)
