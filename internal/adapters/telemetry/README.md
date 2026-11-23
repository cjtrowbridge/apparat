# Workspace telemetry

capture.go records allowlisted operation names, monotonic durations, byte counts and returned rows in a 1,024-observation ring. Snapshots count overwritten observations. Capture files have a schema/build/role header followed by one JSON object per line. Export creates a private partial file and renames the completed capture. ReadPage uses a byte offset and returns up to 24 records. SQLite is independent of capture storage.

- `capture.go`: Retains a fixed ring of payload-free workspace observations.

- `capture_test.go`: Verifies bounded storage, redaction, capture failure and the exported record format.
