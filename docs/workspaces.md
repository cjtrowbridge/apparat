# Operating an owner-hosted workspace

A project keeps its stable registration ID when its display name changes. The selector and heading show `hostname://repository`. Android can use the system device name for a local project title before an Apparat hostname has been saved. Local work stays usable offline. For a registered worker checkout, Files, Editor, Tasks and Terminal use the confirmed owner's scoped TLS API. Chat on a remote checkout is reserved.

## Access

Every approved cluster member using an enrolled device can browse registered repositories and open files. Only a current cluster administrator can save files on another node. File permissions follow the cluster role across checkouts, including existing ones. Pending or blocked users and revoked devices cannot use the workspace API.

On the owning worker, the service account can inspect effective permissions with `apparatd workspace-access status PROJECT_ID USER_FINGERPRINT`. For a missing repository, check its registration and the user’s admission. For a denied save, check the current administrator role. Reopen Files after correcting the registration or membership.

Task execution and terminal use have separate permissions for ordinary members. Grant or revoke `tasks`, `terminal-view` or `terminal-control` with `apparatd workspace-access grant|revoke PROJECT_ID USER_FINGERPRINT SCOPE`. Terminal control also requires terminal view.

## Files, tasks and sessions

The selector loads up to 32 registrations per host, newest first. Older projects and Newest projects replace the visible page on that host. Members see the same registered repositories as administrators. Opening a file as a member shows a read-only view. Directory browsing remains independent of catalog size. Directories stream in pages of 64 raw entries. Continue before the cursor expires after one minute, or reopen the directory. The worker retains up to 64 directory handles. Text and task definition files have a 64 KiB limit. Conditional reads can reuse recent client cache entries after authorization. A remote save carries its base SHA-256 and stable request ID. Leave the draft intact after a disconnect and retry. If another edit wins, compare against the current owner file before rebasing and saving. External tools writing a file concurrently do not participate in Apparat's save lock.

Tasks lists built-in Apparat actions and definitions from VS Code JSONC tasks.json and launch.json. Built-in actions can be run and their text artifacts checked against SHA-256. VS Code definitions show their executor requirements. Task choices page by 32 and run history by 30.

An isolated Linux terminal host needs Bubblewrap with user namespaces, prlimit and a working systemd user manager. Its shell sees the checkout read-only, private temporary storage and a separate network namespace. Only one device controls a session. Other devices belonging to that user can view it and explicitly take control. Android can operate a Linux terminal. Reopening a running terminal resumes its screen. After the process ends, the owner retains the last screen and up to 8 MiB of recent output per session. Retiring old session directories is an operator responsibility.

When Files or a remote terminal is visible, a pinned WebSocket coalesces change notifications. Leaving the view releases its watch. After reconnect the owner supplies a fresh snapshot. Reads, saves and manual refreshes also work through the JSON API.

## Capturing performance

Open Cluster, select this device and open Performance. Save recent telemetry writes private NDJSON under `<node-root>/telemetry/`. A diagnostic session records bounded snapshots for up to 15 minutes or 4 MiB, rotating completed segments at 1 MiB. An administrator can request a combined capture from a confirmed worker. If the worker is unavailable the local capture remains available. Open a completed file to inspect pages of 24 records. Android offers the system document picker to export a copy.

Records include operation class, time, latency, measured bytes, row counts, outcome and an optional correlation ID. The transport measurements count application payload bytes, excluding TLS framing. Each node retains 1,024 recent events in memory and reports overwritten events in the capture header. No project names, filenames, file contents, typed commands or terminal output belong in these measurements. Recording and export perform file I/O off the request path. Captures support manual investigation. They do not change routing, pool policy or request limits.

## Verification

The owner tests exercise denied grants, base conflicts, retry after a lost acknowledgement, revocation and a directory with thousands of entries. The isolated Linux terminal test checks actual command output, controller transfer and resize. The telemetry tests exercise overflow, file permissions and failed writes. Release qualification uses the repository-wide build pipeline. A physical device acceptance check should include a separate editor, an external change and loss of the live connection.

In Files, each row starts with 📁 or 📄, followed by its Git marker when available and its name. Select anywhere on the row to enter the folder or open the file in the editor. An owner read failure appears above the listing.

Release acceptance exercises both roles against two isolated worker APIs, including signed changes on existing sessions, paged catalogs, nested file reads, legacy write grants and file bytes after a save or denial. A separate handler test verifies the exact anonymous JSON response across catalog, file operations and subscription requests with absent, expired and incorrectly bound credentials.
