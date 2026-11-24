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
- `claim.go`: Confirms the selected hostname and invited worker fingerprint in local dialogs, signs with both unlocked identities and records the acknowledged worker.
- `personal_pair.go`: Opens or joins one explicitly approved two-code pairing session. The Linux sender selects an enabled API network grant. The receiver unlocks its portable identity locally and records the signed worker pins.
- `reconcile.go`: Coalesces peer sync after local activity, refreshes on the worker-configured interval, retries unavailable workers, and invalidates denied reader views.

- `cluster_pages.go`: Projects current worker governance and revision-bound commands.
- `cluster_task_pages.go`: Reads pool queue history and submits or cancels tasks.

- `pool_overview.go`: reads the current user's verified worker replicas and latest task states for a routing pool. Administrator views expand direct, public and group grants for approved users. Adding a user rechecks live governance and signs a revision-bound grant for the selected worker. Offline browsing shows the last synchronized state.

- `worker_trust.go`: Reviews worker identity changes before login and resolves host account directories through authenticated worker sessions.
- `worker_trust_test.go`: Verifies explicit consent and default rejection for the fingerprint comparison dialog.

- `project_hosts.go`: Resolves provider requests on their owning worker and retrieves remote checkout metadata.

- `owner_workspace.go`: Routes remote project operations to their registered owner and keeps drafts on the editing device.

- `performance.go`: Runs private workspace captures away from the UI and the SQLite write path.

- `runtime_identity.go`: Handles identity setup and locks for a running node.

- `workspace_catalog.go`: Reads bounded VS Code JSONC catalogs on the checkout owner without running their commands.

- `workspace_terminal.go`: Maps terminal viewing and control to separate explicit grants on the project owner.

- `workspace_cache.go`: Keeps a bounded client cache of recently verified owner reads.

- `workspace_live.go`: Subscribes only while a foreground client is using an owner-backed directory.

- `owner_workspace_test.go`: Verifies scoped remote operations on an actual SQLite registration and working file.

- `workspace_roles_acceptance_test.go`: Exercises signed role changes, paginated registrations, nested file access, disk-verified saves and revocation through two isolated worker APIs.

### Shared chat sessions

- `chat.go`: exposes trusted in-process create, inspect, submit and complete operations on an initialized GUI or headless Runtime. Pass a registered filesystem checkout ID, stable session/turn ULIDs and the project or global scope. Global sessions require the caller to select the owning repository. Duplicate delivery returns the original turn. A changed payload or a second outstanding turn on the same session returns an error. Other sessions can independently await replies. A pending turn holds no execution slot.
- `chat_test.go`: exercises owner registration, concurrent retries, restart, reply correlation and preservation of externally changed transcripts. Run through the governed Go test entrypoint for this package. Remote grants, provider dispatch and presentation belong to later harness work.
