# Platform

This package tree owns local runtime paths, OS integration, signal handling, service-manager boundaries, filesystem conventions, and platform capability detection.

## Boundaries

- May expose platform facts and lifecycle hooks to `internal/app`.
- Must not define product authorization, queue routing, project semantics, or HUD interaction rules.
- Must keep platform-specific behavior isolated behind small interfaces so Steam Deck, Debian, Windows, macOS, Android, and headless Linux remain testable independently.
