---
plan_id: 2026-11-16-15-20-00_repair-github-updates
title: Preserve installed signing identity for GitHub updates
summary: Pin the installed Android signing certificate, advance release versions and verify a data-preserving upgrade.
status: current
created_at: 2026-11-16-15-20-00
---

# Preserve installed signing identity for GitHub updates

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

User authorized the recommended fix: keep the Pixel 10 installed signing key and repair the GitHub update process. Bind to Phase 17 Android updates, preserving Phase 8 state. Canonical test device remains Pixel 4a; no uninstall or data clear, no Pixel 10 direct install. Do not weaken signature or downgrade checks. Existing main worktree contains prior authorized implementation; preserve it. Publication/commit boundaries follow the host commit playbook; complete and validate the candidate before the final approval boundary if required.

- [x] Pin the installed signer, fail on key loss/mismatch, advance manifest to version 2, and add a release verification gate requiring a newer version than the supplied published APK.
- [x] Distinguish non-newer versions from signer/package failures in updater UI without exposing sensitive data.
- [x] Run relevant tests/full build; compare signer/version to the Pixel 10 baseline and published APK; test a data-preserving Pixel 4a upgrade, document and prepare the GitHub candidate.
- [ ] Publish the validated APK and matching source checkpoint through approved Git workflow; fetch published bytes and verify hash/signature/version.

Baseline: GitHub APK is valid but versionCode=1 with signer fccb1e58a4b5edbe19c101248e122237b5338a6aaf01828d1abf5f8034a11e36. Locally installed Pixel builds are versionCode=1 with signer 706c5f922a16d569c0441ac015816198fd7b688ba6faa891193276ef56a32ad0. Candidate must use the latter signer and version>1. Three hypotheses checked: corrupted APK refuted by apksigner; non-increasing version confirmed; mismatched signer confirmed. Raw evidence `.tmp/update-diagnosis/` and `.tmp/update-repair/`. Retain the existing ignored keystore, never commit its private material. Future builds require restoration of that key on other hosts. This development key remains separate from deferred production/store signing custody.

Validation: canonical full build, 57 Python tests, native GUI/HUD tests under Xvfb and full `make verify` passed. Candidate 0.1.1/versionCode 2 matches the retained Pixel 10 certificate; both published-baseline and Pixel-10-baseline publication checks pass. SHA-256: `def2c5e17c907d345da3bfc8da5c213f7502e30d2ceb041c49697dedbf0aa18e`. Pixel 4a upgraded from version 1 to 2; a follow-up in-place installation preserved an app-private continuity marker, then removed only that marker. Initial marker command had a shell-redirection quoting error; the corrected tee-based check is authoritative. Pixel updater visibly rejects current GitHub APK with the distinct signer explanation. No uninstall, data clear or Pixel 10 installation. Final candidate remains at `releases/android/arm64/apparat/latest.apk`; raw evidence is ignored under `.tmp/update-repair/`.

Review: signature and monotonic-version guards remain intact; missing/mismatched local keys fail before compilation instead of silently creating a new identity. Manual manifest increments retain deterministic build inputs. The publication validator requires a freshly downloaded previous APK; it does not itself publish or eliminate concurrent publication races. Production key custody, signed manifests and rollback remain Phase 17. Users of the former GitHub signing identity cannot upgrade in place to this selected installed-device identity; this repair specifically preserves the user's Pixel builds.

Remaining boundary: host commit playbook requires explicit commit/push approval and selection of new files. Prepared new-file inventory: `.tmp/update-repair/new-files.txt`. Prior authorized implementation, submodule revisions, icon tabs and headings remain uncommitted alongside this repair, so publication must include the matching source checkpoint. No commit/push performed; GitHub is not fixed until publication and fetched-byte verification complete.

Publication authorization: user explicitly approved committing and pushing the prepared changes, including the listed new files, with the proposed commit message. Proceed on existing main; retain private signing material outside Git.
