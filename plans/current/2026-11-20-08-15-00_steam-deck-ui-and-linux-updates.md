---
plan_id: 2026-11-20-08-15-00_steam-deck-ui-and-linux-updates
title: Fix Steam Deck presentation and Linux updates
summary: Move the embedded gallery gear, silence untargeted dictation, fill Gamescope and implement verified in-app Linux updates.
status: current
created_at: 2026-11-20-08-15-00
---

# Fix Steam Deck presentation and Linux updates

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

The user's four requested changes authorize implementation and local Deck
validation. Bind to ROADMAP Phase 17 Steam Deck/Linux GUI Gamescope sizing,
packaging, diagnostics, update and rollback, Phase 9B targeted dictation and the
existing optional STMPCK embedding milestone. README preserves Settings last;
the requested gear is the embedded gallery's settings control, not the host tab.
Keep the gallery's gear immediately after discovery in its centered local toolbar.
Keep the standalone dependency unchanged; extend the existing host embedding patch.

Use the existing GUI callback/status path for Linux update controls. Fetch the
public GitHub main commit, then its existing `releases/verified.json` and binary
at that same immutable commit. Require ordinary accepted/passed target metadata,
correct path/architecture, SHA-256 and native ELF before explicit installation.
Bound downloads, cancel on shutdown, preserve the installed binary on failure,
keep one previous executable and atomically replace on the same filesystem.
Restart only after GUI/core shutdown releases the node; Steam's stable executable
path needs no re-registration. This is the existing development channel over
HTTPS, not release-grade signed-manifest trust or automatic schema rollback.

No changes to other user work, TODO items, credentials, Steam shortcuts, system
partition or dependencies. Full release matrix and physical input/audio acceptance
are not implied by this scoped checkpoint. Canonical builds use no-flag
`scripts/build.py`; no alternate application compile path is permitted.

## Work and acceptance

- [x] 1. Move the embedded gear and its hit target after discovery; remove untargeted dictation text; choose native display geometry/fullscreen in Gamescope while preserving desktop resizing and saved placement elsewhere.
- [x] 2. Implement bounded Linux update check, staged verification, explicit installation and graceful restart with failure preservation and a previous binary. Exercise immutable revision, cancellation, corrupt/oversized/wrong-target rejection, no-op and atomic replacement in focused tests.
- [ ] 3. Run governed verification and the canonical build; deploy the qualified GUI on the Deck and capture actual Gamescope presentation plus update behavior. Record any unavailable acceptance explicitly.
- [ ] 4. Update closest READMEs, root README, roadmap/platform evidence, install playbook and append-only journal; refresh plan indexes and run documentation/source/whitespace checks.

## Evidence and stopping condition

Baseline: Deck native Apparat window is 640x480, Gamescope scales it with side
bars; app startup and node persistence passed the prior install plan. Desktop
Settings has no updater callback. Repository visibility was verified public
through authenticated GitHub on 2026-11-20; the Deck has no GitHub CLI and needs
none for this channel. Existing manifest contains accepted ordinary target
SHA-256 values. No release-grade signature or downgrade policy exists yet.

Acceptance requires working controls, digest-verified replacement with unchanged
Steam registration, full Deck display geometry and rendered evidence, passing
focused tests and canonical qualification. Stop when these scoped outcomes pass;
record limitations rather than extending into platform-wide certification.

### Linux delivery authority and target eligibility — 2026-11-20

The user declined Pixel testing while away from home and then explicitly said
“no just do it” in response to deferring delivery. Proceed with Linux build and
deployment now, including the canonical pipeline's reviewed commit/push. Keep
Android qualification separate and preserve its already-published artifact.
The no-flag build must report Android ineligible when its mandatory authorized
device is absent; it must still qualify and deliver the eligible Linux matrix.
Update `build_orchestrator.py`, `artifact_acceptance.py`, `build_delivery.py`
and their tests/docs to make that eligibility truthful without an alternate build
path or weakened acceptance for any built target. The user instruction takes
precedence over the previous all-target stopping condition. Settings-last and
canonical-main requirements remain intact. ROADMAP Phase 17 continues to require
independent evidence for Android and each other platform.

An initial automatic approval review rejected the publishing build before this
latest authorization. No rejected command was executed; finish and review source
before retrying the authorized no-flag pipeline.

### Verification before delivery

- Full `go test ./...` passes. GUI and self-update tests pass with `-race` under
  the bootstrap-provisioned Xvfb. Tests execute the complete check/download,
  explicit install, restart request and up-to-date states using an isolated
  executable; separate adapter tests reject wrong target, path, state, schema,
  digest, ELF, changed files, cancellation and oversized downloads.
- All 64 Python pipeline tests pass, including Linux delivery that preserves
  the older Android artifact/input evidence. Source governance (195 files),
  400-line and documentation checks pass.
- Host prerequisite inspection found Python 3.10 and missing native development /
  Xvfb / ImageMagick tools. Provisioned only ignored local sysroot files through
  the existing reviewed `bootstrap.py --native-source` path, staged from Ubuntu
  packages. Replaced Python 3.11-only digest calls with bounded chunked reads in
  existing build/speech scripts; no system Python or system package was changed.
- Unified patch context needs a scoped Git whitespace attribute for literal
  space-before-tab syntax; generated Go retains normal tab indentation.
- Target screenshot, canonical artifact qualification and Deck update/restart
  remain the next acceptance work. No application artifact has been built through
  an alternate entrypoint or certified by package tests.

### Build recovery evidence

Main advanced to `dc43070` during work. Preserved local changes in a Git stash,
fast-forwarded this same designated main and reconciled conflicts. Retained the
new upstream Settings-entry Android discovery, scrolling, responsive defaults and
model/dictation behavior. The user's original TODO bytes remain unstaged; the
LICENSE is identical to the newly tracked upstream file. No other checkout or
branch was created. GUI/updater race tests pass after reconciliation.

The first canonical compile succeeded for both Linux artifacts. GUI acceptance
failed before any window in EGL initialization; the rejected run is
`.tmp/build-runs/tmp_v29_ccm/report.json`. No promotion or publication occurred.
Hypotheses: application fullscreen change, NVIDIA provider, inherited host EGL
platform. Host is not Gamescope, and forcing Mesa alone still reproduced the
same failure. Explicit X11 EGL selection passes all actual GUI checks, including
STMPCK play and divider screenshots: host `EGL_PLATFORM=wayland` was reaching
Xvfb's X11 display. Keep the acceptance subprocess platform explicit and retry
only through the canonical build.

A separate intermittent preflight failure recurred while pathlib descended into
an ephemeral private `.aws` mount. The physical-line checker now prunes excluded
and private mounts before descent and still raises included-directory errors;
65 pipeline tests pass, including a regression for traversal pruning. This fix
supports acceptance item 3 without weakening the 400-physical-line gate.
