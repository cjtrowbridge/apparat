# Platform

This package tree owns local runtime paths, OS integration, signal handling, service-manager boundaries, filesystem conventions, and platform capability detection.

Phase 3 entrypoints use platform signals for graceful shutdown and runtime configuration to keep durable state outside the project source tree by default.

## Boundaries

- May expose platform facts and lifecycle hooks to `internal/app`.
- Must not define product authorization, queue routing, project semantics, or HUD interaction rules.
- Must keep platform-specific behavior isolated behind small interfaces so Steam Deck, Debian, Windows, macOS, Android, and headless Linux remain testable independently.

## Source inventory

- `doc.go`: Documents the internal/platform package boundary.

- `lock.go`: Owns an OS-backed exclusive node lock and safe diagnostic metadata.

- `lock_unix.go`: Implements advisory exclusive locks on Unix and Android.

- `lock_windows.go`: Implements exclusive node ownership using a Windows byte-range lock.

- `migrate.go`: Migrates one unambiguous legacy runtime root without merging data.

- `lock_test.go`: Exercises real node ownership and non-destructive legacy-root migration.
