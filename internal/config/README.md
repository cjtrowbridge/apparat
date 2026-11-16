# Config Package

This package owns runtime configuration parsing and path resolution.

It defines GUI/headless/auto modes, command flags, environment precedence, one default node root and explicit independent roots, and derived paths such as database, logs, identity, cache, artifacts, backups, recovery, and `last_run.log`.

Configuration code should stay deterministic and testable without opening databases, starting the GUI, or touching network services.

## Source inventory

- `config.go`: Defines configuration values and their defaults.
- `config_test.go`: Verifies defines configuration values and their defaults.
