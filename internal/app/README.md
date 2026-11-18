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

- `preferences.go`: Exposes typed local preference queries and save commands on the owning runtime.

- `projects.go`: registers and reopens local project checkouts, with per-repo account references and device-local defaults.
- `projects_test.go`: Exercises registry idempotency, reopening, external edits, reconnection and preservation on removal. Set APPARAT_TEST_PHASE9_GITHUB=1 only when authorized to test the two real example repositories.

workspace.go resolves every command by registered project ID and coordinates draft recovery. project_tasks.go executes version 1 snapshot-text tasks and publishes digest-verified artifacts.

- `workspace.go`: Runs project-scoped workspace commands against the registered checkout.

- `project_tasks.go`: Executes versioned owner-local text snapshot tasks declared by a project.

- `workspace_test.go`: Verifies editable working trees, selection isolation and durable workspace records.

- `workspace_save.go`: Records file transaction identity, base, outcome and rejection before allowing a replay.
