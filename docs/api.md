# HTTPS REST API Contract

The MVP API is authenticated HTTPS REST over externally configured WireGuard or trusted LAN. Authentication and authorization are mandatory on every non-public endpoint.

This file is the target API contract. The OpenAPI source is a planning skeleton until the secure REST phases implement handlers, clients, conformance tests, and authorization evidence.

All cross-device project access, Task invocation, queue submission, worker leasing, progress/heartbeat, and result delivery use this API. No device accesses another device's project filesystem or SQLite database directly.

REST is an external adapter, not the GUI-to-core boundary. The GUI invokes the transport-neutral internal application commands and queries directly in-process. REST handlers invoke selected operations from that same internal API; neither adapter accesses SQLite directly or implements parallel product rules.

The first implementation step is a loopback-only, read-only REST subset for health, readiness, safe local device state, services, and capabilities. Before cluster identity, authentication, authorization, signing, replay protection, and audit are active, the server rejects non-loopback binding and exposes no command or setter. The secure REST phase preserves those query semantics, admits configured LAN/WireGuard access, and maps authenticated mutations onto the same internal commands already used by the GUI.

OpenAPI source: [`openapi/apparat-v1.yaml`](./openapi/apparat-v1.yaml).

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
