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
