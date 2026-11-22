"""Install and exercise the Linux worker on the optional trashcan0 SSH test host.
The canonical build calls qualify with a locally qualified artifact and digest.
"""
import argparse
import shutil
import subprocess

OPTIONS = ["-oBatchMode=yes", "-oPreferredAuthentications=publickey",
           "-oStrictHostKeyChecking=yes", "-oConnectTimeout=4"]
HOST = "trashcan0"


def qualify(binary, digest):
    if not shutil.which("ssh"):
        return {"state": "pending", "reason": "SSH unavailable"}
    ssh = ["ssh", *OPTIONS, HOST]
    connected = subprocess.run([*ssh, "uname -sm"], capture_output=True, text=True, timeout=15)
    if connected.returncode == 255:
        return {"state": "pending", "reason": "trusted key-only SSH unavailable"}
    if connected.returncode or connected.stdout.strip() != "Linux x86_64":
        raise ValueError("Unexpected trashcan0 platform")
    command = """
set -eu
umask 077
mkdir -p "$HOME/.local/bin"
cd "$HOME/.local/bin"
tmp=$(mktemp -d .apparat-test.XXXXXXXX)
trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/apparatd"
printf '%s  %s\n' DIGEST "$tmp/apparatd" | sha256sum -c - >/dev/null
chmod 700 "$tmp/apparatd"
"$tmp/apparatd" --runtime-dir "$tmp/node" --smoke-test >/dev/null
if ! cmp -s "$tmp/apparatd" apparatd; then
    mv -f "$tmp/apparatd" apparatd
    if systemctl --user is-active --quiet apparatd.service; then
        systemctl --user restart apparatd.service
    fi
fi
printf '%s  %s\n' DIGEST apparatd | sha256sum -c - >/dev/null
"""
    with binary.open("rb") as stream:
        subprocess.run([*ssh, command.replace("DIGEST", digest)], stdin=stream, check=True, timeout=150)
    return {"state": "passed", "device": HOST}


if __name__ == "__main__":
    argparse.ArgumentParser(description=__doc__).parse_args()
    print("Use python3 scripts/build.py")
