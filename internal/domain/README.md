# Domain

This package tree owns Apparat's product rules and durable concepts.

Domain code defines identity, projects, queues, tasks, capabilities, workload classes, commands, events, policies, and validation rules without depending on UI, SQLite, HTTP, WireGuard, model runtimes, or platform APIs.

## Boundaries

- May depend on the Go standard library and other domain packages.
- Must not depend on `internal/adapters`, `internal/platform`, Ebitengine, SQLite drivers, network clients, or source-reference checkouts.
- Must remain testable without a window, GPU, database, network, model runtime, microphone, or filesystem layout.

## Source inventory

- `doc.go`: Documents the internal/domain package boundary.

- `id.go`: Creates and validates canonical ULID identifiers without runtime services.

- `service.go`: Defines local service intent and safe immutable query values.

- `id_test.go`: Tests canonical durable identifiers and deterministic intent validation.

- `preferences.go`: Defines validated local display preferences and normal desktop window geometry.

- `project.go`: describes a local checkout registration and its observed branch and task labels.

project.go also defines the project-scoped workspace request and bounded text view shared by the runtime and UI.

- `cluster_page.go`: Defines the initial routing pools, detail routes and explicit worker commands.
