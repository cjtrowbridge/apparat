"""Publish the qualified host matrix on main and verify the actual download channel.
Called under the canonical pipeline lock. Only reviewed index changes and the
explicit generated paths enter the delivery commit. A rejected push is retryable.
"""
import argparse
import hashlib
import json
import re
import subprocess
import urllib.error
import urllib.parse
import urllib.request

try:
    from scripts import android_release, build, build_evidence as evidence
except ImportError:
    import android_release
    import build
    import build_evidence as evidence

MANIFEST = 'android/apparat/AndroidManifest.xml'
APK = 'releases/android/arm64/apparat/latest.apk'
REPORT = 'releases/verified.json'
INPUTS = ('cmd', 'internal', 'android', 'scripts', 'go.mod', 'go.sum', 'third_party',
          '.gitmodules', 'agentic-pipelines', 'source-governance.yaml')
REPOSITORY = 'cjtrowbridge/apparat'


def git(root, *args):
    return evidence.git(root, *args).decode().strip()


def fetch_main(root):
    git(root, 'fetch', '--no-tags', 'origin', 'refs/heads/main')
    return git(root, 'rev-parse', 'FETCH_HEAD')


def ensure_inputs_staged(root):
    changed = git(root, 'diff', '--name-only', '--', *INPUTS)
    new = git(root, 'ls-files', '--others', '--exclude-standard', '--', *INPUTS)
    if changed or new:
        raise ValueError(f'Stage the intended build inputs before delivery:\n{changed}\n{new}')


def metadata(path):
    directory = build.default_sdk_root()/'build-tools'/build.ANDROID_BUILD_TOOLS
    return android_release.apk_metadata(directory, path)


def published_metadata(root, remote, storage):
    previous = storage/'published.apk'
    with previous.open('wb') as dest:
        subprocess.run(['git','-C',root,'show',remote+':'+APK],stdout=dest,check=True)
    baseline = metadata(previous)
    if baseline['signers'] != [android_release.expected_signer()] or baseline['package'] != android_release.PACKAGE:
        raise ValueError('Published APK package or signing lineage differs')
    return baseline


def allocate(root, remote, baseline):
    path = root/MANIFEST
    content = path.read_text()
    old = git(root,'show',remote+':'+MANIFEST)
    version = baseline['version'] + 1
    match = re.search(r'android:versionName="(\d+\.\d+\.)(\d+)"',old)
    if not match:
        raise ValueError('Expected three-part published Android version name')
    name = match[1] + str(int(match[2])+1)
    content = re.sub(r'android:versionCode="\d+"', f'android:versionCode="{version}"',content)
    content = re.sub(r'android:versionName="[^"]+"', f'android:versionName="{name}"',content)
    path.write_text(content)
    git(root,'add','--',MANIFEST)
    print(f'Android release {name}, version code {version}',flush=True)


def verify_download(output, expected):
    endpoint = f'repos/{REPOSITORY}/contents/{output}?ref=main'
    address = subprocess.check_output(['gh','api',endpoint,'--jq','.download_url'],text=True).strip()
    location = urllib.parse.urlsplit(address)
    if location.scheme != 'https' or location.hostname != 'raw.githubusercontent.com':
        raise ValueError('Unexpected GitHub artifact download host')
    try:
        with urllib.request.urlopen(address,timeout=120) as response:
            actual = evidence.stream_digest(response)
    except urllib.error.URLError:
        raise ValueError(f'GitHub artifact download failed for {output}') from None
    if actual != expected:
        raise ValueError(f'Private GitHub read-back failed for {output}; retry delivery')


def publish(root, rows, verify=verify_download):
    head = git(root,'rev-parse','HEAD')
    if git(root,'ls-remote','origin','refs/heads/main').split()[0] != head:
        print('Pushing qualified delivery to origin/main',flush=True)
        git(root,'push','origin',head+':refs/heads/main')
    published = fetch_main(root)
    if published != head:
        raise ValueError('Remote main differs from tested release commit')
    for row in rows:
        blob = evidence.git(root,'show',published+':'+row['output'])
        if hashlib.sha256(blob).hexdigest() != row['sha256']:
            raise ValueError('Published artifact differs from qualified output: '+row['output'])
        print('Verifying download '+row['output'],flush=True)
        verify(row['output'],row['sha256'])
    if git(root,'ls-remote','origin','refs/heads/main').split()[0] != head:
        raise ValueError('Remote main changed during verification; reconcile and retry')
    print('Delivery verified on main at '+head,flush=True)


def deliver(root, plans, compile_target, storage):
    evidence.restore(root,storage/'promotion.json')
    ensure_inputs_staged(root)
    remote = fetch_main(root)
    head = git(root,'rev-parse','HEAD')
    git(root,'merge-base','--is-ancestor',remote,head)
    expected = {plan.name for plan in plans if plan.possible}
    for plan in plans:
        if not plan.possible and plan.goos != 'android':
            raise ValueError('Required target unavailable: '+plan.name+' '+str(plan.reasons))
    report = root/REPORT
    prior = json.loads(report.read_text()) if report.exists() else {}
    rows = [r for r in prior.get('targets',[]) if r['state']=='passed' and r['name'] in expected]
    current = evidence.fingerprint(root)
    reusable = (prior.get('delivery_schema') == 1 and prior.get('accepted') and
                {r['name'] for r in rows} == expected and
                all(r.get('inputs') == current and evidence.digest(root/r['output']) == r['sha256'] for r in rows))
    android = any(p.possible and p.goos == 'android' for p in plans)
    if android:
        baseline = published_metadata(root,remote,storage)
        if reusable:
            installed = metadata(root/APK)
            content = (root/MANIFEST).read_text()
            reusable = f'android:versionCode="{installed["version"]}"' in content
            if git(root,'hash-object',APK) != git(root,'rev-parse',remote+':'+APK):
                reusable = reusable and installed['version'] > baseline['version']
    if not reusable:
        if android:
            allocate(root,remote,baseline)
        if evidence.locked_run(root,plans,compile_target,storage):
            return 1
        result = json.loads(report.read_text())
        rows = [r for r in result['targets'] if r['state']=='passed']
        for row in rows:
            row['inputs'] = result['inputs']
        if android:
            android_release.validate_candidate(metadata(root/APK),baseline,android_release.expected_signer())
        result['delivery_schema'] = 1
        existing = {r['name'] for r in result['targets'] if r['state']=='passed'}
        preserved = [r for r in prior.get('targets',[]) if r['state']=='passed' and r['name'] not in existing]
        kept = {r['name'] for r in preserved}
        result['targets'] = [r for r in result['targets'] if r['name'] not in kept] + preserved
        evidence.write(report,result)
    if any(r['inputs'] != evidence.fingerprint(root) for r in rows):
        raise ValueError('Build inputs changed before publication')
    ensure_inputs_staged(root)
    owned = {r['output'] for r in rows} | {REPORT}
    staged_outputs = set(git(root,'diff','--cached','--name-only','--','releases').splitlines())
    if staged_outputs - owned - {'releases/README.md'}:
        raise ValueError('Staged artifact outside this host matrix')
    git(root,'add','--',REPORT,*(r['output'] for r in rows))
    if git(root,'diff','--cached','--name-only'):
        git(root,'commit','-m','Deliver verified Apparat build on main')
    publish(root,rows)
    return 0


if __name__ == '__main__':
    argparse.ArgumentParser(description=__doc__).parse_args()
    print('Use python3 scripts/build.py')
