# Commands

This directory contains executable entrypoints.

Each command must stay thin: parse configuration, create the shared runtime, select command-mode behavior, and delegate product behavior to `internal/` packages.

## Inventory

- `apparat`: GUI console entrypoint.
- `apparatd`: headless worker/service entrypoint.

Command packages should not own durable product rules, SQL, queue policy, transport semantics, or HUD state. Put those concerns under `internal/`.
