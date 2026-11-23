# Persistence Adapter

This package owns SQLite repositories, migrations, backup/restore and read-only inspection. One connection, per-connection foreign keys and bounded busy timeout protect small local transactions. See [database operations](../../../docs/database.md) for operations and recovery.

SQLite is the durable store for identities, metadata, chats, transactions, events, queue state, indexes, and workflow state. It does not replace filesystem project directories or Git repositories.

Persistence code must expose application-level repositories and must not leak raw SQL rows into GUI packages.

## Source inventory

- `dependencies.go`: Records the package dependency boundary.

- `backup.go`: Produces consistent SQLite snapshots and validates disposable restores.

- `cluster.go`: Persists cluster device profiles, change records and cursors in SQLite.

- `cluster_test.go`: Verifies persists cluster device profiles, change records and cursors in SQLite.

- `database.go`: Opens and initializes SQLite storage.

- `database_test.go`: Verifies opens and initializes SQLite storage.

- `messaging.go`: Persists message envelopes and replay state.

- `messaging_test.go`: Verifies persists message envelopes and replay state.

- `schema.go`: Owns ordered forward migrations including transitional development tables.

- `services.go`: Persists service intent and replaces observations transactionally.

- `services_test.go`: Verifies transactional services, stale-probe fencing and backup recovery.

- `retry.go`: Classifies SQLite contention without treating cancellation or corruption as retries.

- `contention_test.go`: Exercises real SQLite lock contention and cancelled transaction admission.

- `preferences.go`: Owns migration 4 and independent atomic SQLite records for display preferences and window placement.

- `projects.go`: owns the schema-v6 registry and account preferences. Removing a registration only deletes its SQLite row.

workspace.go owns migration 7 and stores project-scoped drafts, local pending messages and run outcomes. Unregistering a project cascades to its workspace records.

- `workspace.go`: Stores owner-local drafts, pending chat messages and manual run outcomes.

- `user_devices.go`: Migration 9 records signed public profiles by user fingerprint and device authorizations by user and device fingerprint. Profile revisions never regress; equal conflicting revisions fail. SQLite receives public records only.
- `claim.go`: Preserves migration 10 for existing databases. First authentication now creates ownership in the login transaction.
- `claim_test.go`: Verifies target substitution fails and competing first claims produce one approved admin.
- `policy.go`: Migration 11 and signed administrative policy transactions. Replicas verify the authority, actor, revision and previous hash before applying changes. The last administrator cannot be removed.
- `policy_sync.go`: Rebuilds committed policy heads from SQLite on startup, checks admin visibility on every request and pages at most 64 events. An up-to-date authorized caller avoids reading the event table.
- `claimed_workers.go`: Preserves migration 12 and stores the authenticated worker fingerprint and address for reconnection.
- `admission.go`: Migration 13 and dual-key user registration. Unknown users start pending on open nodes. Blocked users and revoked device keys cannot obtain service access.
- `pool_tasks.go`: Migration 14 creates one task table across pools, an attempt history, artifact references and ordered per-pool revisions. Hash-text admission checks membership, dispatch rechecks current grants, and stale or cancelled leases cannot commit results.
- `pool_task_views.go`: Checks current admission and device authorization before returning a creator's task revisions or artifact reference.
- `pool_tasks_test.go`: Covers denied pool access, duplicate submissions, expired leases, stale completion, result integrity and cross-user visibility.
- `queue_heads.go`: Hydrates per-user pool cursors on startup and publishes new heads after commit. A synchronized reader checks its current grant and cached cursor without reading the event journal.
- `governance_view.go`: Reads the current cluster authority, admission switch, signed user profiles and pool memberships for a currently approved administrator in a single read transaction.
- `replica.go`: Builds reader-specific governance snapshots, checks device signatures, hydrates the in-memory view revision, and reads per-pool local cursors.
- `replica_apply.go`: Applies signed snapshots and reader-specific queue pages in SQLite transactions. Cursors and data advance together. A changed role replaces the cached view and removes inaccessible pools.
- `replica_policy.go`: Verifies proposal signatures, authority signatures, owner handoff, revision ordering, and chain hashes before committing a page of admin receipts locally.
- `groups.go`: Applies signed group creation and membership changes. Pool access queries use the same expanded grants for individual and group principals.
- `result_receipt.go`: Signs each completed hash result for its specific creator, pool, task and worker lease. Consumers can verify the receipt and artifact bytes independently.
- `device_directory.go`: adds platform metadata and revision triggers, stores signed GUI pool assignments and constructs the admission-filtered roster. Worker membership reflects its hash-text service policy. The pinned worker test checks assignment and removal through sync.

- `remote_workspace.go`: Stores per-project host permissions, client drafts and bounded save receipts.
