# Application Runtime

This package owns the shared Apparat application runtime boundary.

It coordinates configuration, identity, persistence, modules, queues, services, and command dispatch before a GUI or headless adapter starts.

Phase 3 startup initializes binary-specific runtime directories, reset-on-start `last_run.log` diagnostics, append-only JSONL logging, SQLite migrations, identity status, cluster directory repositories, local messaging repositories, and doctor diagnostics.

Every initialized runtime must create or replace `last_run.log` at `config.LastRunPath`, include that path in doctor diagnostics, and emit enough component-level events to debug startup failures without needing the durable JSONL history first.

## Boundaries

- May depend on `internal/domain`, `internal/adapters`, and `internal/platform`.
- Must expose mode-neutral application services to GUI, CLI, API, scheduler, and transport adapters.
- Must not import from `cmd/` or any package under `third_party/salvagecore`.
- Must keep GUI and headless modes as adapters around one runtime model.
