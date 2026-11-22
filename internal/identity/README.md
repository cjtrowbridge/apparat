# Identity Package

This package owns local identity file handling and diagnostics.

It classifies identity state and contains user/device key, manifest, recovery, and consistency behavior as the identity system grows.

Identity code must redact private keys, passphrases, tokens, raw prompts, raw voice data, and other sensitive material before logging or returning diagnostics.

## Source inventory

- `identity.go`: Creates and persists signing identities with encrypted private-key storage.
- `identity_test.go`: Verifies creates and persists signing identities with encrypted private-key storage.

- `validation.go`: Validates bounded identity files and replaces them atomically.

- `validation_test.go`: Exercises strict identity parsing, key consistency and private-file permissions.

- `session.go`: Loads and locks the node signing session. A verified new generation becomes active through one atomic pointer change. Restoration checks any surviving valid manifest and retains the old generation.
- `setup.go`: Drives recovery, creation, masked password or system protection and optional encrypted portable backup using the platform's native dialogs.
- `access_unix.go`: Restricts the identity directory to the owner before creation.
- `access_windows.go`: Applies an inheritable current-user-only DACL before creation.
- `session_test.go`: Verifies cancellation, restart, locking, wrong-key rejection, recovery and explicit replacement with signing challenges.

- `personal.go`: owns the signed screen name and password-encrypted portable user key. The GUI stores the bundle in `user-identity/user.json` under its node root, separately from device keys.
- `personal_setup.go`: prompts for creation, restore, recovery export and unlock after native device setup. Cancel keeps the local console usable. The headless runtime has no personal session.
- `personal_test.go`: exercises restore, signed profile changes, incorrect passwords and rejection of a bare node backup.
- `device_binding.go`: Issues and verifies a domain-separated signed authorization for a distinct device public key and reads the corresponding signed user profile.
- `claim.go`: Proves possession of both keys for a specific worker, invitation and cluster. Changing the target invalidates the signatures.
- `policy.go`: Signs proposed administrative edits with both identities, binding the command ID, cluster and base revision.
