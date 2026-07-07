# Domain

This package tree owns Apparat's product rules and durable concepts.

Domain code defines identity, projects, queues, tasks, capabilities, workload classes, commands, events, policies, and validation rules without depending on UI, SQLite, HTTP, WireGuard, model runtimes, or platform APIs.

## Boundaries

- May depend on the Go standard library and other domain packages.
- Must not depend on `internal/adapters`, `internal/platform`, Ebitengine, SQLite drivers, network clients, or source-reference checkouts.
- Must remain testable without a window, GPU, database, network, model runtime, microphone, or filesystem layout.
