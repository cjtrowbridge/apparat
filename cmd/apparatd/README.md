# `apparatd`

`cmd/apparatd` is the headless worker and service entrypoint.

It loads configuration with binary name `apparatd`, creates the shared application runtime, supports `--smoke-test` and `--doctor`, and runs without initializing Ebitengine.

The release pipeline builds this command without the `gui` build tag into:

```text
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

Use this binary for display-free devices, service-manager integration, and headless smoke validation.
