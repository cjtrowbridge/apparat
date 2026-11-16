# Logging Package

This package owns structured runtime diagnostics.

It provides append-only JSONL logging with rotation and reset-on-start `last_run.log` diagnostics for immediate debugging.

Logging APIs must redact secrets, tokens, private keys, passphrases, prompts, model outputs, raw voice data, and sensitive payloads by default.

## Source inventory

- `logging.go`: Configures structured logging and redacts sensitive fields.
- `logging_test.go`: Verifies configures structured logging and redacts sensitive fields.

- `redaction.go`: Recursively projects log fields into bounded JSON-safe redacted values.
