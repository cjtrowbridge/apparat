#!/usr/bin/env python3
"""
Prepare isolated Android engine sources without editing tracked submodules.
Exports the exact selected Git revision and applies the display-read guard there.
Keeps invalid metrics unavailable and records source/patch identity for diagnostics.
"""
import argparse
import hashlib
import io
from pathlib import Path
import shutil
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parents[1]
PATCH_VERSION = 'display-values-v1'


def guard_display_values(text):
    old = '\tif v == nil {\n\t\treturn 0, 0, 1, false\n\t}'
    new = '\tif v == nil || v.width <= 0 || v.height <= 0 || v.scale <= 0 {\n\t\treturn 0, 0, 1, false\n\t}'
    if text.count(old) != 1:
        raise RuntimeError('Unsupported Ebitengine displayInfo source; expected one nil-value guard')
    return text.replace(old, new, 1)


def prepare_engine():
    source = ROOT / 'third_party/game/ebiten'
    revision = subprocess.check_output(['git', '-C', str(source), 'rev-parse', 'HEAD'], text=True).strip()
    work = ROOT / '.tmp/mobile-engine'
    if work.exists():
        shutil.rmtree(work)
    work.mkdir(parents=True)
    archive = subprocess.check_output(['git', '-C', str(source), 'archive', revision])
    with tarfile.open(fileobj=io.BytesIO(archive)) as bundle:
        bundle.extractall(work, filter='data')
    target = work / 'internal/ui/ui_mobile.go'
    target.write_text(guard_display_values(target.read_text()))
    digest = hashlib.sha256(target.read_bytes()).hexdigest()
    (work / '.apparat-source').write_text(f'{revision}\n{PATCH_VERSION}\n{digest}\n')
    print(f'Android engine: revision={revision} patch={PATCH_VERSION}', flush=True)
    return work


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.parse_args()
    print('Used by the governed Android build; no standalone compilation entrypoint.')
