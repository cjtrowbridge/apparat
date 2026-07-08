# Internal Packages

This tree contains Apparat implementation packages that are not public Go APIs.

Packages in `internal/` should preserve the ports-and-adapters boundary: product rules and durable state belong in application/domain packages, while external systems are isolated behind adapters.

Every package with source files needs a local `README.md` that explains ownership, boundaries, and operational assumptions.

## Inventory

- `adapters`: GUI, persistence, and future external-system adapters.
- `app`: shared runtime orchestration.
- `cluster`: local cluster directory repositories and capability records.
- `config`: runtime configuration and directory resolution.
- `database`: SQLite lifecycle and migrations.
- `domain`: durable product vocabulary and rules.
- `hud`: mockable HUD state and input-facing view model.
- `identity`: local user/device identity files and diagnostics.
- `logging`: JSONL and last-run diagnostic logging.
- `messaging`: durable inbox/outbox/replay primitives.
- `platform`: OS lifecycle and platform capability boundaries.
