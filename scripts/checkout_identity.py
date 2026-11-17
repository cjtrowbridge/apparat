"""Verify the designated Apparat main checkout before editing or building.
Reads explicit per-user registration outside the source tree.
Never enrolls a checkout as a side effect of checking it.
"""
import argparse
import json
from pathlib import Path
import subprocess


def identity(root):
    def git(*args):
        return subprocess.check_output(['git', '-C', root, *args], text=True).strip()
    top = Path(git('rev-parse', '--show-toplevel')).resolve()
    common = Path(git('rev-parse', '--path-format=absolute', '--git-common-dir')).resolve()
    if top != root.resolve() or git('symbolic-ref', 'HEAD') != 'refs/heads/main':
        raise ValueError('Apparat requires the designated main checkout')
    git('rev-parse', '--verify', 'HEAD^{commit}')
    return [str(top), str(common)]


def check(root, register=False):
    record = Path.home() / '.config/apparat-build/checkout.json'
    expected = identity(root)
    if register:
        record.parent.mkdir(parents=True, exist_ok=True)
        with record.open('x') as out:
            json.dump(expected, out)
    if not record.is_file() or json.loads(record.read_text()) != expected:
        raise ValueError(f'Checkout mismatch or missing {record}; explicitly register the designated main')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--register', action='store_true', help='Register this confirmed main once; never overwrite a registration')
    args = parser.parse_args()
    check(Path(__file__).resolve().parents[1], args.register)
