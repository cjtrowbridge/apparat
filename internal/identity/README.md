# Identity Package

This package owns local identity file handling and diagnostics.

It classifies identity state and contains user/device key, manifest, recovery, and consistency behavior as the identity system grows.

Identity code must redact private keys, passphrases, tokens, raw prompts, raw voice data, and other sensitive material before logging or returning diagnostics.
