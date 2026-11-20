# Playbook: Install Apparat on Steam Deck and Add It to Steam

*Status: Stable*

## Objective

Keep key-based SSH access available, install the qualified native Linux GUI in
the Deck's persistent home storage and open it from Steam in Gaming Mode.

## Prerequisites

- Read `AGENTS.md`, `README.md`, `ROADMAP.md` Phase 17, `releases/README.md` and
  `docs/platform-matrix.md`. Follow the host planning playbook.
- Work from the designated main checkout; run `python3 scripts/checkout_identity.py`
  before first-party edits and `python3 scripts/bootstrap.py --check` before work.
- The Deck is awake on the same reachable network. Confirm its current IP and
  username with the owner. This checkpoint used `deck@10.0.0.197`.
- Host tools: OpenSSH client, SCP, Python 3 and `sha256sum`. Deck tools:
  Python 3, `desktop-file-validate`, running Steam and the installed Valve
  `/usr/bin/steamos-add-to-steam` helper. Inspect that helper on the target;
  its accepted inputs and behavior can change with SteamOS.
- Use `releases/linux/amd64/apparat/latest` for an x86_64 Deck. It is the GUI,
  not `apparatd`. Use an existing tracked, qualified artifact when installing.
  If a new build is necessary, use only no-flag `python3 scripts/build.py` with
  its normal qualification and delivery requirements. Installation alone does
  not authorize an unrelated release or publication.

## Plan Binding

Bind installation to approved active plan items for persistent SSH access,
verified native installation, Steam registration and startup evidence. Map
them to Phase 17's Steam Deck/Linux GUI packaging, desktop integration,
diagnostics, update and rollback requirements. Existing explicit user approval
for these outcomes is execution authority; do not ask for it again.

Protect other applications, Steam entries, credentials and node state. Leave
the SteamOS system partition read-only. No system package installation, root
filesystem changes, application source copies or maintained alternate checkout
are required. Full controller, audio, workspace and suspend/resume acceptance
remain separate checks, not implied by launcher startup.

## Step-by-Step Instructions

### 1. Establish persistent SSH access

On the Deck in Desktop Mode, open Konsole. If the account has no password,
run `passwd` interactively, then enable SSH now and at subsequent boots:

```bash
sudo systemctl enable --now sshd
systemctl is-enabled sshd
systemctl is-active sshd
ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub
```

Expected service results are `enabled` and `active`. On the host, preserve
`~/.ssh/config` with a backup and add or update just this host stanza:

```sshconfig
Host steam-deck
    HostName 10.0.0.197
    User deck
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes
    AddKeysToAgent yes
    ServerAliveInterval 30
    ServerAliveCountMax 3
```

Use the current confirmed address. A router DHCP reservation avoids future IP
changes. Use `steam-deck.local` only after confirming it resolves reliably from
this host; it did not resolve during the recorded installation. Keepalives do
not wake the Deck or make it reachable during sleep.

Reuse the existing host key pair. If none exists, generate one interactively
with `ssh-keygen -t ed25519`; do not overwrite any key. Install only the public
key with `ssh-copy-id -i ~/.ssh/id_ed25519.pub steam-deck`, verifying the host
fingerprint against Konsole before trusting the first connection. Enter account
passwords only in the interactive terminal, never in scripts, logs or chat.
An existing working authorized key needs no reinstall. Keep the host SSH directory
mode 700 and its config/private keys mode 600.

Verify using the host terminal outside a restricted sandbox:

```bash
ssh -o BatchMode=yes -o ConnectTimeout=8 steam-deck \
  'id -un; cat /etc/hostname; uname -m; systemctl is-enabled sshd; systemctl is-active sshd'
```

Expected: `deck`, `steam-deck`, `x86_64`, `enabled`, `active`. Connection refused
usually means sshd is stopped; a timeout suggests sleep, address or network
reachability. Permission denied requires checking the username, public key and
authorized-key permissions. Never replace a changed known-host key blindly.

### 2. Stage and verify the release

From the repository root:

```bash
git status --short
git log -1 --oneline -- releases/linux/amd64/apparat/latest
git diff --exit-code -- releases/linux/amd64/apparat/latest
sha256sum releases/linux/amd64/apparat/latest
ssh steam-deck 'df -h ~; pgrep -a -x apparat; pgrep -a -x apparatd'
```

`pgrep` exits 1 when no matching process exists. Stop an existing Apparat owner
gracefully before upgrading. Do not launch a GUI and worker on the same node
root. Back up populated node state offline before changing to a binary with
different migrations; a binary backup alone cannot undo database migrations.

Create the target directory, reject an unexpected destination symlink and copy
into a staging file, leaving the installed executable intact:

```bash
ssh steam-deck 'mkdir -p ~/.local/bin; test ! -L ~/.local/bin/apparat'
scp releases/linux/amd64/apparat/latest steam-deck:.local/bin/apparat.incoming
ssh steam-deck 'sha256sum ~/.local/bin/apparat.incoming; ldd ~/.local/bin/apparat.incoming'
```

Compare SHA-256 with the host value exactly. Stop on any mismatch or missing
library. Do not activate an incomplete transfer. Repair a mismatched copy by
recopying the same qualified artifact; do not compile a second application on
the Deck or disable SteamOS read-only protection to work around prerequisites.

### 3. Activate atomically and write the desktop entry

Run the following on the Deck via `ssh steam-deck`, only after the hash matches:

```bash
backup="$HOME/.local/share/apparat/deployment-backups/$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -m 700 -p "$backup"
mkdir -p ~/.local/share/applications
test ! -L ~/.local/share/applications/apparat.desktop
if test -e ~/.local/bin/apparat; then
    cp -p ~/.local/bin/apparat "$backup/apparat"
fi
if test -e ~/.local/share/applications/apparat.desktop; then
    cp -p ~/.local/share/applications/apparat.desktop "$backup/apparat.desktop"
fi
chmod 755 ~/.local/bin/apparat.incoming
mv ~/.local/bin/apparat.incoming ~/.local/bin/apparat
cat > ~/.local/share/applications/apparat.desktop.incoming <<'DESKTOP'
[Desktop Entry]
Version=1.0
Type=Application
Name=Apparat
Comment=Personal cluster console
Exec=/home/deck/.local/bin/apparat
TryExec=/home/deck/.local/bin/apparat
Path=/home/deck
Icon=applications-development
Terminal=false
Categories=Utility;
StartupWMClass=Apparat
DESKTOP
chmod 644 ~/.local/share/applications/apparat.desktop.incoming
mv ~/.local/share/applications/apparat.desktop.incoming ~/.local/share/applications/apparat.desktop
desktop-file-validate ~/.local/share/applications/apparat.desktop
```

Adapt all absolute paths if the user is not `deck`. Execute sequentially and
stop on an error. The executable's stable path is independent of the repository
or removable media; node state remains at `~/.local/share/apparat/node` by default.
The desktop entry does not depend on SSH's PATH or an SSH display variable.

### 4. Register through Steam's installed helper

Ensure the intended Steam account is logged in. Inspect existing shortcuts
under `~/.local/share/Steam/userdata/<account-id>/config/shortcuts.vdf` and the
library UI first. The numeric account directory is device-specific, not a
constant. Preserve a backup of the selected account's shortcuts file before
registration, but do not edit or restore it while Steam is running.

If Apparat already exists with the correct executable and desktop-entry paths,
skip registration. A repeat binary update at the same path needs no new shortcut.
For a first install run on the Deck:

```bash
DISPLAY=:0 steamos-add-to-steam /home/deck/.local/share/applications/apparat.desktop
```

The observed helper URL-encodes the path and passes `steam://addnonsteamgame/`
to the running Steam client. Steam creates and persists the native shortcut.
No VDF writer, Python dependency or Steam restart was required on SteamOS 3.9.5.
If the helper is missing or fails, use Steam's **Add a Non-Steam Game** dialog
in Desktop Mode, browse to `/home/deck/.local/bin/apparat`, and name it Apparat.
Do not retry registration blindly: inspect for an entry first to avoid duplicates.

In Gaming Mode open **Library → Non-Steam → Apparat → Play**. Use the native
Linux executable; leave forced Proton compatibility off. For controller input,
start with Steam's Gamepad template rather than keyboard/mouse emulation and
verify the application's actual mappings with physical controls.

## Verification

- Confirm exactly one Apparat entry in the intended account, executable
  `/home/deck/.local/bin/apparat`, empty launch options and native compatibility.
  Steam may choose the executable directory as StartDir rather than Desktop
  Entry's Path; the app resolves persistent state independently of that directory.
- `ssh steam-deck 'sha256sum ~/.local/bin/apparat'` must match the source artifact.
- `desktop-file-validate` must pass. `ldd` must show no missing dependency.
- Launch from the Steam library, not by assigning DISPLAY and executing Apparat
  directly. For remote automation, read the actual shortcut's unsigned `appid`
  (binary VDF, not text), calculate `(appid << 32) | 0x02000000` and run
  `DISPLAY=:0 steam steam://rungameid/<calculated-id>` on the Deck. Never hardcode
  another installation's ID or mutate VDF during readback.
- Check `pgrep -a -x apparat`, the Apparat window and startup events at
  `~/.local/share/apparat/node/last_run.log`. On the recorded Gaming Mode session
  `DISPLAY=:1 xwininfo -root -tree` showed the Apparat game window; discover the
  actual display when these differ. Report error fields without raw payloads.
- Exit and reopen from the same Steam entry to verify persistent installation
  and node reuse. Physical controller navigation, Steam+X, microphone and
  suspend/resume need their own evidence before claiming complete Deck support.

## Update, rollback and removal

After installing a build with the Linux updater, use Settings → Updates → Check for update, then Install update and Restart Apparat. The existing Steam entry keeps its fixed executable path. The updater retains `<executable>.previous`; compatible node data or an offline backup is still required for schema rollback.

For manual recovery or initial delivery of an updater-capable build, stop Apparat and repeat the staged transfer, digest comparison,
backup and atomic replacement. Keep the same path and Steam entry. Verify
startup and persistent data again. Do not recopy source or regenerate keys.

On activation/startup failure, stop Apparat and copy the saved executable to an
incoming file in `~/.local/bin`, set mode 755 and atomically rename it to
`apparat`. Restore the desktop entry if changed. If migrations ran, restore only
a compatible, verified offline node backup following `docs/database.md`; never
delete the runtime lock to defeat an owner or blindly run an older schema reader.
For a first installation there may be no previous executable to restore.

Remove just Apparat through Steam's **Remove Non-Steam Game** UI if desired,
then remove its desktop entry and binary. Preserve node data unless deletion
was separately requested. Leave other shortcuts and the enabled SSH service
alone. Restore a whole shortcuts backup only with Steam fully stopped and only
when no newer unrelated library edits would be lost.

## Lifecycle Compliance

Prompt → select/create plan → establish existing or new execution authority →
execute approved installation → update plan and docs → verify. Record source
revision, hash, target, backup locations, shortcut identity and observed versus
pending acceptance in the plan, platform matrix and append-only journal.
Refresh plan indexes, run `make check-docs` and `git diff --check`, and review
status/diff. Commit or push only under the applicable existing authorization.
