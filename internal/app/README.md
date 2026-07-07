# Application Runtime

This package owns the shared Apparat application runtime boundary.

It coordinates configuration, identity, persistence, modules, queues, services, and command dispatch before a GUI or headless adapter starts.

Phase 3 startup initializes runtime directories, append-only JSONL logging, SQLite migrations, identity status, cluster directory repositories, local messaging repositories, and doctor diagnostics.

## Boundaries

- May depend on `internal/domain`, `internal/adapters`, and `internal/platform`.
- Must expose mode-neutral application services to GUI, CLI, API, scheduler, and transport adapters.
- Must not import from `cmd/` or any package under `third_party/salvagecore`.
- Must keep GUI and headless modes as adapters around one runtime model.
