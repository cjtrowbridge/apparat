"""Prepare static whisper.cpp libraries for the canonical GUI build.
Uses the pinned checkout and CMake's incremental build in the ignored build tree.
"""
import argparse
import hashlib
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT/'third_party/speech/whisper.cpp'


def prepare(ndk=None):
    if subprocess.check_output(['git', '-C', SOURCE, 'status', '--porcelain']).strip():
        raise ValueError('Whisper checkout must match its pinned revision')
    model = ROOT/'internal/adapters/speech/ggml-tiny.en-q5_1.bin'
    with model.open('rb') as f:
        if hashlib.file_digest(f, 'sha256').hexdigest() != 'c77c5766f1cef09b6b7d47f21b546cbddd4157886b3b5d6d4f709e91e66c7c2b':
            raise ValueError('Bundled Tiny digest mismatch')
    cmake = shutil.which('cmake') or ROOT/'.tools/cmake/cmake/data/bin/cmake'
    if not Path(cmake).is_file():
        subprocess.run([sys.executable, ROOT/'scripts/bootstrap.py', '--install-cmake'], check=True)
    dest = ROOT/'.tmp'/('whisper-android' if ndk else 'whisper-host')
    flags = ['-DBUILD_SHARED_LIBS=OFF', '-DGGML_NATIVE=OFF', '-DGGML_OPENMP=OFF',
             '-DGGML_BLAS=OFF', '-DGGML_ACCELERATE=OFF', '-DGGML_METAL=OFF',
             '-DWHISPER_BUILD_TESTS=OFF', '-DWHISPER_BUILD_EXAMPLES=OFF',
             '-DWHISPER_BUILD_SERVER=OFF', '-DCMAKE_POSITION_INDEPENDENT_CODE=ON',
             '-DCMAKE_BUILD_TYPE=Release', f'-DCMAKE_ARCHIVE_OUTPUT_DIRECTORY={dest}/lib']
    if ndk:
        flags.extend([f'-DCMAKE_TOOLCHAIN_FILE={ndk}/build/cmake/android.toolchain.cmake',
                      '-DANDROID_ABI=arm64-v8a', '-DANDROID_PLATFORM=android-23', '-DANDROID_STL=c++_static'])
    print(f'Preparing local speech runtime in {dest}', flush=True)
    subprocess.run([cmake, '-S', SOURCE, '-B', dest, *flags], check=True)
    subprocess.run([cmake, '--build', dest, '--config', 'Release', '--target', 'whisper', '--parallel', '6'], check=True)


if __name__ == '__main__':
    argparse.ArgumentParser(description=__doc__).parse_args()
    print('Used internally by scripts/build.py for GUI candidates.')
