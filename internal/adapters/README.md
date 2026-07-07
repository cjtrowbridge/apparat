# Adapters

Adapters connect Apparat's application and domain layers to external systems.

Examples include the Ebitengine HUD, SQLite repositories, HTTPS REST, WireGuard inspection, inference services, speech services, BOINC, Signal, Meshtastic, and filesystem or Git operations.

## Boundaries

- May depend on `internal/app`, `internal/domain`, and `internal/platform` only through narrow interfaces.
- Must keep blocking I/O, external process calls, model execution, SQL, and network calls outside the render/update path.
- Must return typed outcomes so the application layer can persist, retry, reject, display, or audit failures.
- Must not encode product authorization policy that belongs in the application or domain layer.
