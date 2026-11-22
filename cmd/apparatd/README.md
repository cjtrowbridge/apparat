# `apparatd`

`cmd/apparatd` is the headless worker and service entrypoint.

It loads configuration with binary name `apparatd`, creates the shared application runtime, supports `--smoke-test` and `--doctor`, and runs without initializing Ebitengine.

At startup it prints the selected runtime root and `last_run.log` path before entering the service loop. By default on Linux, the headless runtime root is `~/.local/share/apparat/node` unless `XDG_DATA_HOME`, `APPARAT_RUNTIME_DIR`, or `--runtime-dir` overrides it.

The release pipeline builds this command without the `gui` build tag into:

```text
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

Use this binary for display-free devices, service-manager integration, and headless smoke validation.

## Source inventory

- `main.go`: Parses command options and runs the headless worker.

Phase 8 adds `--demo-services`, `--services`, `--backup`, and `--restore-backup`. GUI/headless are alternative owners of the same locked node root; see root README for migration and recovery. API hosting requires an explicit grant. Linux headless configuration uses api-list, api-grant RULE, api-revoke RULE and api-disable.

- `updates.go`: Coalesces request and owner signals, verifies the Linux worker release and initiates graceful restart.
