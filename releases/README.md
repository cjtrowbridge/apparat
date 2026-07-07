# Release Artifacts

Build outputs use this canonical path:

```text
releases/<goos>/<goarch>/latest[.exe]
```

The `<goos>` and `<goarch>` directory names use Go's `GOOS` and `GOARCH` values, such as `linux/amd64`, `linux/arm64`, `windows/amd64`, `darwin/arm64`, or `android/arm64`.

Generated `latest` and `latest.exe` binaries are ignored by Git. This directory keeps the release layout discoverable while preventing local binaries from entering source control.
