"""Install the qualified Linux amd64 worker and reserve its first claim for the installer.
Run on the admin's Linux amd64 Apparat workstation using a trusted SSH host alias.
"""
import argparse
import hashlib
import json
import re
import shlex
import subprocess
from pathlib import Path

from trashcan_acceptance import OPTIONS

ROOT = Path(__file__).resolve().parents[1]
TARGET = "linux/amd64/apparatd"
INSTALL = r'''
set -eu
umask 077
test "$(uname -sm)" = "Linux x86_64"
systemctl --user show-environment >/dev/null
mkdir -p "$HOME/.local/bin"
cd "$HOME/.local/bin"
exec 9>.apparat-install.lock
flock -x 9
tmp=$(mktemp -d .apparat-install.XXXXXXXX)
trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/apparatd"
printf '%s  %s\n' "$1" "$tmp/apparatd" | sha256sum -c - >/dev/null
chmod 700 "$tmp/apparatd"
"$tmp/apparatd" --runtime-dir "$tmp/test" --smoke-test >/dev/null
if ! "$tmp/apparatd" trust-host >/dev/null 2>&1; then
    "$tmp/apparatd" identity-setup --generate-if-missing
fi
if ! systemctl --user is-active --quiet apparatd.service; then
    "$tmp/apparatd" --services >/dev/null
fi
"$tmp/apparatd" trust-host --install "$2"
changed=0
if ! cmp -s "$tmp/apparatd" apparatd; then
    mv -f "$tmp/apparatd" apparatd
    changed=1
fi
if ! systemctl --user cat apparatd.service >/dev/null 2>&1; then
    units="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
    mkdir -p "$units"
    cat > "$tmp/apparatd.service" <<'UNIT'
[Unit]
Description=Apparat headless cluster worker
After=network.target

[Service]
ExecStart=%h/.local/bin/apparatd
Restart=on-failure
RestartSec=5
UMask=0077

[Install]
WantedBy=default.target
UNIT
    install -m 600 "$tmp/apparatd.service" "$units/apparatd.service"
    systemctl --user daemon-reload
fi
loginctl enable-linger "$(id -un)"
systemctl --user enable apparatd.service
if [ "$changed" = 1 ]; then
    systemctl --user restart apparatd.service
else
    systemctl --user start apparatd.service
fi
systemctl --user is-active --quiet apparatd.service
'''


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("host", help="SSH destination for the worker service account")
    args = parser.parse_args()
    binary = ROOT / "releases" / TARGET / "latest"
    report = json.loads((ROOT / "releases/verified.json").read_text())
    with binary.open("rb") as stream:
        digest = hashlib.file_digest(stream, "sha256").hexdigest()
        if not any(t["name"] == TARGET and t["state"] == "passed" and t["sha256"] == digest
                   for t in report["targets"]):
            parser.exit(1, "Worker executable does not match its qualified release.\n")
        owner = subprocess.check_output([binary, "user-fingerprint"], text=True).strip()
        if not re.fullmatch(r"[A-Za-z0-9_-]{43}", owner):
            parser.exit(1, "Invalid installing identity.\n")
        stream.seek(0)
        command = shlex.join(["sh", "-c", INSTALL, "install-worker", digest, owner])
        subprocess.run(["ssh", *OPTIONS, "--", args.host, command], stdin=stream, check=True)
    print("Worker installed. Open Cluster → Add New Device to claim it.")


if __name__ == "__main__":
    main()
