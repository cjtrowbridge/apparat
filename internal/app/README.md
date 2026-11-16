# Application Runtime

This package owns the shared Apparat application runtime boundary.

It coordinates configuration, identity, persistence, modules, queues, services, and command dispatch before a GUI or headless adapter starts.

Phase 8 startup migrates one unambiguous root, locks ownership, initializes logs/SQLite/identity, composes services and REST, and publishes readiness last. Failures unwind in reverse order. One sequential five-second supervision loop probes mock services with cooperative one-second deadlines; late results are revision-fenced. Queries are shared directly by GUI/headless/REST. No HUD is owned here.

Every initialized runtime must create or replace `last_run.log` at `config.LastRunPath`, include that path in doctor diagnostics, and emit enough component-level events to debug startup failures without needing the durable JSONL history first.

## Boundaries

- May depend on `internal/domain`, `internal/adapters`, and `internal/platform`.
- Must expose mode-neutral application services to GUI, CLI, API, scheduler, and transport adapters.
- Must not import from `cmd/` or any package under `third_party/salvagecore`.
- Must keep GUI and headless modes as adapters around one runtime model.

## Source inventory

- `runtime.go`: Initializes and coordinates shared local application runtime resources.
- `runtime_test.go`: Verifies initializes and coordinates shared local application runtime resources.

- `queries.go`: Exposes transport-neutral node and service inspection plus local commands.

- `services.go`: Coordinates service intent, bounded probes and coalesced change hints.

- `cleanup_test.go`: Verifies every initialized startup stage unwinds resources before returning.

- `lifecycle_test.go`: Injects startup failures and checks shared state through actual adapters.

- `services_unit_test.go`: Tests service orchestration without SQLite, networking, GUI or real clocks.
