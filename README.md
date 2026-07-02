# Apparat

Apparat is a controller-first console for building and operating a personal-area network: a cluster made from some or all of one person's computers, handhelds, phones, servers, single-board computers, and compute devices.

The cluster coordinates projects, typed compute and inference services, durable message queues, automation, and device capabilities. Apparat initially uses a game engine to deliver a portable HUD rather than to build a game. Gamification comes later. The first target is Steam Deck, followed by Debian/Linux, Windows, macOS, and Android; both headless workers and the full UI/UX.

The detailed implementation sequence lives in [ROADMAP.md](./ROADMAP.md).

## Vision

Personal computing is increasingly distributed across devices with very different strengths: a handheld is convenient, a workstation has storage and a keyboard, a server is always online, a Jetson has inference hardware, an old cell phone is just as capable as a Jetson, only slower.

Apparat treats those devices as one personal cluster. From one HUD, a user should be able to:

- Chat with real friends and deliberately share unused inference capacity with them.
- See which devices are online, what they can do, and what they are currently doing.
- Open project workspaces containing chats, files, artifacts, and safe Git operations.
- Route text, image, video, speech, and research workloads only to devices, pools, services, models, or fallback queues that support the requested workload class.
- Run scheduled, webhook-driven, event-driven, and manually approved automation.
- Delegate explicitly budgeted compute to validated public-interest research.
- Continue useful work while devices are offline and reconcile durable state after reconnection.
- Use controller-first voice and text interactions without giving up keyboard, mouse, touch, or headless operation.

## MVP

The MVP is deliberately smaller than the complete vision. It proves the architecture with one Steam Deck console and one laptop or headless Linux worker.

The canonical first vertical slice is:

1. Both devices are connected through an externally configured WireGuard network.
2. A temporary static peer manifest may be used for the earliest smoke test.
3. The final proof uses authenticated enrollment and a signed cluster directory.
4. The Steam Deck submits an idempotent echo or mock-inference job through HTTPS REST.
5. The worker persists the authoritative queue entry in SQLite.
6. The worker executes the job and persists its result.
7. Either device may restart or temporarily disconnect.
8. The Steam Deck reconnects, resumes from durable local state, and retrieves the result.
9. The HUD and structured logs show the job's owner, correlation ID, attempts, state transitions, and final outcome.

This slice proves the HUD, controller input, identity, networking, API, persistence, queue ownership, offline recovery, and headless runtime before real model runtimes or automation make failures harder to diagnose.

## HUD

The game engine renders the application HUD and input system. For the MVP, there is no RPG world, quest system, progression system, or active gameplay layer. Comrades and Research are visible as future-facing tabs but are not actively developed beyond navigable placeholders.

The canonical tab order is:

1. Comrades
2. Projects
3. Research
4. Cluster
5. Routing
6. Tasks
7. Settings

### Comrades

Comrades will eventually contain real friends and trusted groups. It is the social and cooperative-compute surface for:

- Direct and group chat.
- Friend requests, trust, blocking, and revocation.
- Explicit compute-sharing grants.
- Shared typed-inference queue status.
- Usage, quotas, availability, and audit history.

A **comrade queue** is an owner-controlled low-priority inference queue that may include some or all of the owner's compatible compute devices or pools. The owner may grant some or all comrades permission to submit specific workload classes to that queue.

Comrade work is lower priority than the owner's personal work by default and should run only within owner-defined idle-capacity, schedule, model, concurrency, quota, privacy, and power limits. Access is revocable and auditable.

The default shared capability is inference-only. A comrade queue does not grant access to project files, chats, secrets, arbitrary tools, or remote shell execution.

The Comrades tab is first in the HUD so the long-term product shape remains visible, but chat and resource sharing are explicitly outside the active MVP.

### Cluster

Cluster shows:

- Enrolled and known devices.
- Device roles, trust, connectivity, and last-seen state.
- CPU, memory, storage, accelerator, and service capabilities.
- Supported workload classes such as text generation, image generation, video generation, speech-to-text, text-to-speech, and BOINC research compute.
- Running and queued work.
- Queue depth, utilization, failures, and recent activity.
- Logs, health checks, diagnostics, and cluster operations.

### Projects

Projects contains project folders and opens workspace views with:

- VS Code-like project chats.
- File browsing and editing.
- Artifact galleries and generated outputs.
- Project-specific inference routes.
- Safe Git status, diff, stage, commit, branch, history, and conflict views.
- Offline drafts and owner-device transaction state.

Project operations use constrained application APIs. Apparat does not expose an unrestricted remote shell.

### Routing

Routing manages:

- Device and pool queues.
- Inference services and model inventories.
- Project, chat, workflow, and task routing profiles.
- Required workload class.
- Required capabilities and privacy boundaries.
- Queue priority and timeout.
- Ordered fallback destinations.
- Service health and availability.

The MVP uses explicit routes and ordered fallbacks. Dynamic optimization by load, latency, cost, or quality is deferred.

### Tasks

Tasks manages:

- Manual tasks.
- Cron-like schedules.
- Webhooks.
- Internal application and cluster events.
- Durable multi-step workflows.
- Typed workload submissions and awaited results.
- Human approval steps.
- Retry, timeout, failure, and run history.
- Future Signal and Meshtastic triggers.

Tasks remain durable across application and device restarts.

### Settings

Settings contains:

- User, device, and cluster identity.
- Device enrollment, revocation, and recovery.
- WireGuard and LAN endpoint configuration.
- Storage and database diagnostics.
- Controller, keyboard, touch, audio, ASR, and TTS preferences.
- Logging, backup, repair, and platform diagnostics.

### Research

Research is the third HUD tab, placed alongside the people and projects that give the cluster its purpose. It will allow users to delegate explicitly budgeted compute resources to validated BOINC projects and other future public-interest research systems.

Research will show:

- Candidate and validated research projects.
- Project purpose, operator, software, resource requirements, and validation evidence.
- Opt-in device and pool assignments.
- CPU, accelerator, memory, storage, bandwidth, power, thermal, schedule, and priority budgets.
- Active work units, progress, contribution history, results, failures, and estimated impact.
- Isolation, provenance, and audit state.

BOINC work must remain lower priority than the owner's personal workloads unless the owner explicitly chooses otherwise.

Project validation will eventually participate in Apparat's gameplay mechanics. Those mechanics may make reviewing evidence, validating project legitimacy, selecting trusted research, contributing work, and tracking collective progress more engaging. The exact validation and gameplay design remains open and must not imply that points or popularity replace technical, security, legal, or scientific review.

The Research tab is visible in the HUD but BOINC integration and gameplay are not actively developed in the MVP.

## Controller Input

The initial Steam Deck mapping is:

- `L1`: previous top-level tab.
- `R1`: next top-level tab.
- D-pad or left stick: move focus.
- `A`: activate the focused control.
- `B`: back or cancel.
- `X`: contextual actions.
- Menu: command palette.
- Right stick: scroll long views.
- `R2`: hold to talk.

Holding `R2` begins local audio capture. Releasing it ends capture and submits the buffered audio to the selected ASR route. The HUD must visibly distinguish recording, cancellation, upload or queueing, transcription, failure, and completion.

Steam Deck text fields support the system `Steam+X` on-screen keyboard shortcut and expose a visible keyboard action beside text inputs.

Controller, keyboard, mouse, and touch use one shared focus/navigation model. Essential HUD operations must never be mouse-only.

## System Model

### Devices And Roles

GUI and headless modes share the same domain, persistence, identity, networking, queue, task, and logging layers.

A device may hold several roles:

- GUI console
- Headless worker
- Service host
- Queue owner
- Project owner
- Scheduler owner
- Enrollment authority

No permanently online central server is required. One device may authorize enrollment, but enrolled devices cache signed directory and peer records so the cluster can degrade gracefully when devices are offline.

### Projects

Ordinary filesystem directories and Git repositories remain authoritative for project files.

SQLite stores project metadata, ownership, chats, events, queue routes, artifacts, indexes, transactions, drafts, and sync cursors. It does not replace Git or become the canonical store for every project file.

Each MVP project has one authoritative owner device. Remote mutations are submitted as idempotent transactions. Offline edits remain local drafts or Git commits until the owner accepts them. Rejected or conflicting changes retain their editable content and a durable failure reason.

CRDT-based multi-writer editing is a long-term possibility, not an MVP requirement.

### Queues And Jobs

Every direct-device or pool queue has one authoritative owner device. Requesters retain durable outbound submissions and authorized cached status or result snapshots, not a full mirrored authoritative queue.

Delivery is at-least-once. Duplicate safety comes from stable message IDs, job IDs, correlation IDs, and idempotency keys.

Jobs record:

- Owner and requester devices.
- Project and workflow linkage.
- Required service, model, and capabilities.
- Workload class.
- Priority and deadline.
- Attempt count and retry policy.
- Lease or assignment state.
- Cancellation and failure state.
- Result and artifact references.
- Retention policy.

Pool members execute only work leased or assigned by the pool owner and return signed results to that owner.

#### Comrade Queues

Comrade queues are a future specialization of owner-authoritative pool queues:

- The resource owner selects eligible devices or pools.
- The owner selects eligible comrades or groups.
- The owner defines allowed workload classes, service runtimes, models, priorities, schedules, quotas, concurrency, and retention.
- Personal work preempts or outranks comrade work by default.
- The queue owner remains authoritative for admission, scheduling, execution, cancellation, audit, and results.
- Access can be paused or revoked without changing the comrade's social identity.
- Shared inference permission does not imply file, project, tool, secret, or administrative access.

### Compute Workload Classes

Apparat does not treat all inference as one interchangeable bucket. Every service, device, queue, route, and job declares one or more explicit workload classes.

The initial taxonomy is:

- `text_generation`
  - Chat completion, completion, reasoning, code generation, summarization, and related text-model work.
- `image_generation`
  - Text-to-image, image-to-image, editing, upscaling, and related image-model work.
- `video_generation`
  - Text-to-video, image-to-video, editing, interpolation, and related video-model work.
- `speech_to_text`
  - ASR/STT transcription, language detection, timestamps, and related audio-input work.
- `text_to_speech`
  - Speech synthesis, voice selection, streaming audio generation, and related audio-output work.
- `research_boinc`
  - Validated BOINC work units. This is schedulable research compute rather than model inference, but it uses the same capability, policy, queue, and resource-budget framework.

The taxonomy is versioned and extensible. Future classes such as embeddings, reranking, classification, vision analysis, audio generation, simulation, or compilation may be added without pretending they are text generation.

A device may advertise several independent capabilities. For example, one workstation may support text generation through llama.cpp, image generation through a diffusion service, speech-to-text through whisper.cpp, and BOINC work during an overnight schedule.

Each advertised capability records:

- Workload class and schema version.
- Service/runtime and endpoint.
- Supported models or BOINC projects.
- Input and output modalities and limits.
- Required hardware and available accelerators.
- Memory, storage, concurrency, and queue limits.
- Streaming, cancellation, progress, and artifact support.
- Health, availability, load, and last validation time.
- Privacy, authorization, power, schedule, and owner-policy constraints.

Every job declares its required workload class plus any model, modality, resource, privacy, and feature requirements. Routing excludes incompatible devices before applying queue priority or fallback order.

Queues may accept one workload class or an explicit allowlist. A queue containing heterogeneous devices may route only to members that satisfy the submitted job's requirements.

The first production text-generation adapter targets OpenAI-compatible HTTP APIs, followed by Ollama and llama.cpp. Image, video, STT, TTS, and BOINC adapters are developed and validated independently because they have different inputs, outputs, resource profiles, progress semantics, cancellation behavior, and artifacts.

Routing profiles may be attached to projects, chats, workflows, or individual task steps. Large outputs use artifact references and bounded transfers rather than oversized queue envelopes.

### Automation

Every task has one authoritative scheduler owner during the MVP. Task definitions and run state are durable.

Long-running workflows persist their current step, correlation IDs, idempotency keys, pending jobs, retries, timeouts, and resume points. Sensitive or destructive actions require explicit authorization and may require human approval.

Automatic scheduler election and failover are deferred.

### Research Computing

Research compute is a future owner-authorized workload class, separate from model inference and from personal and comrade queues.

BOINC projects may receive compute only after they pass the configured validation process. Owners explicitly select eligible devices, schedules, resource budgets, thermal limits, network limits, and priority.

Research work must be isolated from Apparat project data, identity secrets, personal inference queues, and comrade workloads. Work-unit provenance, BOINC project identity, resource usage, validation status, results, and failures must be auditable.

### Voice

ASR and TTS are separate service capabilities:

```text
R2 audio capture -> ASR queue -> transcribed text -> command or prompt
generated text -> TTS queue -> spoken playback
```

Whisper, whisper.cpp, and Qwen3-ASR are speech-to-text systems. whisper.cpp is the first portable local ASR reference.

Initial TTS should use an OS-native or lightweight service-backed adapter. Qwen3-TTS is a future routed service because its reference implementation is Python/PyTorch and model-heavy.

## Connection Layer

### Primary Transport

The primary online connection layer is a versioned, authenticated HTTPS REST API used over:

- Externally configured WireGuard.
- Trusted local networks.

HTTPS authentication remains mandatory on local networks. WireGuard provides private reachability and encrypted packet transport; it does not replace application authorization.

The MVP detects and uses existing WireGuard configuration but does not create or manage tunnels. App-managed WireGuard is deferred because Linux, Windows, Apple platforms, and Android require different integration approaches.

LAN discovery may suggest endpoints but never grants trust.

### Identity And Enrollment

App identity remains separate from WireGuard identity.

Device authorization binds:

- The Apparat device identity.
- The HTTPS certificate fingerprint.
- The WireGuard public key.
- User or cluster authorization.
- Roles, permissions, and capabilities.

Enrollment is out-of-band through a short-lived QR code or invite containing the cluster fingerprint and a one-time token.

The connection layer uses TLS 1.3 mutual device authentication. Mutating requests do not use TLS 0-RTT. Certificate issuance, expiration, revocation, rotation, lost-device handling, and trust-store updates are first-class lifecycle operations.

The exact X.509 hierarchy and its relationship to Apparat's device keys will be decided in the security architecture phase.

### REST Resources

The initial API surface is:

```text
GET  /v1/health
GET  /v1/device
GET  /v1/capabilities
POST /v1/jobs
GET  /v1/jobs/{id}
POST /v1/jobs/{id}/cancel
GET  /v1/events?after={cursor}&wait={duration}
POST /v1/project-transactions
```

Mutating operations require an `Idempotency-Key`. Asynchronous job submission returns `202 Accepted`, a durable job ID, and a status resource location.

Cursor-based long polling comes before WebSockets. Requests enforce authentication, authorization, schema versions, content types, body limits, deadlines, and bounded concurrency.

`GET /v1/capabilities` returns typed capability descriptors rather than one generic inference flag. Jobs and routes refer to workload classes and capability requirements by stable identifiers.

The production API will be defined through OpenAPI before server and client implementation.

### Signed Envelope

Durable cross-device messages use a transport-independent signed envelope containing:

- Envelope version and message type.
- Message ID and idempotency key.
- Sender identity and recipient target.
- Created time, expiration, and correlation ID.
- Payload type, schema version, length, and hash.
- Inline payload or artifact reference.
- Signature algorithm and signature.

Receivers validate the version, signature, authorization, expiration, body hash, size, replay state, and idempotency before applying work.

HTTPS carries the envelope through JSON REST resources. Constrained transports may use compact binary encodings while preserving the same identity, authorization, correlation, expiration, hash, and signature semantics.

## Future Transports

The transport abstraction describes capabilities such as payload size, online or delayed delivery, direct or broadcast addressing, acknowledgements, fragmentation, store-and-forward support, attachment support, latency, and cost.

### Signal

Signal is a long-term gateway for notifications, approvals, compact commands, and selected human-mediated interactions. It is not assumed to be the cluster's general data plane or an official bot platform. Feasibility, account operation, identity mapping, and maintainability require dedicated research.

### Meshtastic

Meshtastic is a long-term constrained adapter for compact status, alerts, approvals, cancellation, and small task submissions.

Its payload limits, fragmentation, acknowledgements, routing, and store-and-forward behavior require a dedicated protocol and conformance test. Large prompts, artifacts, model data, and project files do not belong on this transport.

Additional transports must reuse the same durable queue, identity, project, task, and signed-envelope semantics rather than creating parallel application models.

## Design Lineage: Salvagecore

An older version of Apparat is available locally at `third_party/salvagecore` as an ignored, temporary reference checkout. It is not part of this repository, is not a tracked dependency, and is not required to build Apparat.

Salvagecore pursued the personal-cluster goal through an RPG-like interface and qTox/Tor networking. That direction remains interesting but is too complex for this MVP.

### Ideas To Retain

- Ports-and-adapters separation.
- Shared GUI and headless runtime.
- Typed command, event, reducer, and store boundaries.
- Thin Ebitengine adapter around domain state.
- Feature-module registration.
- SQLite lifecycle, forward migrations, ULID IDs, UTC millisecond timestamps, repositories, sync cursors, and read-only inspection.
- Device-owned project and queue authority.
- Durable outbox/inbox and idempotent change feeds.
- User identity separate from device identity.
- Go-native Ed25519 signatures.
- Argon2id and XChaCha20-Poly1305 encrypted private-key files.
- Device authorization and transport-key binding concepts.
- Identity diagnostics, repair, recovery, and archived reset.
- Append-only JSONL logging and sensitive-payload redaction.
- Runtime doctor diagnostics.
- Mock-data-first UI development.
- Reusable pane, list, detail, and layout primitives.

### Ideas Not To Inherit Automatically

- RPG framing or game-world simulation.
- qTox/Tor as the primary transport.
- Salvagecore's qTox/Tor-specific Comrades implementation and transport assumptions.
- Tox/Tor-specific identity assumptions.
- The old three-column layout as the only possible HUD.
- Mouse-first raw input handling.
- Its Ebitengine 2.10 alpha pin.
- Claims that non-host builds work before validation.
- Python OpenAI Whisper as the preferred embedded ASR source.
- Full identity-recovery UI as a prerequisite for the first HUD prototype.

Salvagecore's queue, project ownership, identity, persistence, logging, and runtime conclusions are architectural inputs to this project. Its transport and RPG-first product framing are not; Apparat may add selective gameplay mechanics later where they serve understandable validation, participation, and progress.

The [Salvagecore Reference Baseline](./ROADMAP.md#salvagecore-reference-baseline) records the complete inheritance contract: what the predecessor actually implemented, what remained design-only, how retained concepts map into Apparat, which assumptions are rejected, and how the temporary checkout can eventually be removed without losing project knowledge.

## Third-Party Source Strategy

Source submodules are added when Apparat needs a pinned source tree for audit, architecture work, local replacement, integration, or reproducible upstream study.

Normal Go dependencies remain pinned through `go.mod` and `go.sum`. Build tools are pinned through documented tool versions. A source submodule does not automatically become linked into the MVP binary.

### Required Early Sources

| Path | Source | Purpose |
| --- | --- | --- |
| `third_party/game/ebiten` | `https://github.com/hajimehoshi/ebiten.git` | Cross-platform runtime, input, audio, rendering, and mobile binding tools |
| `third_party/game/ebitenui` | `https://github.com/ebitenui/ebitenui.git` | Retained-mode HUD controls and layouts |
| `third_party/game/debugui` | `https://github.com/ebitengine/debugui.git` | Developer diagnostics overlays |
| `third_party/database/modernc-sqlite` | `https://gitlab.com/cznic/sqlite` | Source reference for the cgo-free SQLite driver |
| `third_party/networking/wireguard-go` | `https://git.zx2c4.com/wireguard-go` | Official userspace WireGuard reference |
| `third_party/networking/wgctrl-go` | `https://github.com/WireGuard/wgctrl-go.git` | Go WireGuard control and inspection APIs |
| `third_party/networking/wireguard-tools` | `https://git.zx2c4.com/wireguard-tools` | Linux and Steam Deck configuration reference |
| `third_party/inference/llama.cpp` | `https://github.com/ggml-org/llama.cpp.git` | Future local LLM service adapter reference |
| `third_party/speech/whisper.cpp` | `https://github.com/ggml-org/whisper.cpp.git` | Portable local ASR reference |

Ebitengine should begin from a stable 2.9.x revision unless a focused implementation plan demonstrates a required newer feature.

Every grouping directory requires a README inventory. Every submodule addition requires an intentional revision, license review, purpose statement, update procedure, and declaration of whether it is a source reference or active build dependency.

### Research Before Adding

- Qwen3-TTS.
- Meshtastic protobuf or client sources.
- A Signal gateway implementation.
- BOINC client, RPC, or manager sources selected by the Research integration design.
- Alternative inference, ASR, TTS, artifact, and transport runtimes.

### Excluded From The MVP Source Set

- qTox, TokTok qTox, and go-toxcore-c.
- Tor.
- WebRTC.
- curl.
- OpenSSL.
- libsodium.
- Qwen3-ASR.
- `golang/mobile` as a source checkout.
- termframe.

Go's standard TLS and HTTP libraries cover the initial API. The retained Go-native identity design covers the initial signature and encrypted-key requirements. OpenSSL does not provide PGP semantics, and adding OpenSSL or libsodium would create unnecessary cgo and cross-platform work.

Ebitengine supplies `ebitenmobile`; Android still requires pinned tools and a native wrapper, but not necessarily a `golang/mobile` source submodule.

## Platform Sequence

1. Steam Deck/Linux GUI and controller input.
2. Linux headless worker and service runtime.
3. Secure two-device WireGuard/LAN vertical slice.
4. Linux desktop packaging and service installation.
5. Windows desktop packaging and external-WireGuard validation.
6. macOS packaging, signing, notarization, and external-WireGuard validation.
7. Android native wrapper, Ebitengine AAR, lifecycle, permissions, keyboard, controller/touch, microphone, audio, storage, and background behavior.
8. Platform-specific app-managed WireGuard.

Cross-platform support is claimed only after target-specific builds and behavior have been validated.

## MVP Non-Goals

- RPG gameplay or world simulation.
- qTox/Tor transport.
- Active Comrades chat or shared-compute implementation beyond a HUD placeholder.
- Active Research/BOINC integration or validation gameplay beyond a HUD placeholder.
- Anonymous public networking.
- Embedded WireGuard provisioning.
- Automatic mesh control-plane negotiation.
- CRDT collaboration.
- Multi-owner authoritative queues.
- Automatic scheduler election.
- Dynamic workload routing optimization.
- Arbitrary remote shell or unrestricted tool execution.
- Full Signal or Meshtastic adapters.
- Bundled Qwen3-TTS.
- Android delivery in the first Steam Deck milestone.

## Planned Design Documents

Implementation phases will create:

- `docs/architecture.md`
- `docs/api.md` and an OpenAPI source
- `docs/security.md`
- `docs/database.md`
- `docs/transport-adapters.md`
- `docs/controller-map.md`
- `docs/platform-matrix.md`
- `docs/comrades.md`
- `docs/research.md`
- Third-party grouping inventories
- Build and release documentation

Open questions and implementation phases are tracked in [ROADMAP.md](./ROADMAP.md).
