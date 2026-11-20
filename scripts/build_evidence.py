"""Qualify fresh Apparat candidates before replacing canonical release files.
Preserve per-run evidence and a recovery journal for interrupted promotions.
Invoked only by the no-flag host build orchestrator.
"""
import argparse
import hashlib
import json
import os
import re
from pathlib import Path
import shutil
import subprocess
import tempfile

try:
    from scripts import artifact_acceptance
except ImportError:
    import artifact_acceptance


def digest(path):
    with path.open('rb') as handle:
        return stream_digest(handle)


def stream_digest(handle):
    checksum = hashlib.sha256()
    for block in iter(lambda: handle.read(1024 * 1024), b''):
        checksum.update(block)
    return checksum.hexdigest()


def write(path, data):
    temporary = path.with_suffix('.new')
    temporary.write_text(json.dumps(data, indent=2)+'\n')
    temporary.replace(path)


def git(root, *args):
    return subprocess.check_output(['git', '-C', root, *args])


def fingerprint(root):
    checksum = hashlib.sha256()
    for entry in sorted(git(root, 'ls-files', '-z', '--cached', '--others', '--exclude-standard',
                            '--', 'cmd', 'internal', 'android', 'scripts', 'go.mod', 'go.sum', 'third_party/stmpck', 'source-governance.yaml').split(b'\0')):
        if entry:
            path = root/os.fsdecode(entry)
            checksum.update(entry+b'\0')
            if entry == b'android/apparat/AndroidManifest.xml':
                content = re.sub(rb'android:version(Code|Name)="[^"]+"', b'', path.read_bytes())
                checksum.update(content)
            else:
                checksum.update(digest(path).encode() if path.is_file() else b'missing')
    checksum.update(git(root, 'submodule', 'status'))
    engine = root/'third_party/game/ebiten'
    if git(engine,'status','--porcelain','--untracked-files=normal').strip():
        raise ValueError('Dirty engine checkout; desktop and Android would compile different inputs')
    local = root/'build_environment.py'
    if local.exists():
        checksum.update(digest(local).encode())
    return checksum.hexdigest()


def restore(root, journal):
    if journal.exists():
        for dest, backup in json.loads(journal.read_text()):
            path = root/dest
            if backup:
                shutil.copy2(backup, path)
            else:
                path.unlink(missing_ok=True)
        journal.unlink()


def qualify(root, plans, compile_target, *, publish=False):
    storage = root/'.tmp/build-runs'
    storage.mkdir(parents=True, exist_ok=True)
    with (storage/'build.lock').open('a+b') as lock:
        if os.name == 'nt':
            import msvcrt
            lock.seek(0)
            msvcrt.locking(lock.fileno(), msvcrt.LK_NBLCK, 1)
        else:
            import fcntl
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        if publish:
            try:
                from scripts import build_delivery
            except ImportError:
                import build_delivery
            return build_delivery.deliver(root, plans, compile_target, storage)
        return locked_run(root, plans, compile_target, storage)


def locked_run(root, plans, compile_target, storage):
    run = Path(tempfile.mkdtemp(dir=storage))
    print(f'Qualification report: {run}/report.json', flush=True)
    journal = storage/'promotion.json'
    restore(root, journal)
    manifest = root/'releases/verified.json'
    if manifest.is_symlink() or manifest.parent.resolve() != root/'releases':
        raise ValueError('Noncanonical release directory or manifest')
    if manifest.exists():
        previous = json.loads(manifest.read_text())
        for row in previous['targets']:
            if row['state'] == 'passed' and digest(root/row['output']) != row['sha256']:
                raise ValueError(f"Previously verified artifact mismatch: {row['output']}")
    before = fingerprint(root)
    report = {'edition':'ordinary', 'root':str(root), 'ref':git(root,'symbolic-ref','HEAD').decode().strip(),
              'head':git(root,'rev-parse','HEAD').decode().strip(), 'inputs':before,
              'go': subprocess.check_output(['go','version'],text=True).strip(), 'targets':[]}
    good = True
    outputs = []
    for plan in plans:
        dest = plan.output
        rel = dest.relative_to(root/'releases') if dest.is_absolute() else Path(plan.name)
        candidate = run/'candidates'/rel
        row = {'name':plan.name,'output':str(Path('releases')/rel),'state':'unsupported','reasons':plan.reasons}
        report['targets'].append(row)
        if not plan.possible:
            continue
        try:
            if dest.resolve() != root/'releases'/rel:
                raise ValueError('noncanonical artifact path')
            if '..' in rel.parts or rel.is_absolute():
                raise ValueError('escaped release path')
            candidate.parent.mkdir(parents=True)
            if compile_target(plan,candidate) != 0:
                raise ValueError('compiler failed')
            row['sha256'] = digest(candidate)
            row['state'] = 'built, awaiting verification'
            print(f'Exercising {plan.name}', flush=True)
            evidence = run/'tests'/plan.name
            evidence.mkdir(parents=True)
            artifact_acceptance.validate(plan,candidate,evidence,row['sha256'])
            if digest(candidate) != row['sha256']:
                raise ValueError('candidate changed after test')
            row['state'] = 'passed'
            outputs.append((dest,candidate))
        except (ValueError, OSError, RuntimeError, subprocess.SubprocessError) as exc:
            row['error'] = str(exc)
            if row['state'] == 'unsupported':
                row['state'] = 'failed'
            print(f'{plan.name}: {exc}', flush=True)
            good = False
        write(run/'report.json',report)
    report['accepted'] = bool(outputs) and good and before == fingerprint(root)
    write(run/'report.json',report)
    if not report['accepted']:
        return 1
    backups = []
    for dest, candidate in outputs+[(manifest,run/'report.json')]:
        dest.parent.mkdir(parents=True,exist_ok=True)
        backup = run/'previous'/dest.relative_to(root)
        if dest.exists():
            backup.parent.mkdir(parents=True,exist_ok=True)
            shutil.copy2(dest,backup)
        backups.append([str(dest.relative_to(root)),str(backup) if dest.exists() else None])
    write(journal,backups)
    try:
        for dest,candidate in outputs:
            candidate.replace(dest)
        public = dict(report)
        public.pop('root')
        write(manifest,public)
        journal.unlink()
    except BaseException:
        restore(root,journal)
        raise
    return 0


if __name__ == '__main__':
    argparse.ArgumentParser(description=__doc__).parse_args()
    print('Use python3 scripts/build.py')
