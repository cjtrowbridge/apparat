# HTTPS REST API Contract

The MVP API is authenticated HTTPS REST over externally configured WireGuard or trusted LAN. Authentication and authorization are mandatory on every non-public endpoint.

Phase 8 implements six read-only loopback resources. OpenAPI operations marked `x-phase: 8` have executable loopback response-conformance coverage. The scoped worker listener serves signed login, synchronization, administration and the Git provider operations documented below. The worker also serves the scoped workspace protocol. The general catalog and resource-specific project URLs remain future contracts.

All cross-device project access, Task invocation, queue submission, worker leasing, progress/heartbeat, and result delivery use this API. No device accesses another device's project filesystem or SQLite database directly.

REST is an external adapter, not the GUI-to-core boundary. The GUI invokes the transport-neutral internal application commands and queries directly in-process. REST handlers invoke selected operations from that same internal API; neither adapter accesses SQLite directly or implements parallel product rules.

The first implementation step is a loopback-only, read-only REST subset for health, readiness, safe local device state, services, and capabilities. Before cluster identity, authentication, authorization, signing, replay protection, and audit are active, the server rejects non-loopback binding and exposes no command or setter. The secure REST phase preserves those query semantics, admits configured LAN/WireGuard access, and maps authenticated mutations onto the same internal commands already used by the GUI.

OpenAPI source: [`openapi/apparat-v1.yaml`](./openapi/apparat-v1.yaml).

## Worker workspace transport

`GET /v1/workspace/catalog?before=REGISTRATION_ID` returns up to 32 worker registrations. Use the last row’s NextCursor for the next page. Administrators see remote-created registrations. Other approved users see remote-created registrations with an explicit read grant. Responses omit account names.

`POST /v1/workspace` accepts a single JSON command with a registered ProjectID, Action and checkout-relative Path. Authentication uses the existing pinned TLS listener and device-bound worker session. A write includes Text, the last-read Base digest and a stable operation ID. Success returns a WorkspaceView, including entries, choices, a revision or a continuation cursor as appropriate. A stale base returns 412. Approved cluster admins inherit all checkout scopes. An unmatched scope for another approved user returns 403 with workspace_grant_required and the required scope. The worker checks current roles during live subscriptions and terminal supervision. Detailed failure text and project contents stay out of transport telemetry.

`GET /v1/workspace/live` upgrades the same authenticated TLS connection to WebSocket. Its first frame selects up to sixteen authorized project directory or terminal watches. The server returns a reset on subscription, then sequenced invalidations or screens. A client reconnects by requesting a fresh snapshot. [Workspace operation and diagnostics guide](workspaces.md)

## Endpoints

- `GET /v1/health`: health, version, readiness, and clock state.
- `GET /v1/device`: authenticated device identity, roles, and trust state.
- `GET /v1/capabilities`: typed device/service capability descriptors.
- `GET /v1/services`: list authorization-filtered logical services owned by this device.
- `GET /v1/services/{service_id}`: read safe service identity, policy projection, health, revision, and expiry without provider-local connection data.
- `GET /v1/services/{service_id}/capabilities`: list safe capabilities currently observed for one logical service.
- `GET /v1/projects`: list projects authoritatively owned by this API device.
- `GET /v1/projects/{id}`: read an owned project's authorized summary and current availability.
- `GET /v1/projects/{id}/tasks`: list the owned project's Apparat Task entrypoints and trigger summaries.
- `POST /v1/projects/{id}/tasks/{task_id}/runs`: manually request a Task run; a Task does not require a trigger.
- `POST /v1/jobs`: submit an asynchronous workload job through the receiving device's local routing façade.
- `GET /v1/jobs/{id}`: inspect job state.
- `POST /v1/jobs/{id}/cancel`: request cancellation.
- `POST /v1/queues/{queue_id}/jobs`: submit a job to the authoritative queue owner.
- `POST /v1/queues/{queue_id}/claims`: let an authorized inference worker poll or long-poll the owner for a compatible leased task.
- `POST /v1/queues/{queue_id}/leases/{lease_id}/heartbeat`: renew or report bounded progress for an active lease.
- `POST /v1/queues/{queue_id}/leases/{lease_id}/complete`: post a signed result or failure to the owner for authoritative validation and completion.
- `GET /v1/events?after={cursor}&wait={duration}`: cursor-based event long-polling.
- `POST /v1/project-transactions`: submit owner-authoritative project mutations.
- `POST /v1/artifacts`: begin a bounded authenticated artifact upload and receive an upload resource.
- `GET /v1/artifacts/{artifact_id}`: read authorized metadata or download verified bytes with range support.

Mutating requests require `Idempotency-Key`. Asynchronous success returns `202 Accepted`, a durable resource ID, and a `Location` header.

Requests enforce schema version, content type, body size, deadlines, bounded concurrency, authentication, authorization, replay protection, and redaction-safe errors.

The API exposes no generic remote shell, arbitrary process execution, or unrestricted tool endpoint.

## Service And Capability Semantics

A workload class is not a provider or instance identity. `service_id` identifies one owner-local configured provider endpoint, and `capability_id` identifies one observed model/modality/feature set under that service. A device may expose arbitrary services, including multiple instances of the same driver and workload class.

Remote resources expose logical IDs, driver kind, display name, safe health/availability, capability metadata, policy, revision, observation time, and expiration. They never expose localhost URLs, credential references, tokens, raw provider responses, prompts, or results. The owner rejects expired, disabled, unavailable, incompatible, or unauthorized service targeting before execution.

Service advertisement revisions are monotonic per owner. The default lifetime is 120 seconds with refresh by 60 seconds; expired inventory becomes non-routable immediately and may remain visible as stale diagnostics for up to 24 hours.

## Project Catalog And Task Semantics

A Project is a Git repository owned by the device where its working tree lives and runs. `GET /v1/projects` returns only projects owned by the responding device; it does not relabel cached remote projects as local. Each Apparat instance builds its cluster-wide Projects list by combining local results with signed/cached summaries and authorized REST results from every project owner. Offline owners remain visible from cached summaries with explicit stale/unavailable state.

A Pipeline is a Project with one or more Task entrypoints. It has no separate API authority or ownership record. `GET /v1/projects/{id}/tasks` returns those entrypoints. Trigger bindings are separate Task metadata: a Task may have no binding and be run manually, or may be bound to intervals, authenticated webhooks, internal application events, or cluster events.

Project detail, file/Git operations, Task discovery/execution, and mutations are served by the Project owner through constrained project resources. Remote writes use idempotent project transactions; the API never exposes raw filesystem access or an unrestricted shell.

## Queue Owner And Worker-Pull Semantics

The device identified as a queue's owner is the only authority that may admit, order, lease, cancel, or complete that queue's jobs.

1. A requester sends `POST /v1/queues/{queue_id}/jobs` to the owner with stable job/correlation identity and an idempotency key.
2. Before admission, the owner validates requester authorization, body/schema, workload class, capability requirements, policy, limits, quota, retention, deadline, and current queue state. Acceptance is persisted before `202 Accepted`; rejection is explicit and safe to retry only according to its error class.
3. An authorized queue member sends `POST /v1/queues/{queue_id}/claims`, including worker identity, current service/capability IDs, accepted workload classes, available concurrency, and a bounded long-poll duration.
4. The owner either returns `204 No Content` or one lease containing queue/job/attempt/lease IDs, a fencing token, deadline, workload request, and bounded artifact references.
5. The worker may heartbeat while executing and then posts a signed terminal success or failure. Completion is idempotent and is accepted only from the leased worker with the active fencing token.
6. The owner validates the result and artifacts, then records authoritative completion. Lease expiry allows reassignment; late, stale, or duplicated completion cannot complete the logical job twice.

Workers never pull directly from requesters, accept work by reading replicated queue rows, or make their local result authoritative. The queue owner never pushes unleased work directly into a worker process.

## Artifact Semantics

Artifact metadata binds artifact ID, owner, SHA-256 digest, size, MIME type, provenance, authorization, retention, and relevant Project/job/result identity. Uploads and downloads enforce limits, deadlines, content ranges or chunk identity, and authorization. Partial bytes remain non-authoritative until the receiver verifies length and digest and atomically finalizes the artifact.

Workers may return bounded inline results or artifact references. A queue owner validates the active lease/fencing token, artifact owner and authorization, metadata, size, and digest before accepting authoritative completion. Expired or deleted artifacts return an explicit durable state rather than an apparently successful empty download.

## Current loopback operations

Phase 8 exposed the following local diagnostic projections. The discovery release removes the legacy --listen switch and isolates bootstrap requests from these handlers. Default startup opens no listener. The diagnostic router remains available to package tests. Product listeners use the isolated bootstrap handler described below.

GET `/v1/health` returns version, lifecycle state and readiness; `/v1/device` returns OS, architecture and **local bootstrap-file** identity status. `/v1/services` and `/v1/services/{service_id}` return ID, name, mock driver, desired enablement/revision, health, observation time, safe error and capabilities. `/v1/capabilities` and `/v1/services/{service_id}/capabilities` expose service-scoped capability IDs and workload classes. Canonical ULIDs are required in paths. No provider address, filesystem path, credential reference, raw error, fail-injection switch or key is serialized.

Responses are JSON with `Cache-Control: no-store` and a server-generated `X-Request-ID`. Accept negotiation supports JSON and media wildcards with positive quality. Bodies are rejected (400), non-GET methods return 405/Allow, unsupported Accept returns 406, unknown paths/IDs return 404, unavailable/busy returns 503 and cancelled queries return 504. Unexpected failures return safe codes, never raw errors. Admission is eight in-flight handlers; query deadline two seconds; read/write limits three seconds; headers 8 KiB; encoded responses at most 1 MiB; graceful shutdown three seconds. Query implementations must honor context cancellation. No CORS opt-in, command or setter is exposed.

OpenAPI uses JSON-compatible YAML so Go's standard JSON decoder can parse the same contract without another dependency. Handler tests validate the six route schemas and required fields. Future schemas and mTLS declarations remain separately labeled in that file.

## First discovery milestone

The initial anonymous milestone used a dedicated bootstrap handler serving the anonymous GET /v1 response from the contract below (401 with JSON explaining that authentication is required). It cannot dispatch protected or diagnostic handlers. Unknown paths receive JSON 404 and unsupported methods JSON 405. The installed worker’s current public-key, login and protected routes are described under First-owner enrollment and Pinned worker API below. Start the discovery listener only through the [hosting policy](node-identity.md#api-hosting-policy).

## Planned Phase 10 local authentication contract

The following responses apply only on an explicitly enabled network/interface pair. A disabled server or unapproved ingress has no API listener, including for public-key lookup and the anonymous 401. Phase 10 replaces Phase 8's automatic loopback listener with explicit grants. Outbound API clients operate independently of hosting. See [API hosting policy](node-identity.md#api-hosting-policy).

Extend the existing versioned read-only handler. The public-key URL works before and after authentication, including while the signing key is locked. See [node identity and credential issuance](node-identity.md).

| Request | Application response |
| --- | --- |
| GET /v1 without valid credentials, or anonymous access to a protected resource | 401, Bearer challenge and `{"error":{"code":"authentication_required","message":"Authentication is needed to use the API."}}` |
| GET /v1 with a valid local token | 200 and `{"api_version":"v1","authenticated":true}` |
| GET /v1/identity/public-key with or without credentials | 200 with algorithm (Ed25519), encoding (standard base64), public_key, and fingerprint (SHA-256 of raw public bytes in unpadded base64url). Return only the device signing identity. |
| Public key requested with absent, inconsistent or revoked identity | 503 with error code identity_unavailable. Return no partial key material. |
| Existing health, device, service and capability GET routes | Preserve successful projections for authorized callers. Require the same authentication middleware and return 403 for insufficient scope. |

All application responses, including 400, 401, 403, 404, 405, 406, busy and internal failures, use application/json and one error object with stable code and safe message. Preserve request IDs, no-store, timeouts, concurrency bounds and Allow on 405. Invalid and expired tokens follow the missing-credential response. Authenticate before protected resource lookup. Update diagnostic clients and smoke checks together with enforcement. A failed TLS handshake precedes the HTTP response.

### First-owner enrollment on a granted Linux worker

On an explicitly enabled interface the selected worker issues a short-lived challenge at `GET /v1/auth/challenge`. The app posts the signed user profile, device binding and both signatures to `POST /v1/auth` over device-authenticated TLS. The first verified login by the locally confirmed owner atomically creates the approved administrator. The response identifies current admission and role. The client saves the worker fingerprint for repeat connections. First contact probes TLS without presenting client credentials. The operator verifies the device fingerprint against the host console before login. `GET /v1/public-key` exposes its public identity. A pending user receives no service token.

### Pinned worker API

On a granted interface, the headless worker also accepts a TLS 1.3 session authenticated by a device certificate and signed user login. The authenticated worker API is separate from the Phase 8 local diagnostics listener. JSON `GET /v1/admin/view` returns the currently authorized administrator view. Signed writes use `POST /v1/admin/propose`; the current owner device signs the explicit initial handoff through `POST /v1/admin/order-on-worker`.

`POST /v1/sync` accepts a worker view revision and up to 32 pool cursors for the current reader. A changed view returns its signed snapshot first. Otherwise, each requested visible queue returns at most 64 transitions. An unchanged batch returns `{}`. The app verifies the snapshot’s worker signature, cluster, user and device, and verifies ordered administrator receipts against the original owner device through the handoff. Access changes invalidate hidden views. The hash-text task endpoints submit, cancel, read task history and retrieve a result by task ID. New results carry a worker-signed receipt binding the artifact digest to its creator, task, pool and fenced execution attempt.

### Host GitHub account directory

`GET /v1/admin/git/accounts` returns a JSON object with a `usernames` array for github.com. Requests require TLS 1.3, the enrolled device certificate, a live bearer session bound to that device, current SQLite administrator approval, and a host-local confirmation of the owner. The response uses no-store. Tokens, Git credential helpers and other providers are outside this endpoint. The provider operation below can enroll accounts on eligible workers. Local CLI login is also available. Existing workers keep their cluster access and must run `apparatd trust-host OWNER_FINGERPRINT` before sharing this directory.

Apparat pins the long-lived worker signing key. Reissued leaves must verify against that key. Invalid certificate chains abort the handshake. A valid change displays the old and presented fingerprints and defaults to keeping the host disconnected. Approval applies to this installation. The replacement must subsequently confirm administrator access for the same user and cluster before the saved registration moves. Restoring the original worker key preserves historical signing continuity.

### Worker provider operations and project registration

The confirmed worker’s TLS admin route `GET /v1/admin/git/providers` reports the compiled provider IDs, saved usernames, install and enrollment availability, and a reason for unavailable features. The UI always offers anonymous Git. Discovery announcements and cluster policy feeds omit this directory.

`POST /v1/admin/git/operation` accepts a provider ID, a client-generated ULID, and install, enroll, clone, status, or cancel. A clone also takes a GitHub owner/repo and optional account. The worker selects executable, arguments, source hostname and destination. Status is bound to the initiating admin user and device. Active work periodically rechecks authorization. One worker operation runs at a time. Completed status expires after fifteen minutes. Repeating a clone uses its checkout receipt and registration to recover across worker restarts. Enrollment codes are ephemeral and only the initiating device sees them. Linux enrollment stages CLI files in a private tmpfs directory and requires an unlocked default Secret Service collection. Confirm on the GitHub verification page; the host saves its account through the system CLI. An interrupted machine can leave temporary files in its private tmpfs directory until reboot.

`GET /v1/admin/git/projects` lists remote-created checkout summaries from this worker. The app presents the worker hostname and revision in Projects. This compatibility route returns read-only metadata. The GUI uses the paged workspace catalog and the scoped workspace routes for project operations. Additional destination roots remain future work.

#### Client operation sequence

Use the pinned worker listener on an explicitly enabled interface. All four Git routes require the same TLS 1.3 device certificate, live device-bound bearer, current SQLite admin grant and host-local owner confirmation. The GitHub account authorizes a transfer at GitHub. It cannot authenticate the caller to Apparat.

| Step | JSON contract |
| --- | --- |
| Discover | `GET /v1/admin/git/providers` returns an array with `ID`, `Name`, `Tool`, `Installed`, `Accounts`, `CanInstall`, `CanEnroll` and `Reason`. Interpret capability flags independently. |
| Start | POST to `/v1/admin/git/operation` with `{"Provider":"github","Operation":"clone","ID":"01ARZ3NDEKTSV4RRFFQ69G5FAV","Repo":"example/project","Account":""}`. Omit or empty the account for anonymous GitHub access. Use `install` or `enroll` with their own new IDs. |
| Follow | POST the same provider and ID with `Operation` set to `status` or `cancel`. The GUI polls pending jobs every two seconds. HTTP 200 carries a `State` of `pending`, `completed` or `failed`. Read `Message` for progress or failure, `Code` during device approval, `Username` after enrollment and `Project` after a clone. |
| Recover | A different initiating device cannot poll or cancel the job. Reuse the original ID and identical inputs to retry a clone. An unrelated request for an occupied checkout fails. After restart or status expiry, refresh the worker listing to find a completed registration. |

Jobs have a 16-minute deadline. The worker rechecks active authorization every second. It retains up to 128 jobs, including terminal results until expiry. Invalid inputs or provider return 400 at request validation, missing status returns 404, and a busy worker or conflicting reuse of an ID returns 409. Some input and transfer failures arrive in a terminal job status. Missing or device-mismatched bearer credentials receive 401 and insufficient current access or absent owner confirmation receives 403. TLS verification can end the connection before HTTP.

#### Host prerequisites

The installer uses fixed local package commands. Linux supports apt, dnf and zypper under root or an authorized noninteractive sudo rule. Platform adapters also recognize Homebrew on macOS, WinGet on Windows and pkg inside Termux. A standalone mobile app can act as a client. Release 0.1.72 qualified the Linux worker and Linux/Android GUI clients. An installer adapter alone does not qualify another OS as a network worker.

Remote enrollment currently requires a Linux worker with multi-account GitHub CLI, tmpfs at `/dev/shm` and an unlocked default Secret Service collection accessible through Python GI in the service account’s session. The client directs approval to `https://github.com/login/device`. Only the worker handles the token. It verifies the enrolled username, checks protected storage and preserves the previous active CLI account when publishing the configuration. If the chooser reports enrollment unavailable, use the worker account’s local CLI login and refresh. If provider discovery itself fails, verify the API network grant, saved worker fingerprint, owner confirmation and current admin access before retrying.
