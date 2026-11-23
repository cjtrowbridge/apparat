# Apparat

Apparat is a controller-first console for building and operating a personal-area network: a cluster made from some or all of one person's computers, handhelds, phones, servers, single-board computers, and compute devices.

The cluster coordinates projects, typed compute and inference services, durable message queues, automation, and device capabilities. Apparat initially uses a game engine to deliver a portable HUD rather than to build a game. Gamification comes later. The first target is Steam Deck, followed by Debian/Linux, Windows, macOS, and Android; both headless workers and the full UI/UX.

The canonical implementation sequence lives in [ROADMAP.md](./ROADMAP.md). Product implementation proceeds through focused approved execution plans bound to its phase checklist and the detailed contracts linked from this README.

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

The next milestones deliver offline local repos, then trusted cluster membership, then a read-only catalog of projects hosted by members. The eventual two-device queue proof proceeds as follows:

1. Both devices are connected through an externally configured WireGuard network.
2. A temporary static peer manifest may be used for the earliest smoke test.
3. The final proof uses authenticated enrollment and a signed cluster directory.
4. The Steam Deck submits an idempotent echo or mock-inference job to the queue-owning worker through HTTPS REST.
5. The queue owner validates the request and persists the authoritative queue entry in SQLite.
6. An authorized inference worker pulls a leased task from the queue owner through HTTPS REST, executes it, and posts the result back to the owner.
7. Either device may restart or temporarily disconnect.
8. The Steam Deck reconnects, resumes from durable local state, and retrieves the result.
9. The HUD and structured logs show the job's owner, correlation ID, attempts, state transitions, and final outcome.

This slice proves the HUD, controller input, identity, networking, API, persistence, queue ownership, offline recovery, and headless runtime before real model runtimes or automation make failures harder to diagnose.

## HUD

The game engine renders the application HUD and input system. For the MVP, there is no RPG world, quest system, progression system, or active gameplay layer. Comrades and Research are visible as future-facing tabs but are not actively developed beyond navigable placeholders.

The GUI includes the offline STMPCK gallery on Linux and Android. Its 💉 tab sits immediately before Settings. Settings is always last when adding or reorganizing tabs.

The canonical order is Comrades, Projects, Cluster, Research, STMPCK, Settings.

The tab strip uses compact symbol-only buttons: ☭ Comrades, 🔨 Projects, 🖧 Cluster, 🧪 Research, 💉 STMPCK and ⚙️ Settings. All six fit at 320 logical pixels. Tab buttons have no popup tooltips. The active tab has a fixed page heading. STMPCK places its heading in the gallery toolbar, immediately before the star. Emoji artwork is bundled offline with [attribution](internal/adapters/gui/emoji/README.md), independent of device fonts.

Cluster uses a selector panel for device and operations context. Its content panel includes Routing and Tasks; Tasks manages schedules, webhooks, approvals, and run history without becoming a top-level tab.

For controls in the right-hand detail pane, the default layout for choosing among options is a child page in that pane. It starts with a touch-sized 🔙 Back button, then a heading, then the options. On a phone the page uses the full detail area. Back returns one level and restores the previous selection and scroll position. System-owned prompts such as biometric unlock and brief confirmations of consequential actions may interrupt the page.

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

### Node keys

At GUI startup, Apparat loads the saved identity using its configured protection. A password or system unlock prompt may appear. The restore-or-create choice appears only when the device identity is missing. If you have previously used the device, restore its original private key. Cluster peers recognize its signing identity. Generating a different key requires pairing again. Choose **Find existing keys** or confirm **Generate new keys**, then create a password or use the system credential option when available. Save a password-protected recovery copy somewhere private. Cancel keeps signing locked and lets you work offline. Node keys alone do not enroll a peer. See [node identity and recovery](docs/node-identity.md).

First-time personal setup asks whether this is your first Apparat device or you are joining an existing cluster. Joining opens local pairing discovery or restores an encrypted backup. Each pairing screen displays four digits, including any leading zeros. Update both apps before pairing. System unlock is optional on devices with a supported credential provider.

### Cluster

Open Cluster to see the workers saved for your user identity. Selecting a worker shows its fingerprint, cluster and last known address. Add New Device opens local discovery. Routing, Tasks and Inference Types retain their example entries. **➕ Add New Device** is the first row under **Cluster Devices**. It opens **Local network discovery** in the right detail pane, or a full-width detail screen on narrow devices. A client can browse while its own server is off. A host appears only when its bootstrap API and advertisement are active on an explicitly approved network/interface pair. Select a ➕ hostname to open its node details. Results are temporary, unverified candidates. Choose **Claim this host for your cluster** to authenticate with the user and device identities already unlocked on this installation. The first authenticated user becomes the worker’s approved admin. Additional users start pending while registration is open. On first contact Apparat learns the worker key from the selected LAN endpoint and saves its fingerprint for subsequent connections. Linux can host the bootstrap API on an approved IPv4 interface. Pixel devices running Android 13 or later can browse on Wi-Fi. On Linux, select the interface to browse. Windows and macOS have an IPv4 browser and still need on-device qualification.

To pair a Pixel with your existing Linux Apparat identity, enable API hosting for the desired home network and interface in Settings on Linux. In Cluster → Add New Device, browse that interface and select **Pair another personal device**. On the phone, unlock the device identity, dismiss personal identity creation if offered, select the Linux hostname in discovery, and choose **Receive my identity from this device**. Enter the code from each screen on the other and enter your existing recovery password on the phone. The phone retains its own device key. Select a saved worker and choose **Check my access** to contact it directly. Pairing copies the signed worker directory once. After pairing, each unlocked app polls its saved workers and refreshes on opening Cluster or Projects. The worker supplies a signed view scoped to that user and device. Worker details show the last applied revision. Administrators can change registration, the polling interval, user roles, pools and service grants. Pool tasks lets an authorized user submit text to hash and retrieve the verified result. The default interval is five minutes.


After device unlock, the GUI offers to create or restore your personal identity. Choose a screen name, protect the user key with a password and save its encrypted recovery file. Tap the pencil beside your screen name in **Settings → Personal identity** to edit it. Android keeps your identities unlocked when switching apps. When both identities use Android unlock, one prompt opens a new session after any one-time credential migration. System unlock is optional on each personal device. The worker stores public user credentials only.

The planned personal cluster separates a portable password-protected **user identity** from each installation's **device identity**. Personal pairing transfers the encrypted user key to another explicitly confirmed personal app; headless workers see only public user credentials. Cluster policy in SQLite determines admin roles, service access, groups and pools. Any cluster admin may issue a change. A designated authority orders changes, and peers use authenticated revision checks to synchronize on relevant actions and at a policy-controlled background interval (default five minutes). Pool queues use independent revisions and a common local task table. A worker accepts leased jobs; authorized clients retrieve verified artifacts separately. See the [delivery milestones and acceptance conditions](ROADMAP.md#personal-cluster-governance-and-queue-sync).

The planned claim flow registers the person’s public key, screen name and originating hostname, and verifies possession of the signing key. The first authenticated claimant on an unclaimed node becomes its approved admin. Later users await approval. Node admins manage service access, admin roles and new registrations in Cluster. The editable screen name is also intended for chat and separately opt-in STMPCK leaderboards. See the [Phase 10 roadmap](ROADMAP.md#user-claims-and-node-administration).

The first membership workflow will let a user create a cluster, discover it from another device or use an explicit invite, verify and approve enrollment, reconnect, leave, and revoke access. Discovery supplies candidate addresses. Membership requires confirmation and authentication.

The full Cluster view will show:

- Enrolled and known devices.
- Device roles, trust, connectivity, and last-seen state.
- Safe resource telemetry: CPU capacity and utilization; RAM and swap total, used, and available; disk capacity, used, and free space; and GPU/accelerator identity, utilization, and VRAM capacity, used, and available.
- Resource-observation timestamp, freshness, unavailable/unknown state, and the peer address currently reached; telemetry is advisory and never grants trust or authorizes work.
- One stable peer identity with multiple candidate addresses, including LAN, externally managed WireGuard, Tailscale, or another overlay; addresses are reachability hints, not separate devices or sources of trust.
- Installed version, compatible update availability, update channel, and update/rollback progress for every enrolled Apparat node.
- An authorized node can request a target node's update through the cluster API; the target independently verifies the signed release and applies its own safe update, migration, restart, and rollback procedure.
- Supported workload classes such as text generation, image generation, video generation, speech-to-text, text-to-speech, and BOINC research compute.
- Running and queued work.
- Queue depth, utilization, failures, and recent activity.
- Logs, health checks, diagnostics, and cluster operations.
- Routing detail for typed workload queues, compatibility filters, priorities, fallbacks, and service health.

### Projects

Project Repos lists Git working trees the user explicitly registers on this device. It works offline before the device joins any cluster. SQLite remembers the registration and its stable project ID. The checkout supplies the current Git branch and labeled tasks from a JSON-format .vscode/tasks.json. Apparat displays the labels without running repository commands. A moved or inaccessible checkout can be reconnected, and removing a registration leaves its files intact.

Projects keeps three actions pinned above the checkout list as you scroll.

- **➕ Clone remote repo** downloads a shallow checkout from GitHub. Enter an owner/repo or URL, preview the destination and account, then select **Clone**. The default destination is `<node-root>/data/repos/<owner>/<repo>`.
- **➕ Init new repo** creates an empty Git working tree on `main`. Enter a name, optionally choose a parent folder, and select **Initialize repo** to open its Files workspace. The default parent is `<node-root>/data/repos/`.
- **➕ Import local repo** registers a Git working tree in place. Choose the existing repository folder and select **Import repo**.

The Linux folder button uses Zenity when installed, and the path field also accepts absolute paths. Linux uses the system GitHub CLI for private clones, with a default account and per-repo overrides. Android starts with local checkouts and public clones. Phase 11B adds Android GitHub sign-in and private clones. Local checkouts open offline regardless of sign-in state.

After trusted membership is available, the same list will include authorized summaries that other members choose to share. Each entry identifies its host and availability. Two devices with checkouts of the same Git remote own separate projects. The first shared catalog supports browsing. Later workspace operations go through the owning device’s authenticated API.

The Project Repos list uses blue outlines for repos without configured actions and purple outlines for repos with configured actions. Repos with configured actions also show 🧠 on the right. Both open the same workspace view. Each row shows the full checkout location beneath its name, prefixed with `local://` for this device, followed by the current branch and abbreviated checkout commit. On Android, `local://apparat/` starts at this installation’s data directory and `local://~/` starts at primary shared storage. Other volumes keep their own labels.

Phase 9A provides offline file editing, Git operations, durable local drafts and manual Task results before cluster enrollment. Phase 11A will verify those workflows in the mixed local and remote catalog. Phase 11C will add authorized workspace operations on another owner.

The planned workspace includes:

- VS Code-like project chats.
- File browsing and editing.
- Artifact galleries and generated outputs.
- Project-specific inference routes.
- Safe Git status, diff, stage, commit, branch, history, and conflict views.
- Offline drafts and owner-device transaction state.
- Task entrypoints, triggers, typed inputs and outputs, approvals, routing, and run history.

Project operations use constrained application APIs. Apparat does not expose an unrestricted remote shell.

A Pipeline is not a separate repository or workflow authority. It is a project that defines at least one Apparat-executable entrypoint. Each entrypoint is a Task owned with that project. A Task may be run manually with no trigger, or connected to one or more interval, webhook, internal-application, or cluster-event triggers.

Personal pairing saves each GUI peer on both devices. Existing installations also learn GUI peers from their synchronized worker directory. Cluster Devices lists authenticated GUI installations above headless workers. 🎯 marks the installation you are using. Setup asks for a hostname when the OS supplies a generic name such as localhost. On a headless node use `apparatd identity-setup` for that prompt. From any administrator app, open a device, choose **Rename** for the displayed device, enter a hostname and save. The cluster synchronizes the name by device fingerprint. The operating system hostname, network address and credentials stay independent. Phone and laptop icons indicate the platform reported during login. A network icon marks an older app until it reports its platform. Pool pages show the last synchronized assignments. Administrators can open **Select pools** on a headless worker and check or uncheck each pool. **Save changes** commits the selection in one signed SQLite policy revision. **Cancel** or Back discards edits. Connected app peers copy the authorized membership on their next scheduled synchronization. A worker selection controls its hash-text service. GUI devices appear in the directory for pairing and identification. Pool membership is available only to headless workers. If you skip naming your installation during setup, open its 🎯 entry and choose Rename. Spaces in a local name become hyphens. A cluster publication failure leaves the name saved locally and offers a retry.

The Cluster heading gear opens settings for people, groups, pools and synchronization. Worker settings contain registration and update controls. Open Pool tasks on a worker to submit a task, check progress or retrieve a verified result. New clusters start with Chat Pool (high), Project Pool (medium), Comrade Pool (low) and Research Pool (last). Each starts empty. Enabling services and granting access are separate administrator actions.

### Routing (Cluster detail)

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

Tasks are Apparat-executable project entrypoints. A Task belongs to one project and is authoritatively defined and started by that project's owner device. The Task may perform owner-local project operations and submit typed workload steps to routing queues.

Tasks manages:

- Manual execution without a trigger.
- Cron-like schedules.
- Webhooks.
- Internal application and cluster events.
- Durable multi-step workflows.
- Typed workload submissions and awaited results.
- Human approval steps.
- Retry, timeout, failure, and run history.
- Future Signal and Meshtastic triggers.

Attaching a trigger is optional: a Task with no trigger is still a valid manually executable project entrypoint. Task definitions, trigger bindings, run state, and results remain durable across application and device restarts.

### Settings

Settings contains:

- User, device, and cluster identity.
- Device enrollment, revocation, and recovery.
- WireGuard and LAN endpoint configuration.
- Storage and database diagnostics.
- Controller, keyboard, touch, audio, ASR, and TTS preferences.
- Logging, backup, repair, and platform diagnostics.

### Research

Research follows Cluster in the HUD tab order, placed alongside the people, projects, and local infrastructure that give the cluster its purpose. It will allow users to delegate explicitly budgeted compute resources to validated BOINC projects and other future public-interest research systems.

Research will show:

- Candidate and validated research projects.
- Project purpose, operator, software, resource requirements, and validation evidence.
- Opt-in device and pool assignments.
- CPU, accelerator, memory, storage, bandwidth, power, thermal, schedule, and priority budgets.
- Active work units, progress, contribution history, results, failures, and estimated impact.
- Isolation, provenance, and audit state.

BOINC work must remain lower priority than the owner's personal workloads unless the owner explicitly chooses otherwise.

Project validation will eventually participate in Apparat's gameplay mechanics. Those mechanics may make reviewing evidence, validating project legitimacy, selecting trusted research, contributing work, and tracking collective progress more engaging. The exact validation and gameplay design remains open and must not imply that points or popularity replace technical, security, legal, or scientific review.

The Research tab is visible in the HUD but BOINC integration and gameplay are not actively developed in the MVP. Until then, enrolled-project placeholders show a clearly labeled mock personal contribution and sortable mock friend-contribution table; neither represents a live BOINC account, research result, or social-graph record.

## Input And Controls

All GUI targets use one application action model for tab selection, focus movement, activation, cancellation, contextual actions, scrolling, text entry, and push-to-talk. Controller, keyboard, mouse, and touch bindings dispatch the same actions rather than maintaining separate UI behavior.

### Steam Deck Controller

The initial Steam Deck mapping is:

- **`L1`**: previous top-level tab.
- **`R1`**: next top-level tab.
- D-pad or left stick: move focus.
- **`A`**: activate the focused control.
- **`B`**: back or cancel.
- **`X`**: contextual actions.
- Menu: command palette.
- Right stick: scroll long views.
- **`L2`**: hold to talk.

Holding `L2` begins local audio capture. Releasing it ends capture and submits the buffered audio to the selected ASR route. The HUD must visibly distinguish recording, cancellation, upload or queueing, transcription, failure, and completion.

Steam Deck text fields support the system `Steam+X` on-screen keyboard shortcut and expose a visible keyboard action beside text inputs.

### Debian GUI

Debian GUI mode supports the Steam Deck controller mapping unchanged when a compatible controller is attached. Its default keyboard mapping is:

- **`Ctrl+PageUp`**: previous top-level tab.
- **`Ctrl+PageDown`**: next top-level tab.
- `Alt+1` through `Alt+5`: open Comrades, Projects, Cluster, Research, or Settings directly.
- `Tab` and `Shift+Tab`: move to the next or previous focusable control.
- Arrow keys: move within lists, trees, grids, menus, and spatial focus groups.
- **`Enter`**: activate the focused control or submit the current form.
- **`Space`**: activate a button or toggle the focused checkbox, switch, or selection.
- **`Escape`**: close a modal, leave the current focus scope, go back, or cancel the current safe operation.
- Menu key or `Shift+F10`: open contextual actions for the focused item.
- **`Ctrl+Shift+P`**: open the command palette.
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

Arrows and the Steam Deck D-pad switch tabs immediately along the tab bar. Move toward the page to focus its controls. Enter or A activates the focused control or starts editing a field. Escape or B leaves editing with the draft intact. See [Input and Focus Contract](docs/controller-map.md).

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

The shared core exposes a transport-neutral internal application API made of typed commands and queries. The GUI calls that API directly in-process; it does not send HTTP requests to its embedded core. The REST server is an external adapter that maps versioned HTTP resources onto selected operations from the same internal API. GUI adapters and REST handlers do not access SQLite directly or maintain parallel product rules.

SQLite is the authoritative durable store behind the internal application API. Before cluster identity and request signing are established, the REST adapter is loopback-only and exposes selected read operations. The GUI may still use internal commands directly. After authentication, authorization, signing, replay protection, and audit are active, the REST adapter may expose selected existing commands to trusted remote devices without creating REST-specific setters or state transitions.

A device may hold several roles:

- GUI console
- Headless worker
- Service host
- Queue owner
- Project owner
- Enrollment authority

No permanently online central server is required. One device may authorize enrollment, but enrolled devices cache signed directory and peer records so the cluster can degrade gracefully when devices are offline.

During the MVP, the Project owner evaluates that Project's Task trigger bindings and owns its Task runs. Scheduling is an owner-local responsibility, not a separately elected authority. Scheduler failover is a later resilience feature and must not silently move Project or Task ownership.

### Projects

Every project is an ordinary Git repository whose working tree lives on one authoritative owner device. That device owns the project because it stores and runs the repository; changing ownership requires an explicit future migration rather than an incidental cache or clone.

SQLite stores project metadata, ownership, chats, events, queue routes, artifacts, indexes, transactions, drafts, and sync cursors. It does not replace Git or become the canonical store for every project file.

Once catalog sharing is implemented, each enrolled device maintains a cluster-wide project catalog assembled from its local projects and authorized project summaries advertised by other devices. Therefore the Projects list on any enrolled device includes every project in the cluster that the current user/device is authorized to know about, along with its owner, availability, and freshness. Cached metadata may keep an offline owner's project visible as stale or unavailable, but the cache is not the repository and does not become authoritative.

All remote project listing, metadata, file, Git, Task, and transaction operations go to the project owner through authenticated REST APIs. Remote mutations are submitted as idempotent transactions. Offline edits remain local drafts or explicit Git commits until the owner accepts them. Rejected or conflicting changes retain their editable content and a durable failure reason.

A Pipeline is a project with one or more Apparat Task entrypoints. Pipeline identity is therefore the project identity, not a second independently owned object. Project entrypoints, Task definitions, trigger bindings, and authoritative run records live with the repository's owner device.

CRDT-based multi-writer editing is a long-term possibility, not an MVP requirement.

### Queues And Jobs

Every direct-device or pool queue has one authoritative owner device. Requesters retain durable outbound submissions and authorized cached status or result snapshots, not a full mirrored authoritative queue.

Cross-device queue operations use authenticated REST requests to the queue owner:

1. A requester submits work to the owner with a stable job ID and idempotency key.
2. The owner authenticates and authorizes the requester, validates the schema, workload class, requirements, queue policy, limits, quota, and current state, then durably accepts or rejects the request.
3. Authorized inference devices assigned to the queue poll or long-poll the owner for work. The owner selects eligible work and returns a bounded lease; the owner does not push directly into worker memory or databases.
4. The worker executes only the leased task, reports progress or lease renewal as allowed, and posts a signed terminal result or failure back to the queue owner through REST.
5. The owner validates the worker, lease/fencing token, result schema, idempotency, and artifacts before recording authoritative completion. A worker's local completion is not the queue's authoritative result.

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

Pool members pull only work leased by the pool owner and return signed results to that owner. Lease expiry permits recovery and reassignment; result and completion handling remain idempotent so a late or duplicated worker response cannot complete a logical job twice.

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
- Service/runtime and stable logical service identity. Provider endpoints remain owner-local configuration and are not advertised.
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

### Local Inference Service Instances

Every Apparat node may manage zero to many localhost-exposed inference services, including several instances of the same provider and several instances serving the same workload class. A headless or GUI node may therefore expose Ollama, llama.cpp, Automatic1111, ComfyUI, and other approved services simultaneously without collapsing them into one capability record.

The model keeps four identities separate:

- A workload class describes the requested operation, such as text or image generation.
- A driver kind identifies a provider protocol, such as `ollama`, `openai_compatible`, `llama_cpp`, `automatic1111`, or `comfyui`.
- A service instance is one configured local endpoint with a stable `ServiceID`, independent lifecycle, limits, policy, and health.
- A capability is a model, modality, format, feature, or limit currently discovered on one service instance and identified by a stable `CapabilityID`.

Provider plugins are statically compiled Go drivers registered explicitly at the application composition root. Apparat does not use Go dynamic `.so` plugins. Each driver supplies a factory; each configured endpoint creates an independently supervised instance; and typed executors preserve meaningful differences among workload classes. The manager is keyed primarily by `ServiceID`, never by provider name or workload class.

SQLite is authoritative for desired service configuration and the last safe observed state. Desired configuration, observed health/inventory, discovered capabilities, and derived advertisements are separate records. Provider credentials are stored through local secret references, and provider-local URLs, tokens, prompts, results, and raw failure bodies never enter cluster advertisements or normal logs.

Remote peers never connect directly to another device's localhost provider. They address logical device, service, capability, and model identities through the authenticated Apparat gateway. The gateway applies authorization, queue admission, routing, limits, audit, and artifact policy before invoking a local provider.

Service advertisements carry an owner-scoped monotonic revision, observation time, and expiration time. The default service-advertisement lifetime is 120 seconds and owners refresh by 60 seconds while a service remains eligible. Expired advertisements immediately become non-routable but may remain visible as stale diagnostics for up to 24 hours. A newer revision supersedes every older revision from the same owner; re-advertisement after expiry requires a fresh observation and revision.

### Automation

Every Task is an Apparat-executable entrypoint belonging to one project. A project with at least one Task is presented as a Pipeline. The project owner is also the authoritative owner for its Task definitions and run records during the MVP.

A Task may have zero or more trigger bindings. With no trigger it runs only when explicitly invoked in Apparat. Supported trigger bindings include intervals or cron-like schedules, authenticated webhooks, internal application events, and cluster device/service/queue events. A trigger creates a Task run; it is not itself the executable entrypoint.

Long-running workflows persist their current step, correlation IDs, idempotency keys, pending jobs, retries, timeouts, and resume points. Sensitive or destructive actions require explicit authorization and may require human approval.

Automatic scheduler election and failover are deferred.

### Research Computing

Research compute is a future owner-authorized workload class, separate from model inference and from personal and comrade queues.

BOINC projects may receive compute only after they pass the configured validation process. Owners explicitly select eligible devices, schedules, resource budgets, thermal limits, network limits, and priority.

Research work must be isolated from Apparat project data, identity secrets, personal inference queues, and comrade workloads. Work-unit provenance, BOINC project identity, resource usage, validation status, results, and failures must be auditable.

### Voice

ASR and TTS are separate service capabilities:

```text
L2 or right-Ctrl audio capture -> ASR queue -> transcribed text -> command or prompt
generated text -> TTS queue -> spoken playback
```

The GUI includes whisper.cpp and Whisper Tiny (English Only). Open a project document, place the caret or select text, then hold the PTT control and speak. Release to insert the final transcript. A live preview stays outside the document. Escape cancels, and saving a document or sending a message is a separate action. Android asks for microphone permission on first use. Audio stays in memory on the device, in bounded chunks while the control is held.

### Voice commands and the command directory

Open **Settings → Voice Commands → Browse available commands** to browse navigation destinations or search by name or path. Select a result with a pointer, touch, keyboard or D-pad. Outside an editor, hold L2, right Ctrl or the on-screen **PTT** control, speak, then release. Double-tap the PTT button or L2 to lock listening. The first tap must be brief and the second press must follow within 350 ms. Tap either control once to finish the same recording, or press Escape to discard it. “What can I say?” opens the directory. In a focused text field, including action search, PTT inserts the words you say. “Open Projects” in a chat draft stays in the draft.

Opening a game from the gallery or by voice shows its command title at the top. Choose a board, level, or speed where available, or select Start. A saved game offers Resume and New game. You can back out of new-game setup before starting to keep your save.

Say “open Cluster” for the cluster management tab or “open Bubble Volley” for the bubble-matching game in STMPCK.

Try “switch to Settings,” “launch project Apparat,” “open Apparat project,” “show tasks for Apparat,” “open game Snake,” or “open chat Release planning.” “Open Apparat” opens a unique exact match. When names collide, the chooser shows numbered paths including the owning project for each thread. The chooser occupies the middle two-thirds of the screen in each dimension. Tap the dimmed area to dismiss it and return to the same page. Answer “the project,” “the second one” or “cancel,” or select a row. A partial or near match always asks before opening. With no match you can retry or cancel in place.

Whisper transcribes locally. Ordinary code interprets the final words using the same catalog as search. Registered checkouts and their conversations use local workspace data. **Enable spoken action choices** in **Voice Commands** reads the clarification aloud where local speech output is supported (an offline Android voice or installed Linux `espeak-ng` / `espeak`). Text choices are always available. Opening Tasks shows its existing controls. Running a task still uses its separate execution flow.

Open **Settings → Speech Recognition** to choose a dictation model. Tiny, Base, Small and Medium each offer English Only and Multilingual variants. Select a variant to inspect its size and languages. Confirm the displayed download size before installing an additional model. Settings shows whether the selected model is loaded in memory or will load on the next dictation. Apparat verifies each file in `<runtime-root>/data/whisper/` and saves the choice in SQLite after the decoder loads. Installed models work offline.

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

The planned Phase 10 setup creates one durable Ed25519 node identity per runtime root. New GUI nodes default to passphrase encryption and can opt out. New headless nodes default to owner-only unencrypted storage for unattended startup and can opt into encryption. Existing nodes keep their key and protection mode across GUI and daemon launches. Optional OS unlock lets a user unlock an encrypted node with a fingerprint on supported Android devices. Anyone who obtains an unencrypted private key can impersonate the node. See [node identity and platform unlock](docs/node-identity.md).

App identity remains separate from WireGuard identity.

The MVP uses one cluster-local X.509 root CA whose certificate fingerprint is part of the verified cluster identity. The currently authorized enrollment authority controls issuance under that root for the MVP; adding multiple concurrent issuers requires a later explicit hierarchy and conflict-resolution design.

Each device generates a TLS leaf key separately from its Apparat Ed25519 device-signing key. Enrollment binds the leaf public key, certificate serial and fingerprint, Apparat device signing key, WireGuard public key, roles, scopes, validity, and cluster identity in a signed device record. The device signs or proves possession for its certificate request, and peers require both successful mTLS validation to the cluster root and a current authorized device-record binding. Rotation issues a new serial and binding before the old certificate is retired; revocation and lost-device recovery invalidate the device record and certificate authorization.

Device authorization binds:

- The Apparat device identity.
- The HTTPS certificate fingerprint.
- The WireGuard public key.
- User or cluster authorization.
- Roles, permissions, and capabilities.

Enrollment is out-of-band through a short-lived QR code or invite containing the cluster fingerprint and a one-time token.

The connection layer uses TLS 1.3 mutual device authentication. Mutating requests do not use TLS 0-RTT. Certificate issuance, expiration, revocation, rotation, lost-device handling, and trust-store updates are first-class lifecycle operations.

### REST Resources

**Planned hosting default.** A fresh GUI installation, including Windows and Android, can connect to other nodes as a client with its own API server off. To host, open **Settings → API configuration**, select at least one specific network and interface pair, and explicitly enable the server. Every pair starts unchecked, including loopback. Linux headless service installations configure the same permissions locally through the CLI without a modal. Starting the service alone opens zero API listeners. For example, approving Tailscale on its tunnel permits hosting there; approving home Wi-Fi requires a separate rule. Moving the same Wi-Fi adapter to another network closes its listener. These rules cover the public key, diagnostics and future authenticated services such as LLM inference. See [API hosting policy](docs/node-identity.md#api-hosting-policy).

Today's loopback diagnostic API exposes safe read-only queries. Phase 10 first adds local authentication. Anonymous callers receive a JSON authentication message and can read the node public key. An owner-issued local token grants access to protected queries. Later, enrolled LAN and WireGuard peers use mutual TLS and scoped device authorization. See the [route and error contract](docs/api.md#planned-phase-10-local-authentication-contract).

The initial API surface is:

```text
GET  /v1/health
GET  /v1/device
GET  /v1/capabilities
GET  /v1/services
GET  /v1/services/{service_id}
GET  /v1/services/{service_id}/capabilities
GET  /v1/projects
GET  /v1/projects/{id}
GET  /v1/projects/{id}/tasks
POST /v1/projects/{id}/tasks/{task_id}/runs
POST /v1/jobs
GET  /v1/jobs/{id}
POST /v1/jobs/{id}/cancel
POST /v1/queues/{queue_id}/jobs
POST /v1/queues/{queue_id}/claims
POST /v1/queues/{queue_id}/leases/{lease_id}/heartbeat
POST /v1/queues/{queue_id}/leases/{lease_id}/complete
GET  /v1/events?after={cursor}&wait={duration}
POST /v1/project-transactions
POST /v1/artifacts
GET  /v1/artifacts/{artifact_id}
```

Mutating operations require an `Idempotency-Key`. Asynchronous job submission returns `202 Accepted`, a durable job ID, and a status resource location.

Cursor-based long polling comes before WebSockets. Requests enforce authentication, authorization, schema versions, content types, body limits, deadlines, and bounded concurrency.

`GET /v1/capabilities` returns the responding device's safe aggregate typed capability projection rather than one generic inference flag. The service resources expose authorization-filtered logical service and capability identities, safe health, inventory revision, availability, limits, and expiry without exposing provider-local endpoints or credentials. Jobs and routes refer to workload classes and optional `ServiceID`, `CapabilityID`, and model requirements by stable identifiers.

Each device's `GET /v1/projects` response is authoritative only for projects that device owns. A caller builds the cluster-wide Projects list from signed/cached project advertisements and owner responses; a non-owner does not re-publish a cached project as though it owns it. Project detail, Task discovery/execution, and project transactions are sent to the owning device.

Queue submission, worker claim/long-poll, lease heartbeat, and completion requests are sent to the queue owner. The owner is the only API authority that can admit, order, lease, cancel, or complete that queue's jobs. Worker claims include worker identity and current capabilities; completion includes the lease/fencing token and signed result or artifact references.

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

The MVP wire representation is UTF-8 JSON canonicalized with RFC 8785 JSON Canonicalization Scheme rules. Integer UTC millisecond timestamps are required. The payload hash is SHA-256 over the exact inline payload bytes or the canonical artifact metadata, and the sender signs the canonical envelope with the `signature` value omitted using its Apparat Ed25519 device key. Receivers validate the version, device-record key binding, signature, authorization, recipient, expiration, deadline, payload hash, size, replay state, idempotency, and schema compatibility before applying work.

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

### Governance Frameworks

| Path | Revision | Purpose and update policy |
| --- | --- | --- |
| `agentic-pipelines` | `51a17b0cff0bde9baf9473adacb1c0c11a50d4a7` | Selected shared agent-governance baseline and reusable pipeline tooling; update through a reviewed submodule synthesis |

The governance framework is not imported by the Apparat binary. Host build entrypoints invoke its canonical source checker before compilation; Apparat retains ownership of prerequisites, targets and artifact paths. Update its gitlink only after reviewing upstream instructions, host overrides, live path references, and local verification.

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
| `third_party/inference/pi` | `https://github.com/earendil-works/pi.git` | Multi-provider LLM API and agent-toolkit source reference; no runtime integration |
| `third_party/speech/whisper.cpp` | `https://github.com/ggml-org/whisper.cpp.git` | Portable local ASR reference |

Ebitengine uses an exact, validated stable release; the current selected baseline is v2.10.6. Upgrade when concrete fixes or capabilities justify it and Linux/Android regression evidence is recorded. A development revision requires an unmet need plus explicit compatibility evidence; the initial 2.9.x baseline is not a permanent minor-version restriction.

`debugui` remains a source-reference checkout; its update does not activate an application dependency.

Every grouping directory requires a README inventory. Every submodule addition requires an intentional revision, license review, purpose statement, update procedure, and declaration of whether it is a source reference or active build dependency.

The tracked [`third_party` inventory](./third_party/README.md) records the exact gitlink revisions, licenses, update procedure, and current build relationship for the admitted source set.

### Active Go Workspace

The root Go module is `github.com/cjtrowbridge/apparat`.

Phase 0 establishes:

- Go toolchain baseline: `1.26.11`.
- Ebitengine dependency: `github.com/hajimehoshi/ebiten/v2 v2.10.6`, replaced by the matching pinned source checkout.
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

`make tools` installs pinned developer tools into the ignored `.tools/bin` directory. `make verify` runs formatting, unit tests, build-pipeline tests, race tests, code-size checks, documentation-completeness checks, linting, and vulnerability scanning. On Linux it also runs the GUI navigation and action catalog regressions under Xvfb (`make test-hud-navigation`).

### Build Process

### Bootstrap and build governance

Before first setup, explicitly register the designated Apparat main checkout with
`python3 scripts/checkout_identity.py --register`. Build with `python3 scripts/build.py`.
On Linux this now opens the compiled GUI in Xvfb, exercises the headless executable,
and publishes the signed Android APK after automated package, ABI, signer and version checks. When the authorized Pixel 4a is available, it also installs and launches the APK and verifies the installed bytes and Projects screen. Otherwise the Android entry in `releases/verified.json` records device testing as pending.
Missing prerequisites or failed acceptance leave the verified outputs unchanged.
See scripts/README.md for dependencies and recovery. Stage the source intended for release before building. The pipeline commits staged changes and qualified outputs together, pushes main, and verifies GitHub downloads using the authenticated GitHub CLI. Apparat is distributed through its public GitHub repository. The Android updater reads the published version and downloads an approved update without a GitHub login. Linux updates pin verification metadata and artifact bytes to one GitHub commit. An unchanged rerun verifies delivery without incrementing the version.

Run `python3 scripts/bootstrap.py --check` to inspect Go and hook prerequisites without changes. `python3 scripts/bootstrap.py` installs/checks the repository hook. Provision Go 1.26.11 on PATH or copy a verified existing distribution with `--go-source /path/to/go`; `--native-source /path/to/sysroot` copies an existing native development sysroot into ignored local tools. These copies are independent of the original project. On Linux amd64, `python3 scripts/bootstrap.py --install-android` installs checksum-verified API 35, build-tools 35.0.0, NDK 27.2.12479018 and Temurin 21 into `.tools`; `--check --android` validates them without installing. Existing `JAVA_HOME` or PATH JDKs must be OpenJDK 21. ADB is the system Debian package, shared across projects.

Every supported software-build entrypoint runs the read-only bootstrap probe, canonical source gate and physical-line check before compilation. `source-governance.yaml` declares first-party roots and language rules; local API configuration is unrelated. Build prerequisites fail visibly without automatic system installation. Use `python3 scripts/build_gate.py -- COMMAND` for direct compiler/test/tool commands and approved Git commits so the staged hook inherits the local Go path. Make targets already do this.

The VS Code task and primary launch action use Python on Linux/macOS and the existing native PowerShell wrapper on Windows, entering the same canonical build once. Windows/macOS integration is configured but requires independent host validation. `scripts/build_environment_sample.py` owns optional local environment defaults; the root sample preserves the established copy-to-`build_environment.py` interface.


After cloning, initialize Agentic Pipelines:

```bash
git submodule update --init --recursive agentic-pipelines third_party/stmpck
```

`agentic-pipelines/AGENTS.md` is the selected shared governance baseline. Apparat's build targets are the deterministic pipeline entities, and the existing no-flag `scripts/build.py` process remains the sole build engine. The current host-owned VS Code task and play action are Windows-local and enter `scripts/run_build_pipeline.ps1` before delegating to that build script. Linux and macOS retain their documented direct build commands until native VS Code entrypoints can be tested there.

Run build and verification commands from the repository root. Prefer Makefile targets because they apply repo-local Go cache settings and keep generated files out of source directories.

Prerequisites:

- Go `1.26.11` on `PATH`, or use the checked local toolchain path when present.
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

To create a separate prompt-friendly research digest, run `python3 scripts/generate_gitingest.py`. It pins GitIngest 0.3.1 in the ignored `.tools/gitingest-venv`, stages its temporary digest under ignored `.tmp/` before recording provenance, writes the project-focused ignored `gitingest.txt`, and writes its ignored provenance sidecar `gitingest.manifest.json`. The default excludes recursive third-party source, release artifacts, and tool/cache directories while preserving Apparat-owned code, tests, documentation, plans, and build files. Use `--include-submodules` only when a review requires full third-party source bodies. This is intentionally separate from the build pipeline; its first run needs network access.

`make build` runs the Python build pipeline. The pipeline detects the current machine, reports possible and impossible targets with reasons, then builds every possible target. On a Linux host with desktop build prerequisites, it writes both latest local binaries to:

```text
releases/<goos>/<goarch>/apparat/latest[.exe]
releases/<goos>/<goarch>/apparatd/latest[.exe]
```

The path uses Go `GOOS` and `GOARCH` names such as `linux/amd64`, `linux/arm64`, `windows/amd64`, or `darwin/arm64`. Windows builds use `latest.exe`; other targets use `latest`. `apparat` is the GUI binary and `apparatd` is the headless worker/service binary. Latest artifacts under `releases/` are tracked in Git so other devices can pull the current known-good GUI, headless, and Android APK builds directly. Contributors should rebuild them intentionally when build inputs change.

Android Phase 5 builds the GUI APK only. The canonical Android artifact is:

```text
releases/android/arm64/apparat/latest.apk
```

The Linux update card shows bytes received and download percentage when available while staging the new executable. On Android, opening Settings checks the published version manifest. Choosing **Download update** starts the package transfer. Settings shows the connection, downloaded bytes and percentage when GitHub provides a total, file comparison, signature and version checks, and Android installer handoff.

The same no-flag build pass reports whether Android is possible and builds the APK when the prerequisites are present. The Android pipeline requires an OpenJDK 21 distribution (Eclipse Temurin preferred; Oracle JDK is not permitted), Android SDK platform `android-35`, build-tools `35.0.0`, NDK `27.2.12479018`, and Ebitengine's `github.com/ebitengine/gomobile` module. `JAVA_HOME` must expose `java`, `javac`, and `keytool`. Tools are discovered from `JAVA_HOME`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, and `ANDROID_NDK_HOME`, with ignored repo-local `.tools/` fallbacks when present; `.tmp/` holds disposable patched-tool and exported engine source. If a machine needs custom local paths, copy `scripts/build_environment_sample.py` to ignored `build_environment.py` and update environment values there.

On Windows, `python -m scripts.build` is the directly usable no-flag invocation when `make` is unavailable. It supports Eclipse Temurin OpenJDK 21 and the ignored repo-local `.tools/android` SDK layout. Provision Android platform 35, build-tools 35.0.0, NDK 27.2.12479018, system-PATH ADB, and Go/Gomobile caches before the build; these large local dependencies and the retained certificate-pinned development keystore must remain untracked. A successful APK build is artifact evidence only—use a connected authorized device and ADB install/launch evidence before claiming Windows Android runtime support.

The pipeline binds `cmd/apparatmobile` with Ebitengine's generated Android mobile view classes, compiles the tracked wrapper in `android/apparat`, and writes a signed APK to the canonical release path. The wrapper is the current Android render path: on-device evidence shows it opens to the Phase 4 HUD and accepts touch tab selection. It generates an ignored patched local `gomobile-apparat` helper under `.tools/bin` because the pinned Ebitengine `gomobile` package scanner checks for `github.com/ebitengine/gomobile/app` while its regular expression only recognizes `golang.org/x` package symbols. This patch broadens the scanner, supports local Apparat/Ebitengine module replacement for AAR binding, and preserves modern Android SDK metadata (`minSdkVersion=23`, `targetSdkVersion=35`) while compiling/package-building against Android platform 35. Mobile preparation exports the pinned engine into ignored `.tmp/mobile-engine` and guards invalid display metrics there, leaving the submodule untouched. The helper is rebuilt from its pinned source and current patches on each preparation. It also signs the APK with a retained certificate-pinned development keystore and keeps the Android build independent from the ignored `third_party/salvagecore` checkout.

Android Settings checks the published version when opened, including when it is the startup tab. It reuses a recent check for one minute. A newer version enables **Download update**. Private Git delivery requires a host update when the release metadata is inaccessible. Linux Settings provides a single **Check for update** action. Windows and macOS retain manual update instructions.

The APK retains Android 6/API 23 as its minimum; targeting API 35 does not raise that minimum. Newer Android versions apply newer permission, background-execution and layout rules. The wrapper applies system-bar and camera-cutout insets on Android 15+ so HUD controls stay inside usable screen bounds. Reserved areas use the renderer's background color, with light navigation buttons; the current renderer has a fixed dark palette. Newer and older-device operation require separate validation; Pixel 4a Android 13 evidence alone cannot establish either. The current APK is arm64 only.

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

For a persistent Steam Deck installation, copy the qualified Linux amd64 GUI to
`/home/deck/.local/bin/apparat` and add it to Steam with the installed SteamOS
helper. Open **Library → Non-Steam → Apparat** in Gaming Mode. The
[Steam Deck installation playbook](playbooks/how_to_install_apparat_on_steam_deck_and_add_it_to_steam.md)
covers SSH setup, verified transfers, desktop registration, updates and rollback.
The binary, launcher and node state remain beneath the user's home directory;
the SteamOS system partition stays read-only.

On Linux, save your work and select **Settings → Updates → Check for update** once.
Apparat checks the public GitHub development channel. When an update is available,
it downloads, verifies, installs and restarts automatically. Progress and errors appear
below the button. An already-current build stays open and the button reads **Up to date**. The download must match the
published SHA-256 and native ELF target before atomic replacement; the previous
executable is retained at `<executable>.previous`. The executable directory must
be writable. Steam keeps the same launcher and settings/projects remain in the
node directory. Downloads cancel when Apparat closes. This development channel
uses HTTPS and the existing verification report; release signatures, downgrade
prevention and automatic database rollback remain Phase 17 work. Restore a
previous binary only with compatible node state or a verified offline backup.

In Steam Gaming Mode, Apparat opens fullscreen at the compositor's current display
size rather than stretching a saved desktop rectangle. Desktop sessions retain
normal resizing and saved placement. The embedded STMPCK gallery centers its
settings gear immediately after discovery. PTT outside an editable field accepts voice navigation commands.

Phase 8 runs one embedded node behind either GUI or headless presentation. Both default to `~/.local/share/apparat/node` on Linux (respecting `XDG_DATA_HOME`); `APPARAT_RUNTIME_DIR` or `--runtime-dir` selects an independent root. An OS lock rejects competing ownership before writable SQLite, probes or REST start. PID metadata is diagnostic; never delete the lock file to defeat an active owner.

Stop older Apparat processes before first startup. Exactly one populated legacy `apparat` or `apparatd` root is renamed atomically to `node`; two populated roots, an existing destination, or a failed rename stop with recovery instructions. Nothing is merged. Choose a root explicitly to inspect a conflict, back up both roots, and make a deliberate offline selection. An interrupted atomic rename leaves the selected root at either its old or new location; inspect both before retrying. Older binaries do not honor the new lock and must remain stopped.

```bash
make build
releases/linux/amd64/apparatd/latest --runtime-dir /tmp/apparat-demo --demo-services --services
releases/linux/amd64/apparat/latest --runtime-dir /tmp/apparat-demo
# After stopping the GUI, run the same node headlessly:
releases/linux/amd64/apparatd/latest --runtime-dir /tmp/apparat-demo
```

`--demo-services` atomically seeds an empty store with two mock text instances and one mock image instance. Android seeds this demonstration on the first empty-store startup. Existing services are preserved, including deliberate removals. The Cluster tab's **Local services** detail shows live IDs, revisions and health; the adjacent discovery view shows live DNS-SD candidates. Probes run sequentially every five seconds, with one outstanding probe and a one-second cooperative deadline. A 15-second-old observation is stale. Provider routing and Task execution remain future work.

The default opens no REST listener. Linux Settings → API configuration lets the owner select LAN hosting grants. After first startup, Linux headless operators use `apparatd api-list`, `apparatd api-grant RULE` and `apparatd api-disable` against the same node root. The selected LAN serves a JSON 401 on GET /v1 and advertises its ephemeral port through DNS-SD. Cluster → Add New Device browses on the selected Linux interface or the active Android 13+ Wi-Fi network. Legacy listen arguments are rejected. The six Phase 8 diagnostic routes remain internal. The bootstrap listener exposes only GET /v1. See [API operations](docs/api.md). GUI queries call the core directly, never this listener.

Linux headless workers enable automatic self-updates by default. With API traffic, the first completed response when a check is due (including an anonymous 401) triggers a background check. A successful check is due again in 24 hours. A failed or interrupted attempt may retry after one hour. A quiet worker waits for a request, even when overdue. SQLite preserves the setting and schedule across restarts. Assigned or running work defers installation. The worker verifies its own Linux release, preserves the previous executable and restarts with the same node root and network grants.

Run `apparatd updates-status` to see the policy, last attempt, result and next due time. Use `apparatd updates-disable` or `apparatd updates-enable` to change the setting. `apparatd updates-check` queues an immediate owner-requested check and install, including when automatic checks are disabled. For a custom root, set `APPARAT_RUNTIME_DIR` to the service's root before running these commands. Pool-aware draining and update flags are planned.

`--services` prints safe JSON and exits; `--smoke-test` adds a diagnostic line with root/last-run paths. `--doctor` checks SQLite integrity and local identity status. `--backup /new/path.db` creates a consistent SQLite snapshot plus metadata. `--restore-backup /path.db --runtime-dir /new/root` verifies and restores into a previously absent database and exits. Neither operation overwrites existing backup/destination files. Identity keys and other runtime files are separate: preserve them in an offline root backup when planning full-node recovery. See [recovery procedures](docs/database.md).

`last_run.log` is reset on each owning process start; append-only structured history remains in `logs/`. Startup failures unwind initialized resources; shutdown cancels and joins probes, drains REST, closes SQLite and releases ownership. No GUI dependency is loaded by `apparatd`. Source files retain the 400-physical-line gate.

Fast build/test tasks enter the same source gate and keep outputs under ignored `.tmp`; normal `scripts/build.py` remains the release orchestrator. Android eligibility inspection does not compile tools. Actual Android execution prepares the exact engine export and helper tools once before binding. Go 1.26.11 is required; `python3 scripts/bootstrap.py --install-go` installs a checksum-verified Linux amd64 copy locally.

Android GitHub updates require the retained signing key pinned by `android/apparat/signing-certificate.sha256` and an increasing manifest `versionCode`. The build fails if the key is missing or different. The canonical pipeline assigns the next version and checks the candidate against the signed APK on GitHub main. Keep the key private and securely backed up; do not regenerate it on another build machine. The old GitHub signer is a separate lineage; this channel now follows the currently installed Pixel builds.

Android APK downloads require an explicit download action. Downloads are bounded, cancel with Activity destruction and must match the installed package/signature with a strictly newer version before the system installer is offered. Public hash equality does not establish release authenticity. Production signed manifests, key custody and rollback remain Phase 17; no install permission is granted automatically.

Android update downloads allow 60 seconds without incoming data and 15 minutes of total transfer time, capped at 128 MiB. Timeout, cancellation and oversized-download messages are distinct from package/signature rejection. A failed download or local save does not establish that GitHub is unreachable. Builds older than 0.1.2 retain the earlier restrictive limits until upgraded.

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

## Root source entrypoints

- `build.py`: Compatibility dispatcher to the canonical build script.
- `build_environment.sample.py`: Compatibility sample delegating local environment defaults to scripts/build_environment_sample.py.

Fresh installations open Projects. Responsive placement puts the tabs at the bottom in portrait and uses a left rail in landscape when space permits. Saved startup and placement choices take precedence. Body scrolling follows the drag distance and stops at each content edge.

Settings stores theme, UI scale, font size, startup tab and developer overlay preference in the node database. Select a setting to cycle its available values. Desktop also remembers the normal window rectangle and maximized state. Reset window placement centers the window on the current monitor.

Tab placement in Settings cycles through Top, Bottom, Left, Right and Responsive. Responsive uses a left rail when there is room for a desktop detail pane. Short windows use the horizontal strip so every tab stays reachable. SQLite remembers the selected mode across resizes and restarts.

### Using the local project workspace

Tap a registered checkout to open Files. The folder is usable while Git checks the displayed directory in the background. The 🔙 button goes up one level or returns to Projects from the root. Select a UTF-8 file up to 64 KiB to edit it in the Android document editor or the Linux GTK 4 editor (Python 3 with PyGObject and GTK 4). Changes are kept as local drafts. Save writes the reviewed file and returns to its containing directory with refreshed Git status. Tap a folder to enter it or `../` to go up. Clean filenames are white. Red names use `U` for new or untracked files, `C` for changed files and `D` for deletions. Staged changes stay marked until committed. A conflicting save preserves the draft and offers a comparison. The file options also include a fallback line editor.

Changes lists staged and unstaged files. Select a path to review its diff, stage the intended file, and commit it with the local identity `Apparat <apparat@localhost>`. Tasks shows manual entrypoints and run history. Chat groups pending local messages by conversation.

For a local Task, place a versioned declaration in `.apparat/tasks.json`. For example:

```json
{"version":1,"tasks":[{"id":"readme","label":"Snapshot README","enabled":true,"action":"snapshot-text","path":"README.md"}]}
```

Select **Tasks**, open the named entrypoint, review its input and choose **Run**. The snapshot copies the named text file into local artifact storage and records its SHA-256 digest. Open a run's artifact to verify the stored bytes after restart. VS Code shell commands remain descriptive labels. The built-in `git-status` action captures the current checkout status as a text artifact. Tasks run with a 15-second default deadline, a maximum 60-second configured deadline and explicit confirmation. Unsupported actions and secret-dependent definitions cannot run on this executor. Artifacts are retained until explicit deletion or an optional `retainSeconds` deadline. Pending chat messages are grouped by project thread and can reference the last opened run and artifact.
