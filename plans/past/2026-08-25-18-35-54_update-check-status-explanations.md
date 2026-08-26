---
plan_id: 2026-08-25-18-35-54_update-check-status-explanations
title: Update Check Status Explanations
summary: Give Settings update checks typed, honest outcomes with safe, actionable explanations for current, unavailable, and failure states.
status: past
created_at: 2026-08-25-18-35-54
---

# Update Check Status Explanations

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Define an explicit, safe update-status contract for the Settings update control.
  - [x] 1.1 Bind the work to the temporary Phase 5 Android update bridge and preserve the Phase 17/Phase 9 release-manifest and remote-update work as out of scope.
  - [x] 1.2 Replace ambiguous raw status handling with typed states for checking, up to date, update available, unable to check, permission needed, installer opened, and unable to install.
  - [x] 1.3 Give every non-success state a stable, non-sensitive reason code and a short user-actionable explanation.
  - [x] 1.4 Prohibit raw exception messages, response bodies, tokens, private paths, and unverified release data from the Settings surface and ordinary logs.

- [x] 2. Render concise status and wrapped explanation in Settings.
  - [x] 2.1 Keep the button label concise: `Checking...`, `Up To Date!`, or the applicable action/failure state.
  - [x] 2.2 Render a wrapped explanatory status beneath the button for unavailable, permission, installer, and failure outcomes.
  - [x] 2.3 Report `Unable to check for updates` with a clear local explanation when the current desktop build has no configured update checker; do not imply a successful check.
  - [x] 2.4 Preserve controller, keyboard, mouse, and touch activation, 44px touch targets, and visible status changes without blocking the GUI update loop.

- [x] 3. Map platform update paths to the contract.
  - [x] 3.1 Map Android matching APK hashes to `Up To Date!` and retain the existing user-confirmed installation path for differing verified APKs.
  - [x] 3.2 Map Android network, HTTP, download, hash, bridge-registration, permission, and installer failures to distinct safe explanation/reason pairs.
  - [x] 3.3 Keep desktop limited to its honest no-checker outcome; do not add unsigned manifest fetching, automatic installation, or remote node updates in this plan.

- [x] 4. Add evidence, documentation, and verification.
  - [x] 4.1 Add GUI/unit tests covering each status/explanation mapping, button activation, and redaction of raw error content.
  - [x] 4.2 Update the GUI and Android/update documentation with the status contract, current desktop limitation, and later signed-manifest boundary.
  - [x] 4.3 Run focused GUI tests, relevant full Go/build checks, documentation/code-size/whitespace validation, and record that Android tooling is unavailable on this host.
  - [x] 4.4 Build and visually inspect the Settings Updates fieldset on the host, proving `Up To Date!` and explanatory unable-to-check behavior where each outcome can be exercised.
