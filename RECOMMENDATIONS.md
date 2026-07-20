# Apparat Architecture Recommendations And Integrated Roadmap

Reviewed: 2026-07-19

## Purpose And Status

This document reviews the current Apparat codebase at the point where it is primarily a cross-platform UI mockup with local runtime, SQLite, identity, logging, cluster-directory, and messaging foundations. It recommends cleanup and structural work before backend-heavy features make the current seams expensive to change, and it integrates the remaining implementation program from [`ROADMAP.md`](./ROADMAP.md).

This remains advisory rather than atomic implementation authority. Any accepted code, schema, protocol, dependency, build, or release change still requires a focused approved execution plan under `plans/future/` or `plans/current/` and must remain consistent with [`README.md`](./README.md). Until an explicit documentation migration retires or narrows it, [`ROADMAP.md`](./ROADMAP.md) remains canonical when the two documents conflict.

## Status Key And Migration Scope

- `[ ]` not started
- `[x]` completed and verified
- `[?]` implemented but still needs validation
- `[-]` intentionally closed or deferred

This file is being shaped to supersede the unfinished portion of the roadmap without losing goals, dependencies, checklist detail, exit criteria, open decisions, or MVP criteria. It therefore includes:

- The documentation and planning gates that must happen before code changes.
- The structural recommendations produced by the current codebase review.
- Every incomplete or validation-pending item from legacy Phases 3 and 5.
- Every planned task, feature, dependency, and exit criterion from legacy Phases 6 through 14.
- The roadmap's cross-cutting requirements, open decisions, and MVP completion definition.

Completed Phases 0 through 4 remain useful historical evidence in `ROADMAP.md`; they are not duplicated wholesale here. Their optimistic completion claims that conflict with executable evidence are explicitly reopened below. When this document is promoted to canonical roadmap authority, archive the completed historical phases or retain a concise completion ledger rather than silently deleting that history.

## Documentation And Planning Work Must Happen First

The following work precedes implementation. Documentation changes are product and architecture work, not evidence that the corresponding runtime behavior exists.

### Canonical Documentation Gaps

1. **README product contract:** state that each Apparat node can manage and advertise zero to many local inference service instances, including multiple instances of the same provider and multiple instances serving the same workload class. State that remote cluster members invoke them through the authenticated Apparat gateway, never by connecting directly to a provider's localhost endpoint.
2. **Architecture contract:** define the distinction among workload class, driver kind, concrete service instance, and discovered capability/model. Document the statically registered driver/factory boundary, the instance manager keyed by stable service ID, the shared-core ownership model, and the selected GUI/daemon process-ownership rule.
3. **Database contract:** add desired-versus-observed service state, stable service and capability identifiers, local-only endpoint and credential handling, schema-versioned provider configuration, advertisement revision/expiry state, and the migration away from flat capability records that cannot express service-instance identity.
4. **API contract and OpenAPI source:** define logical service/capability addressing, service inventory, job targeting and route explanations, authorization failures, asynchronous progress/results, bounded requests, cancellation, and safe remote health/error projections.
5. **Security contract:** define authorization and policy enforcement at the Apparat gateway, non-disclosure of localhost URLs and credentials, provider-compromise boundaries, service enablement/advertisement policy, and audit requirements.
6. **Transport and signed-envelope contracts:** define service/capability advertisement revisions, timestamps, expiration, stale behavior, payload compatibility, and how alternate transports carry or reject these records.
7. **Controller and GUI contract:** reconcile the documented shortcut and focus model with the five canonical tabs and current executable behavior. Distinguish core-owned service state from GUI-owned selection, filtering, expansion, layout, and transient form state.
8. **Platform contract:** document one-node process ownership, runtime locking or daemon-client operation, platform service supervision, provider endpoint assumptions, and platform-specific inference/runtime limitations.
9. **Roadmap migration contract:** expand typed compute work to cover discovery, verification, explicit enablement, arbitrary same-provider instances, health/inventory refresh, advertisement expiry, per-instance admission, and restart recovery. Preserve legacy phase IDs during migration so plan references remain traceable.
10. **Implemented-versus-planned labels:** clearly mark provider drivers, service persistence, discovery, routing, execution, and most backend behavior as planned. Do not let approved architecture text imply completed functionality.
11. **Directory-level documentation:** update the nearest code-directory `README.md` whenever an implementation creates or changes drivers, repositories, APIs, schedulers, platform services, tests, scripts, or build behavior. Update `scripts/README.md` and useful `--help` output for every changed script.
12. **User/contributor operations:** update the root README whenever normal users or contributors must configure, run, observe, troubleshoot, migrate, back up, restore, or secure new behavior.
13. **Project/Pipeline/Task contract:** define Projects as owner-local Git repositories projected into a cluster-wide catalog; define Pipelines as Projects with Apparat Task entrypoints; define optional trigger bindings separately from manually executable Tasks.
14. **Queue protocol contract:** define queue-owner REST submission and validation, worker-initiated claim/long-poll, owner-issued leases/fencing, worker result posting, and owner-authoritative completion. Prohibit direct remote database access and worker-authoritative queue state.

The concise inference rule to carry across those documents is: **providers are statically compiled drivers; concrete localhost endpoints are durable service instances; models are discovered capabilities; the shared core manages an arbitrary number of instances by stable service ID and exposes them through Apparat's authenticated gateway.**

### Ongoing Documentation Acceptance

- [ ] Documentation
  - [ ] Update README when product behavior changes.
  - [ ] Update API, security, database, transport, controller, and platform contracts with implementation changes.
  - [ ] Keep agent-operation instructions out of the human-facing README.
  - [ ] Keep implemented, validated, deferred, and planned status explicit.
  - [ ] Keep the legacy roadmap synchronized until the authority migration is complete.

### Documentation Truth Reconciliation

The existing documentation is strong as a product contract but sometimes mixes approved intent with implemented evidence:

- [`docs/controller-map.md`](./docs/controller-map.md) still contains stale six- or seven-tab language in places, while the current contract and code use five top-level tabs.
- Legacy Phase 1 and Phase 2 text contains older `Alt+1` through `Alt+7` claims.
- Legacy Phase 2 marks focus movement, activation, back, context menus, command palette, right-stick scrolling, and input equivalence complete, but the GUI loop implements only a subset.
- [`docs/architecture.md`](./docs/architecture.md) describes domain, platform, adapter, module, command, event, reducer, and store boundaries that are mostly aspirational.
- Legacy Phase 3 describes repository interfaces and a shared durable runtime more strongly than the current split-shell implementation supports.
- The inference contract does not yet explicitly require arbitrary same-provider and same-workload service instances, stable service-instance identity, gateway-only remote access, local endpoint non-disclosure, or desired-versus-observed service state.

Required actions:

1. Audit every completed roadmap claim against executable evidence before backend implementation.
2. Mark unverified behavior `[?]`, unimplemented behavior `[ ]`, and intentionally deferred behavior `[-]` instead of preserving optimistic completion states.
3. Label documents or sections as `contract`, `implemented`, `partially implemented`, or `planned` where ambiguity is likely.
4. Add consistency checks for canonical tab IDs/order, shortcuts, artifact paths, package boundaries, service cardinality, and platform-support claims.
5. Keep one canonical source for navigation and binding defaults; generate or test documentation examples from it where practical.
6. Keep this integrated backlog and the legacy roadmap synchronized until the final authority migration is approved.

### Open Decisions To Resolve And Record

Resolve each decision in the named canonical contract before its dependent implementation plan is approved:

- [ ] Core and process ownership
  - [ ] Confirm the shared headless-capable core embedded in both artifacts.
  - [ ] Choose exclusive single-node runtime locking versus a daemon-owned core with a client GUI when both artifacts run simultaneously.
  - [ ] Define explicit separate runtime roots, identities, and service ownership for intentionally independent nodes on one host.
- [ ] Identity, certificates, and authorization
  - [ ] Select the exact X.509 hierarchy.
  - [ ] Decide how TLS leaf keys relate to Apparat device identity keys.
  - [ ] Finalize authorization vocabulary.
- [ ] Network and protocol
  - [ ] Select canonical signed-envelope encoding.
  - [ ] Select endpoint discovery after temporary static configuration.
  - [ ] Define cluster-directory conflict resolution.
  - [ ] Define service-advertisement revision, expiration, and stale-removal semantics.
- [ ] Inference services
  - [ ] Approve the stable `ServiceID` and `CapabilityID` model and provider-driver registry boundary.
  - [ ] Decide discovery defaults and whether verified known services may be enabled or advertised automatically.
  - [ ] Define provider credential storage/reference behavior and local endpoint visibility.
  - [ ] Define per-service admission, routing preference, health, and failure-isolation policy.
- [ ] Data, artifacts, and recovery
  - [ ] Define artifact chunking, resumption, integrity, and retention.
  - [ ] Define optional database encryption and multi-device restore.
- [ ] Runtime and execution safety
  - [ ] Define process/service supervision.
  - [ ] Define safe tool execution and sandboxing.
  - [ ] Decide whether a headless TUI justifies termframe.
- [ ] Alternative transports
  - [ ] Choose Meshtastic source dependencies.
  - [ ] Choose a maintainable Signal gateway strategy.
- [ ] Comrades and shared compute
  - [ ] Define comrade message transport and end-to-end privacy expectations.
  - [ ] Define comrade prompt/result visibility and resource-owner observability.
  - [ ] Define comrade quota, preemption, abuse, moderation, and revocation defaults.
- [ ] Research and validation gameplay
  - [ ] Select the BOINC integration boundary and source dependencies.
  - [ ] Define BOINC workload isolation across Linux, Windows, macOS, and Android-capable devices.
  - [ ] Define research-project validation evidence and governance.
  - [ ] Define validation gameplay, reputation, anti-gaming, and moderation mechanics.

### Execution-Planning Admission Rules

1. Bind every implementation plan to one or more checklist items in the integrated program below and verify agreement with the README and relevant detailed contracts.
2. Keep plans narrow enough to produce one reviewable behavior, migration, protocol increment, platform proof, or structural seam with explicit rollback and verification.
3. Record schema compatibility, protocol compatibility, security assumptions, failure modes, observability, documentation updates, and target-platform evidence in the plan when applicable.
4. Do not approve inference implementation plans until the service-instance identity/cardinality and gateway rules are documented.
5. Do not approve network-facing implementation plans until dependent identity, authorization, envelope, limit, and audit decisions are resolved.
6. Do not approve platform-support claims from compilation alone; require the platform evidence specified by the platform contract and the integrated exit criteria.
7. Generalize commands, events, effects, repositories, modules, or plugin mechanisms only when implemented vertical slices demonstrate a common contract.

### Recommended Program Order

1. **Documentation and decisions:** close the gaps and record the decisions above.
2. **Truth and residual validation:** reconcile completed claims, finish the legacy Phase 3 shared-runtime proof, and finish or explicitly defer the remaining Android Phase 5 evidence.
3. **Structural admission gate:** establish core/GUI boundaries, package seams, persistence lifecycle, supervised startup/shutdown, headless tests, and security hardening from the P0 recommendations.
4. **Secure two-device proof:** enrollment, mutual TLS, signed envelopes, initial API, and durable mock queue across restart and disconnection.
5. **Project vertical slice:** safe filesystem/Git operations, chats, owner-authoritative transactions, drafts, and artifacts.
6. **Typed services and routing:** implement multi-instance local service discovery and persistence, provider drivers, advertisements, typed queues, pools, routing, and the first real text-generation path.
7. **Automation and voice:** add durable task execution, triggers, ASR, TTS, and privacy-preserving push-to-talk.
8. **Platform release proof:** harden Steam Deck/Linux GUI and headless packaging first, then independently validate Windows, macOS, and Android.
9. **Alternative transports and resilience:** add conformance-tested constrained transports, optional WireGuard management, ownership migration, replication, and routing optimization.
10. **Post-MVP social and research systems:** add Comrades sharing and Research/BOINC only after their threat models, resource policies, and isolation boundaries are approved.

Each checkpoint must remain independently reviewable and preserve a running mock HUD. Avoid a branch that replaces the whole architecture before any intermediate state works.

### Roadmap Authority Migration Exit Criteria

`RECOMMENDATIONS.md` can supersede the unfinished legacy roadmap only when:

- Every incomplete, validation-pending, deferred, and open-decision item in the legacy roadmap is mapped here or intentionally retired with a recorded reason.
- Goals, dependencies, detailed checklists, cross-cutting requirements, exit criteria, and MVP criteria have no unresolved semantic conflict between the two documents.
- The README names this document as the canonical implementation sequence and clearly describes the historical status of `ROADMAP.md`.
- Existing execution plans and documentation links that cite legacy phases have stable aliases or are migrated deliberately.
- The roadmap is archived or reduced through an explicit reviewed documentation change; it is not silently deleted.
- Automated documentation checks and plan-index checks pass after the migration.

## Executive Assessment

The codebase has a useful product contract, a working five-tab mock HUD, a shared startup path, basic local persistence primitives, and unusually good repository governance for this stage. The largest risk is not missing backend code. It is that several architectural boundaries are documented but not yet real:

- `internal/app.Runtime` owns one `hud.Shell`, while the GUI creates and mutates a second independent `hud.Shell`. The problem is not that the GUI fails to render a runtime-owned HUD; it is that HUD state does not belong in the headless-capable core at all.
- `internal/domain` and `internal/platform` are one-line placeholders, `internal/adapters/persistence` is a dependency sentinel, and SQL-backed product records live directly in `internal/cluster` and `internal/messaging`.
- The GUI adapter owns navigation, selection, collapsed/detail, voice-capture, gesture, update, and widget-tree state. Most of that is legitimate presentation state, but it is mixed with backend-facing actions and has no explicit read-model/command boundary to the core.
- The documented command/event/effect/reducer flow is not implemented. Most GUI input mutates `hud.Shell` or GUI-owned maps directly.
- Many documented controller, keyboard, accessibility, offline, failure, recovery, and configuration behaviors are absent even though some roadmap items describe them as complete.
- The default Go test path excludes the GUI-tagged tests. The GUI test binary compiles, but executing those tests requires a display, and several hidden GUI tests still select tab index `5` even though the valid five-tab indexes are `0` through `4`.

The best next move is a short stabilization phase, not a broad rewrite. Preserve the working UI and local primitives, but establish the two-artifact topology, separate core state from presentation state, add a minimal core-facing seam, and harden existing lifecycle/persistence behavior before adding HTTP, enrollment, queues, projects, or inference adapters.

## Intended Core And Artifact Model

Apparat should have one shared, headless-capable core implementation compiled into two separate artifacts:

```text
shared headless-capable core
├── apparatd = core + headless/service adapters
└── apparat  = core + GUI/presentation adapter

core read models and events -> HUD projection
GUI/API/scheduler actions   -> core commands
```

The word `core` refers to mode-neutral application and domain packages, not to the `apparatd` executable. The GUI should not have to spawn or connect to a second local process merely to use the core. A GUI device may own local durable state, work offline, and participate as a cluster node, while a remote `apparatd` exposes its own authoritative state through the future authenticated API.

If the product instead requires `apparat` to be a presentation-only client of a separately running `apparatd`, that must be approved as a distinct architecture decision. It would require local daemon discovery, IPC or HTTPS, lifecycle management, version negotiation, unavailable-daemon UX, and reconsideration of the current binary-specific runtime roots. These recommendations assume the shared-core model above.

State ownership must remain explicit:

- **Core-owned state:** identity, trust, enrollment, authorization, device directory, capabilities, projects and durable drafts, queues, jobs, attempts, results, artifacts, tasks, approvals, durable messages, cursors, retries, backend configuration, health, and cached remote synchronization state.
- **GUI-owned state:** active tab, focused control, selected selector item, back stack, open detail panel, scroll offsets, divider widths, hover/pressed state, modal state, gestures, animation, layout, and unsaved widget text.
- **Boundary workflows:** audio capture is GUI/platform state while a submitted transcription job is core state; an editor buffer is GUI state until explicitly saved as a durable draft; last-tab selection is a UI preference even if persisted; update installation is a platform concern rather than backend domain state.

## Fundamental Project, Pipeline, Task, And Queue Model

These ownership rules are product invariants, not implementation options:

| Concept | Authoritative owner | What other devices retain | Cross-device operation |
| --- | --- | --- | --- |
| Project | Device where the Git working tree lives and runs | Authorized signed/cached summary, freshness, owner, and local drafts | REST to the Project owner |
| Pipeline | Same as its Project; it is a Project with one or more Task entrypoints | Same Project summary plus advertised Task/trigger summary | REST to the Project owner |
| Task | Its Project owner | Authorized Task summary and run snapshot | Manual invocation or trigger request by REST to the Project owner |
| Trigger binding | Its Task/Project owner | Authorized schedule/webhook/event summary | Trigger delivery creates a run request at the owner |
| Queue and queued job | Queue-owning device | Requester's durable outbound record and authorized cached status/result | REST submission to the queue owner |
| Lease/attempt | Queue owner; temporarily assigned to one worker | Worker holds only the active bounded lease | Worker polls/long-polls the owner by REST |
| Result | Queue owner after validation | Worker holds a local execution outcome; authorized peers may cache the accepted result | Worker posts signed outcome to the owner by REST |

### Projects Are Owner-Local Git Repositories With A Cluster-Wide Catalog

A Project is an ordinary Git repository owned by the device where its working tree lives and runs. The repository, file paths, Git state, Project Task definitions, and authoritative run state stay on that owner. A cache, remote summary, or optional Git clone on another device does not silently transfer ownership.

Every Apparat device must nevertheless present one cluster-wide Projects list. It combines local Projects with every authorized Project summary advertised by other enrolled devices. When an owner is offline, cached metadata may keep its Project visible with explicit stale/unavailable state. Remote devices browse or operate the Project through the owner's authenticated REST API; they never read its filesystem or SQLite database directly.

Remote writes remain owner-authoritative idempotent transactions. Offline edits are local drafts or explicit Git commits until the owner accepts them. Authorization to discover a Project is distinct from permission to read files, inspect Git state, invoke Tasks, mutate the Project, read artifacts, or access secrets.

### Pipelines Are Projects With Task Entrypoints

A Pipeline is not a second repository, queue, workflow owner, or independently replicated entity. A Project is presented as a Pipeline when it defines at least one Apparat-executable entrypoint.

Each Project entrypoint is a Task. A Task belongs to exactly one Project during the MVP and is defined, validated, started, and recorded by the Project's owner. It may perform constrained owner-local Project operations and submit typed workload jobs to routing queues.

Trigger bindings are separate from Task identity:

- A Task with zero trigger bindings is valid and runs only when explicitly invoked in Apparat.
- Interval/cron schedules, authenticated webhooks, internal application events, and cluster device/service/queue events may each be bound to a Task.
- A trigger creates a Task run; it does not become the Task or move Project/Task authority to the triggering device.
- Task definition, trigger bindings, run/correlation IDs, current step, approvals, queued jobs, retries, results, and run history are durable owner state.

```text
Project (owner-local Git repository)
  -> Task entrypoint[]
       -> zero triggers: manual execution
       -> trigger[]: interval | webhook | app event | cluster event
       -> TaskRun
            -> constrained owner-local actions
            -> typed jobs submitted to queue owners
```

### Queue Owners Validate; Workers Pull And Return Results

Adding a job to a queue means sending an authenticated REST request to the device that owns that queue. The owner is the only authority for admission, ordering, leases, attempts, cancellation, completion, retention, and audit.

```text
requester
  -> REST submit + idempotency key
queue owner
  -> authenticate, authorize, validate, persist accept/reject
inference worker
  -> REST claim/long-poll with current capabilities
queue owner
  -> select compatible job and issue bounded lease + fencing token
inference worker
  -> execute, heartbeat if allowed, REST-post signed result/failure
queue owner
  -> validate worker + lease + result, persist authoritative completion
```

Workers initiate claim/long-poll requests; the owner does not push unleased tasks into worker memory or databases. A lease identifies the queue, job, attempt, worker, deadline, and fencing token. Expiry permits recovery/reassignment. Heartbeat and completion are accepted only for the active worker and fencing token, and completion is idempotent so stale, late, or duplicate outcomes cannot complete a logical job twice.

A direct queue targets one eligible inference device; that device still pulls an owner-issued lease through REST and may happen to be the owner. A pool queue is coordinated by the owner across eligible member devices, each of which pulls compatible leased work.

The worker's local execution result is not authoritative until the owner accepts it. Workers never pull work from requesters, read replicated queue rows as assignments, write the owner's SQLite database, or gain visibility into unrelated queue jobs, Projects, secrets, or provider credentials.

All of these cross-device requests use the versioned authenticated REST API. Signed envelopes, stable IDs, authorization, limits, deadlines, audit, and redaction apply within that transport.

## Fundamental Local Inference Service Model

The current documentation already anticipates typed workload classes, provider adapters, and devices advertising several independent capabilities. It does not yet make the required zero-to-many cardinality explicit. A flat capability list is not sufficient because one device may run two Ollama servers, several Automatic1111 instances, ComfyUI, llama.cpp, and other providers simultaneously, with overlapping models and workload types.

### Keep Four Concepts Separate

1. **Workload class** describes the operation, such as text generation, image generation, speech recognition, speech synthesis, embedding, or video generation.
2. **Driver kind** identifies the provider protocol or integration, such as `ollama`, `automatic1111`, `comfyui`, `openai_compatible`, or `llama_cpp`.
3. **Service instance** is one concrete configured local endpoint with a stable `ServiceID`. It has independent configuration, enablement, health, limits, lifecycle, and advertisement policy. The data model must permit arbitrary instances of every driver kind and workload class.
4. **Capability** describes what one service instance currently offers: a model, modality, format, feature set, schema version, or limit. For example, Flux is normally a model/capability available through a provider instance, not necessarily a provider driver of its own.

Use a hierarchical identity such as `device_id/service_id/capability_id/model_id` when routing needs model-level precision. Do not overload a workload class or provider name as instance identity.

### Use Static Provider Drivers, Not Go Dynamic Plugins

In this project, “plugin” should mean a statically compiled provider driver registered explicitly at a composition root. It should not mean loading Go `.so` files with `plugin.Open`. Go's official [`plugin` package documentation](https://pkg.go.dev/plugin) describes limited platform support, race-detector limitations, and strict toolchain/dependency compatibility requirements, and recommends static compilation or inter-process communication in many cases.

Static drivers fit the product because Ollama, Automatic1111, ComfyUI, llama.cpp, and similar runtimes already run as separate localhost services. Apparat's driver is a connector and policy-controlled gateway to that external process; it is not the model runtime itself. Explicit registration is portable, type-safe, easy to test, and produces the same known driver set in GUI and headless builds:

```go
registry.Register(ollama.NewFactory(httpClient))
registry.Register(automatic1111.NewFactory(httpClient))
registry.Register(comfyui.NewFactory(httpClient))
```

Avoid hidden package-global `init()` registration. Composition roots should make the included drivers and dependencies visible. If third-party out-of-process extensions become a real requirement later, define a versioned authenticated IPC protocol as a separate decision rather than weakening the core type boundary now.

### Use A Factory Per Driver And An Instance Per Endpoint

A narrow contract should validate configuration, open a concrete instance, inspect its inventory and health, and provide typed executors for the workload classes it supports:

```go
type Factory interface {
	Kind() DriverKind
	Validate(InstanceConfig) error
	Open(context.Context, InstanceConfig) (Instance, error)
}

type Instance interface {
	Inspect(context.Context) (Observation, error)
	Executor(WorkloadClass) (Executor, bool)
	Close() error
}

type Executor interface {
	Execute(context.Context, Request, ProgressSink) (Result, error)
}
```

`Request` and `Result` should be workload-specific typed structures or a deliberately versioned tagged union, not an unbounded `map[string]any`. The manager may normalize lifecycle, admission, progress, errors, and health, but it should not erase meaningful differences among text, image, video, and speech execution.

The manager's primary index must be the stable instance identity:

```go
type Manager struct {
	factories map[DriverKind]Factory
	instances map[ServiceID]*ManagedInstance
	byClass   map[WorkloadClass][]ServiceID
}
```

Secondary indexes may support routing by driver, class, model, device, health, or policy. Never make `DriverKind` or `WorkloadClass` the primary map key: either choice silently enforces one instance per provider or workload type.

### Treat Apparat As The Cluster Gateway

Remote peers should use this path:

```text
remote Apparat
  -> authenticated local Apparat API
  -> authorization, routing, durable queue, and policy
  -> local inference manager
  -> localhost inference provider
```

Provider endpoints and credentials remain local to the owning node. Cluster advertisements expose logical IDs and safe capability, health, availability, concurrency, and policy metadata—not `localhost` URLs, tokens, or provider-specific secrets. A useful advertisement shape is:

```text
DeviceAdvertisement
  ServiceAdvertisement[]
    ServiceID, DriverKind, DisplayName
    Health, Availability, Concurrency, Policy
    CapabilityAdvertisement[]
      CapabilityID, WorkloadClass, ModelID
      Modalities, Formats, Features, Limits, SchemaVersion
```

Advertisements need a revision, observation timestamp, expiry, and explicit stale behavior. Routing should reject expired or unauthorized inventory rather than assuming the last advertisement remains true indefinitely.

### Separate Desired, Observed, And Advertised State In SQLite

SQLite is the durable authority for local desired configuration and the last-known safe observation; the in-memory manager is derived runtime state. Query-critical identifiers and policy fields should be normalized, while provider-specific configuration and capability details may use schema-versioned JSON.

Suggested logical tables are:

- `inference_service_instances`: `service_id`, `local_device_id`, `driver_kind`, display name, local endpoint, enabled flag, advertise flag, provider configuration, credential reference, revision, and timestamps.
- `inference_service_observations`: `service_id`, lifecycle state, health state, last probe time, inventory hash, and redacted/safe failure fields.
- `inference_service_capabilities`: `service_id`, `capability_id`, workload class, schema version, model ID, capability data, and observation time.

Desired state includes configured endpoint, enablement, advertisement policy, admission policy, and concurrency. Observed state includes reachability, health, models, formats, provider limits, current availability, last probe, and safe failure information. Credentials must be stored as secret references and must not enter replicated configuration, advertisements, logs, or general JSON blobs.

At startup the core should load desired instances, resolve each registered factory, open enabled instances, probe them with bounded concurrency and deadlines, persist observations and capabilities, derive the signed advertisement, and refresh periodically. Inventory or policy changes advance a revision and invalidate the prior advertisement. Discovery should follow `discovered -> verified -> enabled -> advertised`; a random responding local port must not automatically become cluster-accessible without provider validation and policy approval. An optional policy may auto-promote verified known services.

### Make Every Instance Independently Supervised

Each managed service instance needs its own health schedule, admission semaphore, queue or concurrency limit, request deadline and cancellation, retry classification, circuit-breaker or unavailable state, progress normalization, inventory refresh, and orderly shutdown. One unhealthy Automatic1111 endpoint must not prevent a second Automatic1111 endpoint or an Ollama endpoint from operating.

Use Go goroutines and channels for bounded supervision, wake-ups, and transient progress delivery. Do not use channels as durable job state. SQLite remains authoritative for jobs, attempts, cancellation intent, leases, results, and restart recovery. The scheduler routes through logical service and capability IDs, then resolves the current local instance and typed executor.

### Use One Manager In GUI And Headless Builds

Both artifacts construct the same core inference manager and SQLite repositories. The GUI must not maintain a second discovery registry or infer service truth from widgets; it queries core read models such as `ListLocalServices`, `GetService`, and `ListServiceCapabilities`. GUI filtering, selected service, expanded panels, scroll position, and unsaved form input remain GUI-owned presentation state.

By default, `apparat` and `apparatd` should be alternative process forms of one Apparat node, using one node identity and runtime database with an exclusive process lock. Otherwise two processes on one host could discover the same Ollama endpoint and advertise it twice under competing state. If simultaneous daemon and GUI operation is required, the GUI should become a client of the daemon-owned core. Running multiple independent Apparat nodes on one host should require explicit separate runtime roots, identities, and service ownership.

## Priority Definitions

- **P0 — backend admission gate:** complete before Phase 6 or any other substantial backend implementation.
- **P1 — first vertical-slice hardening:** complete while building the initial two-device slice, before multiplying adapters and durable entities.
- **P2 — scale and release hygiene:** complete as the project gains more platforms, contributors, and release history.

## P0 Recommendations: Establish The Correct Seams

### P0.1 Separate Core State From Presentation State

Current evidence:

- `internal/app.Runtime` constructs `hud.NewShell()` and exposes `Runtime.Snapshot()`.
- `internal/adapters/gui.NewGameWithRuntimeInfo` constructs another `hud.NewShell()`.
- `Runtime.Start` passes only path strings to the GUI, not the runtime's state, command dispatcher, repositories, or event stream.
- Android initializes a runtime and then creates a separate GUI game; the runtime is retained indirectly rather than owned through an explicit mobile lifecycle.

Recommendation:

1. Remove `hud.Shell` and `Runtime.Snapshot() hud.Snapshot` from `internal/app.Runtime`. A headless runtime should not know which tab is active or which control is focused.
2. Make the shared core authoritative only for backend/domain state and durable command execution.
3. Let the GUI own a presentation controller for tabs, focus, selection, back navigation, scrolling, layout, gestures, and transient editor state.
4. Give GUI, headless, API, scheduler, and transport adapters the same small core-facing interface:
   - Lifecycle and health.
   - Typed commands for mutations.
   - Queries/read models for current core state.
   - Subscriptions or change notifications when live state exists.
5. Make `cmd/apparat`, `cmd/apparatd`, and `cmd/apparatmobile` composition roots that construct the same core and inject the appropriate adapters.
6. Let the HUD projection combine core read models with GUI-owned presentation state. The core should expose stable entity IDs but should not store the GUI's current selection.
7. Make core read models copy-safe or immutable at adapter boundaries. GUI snapshots may separately contain GUI-owned slices and maps.

Target flow:

```text
GUI/API/scheduler -> core command -> use case/persistence/effect
core query/change -> read model   -> HUD projection + GUI state -> widgets
```

This is the most important prerequisite for backend work. It prevents backend truth from leaking into widgets without forcing headless processes to carry tabs, focus, scrolling, or other presentation concepts.

### P0.2 Make The Documented Package Boundaries Real

Current evidence:

- `internal/domain` and `internal/platform` contain only package declarations.
- `internal/adapters/persistence` contains no repository implementation.
- `internal/database`, `internal/cluster`, and `internal/messaging` mix domain-shaped records, time, SQL schema creation, queries, and persistence behavior.
- Empty reserved directories such as `internal/input` and `internal/runtime` add conceptual surface without behavior.
- `internal/adapters/gui/shell.go` and dependency-sentinel files appear to be scaffolding rather than active boundaries.
- Runtime and HUD configuration define overlapping mode and state concepts, while many declared HUD configuration values are never consumed by the renderer or input loop.

Recommendation:

1. Treat `internal/app` plus `internal/domain` as the shared headless-capable core compiled into both artifacts.
2. Put durable product vocabulary in `internal/domain`: stable IDs, device/capability/job/message types, validation, state transitions, and typed errors. Add each concept only with the feature that needs it, and keep it free of SQL, files, clocks, GUI libraries, and network packages.
3. Put use cases and port interfaces in `internal/app`: lifecycle, feature commands/queries, transactions, and only the repository, clock, ID, transport, or workload ports required by implemented features.
4. Put SQLite implementations under a real persistence adapter such as `internal/adapters/sqlite` or `internal/adapters/persistence`. Keep migration ownership there and return domain/application types rather than SQL rows.
5. Keep `internal/hud` as a GUI-side presentation model and projection layer. It may consume core read models and own navigation/presentation state, but `internal/app`, `internal/domain`, and `cmd/apparatd` must not import it.
6. Keep Ebitengine/EbitenUI input and rendering under the GUI adapter. `apparatd` must compile and run without GUI or HUD packages.
7. Use `internal/platform` only when the first real platform abstraction is introduced. Remove empty placeholder directories and dead sentinel files until they have a concrete contract.
8. Collapse duplicate types such as `app.Mode` and `config.Mode` unless their distinction becomes meaningful.
9. Define the focused, statically registered inference driver/factory contract because arbitrary provider-instance multiplicity is a fundamental product invariant. Do not turn it into a generic dynamic plugin or feature-module framework; generalize unrelated registration and lifecycle patterns only after multiple real features demonstrate a common need.

Prefer a few honest packages with narrow interfaces over a large aspirational package tree.

### P0.3 Establish A Minimal Core-To-HUD Read-Model Boundary

Current evidence:

- Tab IDs are stable, but sections and rows have no IDs.
- GUI selection uses section indexes even though selector headings and live data can change ordering.
- Settings behavior is attached by comparing lowercased section titles to `"updates"` and `"diagnostics"`.
- Row data is mostly `Label`, `Detail`, `Disabled`, and `Future`, which cannot express actions, loading, error, progress, permissions, timestamps, badges, or accessibility state cleanly.
- Chat, Git, Chat, Run, and Send controls are enabled no-ops.

Recommendation:

1. Add stable IDs to backend entities as their first real features are implemented. Project those IDs into GUI selector items so presentation selection survives reordering and refresh.
2. Add stable IDs and typed action IDs to current HUD sections that own behavior. Stop attaching behavior to display strings such as `"updates"` and `"diagnostics"`.
3. Bind every enabled control to either a GUI-local action or a typed core command. Otherwise mark it disabled/unavailable with a reason.
4. Separate mock fixtures from the production read-model and HUD schemas so live core data can replace fixtures without changing widget contracts.
5. Keep the P0 view schema minimal. Add richer element kinds and reusable loading/error presentation patterns in P1 when the first live feature supplies real requirements.

This boundary lets the core remain presentation-free while allowing live repositories to replace fixtures without coupling widgets to SQL or backend service types.

### P0.4 Grow Core Commands And Events From The First Real Feature

The architecture documents describe commands, events, effects, reducers, stores, and feature modules, but nearly all real core features are still unimplemented. Do not build that framework speculatively.

1. Define a minimal core interface now: lifecycle, health, feature-specific commands, feature-specific queries/read models, and change notification.
2. Implement one real local feature through the entire seam. A strong first inference-facing slice is a durable service-instance inventory with a mock driver, two configured instances of the same driver, independent observations, and one persisted mock job routed to a selected instance.
3. Add command IDs, correlation IDs, actor/target, deadline, and idempotency keys only where that feature's durability or cross-device behavior requires them.
4. Keep filesystem, SQL, network, updater, audio, and workload operations outside the render loop, but introduce a generic effect abstraction only after multiple effects demonstrate a shared contract.
5. Use deterministic clocks and ID sources where the implemented feature needs retry, ordering, or recovery tests.
6. Generalize event, reducer, store, and module patterns only after a second feature confirms what is actually common.

The goal is a proven boundary, not a framework capable of every future phase.

### P0.5 Centralize And Harden SQLite Lifecycle

Fix these reliability problems before adding more tables:

- `PRAGMA foreign_keys = ON` is connection-local, while the current `sql.DB` may open multiple connections. Configure every connection through the driver/DSN or deliberately constrain and document the connection pool.
- Cluster table initialization and messaging table initialization are scattered `Init` calls rather than one ordered migration set.
- `cluster.Directory.PutDevice` updates the profile and appends its change-feed record in separate non-transactional statements.
- `messaging.Store.Seen` treats every insert error as a duplicate and returns `true, nil`, which can hide disk, schema, cancellation, and connection failures.
- Current message direction and payload are untyped strings, and `messaging.Seen` cannot distinguish duplicate delivery from operational failure.
- Migration code does not yet establish schema compatibility tests across real historical database fixtures.

Recommendation:

1. Create one ordered migration registry owned by the persistence adapter. No repository should create production tables opportunistically at runtime.
2. Make every state change plus outbox/event/change-feed append one transaction.
3. Distinguish constraint violations from operational errors and preserve `context.Canceled`/deadline errors.
4. Add repository interfaces around the first implemented application operations, not speculative generic CRUD.
5. Add constraints, statuses, optimistic version checks, indexes, and query-plan tests only when a real feature defines the invariant or hot query.
6. Test the current primitives for restart, partial failure, duplicate delivery, migration upgrade, corruption, backup, and restore in proportion to the behavior they already claim.
7. Inject a clock and ID generator when retry and ordering behavior becomes part of the implemented core feature.

Harden the database lifecycle and existing primitives now; design final job, queue, task, and project schemas with their first real use cases.

For inference state specifically, preserve zero-to-many service cardinality from the first migration: key records by `ServiceID`, keep desired configuration separate from observations and capabilities, normalize query-critical fields, version provider-specific JSON, and store only secret references. Do not encode one provider or one workload class per device in table keys or uniqueness constraints.

### P0.6 Make Runtime Startup And Shutdown Transactional

Current initialization assigns resources incrementally. A later failure can leave a database or logger open, and a retry can initialize another copy. The GUI and Android paths also lack a clear ownership contract for all long-lived resources.

Recommendation:

1. Model runtime components with explicit `Start`, `Ready`, and `Close` behavior.
2. On startup failure, close already-started components in reverse order and return an error that identifies the failed component.
3. Make shutdown idempotent and bounded by context/deadline.
4. Use one root context and a supervised goroutine group for API servers, schedulers, queue workers, subscriptions, and background checks.
5. Do not launch unmanaged goroutines from UI or platform callbacks.
6. Expose readiness separately from liveness. A process can be alive while identity, migrations, or required services are not ready.
7. Keep the shared core and `apparatd` independent of GUI/HUD imports. The GUI artifact should add its runner at the composition root rather than being called from `internal/app`.

### P0.7 Keep Core Tests Headless And Separate GUI Test Infrastructure

Current evidence:

- `go test ./...` reports no tests for the GUI adapter because the meaningful files use the `gui` tag.
- `make verify` does not invoke `test-gui-deps`.
- Running `go test -tags gui ./internal/adapters/gui` in the current environment fails during Ebitengine/GLFW initialization because no display is available.
- A compile-only GUI test binary succeeds.
- Several GUI tests still call `SelectTab(5)` after the application moved to five tabs.

Recommendation:

1. Ensure all core/domain tests run without build tags, a window, Ebitengine, or display libraries.
2. Add an architecture check over `cmd/apparatd` dependencies that rejects HUD, GUI, Ebitengine, and EbitenUI packages.
3. Extract GUI input mapping, focus navigation, selection, gesture arbitration, responsive layout decisions, and view reconciliation into display-free presentation controllers with ordinary Go tests.
4. Keep only thin Ebitengine/EbitenUI integration tests behind the `gui` tag.
5. Make `make verify` run core tests, architecture checks, display-free GUI behavior tests, and a GUI-tagged compile gate on supported build hosts.
6. Provide an explicit optional rendered-test target for environments with Xvfb or a real display.
7. Fix stale tab-index tests and derive cases from canonical tab descriptors instead of hard-coded counts.
8. Add architecture tests that reject domain imports of adapters and reject GUI imports of SQL/network packages.
9. Add fuzz tests for encrypted identity-file parsing, signed-envelope parsing, migration metadata, and other untrusted serialized input as those inputs become implemented.
10. Test inference multiplicity explicitly: two instances of the same driver and workload class must retain distinct identity, configuration, health, capabilities, admission limits, routing, and shutdown behavior.

Coverage percentages are secondary. The core gate is that backend state transitions and failure paths run headlessly; the GUI gate is that presentation behavior runs without requiring a display wherever practical.

### P0.8 Harden Identity, Logging, And Update Surfaces

Identity:

- Validate base64 errors, nonce/salt/ciphertext sizes, private-key length, manifest version, key kind, file permissions, and manifest/key consistency.
- Store explicit file-format and KDF versions and parameters so Argon2 settings can evolve safely.
- Write key and manifest updates atomically with temporary files, sync where needed, and recover predictably from partial rotation.
- Make `Classify` verify content and correspondence rather than file existence alone.
- Separate user identity, device identity, device authorization, and TLS identity in types and storage.

Logging:

- Make redaction recursive and safe for nested maps, slices, structs, error chains, URLs, and headers. Current key-substring filtering only handles the top-level field map.
- Prefer an allowlist of safe structured fields for security-sensitive events.
- Serialize concurrent writes and rotation. Report rotation failures instead of silently ignoring them.
- Define retention and disk-full behavior. Logging failure must not silently corrupt application state or cause unbounded disk growth.
- Keep user-facing errors separate from detailed internal errors so secrets do not escape through `err.Error()`.

Android updater:

- Disable automatic startup downloads and production-facing self-update until a stable signing identity, monotonic version, authenticated update manifest, artifact length/hash, atomic download, and exact installer result path exist.
- Restrict the current raw-branch/debug-signed updater to clearly labeled development builds if it must remain temporarily.
- Verify expected package name, signer, version progression, content length, and hash before invoking installation.
- Download to a temporary file with size limits and atomic promotion; never treat a partial cache file as ready.
- Report pending user action, success, cancellation, and Android's actual failure result rather than equating “installer opened” with success.
- Raise the Android target API deliberately and validate the resulting permission, storage, background, and installer behavior before release packaging.

These are local attack and data-loss surfaces, so they should be hardened before exposing an authenticated network API.

## P1 Recommendations: Complete The User Interaction Contract

These are GUI/presentation recommendations, not requirements for the headless core to understand or store UI state. They may proceed alongside the first backend slice once the core-to-HUD boundary is stable.

### P1.1 Route Every Input Through One Action Mapper

The current GUI polls individual keys and controller buttons directly while `HUDConfig.Bindings` is mostly descriptive. Replace scattered polling with an input mapper that produces the same action types for controller, keyboard, pointer, touch, and accessibility surfaces.

The mapper should handle:

- Press, release, hold, repeat, axis threshold, and gesture ownership.
- Text-input focus so global shortcuts do not steal ordinary editing keys.
- Configurable bindings and conflict detection.
- Device connection/disconnection and controller identity.
- Deterministic repeat timing based on elapsed time, not frame count.
- A single action trace suitable for tests and diagnostics.

### P1.2 Implement Focus, Activation, Back, And Context Semantics

Before adding real forms or backend controls, implement and test:

- Visible focus for tabs, selector items, controls, text fields, dialogs, and overlays.
- `Tab`/`Shift+Tab`, D-pad/stick, arrows, `A`, `Enter`, and `Space` activation.
- `B`, `Escape`, and mouse-back behavior with a predictable focus/back stack.
- `X`, right-click, long-press, Menu, and `Shift+F10` context actions.
- Focus trapping and restoration for modals.
- Focus preservation by stable item ID across data refreshes and responsive layout changes.
- Disabled-control explanations and discoverable keyboard/controller alternatives for drag operations.

### P1.3 Replace Full Widget-Tree Rebuilds With Reconciliation

`rebuildUI` currently reconstructs the full EbitenUI tree for many selections, resizes, toggles, and state changes. That is acceptable for a mockup but will cause allocation churn, lost widget state, focus resets, text-input loss, and gesture races when live updates arrive.

Recommendation:

1. Keep the root shell and stable widgets alive.
2. Rebuild only when the structural view type changes; update text, enabled state, selection, progress, and status in place.
3. Key reusable widgets by stable view ID.
4. Coalesce core change notifications so a burst of backend updates causes at most one HUD projection/reconciliation per frame.
5. Cache wrapped text/layout measurements by content, width, scale, and font.
6. Preserve text drafts, scroll offsets, focus, and selection independently from backend refreshes.
7. Replace fixed frame-count gesture suppression with gesture IDs or explicit deferred-event acknowledgement.

Measure allocations and update time before and after this work rather than assuming widget count alone is the bottleneck.

### P1.4 Define Standard Live-Data UX States

Every backend-backed surface should have consistent behavior for:

- Initial loading and refresh.
- Empty data with a useful next action.
- Offline cached data with last-updated time.
- Permission denied and authentication expired.
- Partial failure where some panels remain usable.
- Retry with visible backoff and cancellation.
- Stale or conflicted data.
- Long-running progress and restart recovery.
- Optimistic actions that are pending owner acceptance.
- Destructive actions requiring confirmation and explaining consequences.

Build these states as reusable view-model patterns before each feature invents its own spinner, error string, and retry button.

### P1.5 Make Configuration Honest And Durable

The HUD configuration model declares many settings the GUI does not honor. Either implement a setting end to end or remove it from the active configuration surface until it is real.

Separate GUI preferences from backend policy, then use layered configuration within each boundary:

1. Compiled safe defaults.
2. Platform/runtime defaults.
3. Durable user preferences.
4. Session-only overrides.

Validate values centrally, version durable settings, preserve unknown future fields safely, and expose reset/export/import behavior. Active tab, layout, scale, and theme remain GUI preferences; authorization, queue, retention, and service policy remain core state. Security policy must not share the same casual preference path as theme and spacing.

### P1.6 Add Performance And Resource Budgets

Record measurable budgets before backend load obscures UI costs:

- Startup-to-ready time for GUI and headless modes.
- Steady-state and peak memory on Steam Deck and Android.
- GUI update/draw time and allocations per frame.
- Core read-model size plus HUD projection and reconciliation time.
- SQLite transaction latency and critical query plans.
- Background goroutine count and shutdown time.
- Log volume and retention size.
- Network request size, concurrency, deadline, and retry limits.

Add small benchmarks for core read-model production, HUD projection, large selector lists, wrapped text, SQLite outbox transactions, and log redaction. Treat budgets as regression signals, not universal micro-optimization targets.

## P2 Recommendations: Simplify Builds And Releases

### P2.1 Separate Fast Developer Builds From Release Orchestration

The current no-flag build path detects and builds every possible target, including Android when its large toolchain is present. Keep a one-command release build, but also provide fast explicit targets such as:

- Host GUI only.
- Host headless only.
- Host smoke artifacts.
- Android only.
- Full release matrix.

Cache Android binding/tool outputs by Go version, dependency lock, source hash, SDK/NDK versions, and patch version. A normal backend edit should not rebuild an unchanged mobile toolchain.

### P2.2 Stop Mutating Source Checkouts During Builds

The Android pipeline temporarily patches `third_party/game/ebiten` and restores it afterward. An interruption can leave the submodule dirty, and exact-text patching is fragile across upstream changes.

Apply required patches to a disposable copied worktree, maintain a clearly pinned patch file/fork, or upstream the change. Build inputs should be immutable; generated work should remain under a disposable build directory.

### P2.3 Reconsider Tracking Large Binary Artifacts In Git

Tracked release artifacts currently account for roughly 80 MB in the working tree before history growth. Repeated binary replacement increases clone size, repository transfer time, and review noise.

Prefer immutable CI/release artifacts with version, commit, checksum, signer, and provenance metadata. If a pull-to-device “latest” workflow remains essential, track a small signed manifest or pointer instead of each binary. This recommendation changes an explicit current product/build decision and therefore requires deliberate approval and migration planning.

### P2.4 Add Reproducibility And Provenance

For release artifacts, record:

- Source commit and dirty-state refusal.
- Go, SDK, NDK, JDK, and dependency versions.
- Build tags and target tuple.
- Artifact hash, size, package/version metadata, and signer fingerprint.
- Reproducibility status or known nondeterministic inputs.
- SBOM and vulnerability-scan result when the release process matures.

## Expected GUI User Behaviors Not Yet Reliably Supported

These behaviors matter to the GUI artifact and product experience, but they are not headless-core readiness requirements. The headless core supplies stable read models, commands, progress, errors, and recovery state; the GUI owns how users navigate and interact with that information.

| Area | Expected behavior | Current gap | Recommended acceptance evidence |
| --- | --- | --- | --- |
| Controller navigation | D-pad/stick moves visible focus; `A` activates; `B` goes back; `X` opens context; Menu opens palette | GUI loop handles L1/R1/R2 but not the full accepted action map | Display-free action tests plus Steam Deck capture |
| Keyboard navigation | Tab traversal, arrows, Enter/Space, Escape, context keys, palette, collection navigation | Only direct tabs, tab cycling, PTT, and partial Escape behavior are wired explicitly | Keyboard matrix tests with text-input precedence |
| Mouse/touch equivalence | Right-click/long-press context, mouse-back, accessible alternatives to drag | Scrolling, tab selection, divider, overlay, and PTT exist; context/back equivalence does not | Pointer/touch action conformance tests |
| Focus | Every focusable element has visible focus that survives refresh/layout changes | HUD tracks a mock focus index; GUI focus is not the canonical application focus model | Stable-ID focus tests across refresh and collapse |
| Scrolling | Wheel, touch, right stick, PageUp/Down, Home/End scroll the intended viewport | Custom wheel/drag scrolling exists; controller and full keyboard scrolling are incomplete | Nested viewport and boundary tests on desktop/mobile |
| Selector navigation | Selection remains on the same logical item when items reorder or headings change | Selection is stored by section index | Reorder/insert/remove projection tests using item IDs |
| Chat composer | Send submits or clearly explains why unavailable; drafts survive redraw | Send is an enabled no-op and full rebuilds can lose widget state | Draft persistence and disabled/unavailable tests |
| Project controls | Git, Chat, and Run lead to real routes or are visibly unavailable | Buttons are enabled no-ops | Every enabled button dispatches a tested command |
| Loading and errors | Users see loading, offline, stale, denied, failed, retry, and partial-success states | Mock text mostly shows static future labels | View-state matrix and failure-injection tests |
| Settings | Visible values are applied, validated, persisted, and resettable | Most settings are hard-coded descriptors not consumed by GUI behavior | Restart persistence and invalid-value tests |
| Voice/PTT | Capture, privacy indication, cancellation, queueing, transcription, failure, and completion are real | Only an in-memory state transition exists; no audio/effect path | Fake-audio state-machine tests before real device validation |
| Identity | First run guides create/import; status verifies key/manifest consistency; repair is safe | Runtime classifies file presence; no integrated user flow | Corrupt/partial/rotation/recovery fixtures and UI states |
| Enrollment | Users can review fingerprint, role, permissions, expiry, confirmation, and revocation | Not implemented | Two-party invite and expiry/replay tests |
| Android update | Version/signature are checked and final installer outcome is reported | Raw APK hash comparison and “Installer opened” status only | Older compatible build upgraded in place with result capture |
| Accessibility | Scale, contrast, focus, readable labels, safe areas, and non-pointer operation are verified | Some fields are modeled; implementation and Android evidence are partial | Automated layout bounds plus device/controller review |
| Destructive actions | Scope and consequence are shown; confirmation and recovery are explicit | Product contract requires this, but no shared interaction exists | Command policy and confirmation-dialog tests |
| Restart recovery | Pending commands, drafts, jobs, and progress resume without duplication | Local primitives exist, but UI and runtime state are not connected | Kill/restart integration tests with stable correlation IDs |
| Local inference services | Users can inspect every discovered/configured local service instance, distinguish duplicates by stable name/ID, see verification/health/capabilities, enable or disable advertisement, and understand failures | The UI has mock capability concepts but no core-backed multi-instance inventory or lifecycle | Two same-provider instances plus another provider projected from one core manager, with independent health and policy changes |

## Recommended Simplifications And Removals

Perform these as part of accepted focused plans, not as an unreviewed cleanup sweep:

1. Remove empty `internal/input` and `internal/runtime` directories until concrete packages need them.
2. Remove unused adapter/sentinel scaffolding once dependency and GUI compile checks have real owners.
3. Replace duplicate mode types with one canonical runtime mode.
4. Split `Game` into a GUI-owned presentation controller plus focused renderer, input, gesture, overlay, and reconciliation helpers; do not move these concerns into the headless core or turn every helper into a public package.
5. Remove unused configuration fields or mark them explicitly planned until a consumer and acceptance test exist.
6. Move mock scenario content to fixtures/build-time mock projections so production view types remain small and backend-ready.
7. Remove hard-coded display-string behavior and hard-coded tab counts.
8. Remove enabled no-op controls.
9. Remove direct time/random dependencies from deterministic domain/application logic through injected ports.
10. Remove build-time mutation of tracked or submodule source trees.

## Shared-Core Readiness Exit Criteria

Apparat is in a strong position to begin backend features when:

- `apparatd` and `apparat` compile the same mode-neutral core implementation, while each artifact adds only its required adapters.
- The shared core, `internal/domain`, `internal/app`, and `cmd/apparatd` do not import HUD, Ebitengine, EbitenUI, or GUI packages.
- The GUI obtains backend state through core queries/read models and change notifications, then combines it with GUI-owned presentation state.
- Active tab, focus, selection, scroll, panel layout, gestures, and transient widget text remain outside the core.
- Domain and application tests run without GUI, SQLite, network, real time, or platform dependencies.
- Implemented external effects are behind explicit ports with typed outcomes and bounded context; speculative adapters and abstractions are not required.
- SQLite state changes and their outbox/event records are atomic, and operational errors cannot masquerade as duplicates.
- Runtime startup failure cleans up partial resources and shutdown is supervised and idempotent.
- Identity files, logs, and updater inputs are treated as security-sensitive untrusted data.
- Documentation distinguishes contract, implemented behavior, partial validation, and future intent.
- The canonical documentation explicitly defines arbitrary local inference service multiplicity, stable service and capability identity, Apparat's authenticated gateway role, local-only provider endpoints/credentials, and desired-versus-observed state before implementation begins.
- A service manager can keep two instances of the same driver and workload class distinct through configuration, probing, capabilities, admission, routing, failure, and shutdown.
- Inference service persistence is keyed by stable `ServiceID`; desired configuration, observed health/inventory, discovered capabilities, and advertisement revision/expiry are not collapsed into one mutable record.
- `apparat` and `apparatd` use the same inference manager and repositories. Process ownership prevents two artifacts from unintentionally advertising the same local provider as competing instances.
- Remote peers address logical service/capability IDs through an authenticated Apparat API; advertisements and safe errors never disclose localhost provider URLs or credentials.
- A local durable mock job travels through the shared core's command, persistence, query/change, routing, and recovery seams and can be rendered by the GUI without putting GUI state into the core.

## GUI-Readiness Exit Criteria

The GUI artifact is ready to consume growing backend state when:

- Default verification exercises display-free GUI behavior and compiles the native GUI adapter.
- Every enabled control dispatches a GUI-local action or real core command; unavailable functionality is disabled with a reason.
- Stable core entity IDs and GUI item IDs preserve selection, focus, drafts, and scroll state across read-model refreshes.
- Loading, empty, offline, denied, failure, retry, cancellation, and restart-recovery states have shared presentation patterns.
- GUI navigation and presentation state remain usable when the core is loading, partially available, offline from remote devices, or returning errors.
- Backend updates can reconcile into the HUD without full widget-tree rebuilds or loss of transient GUI state.
- Multiple local inference service instances—including two instances of the same provider—appear as distinct stable items with independent health, capability, enablement, and advertisement state supplied by the core.

Meeting these separate core and GUI criteria will reduce the cost of every later phase: HTTPS enrollment, project operations, queue routing, automation, speech, Comrades, and Research can reuse the same tested backend seams without making the headless artifact carry presentation concerns.

## Integrated Remaining Implementation Program

This program incorporates all unfinished and future implementation work from the legacy roadmap. Legacy phase numbers are retained as migration aliases, but the documentation, decision, and structural gates above take precedence over phase numbering. Checklist completion requires implementation and the stated evidence; architecture text alone does not complete an item.

### Stage 0: Structural Admission Gate

**Goal:** Make the documented shared-core, GUI, persistence, security, lifecycle, and testing seams real before networked backend growth.

**Dependencies:** Documentation truth reconciliation and resolution of the dependent open decisions above.

- [ ] Complete the applicable P0 recommendations in this document through focused approved plans.
- [ ] Prove the shared core can run without HUD, GUI, Ebitengine, EbitenUI, display, or GUI-tag dependencies.
- [ ] Prove the GUI consumes core read models and commands while retaining presentation-only state locally.
- [ ] Centralize and harden SQLite migrations, transactions, connection configuration, error classification, backup, and restore behavior.
  - [-] Keep WAL opt-in until platform validation is complete.
- [ ] Supervise startup, background work, readiness, shutdown, and partial-start cleanup.
- [ ] Harden identity parsing and atomic writes, recursive log redaction, and updater trust boundaries.
- [ ] Add headless architecture gates, display-free presentation tests, and GUI compilation/rendered-test targets.
- [ ] Demonstrate the shared-core and GUI-readiness exit criteria above.

**Exit criteria**

- The core and GUI readiness criteria in this document pass.
- No unresolved documentation contradiction can change the identity, cardinality, authority, persistence, or security model of the next implementation stage.
- The existing mock HUD remains runnable while core state is reached through the new boundary.

### Legacy Phase 3 Carryover: Shared Runtime Validation

**Goal:** Close the remaining validation gap in the durable local runtime shared by GUI and headless devices.

**Dependencies:** Stage 0 core/process-ownership decision.

- [?] Split GUI and headless runtime adapters.
  - [?] Make the default GUI binary enter a real Ebitengine run loop instead of exiting after runtime initialization.
    - The Ebitengine loop exists behind the `gui` build tag and the release pipeline builds the `apparat` artifact with that tag.
    - Native desktop-library and display-server validation remains target-specific evidence.
  - [ ] Replace the current binary-specific runtime-root assumption with the approved single-node lock or daemon-client ownership model where both artifacts could target the same node.

**Exit criteria**

- GUI and headless artifacts use the same durable core without sharing GUI state.
- The headless path initializes no GUI dependency.
- Native GUI startup is validated on each claimed target rather than inferred from tagged compilation.
- Running both artifacts cannot accidentally create competing identities or duplicate advertisements for the same local inference services.

### Legacy Phase 5 Carryover: Android GUI Validation And Hardening

**Goal:** Finish the outstanding Android GUI evidence without claiming an Android headless target.

**Dependencies:** Existing Phase 5 Android build pipeline and the Stage 0 core/GUI boundary.

- [ ] Prove the Android build works after temporarily moving or hiding `third_party/salvagecore`; the implemented script has no reference to it, but this destructive local proof still needs an explicit checkpoint if desired.
- [ ] Define release signing, store packaging, and automated version generation in a later release-hardening phase.
- [ ] Adapt runtime paths for Android GUI.
  - [ ] Verify the actual Android app-scoped runtime root on device/emulator.
  - [ ] Verify `last_run.log` is recreated on every Android GUI launch.
  - [ ] Verify SQLite, logs, identity, cache, artifacts, backups, and recovery directories are Android-safe.
- [ ] Validate the Android GUI smoke path.
  - [ ] Verify keyboard/controller navigation where the device supports it.
- [ ] Implement Android GUI parity.
  - [ ] Keep `third_party/salvagecore` as reference-only and prove the final Android GUI path does not depend on it.
  - [ ] Add Android safe-area/status-bar/navigation-bar layout handling so the HUD is readable on phone screens.
  - [ ] Add Android scale/density handling so tab buttons and body text remain usable on Pixel-class devices.
  - [ ] Validate additional phone, tablet, portrait, landscape, keyboard, controller, process-liveness, and `last_run.log` behavior on real Android devices.
- [ ] Add an optional integration test target that installs and launches the APK when Android tools and a device/emulator are available.

**Exit criteria**

- [ ] Android safe-area, density, and touch handling make the HUD readable and usable on a Pixel-class device.
- Android runtime paths and lifecycle are evidence-backed on device/emulator.
- Android GUI behavior consumes the same shared core contracts without introducing a second service registry or backend-state model.
- Release signing and distribution remain explicitly deferred to the platform release stage until implemented.

### Stage 1 / Legacy Phase 6: Secure Two-Device HTTPS/WireGuard Vertical Slice

**Goal:** Complete the MVP proof between a Steam Deck and one headless worker.

**Dependencies:** Documentation and Stage 0 gates; completed legacy Phases 1–5 foundations.

- [ ] Add external-network configuration.
  - [ ] Detect expected WireGuard interfaces where possible.
  - [ ] Support explicit peer endpoint configuration.
  - [ ] Support trusted-LAN endpoints through the same HTTPS API.
  - [ ] Make discovery advisory rather than authoritative.
- [ ] Add enrollment.
  - [ ] Generate a short-lived QR/invite.
  - [ ] Display and verify the cluster fingerprint.
  - [ ] Exchange device profile and certificate request.
  - [ ] Bind device identity, TLS certificate, WireGuard key, roles, permissions, and capabilities.
  - [ ] Replicate the signed peer record.
  - [ ] Expire or revoke the enrollment token.
- [ ] Add mutual TLS.
  - [ ] Require authenticated client and server devices.
  - [ ] Validate certificate binding and authorization.
  - [ ] Add rotation and revocation tests.
  - [ ] Reject mutating 0-RTT behavior.
- [ ] Implement the initial REST API.
  - [ ] Health.
  - [ ] Device profile.
  - [ ] Typed workload capabilities.
  - [ ] Owner-local Project listing and detail used to assemble cluster-wide Project catalogs.
  - [ ] Project Task-entrypoint listing and manual run submission.
  - [ ] Submit job.
  - [ ] Read job.
  - [ ] Cancel job.
  - [ ] Queue-owner job submission, worker claim/long-poll, lease heartbeat, and result completion.
  - [ ] Poll events by cursor.
  - [ ] Submit project transaction placeholder.
  - [ ] Enforce schemas, limits, deadlines, authorization, and audit logs.
  - [ ] Address services and capabilities by logical IDs without exposing provider-local endpoints or credentials.
- [ ] Implement the signed envelope.
  - [ ] Sign outgoing messages.
  - [ ] Verify incoming identity, signature, hash, expiration, and authorization.
  - [ ] Reject replay.
  - [ ] Apply duplicate messages idempotently.
- [ ] Implement the echo/mock queue.
  - [ ] Persist requester outbox submission.
  - [ ] Return `202 Accepted` and a durable job resource.
  - [ ] Persist owner acceptance or rejection.
  - [ ] Have an authorized worker pull a mock lease from the owner over REST.
  - [ ] Execute the mock job only under the active lease.
  - [ ] Post the signed mock result back to the owner over REST for validation.
  - [ ] Persist progress and result.
  - [ ] Poll or long-poll status.
  - [ ] Support cancellation, timeout, retry, and failure.
  - [ ] Resume after requester restart.
  - [ ] Resume after owner restart.
  - [ ] Reject jobs whose workload class or requirements have no compatible destination.
- [ ] Demonstrate the proof.
  - [ ] Submit from Steam Deck.
  - [ ] Execute on the headless device.
  - [ ] Disconnect one device.
  - [ ] Restart one or both applications.
  - [ ] Reconnect.
  - [ ] Recover final state and result.
  - [ ] Confirm logs and HUD share correlation and job IDs.

**Exit criteria**

- The complete two-device proof passes repeatedly across restart and temporary disconnection.
- No trust is derived solely from LAN presence or WireGuard reachability.
- Duplicate delivery cannot duplicate the logical job.
- API, logs, events, and HUD use the same stable device, service, job, and correlation identities.

### Stage 2 / Legacy Phase 7: Project Workspace And Git Operations

**Goal:** Make Apparat useful for real project navigation and controlled repository work.

**Dependencies:** Stage 1 transport, authorization, and persistence proof.

- [ ] Add project registration and ownership.
  - [ ] Register existing filesystem/Git folders.
  - [ ] Assign one owner device.
  - [ ] Store metadata and routes in SQLite.
  - [ ] Validate safe project roots and path traversal protection.
  - [ ] Treat the device holding/running the Git working tree as authoritative.
  - [ ] Advertise signed authorization-filtered Project summaries with revision, freshness, and availability.
  - [ ] Merge owner-local and authorized remote summaries into the Projects list on every device.
  - [ ] Keep offline remote Projects visible as stale/unavailable without treating cached metadata as repository authority.
  - [ ] Route all remote Project reads, Git operations, Task operations, and mutations to the owner through REST.
- [ ] Add file management.
  - [ ] Browse directories.
  - [ ] View text files.
  - [ ] Edit supported text files.
  - [ ] Create, rename, move, and delete with explicit confirmation.
  - [ ] Track unsaved and offline drafts.
  - [ ] Define binary and large-file handling.
- [ ] Add safe Git operations.
  - [ ] Status.
  - [ ] Diff.
  - [ ] Stage and unstage.
  - [ ] Commit with explicit scope.
  - [ ] Branch listing and switching.
  - [ ] History and commit details.
  - [ ] Conflict visualization.
  - [ ] No unrestricted shell escape.
- [ ] Add project chats.
  - [ ] Chat list and history.
  - [ ] Prompt editor.
  - [ ] Message/job/artifact relationships.
  - [ ] Project route selection.
  - [ ] Offline pending messages.
- [ ] Add project transactions.
  - [ ] Stable transaction IDs.
  - [ ] Owner-device apply.
  - [ ] Version/conflict checks.
  - [ ] Durable rejection reasons.
  - [ ] Revise, discard, and retry.
  - [ ] Idempotent replay.
- [ ] Define Pipelines and Project Task entrypoints.
  - [ ] Derive Pipeline status from a Project having at least one Apparat Task entrypoint; do not create a separately owned Pipeline entity.
  - [ ] Store Task definitions and authoritative run records with the Project owner.
  - [ ] Define typed inputs/outputs, permissions, constrained execution behavior, routing, approvals, and entrypoint schema version.
  - [ ] Allow manual Task execution with no trigger binding.
  - [ ] Keep trigger bindings separate so intervals, webhooks, application events, and cluster events can invoke the same Task.
  - [ ] Expose authorized Task summaries and manual invocation through the Project owner's REST API.
- [ ] Add artifacts.
  - [ ] Metadata and ownership.
  - [ ] Hash and MIME type.
  - [ ] Bounded upload/download.
  - [ ] Resume and integrity verification.
  - [ ] Retention and cleanup.

**Exit criteria**

- A Steam Deck can open a real project, inspect files and Git state, submit a project chat job, and recover offline drafts without granting arbitrary shell access.
- Every device presents all authorized Projects across the cluster with owner and freshness state; remote operations go to the owning device.
- A Project with a Task entrypoint appears as a Pipeline, and that Task can run manually without any configured trigger.

### Stage 3 / Legacy Phase 8: Typed Local Services, Compute Queues, And Routing

**Goal:** Discover and manage an arbitrary number of local inference service instances, advertise them safely, and route each workload only through authoritative queues and devices that explicitly support its workload class and requirements.

**Dependencies:** Stages 1–2, plus the documented inference identity, persistence, gateway, and process-ownership decisions.

- [ ] Establish the workload-class registry.
  - [ ] Add `text_generation`.
  - [ ] Add `image_generation`.
  - [ ] Add `video_generation`.
  - [ ] Add `speech_to_text`.
  - [ ] Add `text_to_speech`.
  - [ ] Add `research_boinc`.
  - [ ] Define versioning and extension rules for future classes such as embeddings, reranking, classification, vision analysis, audio generation, simulation, and compilation.
  - [ ] Keep workload class independent from runtime/provider and model/project identity.
- [ ] Implement the statically registered inference driver boundary.
  - [ ] Define typed factory, instance, inspection, executor, progress, result, and error contracts.
  - [ ] Register included drivers explicitly at composition roots without package-global `init()` behavior.
  - [ ] Keep requests/results workload-specific or deliberately tagged and schema-versioned; do not reduce them to arbitrary maps.
  - [ ] Keep local providers as separately supervised processes/services rather than loading model runtimes into the HUD.
  - [ ] Keep third-party out-of-process plugin IPC deferred until a concrete requirement and security/versioning contract exist.
- [ ] Implement durable zero-to-many local service instances.
  - [ ] Assign every configured endpoint a stable `ServiceID` independent from driver kind and workload class.
  - [ ] Permit arbitrary instances of every driver kind, including multiple same-provider and same-workload instances.
  - [ ] Persist desired configuration separately from observed health/inventory and discovered capabilities.
  - [ ] Store provider credentials as secret references; keep provider-local endpoints and secrets out of advertisements and replicated records.
  - [ ] Build the in-memory manager from SQLite desired state and index primarily by `ServiceID` with secondary class, driver, model, and health indexes.
  - [ ] Give every instance independent probing, health, concurrency/admission, cancellation, retry, failure isolation, inventory refresh, and shutdown.
  - [ ] Prove two instances of the same driver and workload class remain distinct through restart, routing, failure, and removal.
- [ ] Add local inference discovery and lifecycle.
  - [ ] Probe known provider defaults and explicit endpoints with bounded concurrency and deadlines.
  - [ ] Validate provider identity/protocol before treating a responding endpoint as a service.
  - [ ] Implement `discovered -> verified -> enabled -> advertised` lifecycle states.
  - [ ] Require explicit enablement/advertisement policy unless a documented policy auto-promotes verified known services.
  - [ ] Refresh health and capability inventory without blocking unrelated instances.
  - [ ] Record safe observations and failures without logging response bodies, prompts, results, credentials, or sensitive endpoint data.
- [ ] Add typed service capability inventory.
  - [ ] Workload class and schema version.
  - [ ] Service runtime/provider.
  - [ ] Endpoint.
    - Keep the endpoint in local desired configuration; advertise only logical service identity and safe connection metadata.
  - [ ] Device owner.
  - [ ] Stable service and capability IDs with zero-to-many cardinality per device.
  - [ ] Models or BOINC projects.
  - [ ] Input/output modalities and limits.
  - [ ] Hardware and accelerator requirements.
  - [ ] Memory, storage, concurrency, and queue limits.
  - [ ] Streaming, progress, cancellation, and artifact support.
  - [ ] Health, load, availability, and validation timestamp.
  - [ ] Power, thermal, and schedule constraints.
  - [ ] Privacy and authorization scope.
  - [ ] Advertisement revision, observation timestamp, expiry, and stale behavior.
- [ ] Advertise and expose services through Apparat.
  - [ ] Derive signed device/service/capability advertisements from desired policy and safe observed state.
  - [ ] Address remote work through `device_id/service_id/capability_id/model_id` as required.
  - [ ] Route remote requests through authenticated Apparat authorization, queues, policy, and audit before reaching localhost providers.
  - [ ] Reject direct remote provider access and never advertise localhost URLs or provider credentials.
  - [ ] Expire stale advertisements and reject unavailable, disabled, unauthorized, or incompatible services before execution.
  - [ ] Expose the same local-service read models through headless API and GUI projection.
- [ ] Add text-generation adapters.
  - [ ] Add OpenAI-compatible adapter.
    - [ ] Text/chat generation.
    - [ ] Streaming versus non-streaming behavior.
    - [ ] Timeouts and cancellation.
    - [ ] Usage and error normalization.
    - [ ] Artifact/result storage.
  - [ ] Add Ollama adapter.
    - [ ] Model inventory.
    - [ ] Generation.
    - [ ] Pull/install state where authorized.
    - [ ] Health and cancellation.
  - [ ] Add llama.cpp service adapter.
    - [ ] Keep llama.cpp outside the HUD process.
    - [ ] Discover server capabilities.
    - [ ] Normalize model and generation behavior.
    - [ ] Document per-platform acceleration separately.
- [ ] Add image-generation provider coverage and adapter contract.
  - [ ] Add Automatic1111 and/or ComfyUI drivers through the common static registry as the first approved image-provider integrations.
  - [ ] Support multiple simultaneously configured image providers and multiple instances of the same provider.
- [ ] Define image-generation adapter contract.
  - [ ] Text-to-image and image-to-image inputs.
  - [ ] Image dimensions, formats, model, sampler, and resource requirements.
  - [ ] Progress, cancellation, preview, result artifacts, and metadata.
- [ ] Define video-generation adapter contract.
  - [ ] Text-to-video and image-to-video inputs.
  - [ ] Duration, dimensions, frame rate, format, model, and resource requirements.
  - [ ] Long-running progress, cancellation, checkpoint, result artifacts, and storage limits.
- [ ] Register speech workload contracts.
  - [ ] Define STT audio inputs, language, timestamps, streaming, and transcript output.
  - [ ] Define TTS text inputs, voice, language, streaming, audio format, and output.
  - [ ] Defer concrete STT/TTS adapters to Phase 10 while preserving typed discovery and routing now.
    - This maps to Stage 5 in the integrated sequence.
- [ ] Register BOINC workload contract.
  - [ ] Define BOINC project identity, client/runtime, platform, application, resource, schedule, and validation requirements.
  - [ ] Defer concrete BOINC execution to Phase 14 while preserving typed discovery and routing now.
    - This maps to Stage 9 in the integrated sequence.
- [ ] Implement authoritative queues.
  - [ ] Direct device queues.
  - [ ] Pool-owner queues.
  - [ ] Single-class queues.
  - [ ] Explicit multi-class queue allowlists.
  - [ ] Per-member capability matching inside heterogeneous pools.
  - [ ] Priorities.
  - [ ] Leases.
  - [ ] Deadlines.
  - [ ] Retries and backoff.
  - [ ] Cancellation.
  - [ ] Failure reasons.
  - [ ] Result and artifact references.
  - [ ] Retention.
  - [ ] Send every remote submission to the queue owner through authenticated REST.
  - [ ] Validate authentication, authorization, idempotency, schema, workload requirements, policy, quota, limits, and queue state before durable admission.
  - [ ] Make inference workers claim or long-poll for compatible work from the queue owner; do not push unleased work to workers.
  - [ ] Issue owner-authoritative attempt IDs, leases, deadlines, and fencing tokens.
  - [ ] Accept heartbeat/progress and signed completion only from the active leased worker.
  - [ ] Validate results/artifacts at the owner before authoritative completion.
  - [ ] Reject stale, expired, replayed, superseded, or duplicate completion without double-completing a logical job.
- [ ] Implement pool execution.
  - [ ] Pool membership.
  - [ ] Owner assignment.
  - [ ] Capability filtering.
  - [ ] Member leases.
  - [ ] Signed member results.
  - [ ] Owner validation and authoritative completion.
- [ ] Implement routing profiles.
  - [ ] Project defaults.
  - [ ] Chat overrides.
  - [ ] Workflow/task step routes.
  - [ ] Required workload class.
  - [ ] Required runtime/provider, model/project, modality, hardware, and feature capabilities.
  - [ ] Optional or required concrete service instance targeting without confusing service identity with provider kind.
  - [ ] Privacy boundary.
  - [ ] Priority and timeout.
  - [ ] Ordered fallback.
  - [ ] Clear explanation of the selected route.
  - [ ] Clear reason each incompatible device or queue was excluded.

**Exit criteria**

- A project can submit real text generation through an explicit route, survive retry/restart, fall back deterministically, and retrieve an authoritative result.
- Mock image, video, STT, TTS, and BOINC jobs route only to matching advertised capabilities.
- Unsupported workload classes and incompatible requirements fail clearly before execution.
- One Apparat node can supervise and advertise several simultaneous provider instances, including multiple instances of the same driver and workload class, without identity collision or shared failure state.
- Remote nodes use Apparat's authenticated logical service API and cannot learn or invoke provider-local endpoints directly.
- Queue requesters submit to the owner, eligible inference workers pull leased tasks from that owner over REST, and only owner-validated returned results become authoritative.

### Stage 4 / Legacy Phase 9: Automation, Scheduling, And Webhooks

**Goal:** Run durable cluster tasks even when some devices are offline.

**Dependencies:** Stage 3 typed queues and routing.

- [ ] Add task definitions.
  - [ ] Owner device.
    - The owner is the device that owns the Task's Project and Git working tree.
  - [ ] Project ID and Apparat entrypoint identity/schema.
  - [ ] Trigger.
  - [ ] Steps.
  - [ ] Inputs and outputs.
  - [ ] Permissions.
  - [ ] Retry and timeout.
  - [ ] Approval policy.
  - [ ] Enabled/paused state.
- [ ] Add triggers.
  - [ ] Manual.
    - Manual execution requires no persistent trigger binding.
  - [ ] Hourly/daily/weekly/monthly or cron-like schedule.
  - [ ] Authenticated webhook.
  - [ ] Internal application event.
  - [ ] Cluster device/service/queue event.
- [ ] Add durable workflow execution.
  - [ ] Run ID and correlation ID.
  - [ ] Current step.
  - [ ] Submitted job references.
  - [ ] Await state.
  - [ ] Checkpoint and resume point.
  - [ ] Retry and timeout.
  - [ ] Cancellation.
  - [ ] Success/failure output.
- [ ] Add safe action boundaries.
  - [ ] Allowlisted application actions.
  - [ ] Project-scoped file/Git operations.
  - [ ] Explicit service calls.
  - [ ] Secret references rather than secret values in task definitions.
  - [ ] Human approval where configured.
  - [ ] No unrestricted remote shell.
- [ ] Add run history and diagnostics.
  - [ ] Timeline.
  - [ ] Inputs and outputs with redaction.
  - [ ] Queue/job linkage.
  - [ ] Failure reason.
  - [ ] Retry history.
  - [ ] Resume behavior after restart.

**Exit criteria**

- A scheduler-owned task can trigger, submit inference, await a result, survive restart, resume idempotently, and produce an auditable outcome.
- The same Project Task entrypoint can run manually or through any authorized interval, webhook, application-event, or cluster-event binding without changing Task identity or ownership.

### Stage 5 / Legacy Phase 10: ASR, TTS, And Voice Control

**Goal:** Turn controller and Debian GUI push-to-talk into a reliable routed cluster capability.

**Dependencies:** The display-free input contract plus Stages 3–4.

- [ ] Add audio capture.
  - [ ] Start while `R2` or the configured Debian GUI push-to-talk key is held.
  - [ ] Stop and submit on release.
  - [ ] Use right `Ctrl` as the documented Debian fallback binding.
  - [ ] Cancel a held Debian recording with `Escape` without submitting on release.
  - [ ] Cancel before submission.
  - [ ] Limit duration and memory.
  - [ ] Store temporary audio safely.
  - [ ] Delete according to privacy policy.
- [ ] Add ASR routing.
  - [ ] Local whisper.cpp service.
  - [ ] Remote ASR service.
  - [ ] Project/context-specific route.
  - [ ] Queue, progress, timeout, retry, and cancellation.
  - [ ] Language and model settings.
- [ ] Add transcription behavior.
  - [ ] Populate focused text field.
  - [ ] Open command palette intent.
  - [ ] Submit prompt when explicitly configured.
  - [ ] Allow review/edit before consequential actions.
- [ ] Add TTS.
  - [ ] OS-native or lightweight service adapter first.
  - [ ] Route generated text separately from ASR.
  - [ ] Support play, pause, stop, and interruption.
  - [ ] Add Qwen3-TTS only as a service adapter after research.
- [ ] Add privacy and feedback.
  - [ ] Recording indicator.
  - [ ] Route/device indicator.
  - [ ] Queue/transcription state.
  - [ ] Failure and retry.
  - [ ] Retention and deletion.
  - [ ] No raw recordings in normal logs.

**Exit criteria**

- Holding and releasing `R2` or right `Ctrl` produces editable transcribed text through a selected local or remote route.
- Spoken output can be routed independently.
- Voice state remains visible, cancellable, and privacy-preserving.
- Capture and GUI focus remain presentation/platform state until explicit submission creates a durable core job.

### Stage 6 / Legacy Phase 11: Platform Packaging And Release Pipeline

**Goal:** Validate and ship each supported platform honestly and independently.

**Dependencies:** Stable secure vertical slice, HUD, and shared-core service lifecycle.

- [ ] Steam Deck/Linux GUI.
  - [ ] Controller database and mappings.
  - [ ] Debian keyboard mapping and text-input precedence.
  - [ ] Debian mouse and touchpad behavior.
  - [ ] Debian right-`Ctrl` push-to-talk and cancellation.
  - [ ] Configurable binding persistence and conflict reporting.
  - [ ] Settings UI for viewing and reassigning scroll, pane, pointer-drag, touch-drag, keyboard, and controller bindings.
  - [ ] Settings UI for customizing HUD aesthetics, including fonts, icon glyphs, and distinct button/panel background colors.
  - [ ] Gamescope/fullscreen/window behavior.
  - [ ] Hi-DPI/readability.
  - [ ] `Steam+X` keyboard.
  - [ ] Microphone and audio.
  - [ ] External WireGuard.
  - [ ] Packaging and update path.
- [ ] Linux headless.
  - [ ] Service installation.
  - [ ] User/system service choice.
  - [ ] CLI and authenticated API control.
  - [ ] Health checks and service-manager operations.
  - [ ] Graceful `SIGINT` and `SIGTERM`.
  - [ ] Runtime directories and permissions.
  - [ ] Startup, restart, logs, and doctor.
  - [ ] No display dependency.
  - [ ] Exclusive node-runtime ownership or daemon-client operation as approved.
- [ ] Windows.
  - [ ] Build and package.
  - [ ] Signing.
  - [ ] Runtime paths.
  - [ ] Controller and audio.
  - [ ] Firewall and HTTPS.
  - [ ] External WireGuard.
- [ ] macOS.
  - [ ] Build and package.
  - [ ] Signing and notarization.
  - [ ] Runtime paths and sandbox considerations.
  - [ ] Controller, microphone, and audio.
  - [ ] External WireGuard.
- [ ] Android.
  - [ ] Continue from Phase 5 Android GUI APK support rather than redoing its build-pipeline work.
  - [ ] Harden release signing, versioning, provenance, and upgrade/rollback behavior.
  - [ ] Replace the temporary APK hash-only update check with installed-version versus latest-version display before offering an update.
  - [ ] Validate additional Android ABIs only after `android/arm64` is proven.
  - [ ] Expand device coverage across phone, tablet, controller, keyboard, and touch configurations.
  - [ ] Validate microphone, audio output, scoped storage, background, and battery behavior against real feature use.
  - [ ] Keep Android headless out of scope unless a later Termux/service-worker strategy is approved.
  - [ ] Keep external WireGuard first; defer VPN-service integration to the later transport/platform phase.
- [ ] Release engineering.
  - [ ] Artifact naming and directory layout.
  - [ ] Checksums and provenance.
  - [ ] Version metadata.
  - [ ] Reproducible build inputs.
  - [ ] Platform test matrix.
  - [ ] Upgrade and rollback.
  - [ ] Separate fast host GUI/headless builds from the full release matrix.
  - [ ] Stop mutating tracked or submodule source checkouts during builds.
  - [ ] Decide whether to replace Git-tracked large binaries with immutable release artifacts and a signed latest manifest.

**Exit criteria**

- Each platform is marked supported only after its build, packaging, input, storage, networking, audio, and lifecycle checks pass.
- Platform lifecycle evidence must include inference-service supervision and process ownership where those capabilities are supported.
- Release artifacts carry version, source commit, hash, size, signer/provenance, toolchain, and target metadata.

### Stage 7 / Legacy Phase 12: Alternative Transports And Long-Term Resilience

**Goal:** Carry the same authenticated durable operations across constrained or human-mediated transports.

**Dependencies:** Stable signed envelope, queues, authorization, advertisements, and transport adapter tests.

- [ ] Add transport conformance tests.
  - [ ] Identity and authorization.
  - [ ] Envelope integrity.
  - [ ] Replay and duplicate behavior.
  - [ ] Expiration.
  - [ ] Acknowledgement.
  - [ ] Fragmentation.
  - [ ] Store-and-forward.
  - [ ] Payload and attachment limits.
- [ ] Research and implement Meshtastic adapter.
  - [ ] Choose protobuf/client source.
  - [ ] Define a dedicated compact application port/message type.
  - [ ] Define allowed commands and status messages.
  - [ ] Define fragmentation and reassembly.
  - [ ] Define acknowledgement and retry.
  - [ ] Define authorization and replay protection.
  - [ ] Reject oversized prompts, artifacts, project files, and model payloads.
- [ ] Research and implement Signal gateway.
  - [ ] Validate available maintained integration approaches.
  - [ ] Define account and device operation.
  - [ ] Map Signal identity to Apparat authorization.
  - [ ] Restrict to notifications, approvals, compact commands, and selected task triggers.
  - [ ] Avoid making Signal the authoritative queue or project transport.
- [ ] Add platform-specific WireGuard management.
  - [ ] Linux kernel/tools adapter.
  - [ ] Windows supported embedding adapter.
  - [ ] Apple Network Extension/WireGuardKit adapter.
  - [ ] Android VPN-service/tunnel adapter.
  - [ ] Preserve external-tunnel compatibility.
- [ ] Add long-term resilience.
  - [ ] Scheduler failover.
  - [ ] Queue-owner migration.
  - [ ] Project-owner migration.
  - [ ] Cluster-directory conflict handling.
  - [ ] Advanced replication.
  - [ ] Optional CRDT research.
  - [ ] Dynamic routing optimization.

**Exit criteria**

- Alternative transports carry only operations appropriate to their capabilities while preserving Apparat identity, authorization, queue, project, and task semantics.
- Service and capability semantics remain subject to the same gateway and advertisement rules across every transport.
- Constrained transports cannot bypass service-advertisement expiry, gateway policy, payload limits, or workload authorization.

### Stage 8 / Legacy Phase 13: Comrades, Chat, And Shared Inference

**Goal:** Add trusted real-friend communication and owner-controlled sharing of otherwise idle inference capacity.

**Dependencies:** Stable identity, authorization, queues, multi-instance routing, audit, and at least one suitable authenticated transport.

- [ ] Define the Comrades trust and identity model.
  - [ ] Define friend invitation, acceptance, rejection, blocking, removal, and reauthorization.
  - [ ] Bind social identity to Apparat user/device identity without exposing private cluster topology.
  - [ ] Define direct and group membership.
  - [ ] Define trust state, verification, key changes, and compromised-account recovery.
- [ ] Add Comrades chat.
  - [ ] Direct conversations.
  - [ ] Group conversations.
  - [ ] Durable outbound and inbound messages.
  - [ ] Offline delivery and retry.
  - [ ] Attachments and artifact references.
  - [ ] Delivery, failure, and read state where supported.
  - [ ] Transport-independent message identity and signatures.
  - [ ] Clear separation between social chat and project chat.
- [ ] Define shared-compute grants.
  - [ ] Resource owner.
  - [ ] Eligible comrades and groups.
  - [ ] Eligible devices and pools.
  - [ ] Allowed workload classes, service runtimes, models, and capabilities.
  - [ ] Schedule and idle-capacity rules.
  - [ ] Priority and preemption.
  - [ ] Concurrency and rate limits.
  - [ ] Daily/monthly quotas.
  - [ ] Prompt/result/artifact visibility policy.
  - [ ] Expiration, pause, revocation, and emergency stop.
- [ ] Implement comrade queues.
  - [ ] Create an owner-authoritative queue class for shared inference.
  - [ ] Assign some or all owner devices or pools.
  - [ ] Assign some or all authorized comrades.
  - [ ] Keep personal work higher priority by default.
  - [ ] Admit work only when owner policy and idle-capacity rules allow it.
  - [ ] Restrict default grants to inference generation.
  - [ ] Deny project files, chats, secrets, tools, shell, administration, and unrelated cluster data by default.
  - [ ] Persist admission, scheduling, execution, result, quota, rejection, and revocation events.
  - [ ] Apply grants to logical service/capability IDs without exposing provider-local endpoints or credentials.
- [ ] Add Comrades HUD.
  - [ ] Contact and group list.
  - [ ] Conversation view.
  - [ ] Trust and verification state.
  - [ ] Shared queue list.
  - [ ] Grant editor.
  - [ ] Usage, quota, priority, availability, and audit views.
  - [ ] Owner emergency pause.
- [ ] Add safety and abuse controls.
  - [ ] Resource-owner consent.
  - [ ] Request authentication and authorization.
  - [ ] Model/service allowlists.
  - [ ] Prompt and output size limits.
  - [ ] Malware and unsafe-content boundary decisions.
  - [ ] Quota abuse and denial-of-service protections.
  - [ ] Audit and dispute evidence without unnecessary content retention.

**Exit criteria**

- A verified comrade can submit an authorized inference job to a comrade queue.
- Owner work retains priority.
- The resource owner can inspect usage and immediately pause or revoke access.
- Shared inference does not expose project files, secrets, arbitrary tools, shell access, or unrelated cluster state.
- Shared inference also does not expose provider-local endpoints or credentials.

### Stage 9 / Legacy Phase 14: Research, BOINC, And Validation Gameplay

**Goal:** Allow opt-in personal compute to support validated BOINC projects through a transparent, constrained, and engaging Research surface.

**Dependencies:** Stable device capabilities, scheduling, resource budgets, task execution, audit, packaging, and safe workload isolation.

- [ ] Define the Research trust model.
  - [ ] Define project identity and authoritative BOINC metadata.
  - [ ] Define technical, scientific, operator, security, legal, and reputation evidence.
  - [ ] Define validation states and who may propose, review, challenge, approve, suspend, or remove a project.
  - [ ] Define revalidation after software, operator, endpoint, or policy changes.
  - [ ] Ensure gameplay scores or popularity cannot substitute for required security and legitimacy checks.
- [ ] Define validation gameplay.
  - [ ] Decide how players discover candidate research projects.
  - [ ] Decide how evidence review and validation actions are represented.
  - [ ] Define progression, reputation, achievements, or collective goals without creating pay-to-win or popularity-only trust.
  - [ ] Define anti-gaming, collusion, Sybil, misinformation, and moderation controls.
  - [ ] Keep project validation evidence inspectable outside the gameplay presentation.
- [ ] Research and select the BOINC integration boundary.
  - [ ] Evaluate BOINC client control/RPC interfaces.
  - [ ] Decide whether Apparat supervises an external BOINC client or embeds selected components.
  - [ ] Select source submodules only after the boundary is approved.
  - [ ] Define project attachment, account/authentication, work fetch, pause, resume, detach, and update behavior.
  - [ ] Define BOINC version and project compatibility policy.
- [ ] Define research resource policies.
  - [ ] Opt-in devices and pools.
  - [ ] CPU, GPU/accelerator, memory, storage, and bandwidth budgets.
  - [ ] Power, battery, temperature, fan, and schedule limits.
  - [ ] Personal, task, and comrade workload priority above research by default.
  - [ ] Pause on user activity, low battery, thermal pressure, metered network, or policy violation.
  - [ ] Per-project and aggregate quotas.
- [ ] Isolate BOINC workloads.
  - [ ] Separate runtime directories and credentials.
  - [ ] Restrict project access to Apparat data, identities, projects, secrets, and networks.
  - [ ] Define process, container, OS sandbox, or platform-specific isolation.
  - [ ] Verify downloaded application signatures and provenance where supported.
  - [ ] Record project, application, work-unit, and result provenance.
- [ ] Add Research HUD.
  - [ ] Candidate and validated project catalog.
  - [ ] Validation evidence and status.
  - [ ] Device/pool assignment.
  - [ ] Resource budget editor.
  - [ ] Active work units and progress.
  - [ ] Contribution history and estimated impact.
  - [ ] Failures, suspension, and audit state.
  - [ ] Gameplay progression and collective goals after the validation design is approved.
- [ ] Add operations and recovery.
  - [ ] Start, pause, resume, and stop.
  - [ ] Recover after restart.
  - [ ] Handle project outage or revocation.
  - [ ] Handle invalid, malicious, or compromised project state.
  - [ ] Preserve owner control and immediate shutdown.

**Exit criteria**

- A user can opt a device into a validated BOINC project with explicit resource limits.
- Personal and comrade work preempts or outranks research by default.
- Research work is isolated from Apparat identities, projects, queues, and secrets.
- Validation evidence and provenance remain auditable independently of gameplay rewards.

## Cross-Cutting Acceptance Requirements

These requirements apply to every implementation stage and must be represented in each focused plan where relevant.

- [ ] Security
  - [ ] Least privilege.
  - [ ] Explicit authorization.
  - [ ] Safe defaults.
  - [ ] Secret redaction.
  - [ ] Key rotation and revocation.
  - [ ] No unrestricted remote execution.
  - [ ] Shared-compute grants never imply project, file, secret, tool, shell, or administrative access.
  - [ ] Research workloads remain isolated from personal and comrade data.
  - [ ] Provider-local endpoints and credentials remain local and undisclosed.
- [ ] Reliability
  - [ ] Stable IDs.
  - [ ] Idempotent operations.
  - [ ] Durable state transitions.
  - [ ] Bounded retries.
  - [ ] Cancellation and deadlines.
  - [ ] Restart recovery.
  - [ ] Independent failure and admission boundaries for every local service instance.
- [ ] Observability
  - [ ] Structured logs.
  - [ ] Correlation IDs.
  - [ ] Queue and job traces.
  - [ ] Health endpoints.
  - [ ] User-readable failure reasons.
  - [ ] Safe service-instance health and advertisement diagnostics.
- [ ] Privacy
  - [ ] No raw prompts, model output, voice, private keys, or tokens in default logs.
  - [ ] Explicit retention.
  - [ ] Clear ownership and visibility.
- [ ] Performance
  - [ ] Frame-time budget.
  - [ ] Memory budget.
  - [ ] Startup budget.
  - [ ] SQLite query budget.
  - [ ] Network and constrained-transport payload budgets.
  - [ ] Bounded discovery/probe concurrency and per-service execution budgets.
- [ ] Recovery
  - [ ] Database backup.
  - [ ] Identity repair.
  - [ ] Migration compatibility.
  - [ ] Upgrade rollback.
  - [ ] Artifact integrity verification.
  - [ ] Service desired-state and job recovery without duplicate advertisement or execution.
## Integrated MVP Completion Definition

The MVP is complete only when all applicable cross-cutting requirements and these product outcomes have evidence:

- [ ] Steam Deck HUD and input
  - [ ] Controller navigation works across all primary tabs.
  - [ ] Comrades is present as the first navigable future-facing tab.
  - [ ] Research is present after Cluster as a navigable future-facing tab.
  - [ ] Settings is present as the final navigable tab.
  - [ ] `R2` push-to-talk state works and can route to ASR.
- [ ] Shared runtime and secure connectivity
  - [ ] A headless Linux worker runs without Ebitengine initialization.
  - [ ] Two devices enroll and authenticate mutually.
  - [ ] HTTPS REST works over WireGuard and trusted LAN with the same authorization.
  - [ ] Signed envelopes reject tampering, replay, expiration, and unauthorized work.
  - [ ] GUI and headless artifacts use the same core state and cannot accidentally double-own one node.
- [ ] Durable state and project operation
  - [ ] A durable job survives duplicate delivery, temporary disconnection, and application restart.
  - [ ] A real project can be browsed with safe Git operations.
  - [ ] A durable scheduled or webhook task can submit and await a job.
  - [ ] Every device shows a cluster-wide authorized Project list while each Git repository and its Task entrypoints remain authoritative on its owner device.
  - [ ] A Pipeline is represented as a Project with one or more Tasks, and a Task can be invoked manually without a trigger.
- [ ] Typed compute routing
  - [ ] A real OpenAI-compatible text-generation job can be routed through an authoritative queue.
  - [ ] Device capability records distinguish text generation, image generation, video generation, STT, TTS, and BOINC support.
  - [ ] Jobs cannot route to devices that do not advertise the requested workload class and requirements.
  - [ ] One node can manage and advertise multiple simultaneous inference service instances, including two instances of the same provider/workload class.
  - [ ] Remote jobs reach provider services only through authenticated Apparat gateway policy and logical service/capability IDs.
  - [ ] Queue owners validate REST submissions; inference workers pull leases and return results by REST; only the owner records authoritative completion.
- [ ] Diagnostics and release
  - [ ] Logs and diagnostics explain failures without leaking sensitive payloads.
  - [ ] The Steam Deck/Linux release is packaged and validated.
