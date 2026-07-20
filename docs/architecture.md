# Architecture Contract

Apparat is a ports-and-adapters Go application with one shared runtime and several presentation, transport, persistence, platform, and workload adapters.

## Status

This file is the target architecture contract. The current repository implements startup, SQLite, identity, logging, messaging, and a five-tab GUI mockup, but several package, command/query, lifecycle, and adapter boundaries below remain planned. `ROADMAP.md` defines the evidence required before a planned boundary is treated as implemented.

## Package Boundaries

- `cmd/apparat`: GUI console entry point.
- `cmd/apparatd`: headless worker/service entry point.
- `internal/app`: shared runtime orchestration, command dispatch, module registration, service lifecycle, and mode selection.
- `internal/domain`: product rules, identity concepts, project ownership, queue ownership, workload classes, commands, events, policies, and validation.
- `internal/hud`: presentation-owned mockable HUD state, tab order, focus, input actions, voice states, mock projections, and diagnostics; it must not own durable backend truth.
- `internal/adapters`: external-system integrations such as GUI, persistence, API, transport, model, speech, BOINC, Git, filesystem, and service adapters.
- `internal/platform`: OS lifecycle, paths, service-manager hooks, signals, and platform capability discovery.
- `third_party`: source-reference checkouts isolated from root Go package discovery unless a later approved plan activates an adapter.

Domain packages must not import GUI, SQLite, HTTP, WireGuard, model runtime, BOINC, Signal, Meshtastic, or source-reference packages.

## Roles

- GUI console: renders HUD state and dispatches user commands.
- Headless worker: runs without Ebitengine initialization and hosts queues, services, tasks, or API endpoints.
- Service host: advertises typed workload capabilities and executes leased work.
- Queue owner: receives REST submissions, authoritatively validates and admits jobs, answers worker lease requests, schedules, cancels, audits, and records authoritative results for one queue.
- Project owner: stores and runs a project's Git working tree, advertises its project summary, serves authorized project REST operations, owns its Task entrypoints, authoritatively applies project transactions, and keeps rejected edits durable.
- Enrollment authority: issues short-lived invites and signs directory updates.

One device may hold several roles.

During the MVP, evaluating trigger bindings and persisting Task runs is part of the Project-owner role. There is no independently authoritative scheduler role or elected scheduler. Later failover may move an explicitly defined ownership lease, but cannot create a second Task authority.

## Project, Pipeline, And Task Ownership

- A Project is a Git repository with one authoritative owner: the device on which its working tree lives and runs.
- Every enrolled device projects one cluster-wide Projects list containing its local projects plus all authorized remote project summaries. Remote projects retain their owning device identity and online/stale/unavailable state.
- Project summaries may be replicated or cached for discovery and offline display. Files, Git state, Task definitions, and authoritative run state remain on the owner and are accessed through that owner's authenticated REST API.
- A Pipeline is a Project that declares at least one Apparat-executable entrypoint. Pipeline identity and ownership are the Project's identity and ownership; it is not a separate repository or scheduler object.
- A project entrypoint is a Task. A Task belongs to exactly one Project in the MVP and is authoritatively defined, validated, invoked, and recorded by that Project's owner.
- A Task may have zero or more trigger bindings. Zero means manual execution only. Trigger bindings may represent intervals/cron-like schedules, authenticated webhooks, internal application events, or cluster events.
- Trigger delivery creates or requests a Task run; it does not move the Task definition or Project authority to the triggering device.
- Task steps may call owner-local constrained project operations and may submit typed jobs to separately owned routing queues.

## Owner-Directed REST Request Flows

All cross-device application requests use the authenticated REST API. Devices do not read or write another device's SQLite database or project working tree directly.

Project flow:

1. Project owners advertise signed, authorization-filtered summaries.
2. Each device merges local summaries with cached/refreshable remote summaries into its cluster-wide Projects projection.
3. A remote read, Git action, Task invocation, or mutation is sent to the Project owner.
4. The owner authenticates, authorizes, validates, executes or persists the request, and returns the authoritative state or durable rejection.

Queue flow:

1. A requester submits a job by REST to the queue-owning device with stable job, message, correlation, and idempotency identity.
2. The owner validates the requester, schema, workload/capability requirements, policy, limits, quota, and queue state before durably accepting or rejecting it.
3. An authorized inference worker in the queue polls or long-polls the owner for work and presents its identity and current capabilities.
4. The owner chooses compatible work and returns a lease with a lease ID, attempt identity, fencing token, deadline, and bounded payload/artifact references.
5. The worker executes the lease and sends heartbeat/progress only as allowed, then posts a signed terminal result or failure to the owner.
6. The owner validates the worker, active lease/fencing token, result schema, artifacts, and idempotency before recording authoritative completion.

The queue owner never relies on a worker's local database as queue truth. Lease expiration permits recovery or reassignment, and late or duplicated completion requests cannot complete the logical job twice.

## Binary And Runtime Boundaries

`apparat` and `apparatd` are built as separate release artifacts under `releases/<goos>/<goarch>/<binary>/latest[.exe]`.

- `apparat` is compiled with the GUI build tag by the release pipeline and enters the Ebitengine run loop during normal execution.
- `apparatd` is compiled without the GUI build tag and must remain safe on devices without desktop libraries or a display server.
- The current implementation uses binary-specific default runtime roots, so GUI and headless smoke runs produce separate databases, logs, caches, and `last_run.log` files unless `--runtime-dir` overrides them. This is current evidence, not the final node-ownership model.
- `last_run.log` is reset on each process start for immediate debugging, while append-only JSONL logs under `logs/` retain durable structured history.

The target one-node contract is:

- `apparat` and `apparatd` embed the same headless-capable core and are alternative process forms of one logical node.
- One default node runtime root contains one device identity, SQLite database, service inventory, artifact store, and operational state.
- An exclusive node-runtime lock enforces one authoritative process/SQLite writer. Starting the other artifact against an owned root fails safely and reports the owner.
- Simultaneous GUI and daemon operation requires a separately approved daemon-client mode in which the daemon owns the core and database and the GUI is a versioned client.
- Multiple intentionally independent nodes on one host require explicit runtime roots, identities, ports, and local-service ownership. They may not implicitly discover and advertise the same provider endpoint.

GUI navigation, focus, selection, layout, scrolling, gestures, animation, and unsaved editor text remain presentation state. Identity, trust, directories, Projects, durable drafts, Tasks, queues, jobs, services, capabilities, artifacts, retries, backend configuration, and cached remote state belong to the shared core. Audio capture remains platform/presentation state until explicit submission creates a durable core job.

## Local Inference Driver And Instance Boundary

Apparat uses statically compiled provider drivers registered explicitly at composition roots; it does not load Go dynamic plugins or use hidden package-global registration. The target contract separates:

- `WorkloadClass`: provider-independent operation type.
- `DriverKind`: provider protocol such as `ollama`, `openai_compatible`, `llama_cpp`, `automatic1111`, or `comfyui`.
- `ServiceID`: stable identity for one configured localhost endpoint with independent desired state, health, admission, supervision, and advertisement policy.
- `CapabilityID`: stable identity for one observed model, modality, format, feature set, or limit offered by a service.

A driver factory validates configuration and opens instances. Each instance exposes inspection and typed workload executors. One manager keyed by `ServiceID` supervises arbitrary instances, including multiple instances of the same driver and workload class; secondary indexes support lookup by class, driver, model, health, and policy. Failures, concurrency, cancellation, inventory refresh, and shutdown are isolated per instance.

Remote devices invoke logical service/capability identities through the authenticated Apparat gateway. The gateway applies authorization, queue ownership, routing, limits, audit, and artifact policy before reaching a localhost provider. Provider URLs and credentials remain local.

Service advertisements are derived from safe desired policy and observed state. Each carries an owner-scoped monotonic revision, observation timestamp, and expiration. The default lifetime is 120 seconds with refresh by 60 seconds. Expired advertisements are immediately non-routable, remain visibly stale for up to 24 hours, and require a newer fresh observation to become routable again.

## Artifact Boundary

Artifact bytes live in the owning node's runtime artifact store; SQLite contains metadata, ownership, SHA-256 digest, size, MIME type, authorization, provenance, job/Project links, transfer state, and retention policy. Cross-device transfer uses authenticated bounded HTTPS with resumable ranges or chunks and end-to-end digest verification. A queue result is not authoritative until the queue owner validates its active lease, artifact authorization, declared metadata, size, and digest. Cleanup is owner-policy-driven and must not leave authoritative records pointing silently at deleted bytes.

## Modules, Commands, Events, And Store

Feature modules register stable IDs, routes, commands, reducers, effects, repositories, health checks, and view-model producers.

The canonical flow is:

1. Controller, keyboard, pointer, API, scheduler, or transport input becomes a typed command.
2. The application validates identity, authorization, state, ownership, workload class, and capability requirements.
3. Durable intent is recorded before retryable external effects.
4. Adapters execute effects outside render/update paths.
5. Typed events record progress, success, failure, cancellation, timeout, or rejection.
6. Reducers update durable records and snapshots.
7. The HUD renders snapshots only.

## Workload Classes

Initial classes are:

- `text_generation`
- `image_generation`
- `video_generation`
- `speech_to_text`
- `text_to_speech`
- `research_boinc`

Workload classes are versioned, provider-independent, and extensible. Runtime names, model IDs, BOINC project IDs, and service endpoints are capability fields, not workload classes.

## Capability Contracts

Devices, services, queues, routes, and jobs declare authorization-appropriate subsets of:

- Workload class and schema version.
- Runtime/provider and stable logical service identity. Provider endpoints appear only in owner-local desired configuration and never in remote advertisements.
- Supported models, projects, modalities, formats, limits, and artifacts.
- Hardware, accelerator, memory, storage, concurrency, and queue limits.
- Streaming, progress, cancellation, health, availability, load, and validation time.
- Privacy, authorization, power, schedule, quota, retention, and owner policy.

Routing first excludes incompatible capabilities, then applies priority and fallback order.

## Salvagecore Decisions

- Copy: no wholesale source tree.
- Adapt: ports/adapters, shared GUI/headless runtime, command/event/store flow, module registration, SQLite lifecycle ideas, identity diagnostics, JSONL logging, mock-data-first UI development.
- Rewrite: HUD input model, tab layout, routing, identity, persistence, API, and service adapters in Apparat-native packages.
- Reject: RPG-first product framing, qTox/Tor primary transport, mouse-first input as the foundation, Ebitengine alpha as the default, and Python Whisper as embedded ASR.
