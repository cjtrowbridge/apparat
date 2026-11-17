"""Prepare the embedded STMPCK module from its pinned dependency and host patch.
Export into a temporary directory, then install the complete generated module.
"""
import argparse
import hashlib
import io
import os
from pathlib import Path
import subprocess
import tarfile
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def prepare():
    source = ROOT/'third_party/stmpck'
    patch = ROOT/'scripts/patches/stmpck-embed.patch'
    pin = subprocess.check_output(['git','-C',ROOT,'ls-files','--stage','third_party/stmpck'],text=True).split()[1]
    stamp = pin + hashlib.sha256(patch.read_bytes()).hexdigest()
    cache = ROOT/'.tmp/stmpck-modules'
    cache.mkdir(parents=True,exist_ok=True)
    target = cache/stamp
    if not target.exists():
        archive = subprocess.check_output(['git','-C',source,'archive',pin,'internal','go.mod','go.sum'])
        with tempfile.TemporaryDirectory(dir=cache) as work:
            module = Path(work)/'module'
            with tarfile.open(fileobj=io.BytesIO(archive)) as bundle:
                bundle.extractall(module,filter='data')
            subprocess.run(['git','apply','--no-index','--directory=module',str(patch)],cwd=work,env={**os.environ,'GIT_CEILING_DIRECTORIES':str(cache)},check=True)
            module.rename(target)
    link = ROOT/'.tmp/stmpck-module'
    with tempfile.TemporaryDirectory(dir=cache) as work:
        alias = Path(work)/'current'
        alias.symlink_to(target,target_is_directory=True)
        alias.replace(link)


if __name__ == '__main__':
    argparse.ArgumentParser(description=__doc__).parse_args()
    prepare()
