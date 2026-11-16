---
plan_id: 2026-11-16-16-00-00_repair-slow-update-downloads
title: Repair slow Android update downloads
summary: Restore tolerant download behavior and distinguish transfer failures from reachability failures.
status: past
created_at: 2026-11-16-16-00-00
---

# Repair slow Android update downloads

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Continuation of the authorized updater repair, bound to ROADMAP Phase 17 Android distribution. Preserve signing identity, monotonic versions, explicit installation, data and user device preferences. Pixel 10 access remains read-only; no installation approval received.

- [x] Correct the 15-second idle / 120-second total download regression, with bounded transfer and accurate failure reasons.
- [x] Verify actual transfer code against slow-progress, stalls, cancellation and size-limit cases; build and review candidate.
- [x] Record evidence and remaining device/publication acceptance without claiming the existing installed updater is repaired.

Evidence: actual Pixel 10 installed APK hash equals the retained version 1 baseline and its signer matches published version 2. Pixel DNS resolves GitHub, ping returns, host GET and HEAD return valid version 2 / HTTP 200. Actual phone retry downloaded a partial APK then reported unable_network and removed partial bytes. Logs omit the IOException cause, so exact idle-versus-total timeout remains unproven. Prior committed downloader allowed 60-second reads with no total deadline; Phase 8 introduced 15-second reads and a 120-second total cap. All IOExceptions misleadingly map to service-unreachable. Hypotheses: signer/version mismatch refuted for current published bytes; endpoint unavailability refuted during observed successful partial transfer; slow/stalled-transfer sensitivity supported by partial progress and restrictive limits. No assertion of a universal GitHub outage or exact timeout cause.

Use a small pure-Java transfer helper to test real copy/timeout boundaries with deterministic clocks and streams; retain the existing HTTP connection and installer owners. Restore 60-second idle tolerance, allow 15 minutes total, retain 128 MiB size/cancellation bounds and signature checks. Publish/install only under the corresponding existing or newly obtained authority; this plan does not override read-only Pixel 10 access.

## Verification and disposition

Canonical build, source/physical-size/docs gates, production Java transfer tests, GUI/HUD tests and Python tests passed. Candidate 0.1.2/code 3 uses the pinned signer and passes publication comparison against live GitHub version 2. Pixel 4a installed in place, then downloaded and verified the real GitHub APK in 5.465 seconds and correctly reported no_newer_update. That test validates transport and verification, not an installer handoff to an older APK. Prior Pixel 4a logs show version 2 failed after 124.343 seconds, supporting the total-deadline diagnosis; no exact IOException cause was logged then. The real network improved during testing; controlled production-code tests separately establish successful progress beyond 120 seconds with 40-second simulated read intervals. Tested idle-read exception, total cap, cancellation and 128 MiB boundary; no content or sensitive values logged. Do not claim a measured slow real-device download or that changing source repairs already installed versions.

Pixel 10 was inspected only; its installed code 1 APK hash/signature exactly match the retained baseline. No Pixel 10 install, uninstall or data clear. GitHub still serves code 2. Candidate remains `releases/android/arm64/apparat/latest.apk` and ignored `.tmp/pixel10-update-diagnosis/candidate-0.1.2.apk`; Linux build outputs restored to HEAD. Evidence in `.tmp/pixel10-update-diagnosis/`.

- [x] Publish the reviewed 0.1.2 source/APK checkpoint with approval, verify fetched bytes, then archive this plan and reconcile the roadmap. Existing phone versions retain their timeout behavior until they install the correction.

User now explicitly approves committing/pushing 0.1.2 including the three new files, and updating the connected Pixel 10 successfully. This supersedes the earlier read-only device boundary. Preserve app data; prefer the actual GitHub updater and verify installed bytes afterward.

## Publication and Pixel 10 completion

Published `c13b77f` to origin/main. Exact GitHub updater URL returns 0.1.2/code 3 with SHA-256 `681fb88f24bac4bb813c830a6b821b3f9b0af088dde411e7f90e35713e1f0867`; fetched package/signature/version gate passes against prior code 2. Pixel 10 used the actual app download and Android installer flow (permission_needed, installer_opened), upgraded from code 1 to code 3 and launched successfully. Pulled installed base.apk has the same SHA-256 as the published APK. App-private continuity marker survived and was then removed; no uninstall or data clear. Post-upgrade Settings displays Up To Date and exact downloaded/installed APK equality. No ADB install fallback was used. Evidence includes `pixel10-success.log`, `pixel10-new-settings.png`, `installed-0.1.2.apk` and `published-0.1.2-check.log` in ignored `.tmp/pixel10-update-diagnosis/`. All outcomes complete; historical pending/read-only statements above describe earlier checkpoints.
