# Logging Package

This package owns structured runtime diagnostics.

It provides append-only JSONL logging with rotation and reset-on-start `last_run.log` diagnostics for immediate debugging.

Logging APIs must redact secrets, tokens, private keys, passphrases, prompts, model outputs, raw voice data, and sensitive payloads by default.
