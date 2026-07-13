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

## Input And Controls

All GUI targets use one application action model for tab selection, focus movement, activation, cancellation, contextual actions, scrolling, text entry, and push-to-talk. Controller, keyboard, mouse, and touch bindings dispatch the same actions rather than maintaining separate UI behavior.

### Steam Deck Controller

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

### Debian GUI

Debian GUI mode supports the Steam Deck controller mapping unchanged when a compatible controller is attached. Its default keyboard mapping is:

- `Ctrl+PageUp`: previous top-level tab.
- `Ctrl+PageDown`: next top-level tab.
- `Alt+1` through `Alt+7`: open Comrades, Projects, Research, Cluster, Routing, Tasks, or Settings directly.
- `Tab` and `Shift+Tab`: move to the next or previous focusable control.
- Arrow keys: move within lists, trees, grids, menus, and spatial focus groups.
- `Enter`: activate the focused control or submit the current form.
- `Space`: activate a button or toggle the focused checkbox, switch, or selection.
- `Escape`: close a modal, leave the current focus scope, go back, or cancel the current safe operation.
- Menu key or `Shift+F10`: open contextual actions for the focused item.
- `Ctrl+Shift+P`: open the command palette.
- `PageUp`, `PageDown`, `Home`, and `End`: navigate long views and collections.
- Right `Ctrl`: hold to talk; release to submit captured audio to the selected ASR route.

Pressing `Escape` while right `Ctrl` is held cancels the recording, and releasing right `Ctrl` after cancellation must not submit it. Push-to-talk and other bindings are configurable, but these defaults remain the documented fallback and automated-test contract.

The default mouse mapping is:

- Left click: focus and activate the selected control according to its normal desktop behavior.
- Right click: open contextual actions.
- Wheel, touchpad scroll, or touch drag: scroll the focused or pointed scroll container.
- Mouse back button: back or cancel when available.
- Drag: resize approved panes, adjust sliders, select text, or perform an explicitly represented reorder operation.

No essential operation may require pointer-only drag behavior. Complex drag operations must also expose keyboard and controller actions such as move up, move down, move before, move after, or resize by step. An additional mouse button may be configured for push-to-talk, but no mouse-specific push-to-talk binding is required by default.

When a text editor or text field owns focus, ordinary Debian text-editing and clipboard behavior takes precedence over unmodified navigation keys. Global bindings use explicit modifiers so normal typing does not trigger application navigation.

### Debian Headless

Debian headless mode has no HUD focus map and must not initialize Ebitengine. It is operated through:

- Documented CLI commands and configuration.
- The authenticated HTTPS REST API.
- Service-manager operations and health checks.
- `SIGINT` or `SIGTERM` for graceful shutdown.

An interactive terminal UI is not part of the MVP. Adding one later requires a separate input contract and is the admission gate for reconsidering termframe.

### Shared Input Requirements

- Every visible focusable control has a visible focus state.
- Modal surfaces trap focus and restore it when closed.
- Disabled controls are skipped or explained consistently.
- Scrolling has deterministic entry, movement, boundary, and exit behavior.
- Essential HUD operations are never mouse-only.
- Input bindings dispatch commands; they do not mutate durable state directly.
- Binding conflicts, rebinding, and platform-reserved shortcuts are surfaced through Settings rather than silently ignored.

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
R2 or right-Ctrl audio capture -> ASR queue -> transcribed text -> command or prompt
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

`debugui` remains a source-reference checkout for now. Its current source revision tracks Ebitengine 2.10 alpha work, so it is intentionally not an active Go dependency while Apparat starts on stable Ebitengine 2.9.x.

Every grouping directory requires a README inventory. Every submodule addition requires an intentional revision, license review, purpose statement, update procedure, and declaration of whether it is a source reference or active build dependency.

The tracked [`third_party` inventory](./third_party/README.md) records the exact gitlink revisions, licenses, update procedure, and current build relationship for the admitted source set.

### Active Go Workspace

The root Go module is `github.com/cjtrowbridge/apparat`.

Phase 0 establishes:

- Go toolchain baseline: `1.26.4`.
- Ebitengine dependency: `github.com/hajimehoshi/ebiten/v2 v2.9.9`.
- EbitenUI dependency: `github.com/ebitenui/ebitenui v0.7.4-0.20260422023258-b1c31d852489`.
- SQLite dependency: `modernc.org/sqlite v1.53.1-0.20260625155647-5d243466fa05`.
- Developer tool pins: `golangci-lint v2.12.2` and `govulncheck v1.5.0`.

`third_party/` is isolated from root Go package discovery by its own lightweight `go.mod`. This keeps reference checkouts, external tests, GPL reference trees, and temporary predecessor material out of application builds unless a later approved plan explicitly activates an adapter.

Use:

```bash
make tools
make verify
make build
```

`make tools` installs pinned developer tools into the ignored `.tools/bin` directory. `make verify` runs formatting, unit tests, build-pipeline tests, race tests, code-size checks, documentation-completeness checks, linting, and vulnerability scanning.

### Build Process

Run build and verification commands from the repository root. Prefer Makefile targets because they apply repo-local Go cache settings and keep generated files out of source directories.

Prerequisites:

- Go `1.26.4` on `PATH`, or use the checked local toolchain path when present.
- Python 3 for repository scripts.
- Pinned developer tools installed with `make tools` before full verification.
- Network access the first time Go modules or vulnerability data must be downloaded.
- Linux GUI builds need native desktop development headers for Ebitengine/GLFW: `libx11-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libgl1-mesa-dev`, `libxxf86vm-dev`, and `libasound2-dev`.

Common commands:

```bash
make verify
make build
make run-built
make run-built-headless
python3 scripts/build.py
python3 scripts/build.py --help
```

`make build` runs the Python build pipeline. The pipeline has one no-flag entry point: it detects the current machine, reports possible and impossible targets with reasons, then builds every possible target. On a Linux host with desktop build prerequisites, it writes both latest local binaries to:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

The path uses Go `GOOS` and `GOARCH` names such as `linux/amd64`, `linux/arm64`, `windows/amd64`, or `darwin/arm64`. Windows builds use `latest.exe`; other targets use `latest`. `apparat` is the GUI binary and `apparatd` is the headless worker/service binary. Latest artifacts under `releases/` are tracked in Git so other devices can pull the current known-good GUI, headless, and Android APK builds directly. Contributors should rebuild them intentionally when build inputs change.

Android Phase 5 builds the GUI APK only. The canonical Android artifact is:

```text
releases/android/arm64/apparat/latest.apk
```

The same no-flag build pass reports whether Android is possible and builds the APK when the prerequisites are present. The Android pipeline requires an OpenJDK 21 distribution (Eclipse Temurin preferred; Oracle JDK is not permitted), Android SDK command-line tools, platform `android-35`, build-tools `35.0.0`, NDK `27.2.12479018`, and Ebitengine's `github.com/ebitengine/gomobile` module. `JAVA_HOME` must expose `java`, `javac`, and `keytool`. Tools are discovered from `JAVA_HOME`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, and `ANDROID_NDK_HOME`, with ignored repo-local `.tools/` fallbacks when present; `.tmp/` is used only for disposable patched-tool source. If a machine needs custom local paths, copy `build_environment.sample.py` to ignored `build_environment.py` and update environment values there.

The pipeline binds `cmd/apparatmobile` with Ebitengine's generated Android mobile view classes, compiles the tracked wrapper in `android/apparat`, and writes a signed APK to the canonical release path. The wrapper is the current Android render path: on-device evidence shows it opens to the Phase 4 HUD and accepts touch tab selection. It generates an ignored patched local `gomobile-apparat` helper under `.tools/bin` because the pinned Ebitengine `gomobile` package scanner checks for `github.com/ebitengine/gomobile/app` while its regular expression only recognizes `golang.org/x` package symbols. This patch broadens the scanner, supports local Apparat/Ebitengine module replacement for AAR binding, and preserves modern Android SDK metadata (`minSdkVersion=23`, `targetSdkVersion=30`) while compiling/package-building against Android platform 35. It also signs the APK with a generated debug keystore and keeps the Android build independent from the ignored `third_party/salvagecore` checkout.

During Phase 5, the Settings tab also exposes a temporary `Updates` fieldset with an EbitenUI `Check for update` button. On Android, that button calls the wrapper updater through the Gomobile bridge. The action downloads the tracked GitHub `latest.apk`, compares that file's SHA-256 with the installed APK, opens Android's per-app unknown-source permission screen only when an update is needed and permission is missing, then launches the system package installer for user approval. A silent startup check follows the same path but only surfaces user-facing state when action is needed. A later Android release-hardening phase replaces this hash-only bridge with installed-version versus latest-version display before offering an update.

Android headless is intentionally out of scope for Phase 5: the build report marks `android/arm64/apparatd` impossible with a clear message. Users who want headless behavior on Android should use a future Termux/service-worker strategy rather than expecting an APK for `apparatd`.

Use `make run-built` for the GUI artifact smoke test and `make run-built-headless` for the headless artifact smoke test.

`python3 build.py` at the repository root is a compatibility wrapper that delegates to `python3 scripts/build.py`. The canonical script location remains `scripts/build.py`; script inventory and troubleshooting details live in [`scripts/README.md`](./scripts/README.md).

Build troubleshooting:

- If Go tries to write under a read-only home cache, rerun through `make build` or set `GOCACHE` and `GOMODCACHE` to writable paths.
- If module downloads fail, allow network access or pre-populate the Go module cache.
- If the GUI artifact fails with `X11/Xlib.h` or similar missing headers, install the Linux GUI development packages listed above.
- If only the headless worker is needed, run `make build` first and then `make run-built-headless`.
- If Android preflight fails, install or point `JAVA_HOME`, `ANDROID_HOME`/`ANDROID_SDK_ROOT`, and `ANDROID_NDK_HOME` at the pinned toolchain versions, optionally through ignored `build_environment.py`, then rerun `make build`.
- If Android device validation fails, verify `adb devices` outside restricted sandboxes and capture `adb logcat` before treating the APK as launched.
- If documentation checks fail, add or update the closest relevant directory `README.md` and ensure new scripts are listed in `scripts/README.md`.

### Local Runtime

Phase 3 adds shared local runtime startup for GUI and headless modes:

- `cmd/apparat` is the GUI entry point.
- `cmd/apparatd` is the headless worker/service entry point and does not initialize Ebitengine.
- `--smoke-test` initializes the shared runtime, prints a non-window diagnostic line including `root=` and `last_run=`, and exits for build and CI checks.
- `--doctor` validates runtime directories, logging, SQLite, identity status, cluster directory, and local messaging setup; its output includes the exact `last_run.log` path.
- `--runtime-dir` overrides the runtime data root; otherwise `apparat` and `apparatd` use separate platform data directories outside the source tree.
- Default Linux runtime roots are `~/.local/share/apparat/apparat` for the GUI and `~/.local/share/apparat/apparatd` for the headless worker, unless `XDG_DATA_HOME` is set.
- Normal GUI and headless startup prints the selected runtime root and `last_run.log` path before entering the long-running process.
- Runtime subdirectories include database, logs, identity, cache, artifacts, backups, and recovery.
- `last_run.log` is recreated in the runtime root at every process start and records verbose startup, component, doctor, smoke-test, failure, panic, and shutdown diagnostics for immediate debugging.
- Append-only JSONL logs remain under the runtime `logs/` directory for durable structured history.
- GUI builds compiled with the `gui` build tag enter the Ebitengine run loop; headless builds keep the non-window path available for worker and service validation environments.

Contributor verification includes a source-size gate: code files must be at most 400 physical lines unless they are excluded generated/vendor/reference artifacts. Over-limit files should be decomposed into smaller package files with local README context where needed.

Local startup creates an append-only JSONL log, opens SQLite with foreign keys, applies checksumed forward migrations, initializes local cluster-directory tables, and initializes durable inbox/outbox/replay/cursor primitives.

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

Ebitengine supplies mobile runtime packages through the admitted Ebitengine source tree and the pinned `github.com/ebitengine/gomobile` module. Phase 5 uses an Apparat-owned Android wrapper because direct `GoNativeActivity` builds can initialize runtime state but remain on the Android splash/default icon instead of attaching Ebitengine's `EbitenView`. The wrapper path renders the current HUD through generated Ebitengine mobile view classes and tracked Apparat Java/manifest sources rather than depending on salvagecore's temporary `third_party/cicd/mobile` reference.

## Platform Sequence

1. Steam Deck/Linux GUI and controller input.
2. Debian/Linux GUI keyboard, mouse, optional-controller, and push-to-talk input.
3. Linux headless worker and service runtime.
4. Android GUI APK build pipeline for `releases/android/arm64/apparat/latest.apk`; no Android headless artifact in this phase.
5. Secure two-device WireGuard/LAN vertical slice.
6. Linux desktop packaging and service installation.
7. Windows desktop packaging and external-WireGuard validation.
8. macOS packaging, signing, notarization, and external-WireGuard validation.
9. Android release hardening beyond the first GUI APK.
10. Platform-specific app-managed WireGuard.

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
