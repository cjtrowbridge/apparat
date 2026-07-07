# Persistence Adapter

This package tree will contain SQLite repositories, migrations, backup, repair, and read-only inspection boundaries.

SQLite is the durable store for identities, metadata, chats, transactions, events, queue state, indexes, and workflow state. It does not replace filesystem project directories or Git repositories.

Persistence code must expose application-level repositories and must not leak raw SQL rows into GUI packages.
