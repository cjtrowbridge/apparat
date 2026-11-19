#!/usr/bin/env python3
"""
Install verified Android SDK archives and Temurin 21 into ignored local tools.
Downloads only on explicit bootstrap request; preserves existing installations.
Validates digests and archive paths before extracting executable tool inputs.
"""
import argparse
import hashlib
import json
from pathlib import Path
import platform
import shutil
import stat
import tarfile
import urllib.request
import zipfile

ROOT = Path(__file__).resolve().parents[1]
PACKAGES = (
    ('platform-35_r02.zip', '0bb560a90a7a2cbd0dd8348224d518b638fe7949', 'platforms/android-35'),
    ('build-tools_r35_linux.zip', '2cfaa0bbb2336e9ec18ed3ecea84fa2e2af607bc', 'build-tools/35.0.0'),
    ('android-ndk-r27c-linux.zip', '090e8083a715fdb1a3e402d0763c388abb03fb4e', 'ndk/27.2.12479018'),
)


def download(url, destination, checksum, algorithm):
    digest = hashlib.new(algorithm)
    print(f'Downloading {destination.name}', flush=True)
    with urllib.request.urlopen(urllib.request.Request(url, headers={'User-Agent': 'Apparat-bootstrap/1.0'}), timeout=120) as response, destination.open('wb') as output:
        while True:
            chunk = response.read(1024 * 1024)
            if not chunk:
                break
            digest.update(chunk)
            output.write(chunk)
    if digest.hexdigest() != checksum:
        destination.unlink()
        raise RuntimeError(f'Checksum mismatch: {destination.name}')


def unpack_zip(archive, stage):
    with zipfile.ZipFile(archive) as bundle:
        for entry in bundle.infolist():
            target = stage / entry.filename
            if not target.resolve().is_relative_to(stage.resolve()):
                raise ValueError('Archive path escapes staging directory')
            mode = entry.external_attr >> 16
            if stat.S_ISLNK(mode):
                link = bundle.read(entry).decode()
                if not (target.parent / link).resolve().is_relative_to(stage.resolve()):
                    raise ValueError('Archive symlink escapes staging directory')
                target.parent.mkdir(parents=True, exist_ok=True)
                target.symlink_to(link)
            else:
                bundle.extract(entry, stage)
                if mode and not entry.is_dir():
                    target.chmod(mode & 0o777)


def install_cmake():
    if shutil.which('cmake') or (ROOT/'.tools/cmake/cmake/data/bin/cmake').is_file():
        return
    if platform.system() != 'Linux' or platform.machine() != 'x86_64':
        raise RuntimeError('Install CMake 3.21 or later for local speech builds')
    dest = ROOT/'.tools/cmake'
    dest.mkdir(parents=True, exist_ok=True)
    wheel = dest/'cmake.whl'
    url = ('https://files.pythonhosted.org/packages/59/e8/096984b89133681533650b9078c5ed1c5c9b534e869b5487f22d4de1935c/'
           'cmake-3.31.6-py3-none-manylinux_2_17_x86_64.manylinux2014_x86_64.whl')
    download(url, wheel, '1c8b05df0602365da91ee6a3336fe57525b137706c4ab5675498f662ae1dbcec', 'sha256')
    unpack_zip(wheel, dest)
    wheel.unlink()


def install_android():
    if platform.system() != 'Linux' or platform.machine() not in ('x86_64', 'amd64'):
        raise RuntimeError('Automatic Android archive setup supports Linux amd64 only; configure SDK/JDK paths on other hosts')
    sdk = ROOT / '.tools/android'
    sdk.mkdir(parents=True, exist_ok=True)
    for name, checksum, relative in PACKAGES:
        destination = sdk / relative
        marker = destination / '.apparat-archive-sha1'
        if marker.exists() and marker.read_text().strip() == checksum:
            print(f'Already installed: {relative}', flush=True)
            continue
        if destination.exists():
            raise RuntimeError(f'Unmanaged or partial installation: {destination}')
        archive = sdk / name
        download('https://dl.google.com/android/repository/' + name, archive, checksum, 'sha1')
        stage = sdk / (name + '.extract')
        if stage.exists():
            shutil.rmtree(stage)
        stage.mkdir()
        unpack_zip(archive, stage)
        children = list(stage.iterdir())
        if len(children) != 1 or not children[0].is_dir():
            raise RuntimeError('Unexpected SDK archive layout')
        destination.parent.mkdir(parents=True, exist_ok=True)
        children[0].rename(destination)
        marker.write_text(checksum + '\n')
        archive.unlink()
        stage.rmdir()
    install_jdk()


def install_jdk():
    destination = ROOT / '.tools/jdk-21'
    if (destination / 'release').is_file():
        print('JDK already installed; bootstrap will validate it', flush=True)
        return
    url = 'https://api.adoptium.net/v3/assets/latest/21/hotspot?architecture=x64&image_type=jdk&os=linux'
    with urllib.request.urlopen(urllib.request.Request(url, headers={'User-Agent': 'Apparat-bootstrap/1.0'}), timeout=60) as response:
        package = json.load(response)[0]['binary']['package']
    archive = ROOT / '.tools/temurin-21.tar.gz'
    download(package['link'], archive, package['checksum'], 'sha256')
    stage = ROOT / '.tools/temurin-extract'
    stage.mkdir(exist_ok=True)
    with tarfile.open(archive) as bundle:
        bundle.extractall(stage, filter='data')
    children = list(stage.iterdir())
    if len(children) != 1:
        raise RuntimeError('Unexpected JDK archive layout')
    children[0].rename(destination)
    (destination / '.apparat-download.json').write_text(json.dumps(package, indent=2) + '\n')
    stage.rmdir()
    archive.unlink()


def install_go(version):
    if platform.system() != 'Linux' or platform.machine() not in ('x86_64', 'amd64'):
        raise RuntimeError('Automatic Go installation supports Linux amd64; use --go-source elsewhere')
    destination = ROOT / '.tools' / version
    if destination.exists():
        print('Go already installed; bootstrap will validate it', flush=True)
        return
    with urllib.request.urlopen('https://go.dev/dl/?mode=json&include=all', timeout=60) as response:
        releases = json.load(response)
    release = next(item for item in releases if item['version'] == version)
    package = next(item for item in release['files'] if item['os'] == 'linux' and item['arch'] == 'amd64' and item['kind'] == 'archive')
    archive = ROOT / '.tools' / package['filename']
    archive.parent.mkdir(parents=True, exist_ok=True)
    download('https://go.dev/dl/' + package['filename'], archive, package['sha256'], 'sha256')
    stage = ROOT / '.tools' / (version + '.extract')
    if stage.exists():
        shutil.rmtree(stage)
    stage.mkdir()
    with tarfile.open(archive) as bundle:
        bundle.extractall(stage, filter='data')
    (stage / 'go').rename(destination)
    (destination / '.apparat-download.json').write_text(json.dumps(package, indent=2) + '\n')
    stage.rmdir()
    archive.unlink()


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.parse_args()
    print('Use scripts/bootstrap.py --install-android to authorize local installation.')
