---
plan_id: 2026-11-20-00-20-00_steam-deck-local-install
title: Install Apparat persistently on Steam Deck
summary: Install the qualified Linux GUI in the Deck home directory and register a native Gaming Mode launcher with a reusable deployment playbook.
status: past
created_at: 2026-11-20-00-20-00
---

# Install Apparat persistently on Steam Deck

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

User authorization: the 2026-11-20 request to install the Linux binary in
`~/.local/bin`, register a non-Steam app and document future deployments;
the preceding request authorizes persistent SSH access.
Bind to ROADMAP Phase 17 / Steam Deck and Linux GUI / “Packaging, desktop
integration, diagnostics, update, and rollback” and Gamescope startup evidence.
This is a bounded installation checkpoint, not completion of Phase 17 or physical
controller, microphone, suspend/resume or full workspace qualification.
README selects Steam Deck first and persistent default node state beneath
`~/.local/share/apparat/node`; preserve that contract and all existing user work.

Use the tracked, previously qualified `releases/linux/amd64/apparat/latest`
artifact, SSH/SCP and the installed Valve `steamos-add-to-steam` helper. No new
application build, publication, branch, dependency or system package is needed.
Keep the SteamOS system partition read-only. Preserve existing binaries,
desktop files, Steam shortcuts and node data. Main uncertainty: launcher
registration and native runtime library/display compatibility on this Deck.

## Work and acceptance

- [x] 1. Verify boot-enabled SSH/key access, stage and hash-check the qualified GUI, atomically install `/home/deck/.local/bin/apparat` with rollback where needed.
- [x] 2. Create and validate a desktop entry, register it through Steam's installed helper, and verify exactly one matching non-Steam entry and native Gaming Mode launch.
- [x] 3. Write the reusable playbook and AGENTS route, update README/platform evidence and append the journal; run `make check-docs`, plan-index and whitespace checks.

## Evidence and stopping condition

Initial evidence: passwordless SSH succeeds as `deck`, `/etc/hostname` is
`steam-deck`, SteamOS 3.9.5, x86_64, `sshd` enabled/active. Steam is in Gaming
Mode. The source binary is tracked at delivery commit `dd1594c` with SHA-256
`7e2a9de3638342df95b3af66b2b834a52299c021e99836008b493c98243f96c4`.
The target has 82 GiB free in persistent `/home`. The installed Valve helper
accepts desktop entries and submits `steam://addnonsteamgame/` to running Steam.
Initial shortcuts file is 13 bytes; do not edit it while Steam owns it.

Stop after verified installation, a Steam-created launcher and observed startup,
with an explicit account of any remaining physical acceptance. Record deployment
hash, installed paths, shortcut identity and rollback instructions here and in
the playbook. Archive when the scoped outcomes and documentation checks pass.

### Installation acceptance — 2026-11-20

- Host bootstrap inspection: 3 passed, 0 failed. Checkout identity passed before
  first-party edits. Reused qualified release; no build was needed or claimed.
- Installed binary SHA-256 matches the source above; ELF dependencies resolve,
  mode 755. Desktop file validation passes, mode 644. Initial shortcuts backup:
  `/home/deck/.local/share/apparat/deployment-backups/20261120T160857Z/shortcuts.vdf`.
  Fresh installation had no prior executable or desktop entry.
- Steam helper created exactly one Apparat entry in account directory 80647528,
  appid 3167627568, URI `steam://rungameid/13604856810501570560`, native executable
  `/home/deck/.local/bin/apparat`, no launch options. Existing library entries
  were not edited directly. Repeat deployments skip registration at the same path.
- First launch PID 4991, graceful SIGTERM shutdown, second Steam launch PID 5147.
  Gamescope game display :1 showed an Apparat X window, 640x480. Both starts
  reached ready; the second recorded zero error/fatal events. Durable node reused
  at `/home/deck/.local/share/apparat/node`. Left the second launch open.
- SSH batch login succeeds with `IdentityAgent=none`; sshd enabled and active.
  Hostname discovery still times out. Alias uses current IP 10.0.0.197; reserve
  DHCP on the router or update HostName after address changes. No router edits,
  reboot, suspend, physical controller, microphone or broader workspace trial.
- Added the repeatable playbook, AGENTS route, README usage and platform evidence;
  journal appended. `make check-docs`, plan index checks and `git diff --check`
  pass. No source or product behavior changed and Phase 17 remains open.

Acceptance is satisfied for this bounded deployment; retain the native helper
path and stop. No independent VDF writer, launcher service or dependency was
needed. Source/installed bytes, one saved shortcut and two ready startups are
the comparison evidence. No application performance improvement is claimed.
