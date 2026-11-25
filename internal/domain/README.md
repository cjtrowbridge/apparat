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

- `device_name.go`: validates Apparat host labels and resolves display names without changing endpoint selection.

The canonical delivery allocates the number in release.txt before compilation. ReleaseVersion returns that embedded number for the running process, including offline node health and Linux Settings.

- `git_provider.go`: Defines authenticated provider capabilities and host operation progress.

- `chat.go`: defines session identity, scope, descriptive metadata and ordered turns. A submitted turn with no reply stays pending indefinitely. Redelivery of the same turn and content is idempotent.

- `chat_model.go`: provider-neutral requests, job IDs, assistant tool calls and correlated results shared by synchronous and deferred providers.
- `inventory.go`: typed timestamped worker observations, installed model metadata, capacity counters and fresh text capability checks.

- `model_choice.go`: Identifies a chat model independently of the worker chosen to serve a turn.

- `inference_claim.go`: binds jobs and digests to an inference roster and arbitrates in-memory reservations. The table starts in recovery. Authenticated transitions reject conflicting claims, stale withdrawals and attempts to withdraw an owned job.
- `inference_timing.go`: records bounded full-roster confirmation and applied-event timing samples. Failed rounds have a separate counter. Five samples of each type enable a three-times-p99 recheck interval, with a 250 ms floor and gradual decreases.
- `inference_job.go`: defines roster identities, model-step records, canonical digests and the four-priority ordering.

- `global_artifact.go`: Global session replica envelope and deterministic transcript writer.
- `inference_claim_test.go`: Checks adaptive timing, cold-start votes, withdrawal replay and owner fencing.
