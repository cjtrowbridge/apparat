# Operating an owner-hosted workspace

A project keeps its stable registration ID when its display name changes. The selector and heading show `hostname://repository`. Local work stays usable offline. For a registered worker checkout, Files, Editor, Tasks and Terminal use the confirmed owner's scoped TLS API. Chat on a remote checkout is reserved.

## Access

Run the following on the owning worker using its service account and normal node root. Read the project ID from its registration and use the approved personal identity fingerprint. Repeat for each permission required.

```sh
apparatd workspace-access grant PROJECT_ID USER_FINGERPRINT read
apparatd workspace-access grant PROJECT_ID USER_FINGERPRINT write
apparatd workspace-access revoke PROJECT_ID USER_FINGERPRINT write
```

Read permits directory listings, text, task definitions and verified run results. Write and tasks also require read. Terminal access uses terminal-view, with terminal-control additionally required to start, send input, take over, resize or stop. Access to GitHub credentials and permission to clone are separate from checkout grants.

## Files, tasks and sessions

The selector loads up to 32 registrations per host, newest first. Older projects and Newest projects replace the visible page on that host. An approved non-admin sees worker checkouts explicitly shared for reading. Directory browsing remains independent of catalog size. Directories stream in pages of 64 raw entries. Continue before the cursor expires after one minute, or reopen the directory. The worker retains up to 64 directory handles. Text and task definition files have a 64 KiB limit. Conditional reads can reuse recent client cache entries after authorization. A remote save carries its base SHA-256 and stable request ID. Leave the draft intact after a disconnect and retry. If another edit wins, compare against the current owner file before rebasing and saving. External tools writing a file concurrently do not participate in Apparat's save lock.

Tasks lists built-in Apparat actions and definitions from VS Code JSONC tasks.json and launch.json. Built-in actions can be run and their text artifacts checked against SHA-256. VS Code definitions show their executor requirements. Task choices page by 32 and run history by 30.

An isolated Linux terminal host needs Bubblewrap with user namespaces, prlimit and a working systemd user manager. Its shell sees the checkout read-only, private temporary storage and a separate network namespace. Only one device controls a session. Other devices belonging to that user can view it and explicitly take control. Android can operate a Linux terminal. Reopening a running terminal resumes its screen. After the process ends, the owner retains the last screen and up to 8 MiB of recent output per session. Retiring old session directories is an operator responsibility.

When Files or a remote terminal is visible, a pinned WebSocket coalesces change notifications. Leaving the view releases its watch. After reconnect the owner supplies a fresh snapshot. Reads, saves and manual refreshes also work through the JSON API.

## Capturing performance

Open Cluster, select this device and open Performance. Save recent telemetry writes private NDJSON under `<node-root>/telemetry/`. A diagnostic session records bounded snapshots for up to 15 minutes or 4 MiB, rotating completed segments at 1 MiB. An administrator can request a combined capture from a confirmed worker. If the worker is unavailable the local capture remains available. Open a completed file to inspect pages of 24 records. Android offers the system document picker to export a copy.

Records include operation class, time, latency, measured bytes, row counts, outcome and an optional correlation ID. The transport measurements count application payload bytes, excluding TLS framing. Each node retains 1,024 recent events in memory and reports overwritten events in the capture header. No project names, filenames, file contents, typed commands or terminal output belong in these measurements. Recording and export perform file I/O off the request path. Captures support manual investigation. They do not change routing, pool policy or request limits.

## Verification

The owner tests exercise denied grants, base conflicts, retry after a lost acknowledgement, revocation and a directory with thousands of entries. The isolated Linux terminal test checks actual command output, controller transfer and resize. The telemetry tests exercise overflow, file permissions and failed writes. Release qualification uses the repository-wide build pipeline. A physical device acceptance check should include a separate editor, an external change and loss of the live connection.
