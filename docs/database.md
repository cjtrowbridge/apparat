# Database Contract

SQLite stores durable application metadata and workflow state; ordinary project files remain filesystem and Git data.

## Layers

- Identity, certificates, trust, enrollment, and directory records.
- Device and typed capability records.
- Project metadata, chats, artifacts, transactions, drafts, and cursors.
- Typed services, queues, jobs, leases, attempts, results, and artifacts.
- Events, outbox, inbox, idempotency keys, replay records, and audit logs.
- Tasks, schedules, webhooks, run history, and approval state.
- Research project catalog, validation state, budgets, work-unit references, and provenance.

## Ownership

Each MVP project, queue, and task has exactly one authoritative owner device. Remote changes are idempotent submitted transactions.

## Migrations And Data Types

Migrations are forward-only and checksumed. IDs are ULID-compatible sortable strings. Timestamps are UTC integer milliseconds; human schedule timezone data is stored explicitly when needed.

Phase 3 implements the first SQLite lifecycle:

- Open, close, and ping via the cgo-free `modernc.org/sqlite` driver.
- Enable foreign keys at connection startup.
- Keep WAL opt-in until platform behavior is validated; Phase 3 does not claim universal WAL support.
- Apply forward-only migrations with stored SHA-256 checksums.
- Detect migration checksum mismatches.
- Provide read-only diagnostics for user version, foreign-key status, and migration count.
- Provide sortable local IDs and UTC millisecond helpers.

## Backup, Repair, Restore, And Encryption

The MVP defines backup, repair, restore, and integrity checks before broad replication. Optional at-rest encryption remains a Phase 1 decision input, not a default until key storage and recovery are designed.
