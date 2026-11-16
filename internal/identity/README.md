# Identity Package

This package owns local identity file handling and diagnostics.

It classifies identity state and contains user/device key, manifest, recovery, and consistency behavior as the identity system grows.

Identity code must redact private keys, passphrases, tokens, raw prompts, raw voice data, and other sensitive material before logging or returning diagnostics.

## Source inventory

- `identity.go`: Creates and persists signing identities with encrypted private-key storage.
- `identity_test.go`: Verifies creates and persists signing identities with encrypted private-key storage.

- `validation.go`: Validates bounded identity files and replaces them atomically.

- `validation_test.go`: Exercises strict identity parsing, key consistency and private-file permissions.
