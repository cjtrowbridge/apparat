# Release Artifacts

Build outputs use this canonical path:

```text
releases/<goos>/<goarch>/<binary>/latest[.exe]
```

The `<goos>` and `<goarch>` directory names use Go's `GOOS` and `GOARCH` values, such as `linux/amd64`, `linux/arm64`, `windows/amd64`, `darwin/arm64`, or `android/arm64`. The `<binary>` directory is `apparat` for the GUI console and `apparatd` for the headless worker/service.

Generated `latest` and `latest.exe` binaries are ignored by Git. This directory keeps the release layout discoverable while preventing local binaries from entering source control.
