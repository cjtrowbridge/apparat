# Config Package

This package owns runtime configuration parsing and path resolution.

It defines GUI/headless/auto modes, command flags, environment precedence, binary-specific default runtime roots, and derived paths such as database, logs, identity, cache, artifacts, backups, recovery, and `last_run.log`.

Configuration code should stay deterministic and testable without opening databases, starting the GUI, or touching network services.
