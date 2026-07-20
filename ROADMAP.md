# Apparat Roadmap

This roadmap translates the product contract in [README.md](./README.md) into ordered implementation phases.

Roadmap items describe goals, dependencies, and completion evidence. They are not atomic execution authority. Before changing application code, dependencies, schemas, protocols, or build systems, create and approve a focused execution plan under `plans/future/` or `plans/current/`.

## Status Key

- `[ ]` not started
- `[x]` completed and verified
- `[?]` implemented but still needs validation
- `[-]` intentionally closed or deferred

## Product Baseline

The roadmap assumes these decisions:

- **Product identity**
  - Apparat is a personal-area-network cluster console, not an RPG.
  - The application uses a game engine for a portable controller-first HUD.
  - Gamification may be added later where it supports validation, participation, and progress.
- **Platform and runtime targets**
  - Steam Deck is the first GUI target.
    - Controller-first interaction is a core acceptance requirement for the first visual shell.
    - `L1` and `R1` switch top-level tabs.
    - `R2` is hold-to-talk; releasing it submits captured audio to ASR.
  - Debian/Linux GUI mode is a first-class desktop target.
    - `Ctrl+PageUp` and `Ctrl+PageDown` switch top-level tabs.
    - `Alt+1` through `Alt+5` open the five canonical tabs directly; Routing and Tasks are selected within Cluster.
    - `Tab`, `Shift+Tab`, arrow keys, `Enter`, `Space`, `Escape`, Menu or `Shift+F10`, and `Ctrl+Shift+P` provide focus, activation, cancellation, contextual-action, and command-palette controls.
    - Holding right `Ctrl` starts push-to-talk; releasing it submits audio, while `Escape` cancels the held recording.
    - Mouse and touchpad input support ordinary desktop activation, context menus, scrolling, and explicit drag operations without making any essential workflow pointer-only.
    - A connected controller uses the Steam Deck action mapping.
  - Debian/Linux headless devices are first-class workers and service hosts.
    - Headless control uses documented CLI commands, authenticated HTTPS REST, service-manager operations, health checks, and graceful `SIGINT`/`SIGTERM` handling.
    - No interactive terminal UI is required for the MVP.
  - Windows, macOS, and Android follow after the Linux vertical slice.
- **HUD information architecture**
  - The canonical tab order is:
    1. Comrades
    2. Projects
    3. Cluster
    4. Research
    5. Settings
  - Cluster uses a selector panel and content panel for device and operations context. Routing and Tasks are Cluster selector items, not top-level tabs.
  - Comrades is visible first but remains a navigable placeholder during the MVP.
    - It eventually supports real-friend chat.
    - It eventually supports revocable low-priority inference sharing through owner-controlled comrade queues.
  - Research follows Cluster and remains a navigable placeholder during the MVP.
    - It eventually delegates explicitly budgeted compute to validated BOINC projects.
    - Research-project validation eventually participates in gameplay mechanics.
  - Settings remains the final tab.
- **Data and authority**
  - Every Project is an ordinary Git repository owned by the device where its working tree lives and runs.
  - Every device presents one cluster-wide authorized Projects list assembled from owner-local Projects and signed/cached summaries from every other device.
  - A Pipeline is a Project with one or more Apparat Task entrypoints; it is not a separately owned entity.
  - A Task may run manually with no trigger or be bound to intervals, authenticated webhooks, internal application events, or cluster events.
  - SQLite stores identities, metadata, chats, transactions, events, queue state, indexes, and durable workflow state.
  - Projects and queues have one authoritative owner device during the MVP; Project Tasks and their run records are owned with the Project.
  - All cross-device Project, Task, queue-submission, worker-claim, heartbeat, and result operations use authenticated HTTPS REST directed to the owning device.
  - Queue owners validate and admit requests; authorized inference workers pull leased tasks and post signed results back; only the owner records authoritative completion.
  - Cross-device delivery is at-least-once.
    - Stable IDs and idempotent application provide duplicate safety.
- **Connection and trust**
  - The primary online connection is authenticated HTTPS REST over:
    - Externally configured WireGuard.
    - A trusted local network.
  - HTTPS authentication remains mandatory on LAN.
  - TLS 1.3 mutual device authentication, signed envelopes, and explicit authorization protect application operations.
  - WireGuard identity and Apparat identity remain separate but cryptographically bound.
  - HTTPS/WireGuard is the first full-capability transport.
  - Signal, Meshtastic, and other transports are future adapters carrying the same durable signed messages.
- **Typed compute routing**
  - Compute routing uses explicit workload classes rather than one generic inference flag.
  - Initial workload classes are:
    - Text generation.
    - Image generation.
    - Video generation.
    - Speech-to-text.
    - Text-to-speech.
    - BOINC research compute.
  - BOINC is schedulable research compute rather than model inference.
    - It still uses the same capability, policy, queue, and resource-budget framework.
  - Devices may advertise several independent workload capabilities.
  - Jobs and routes are eligible only for queues and devices that satisfy their declared workload class and capability requirements.
  - Adapter order begins with:
    1. OpenAI-compatible text generation.
    2. Ollama.
    3. llama.cpp.
  - whisper.cpp is the first local ASR reference.
  - OS-native or lightweight service-backed TTS precedes Qwen3-TTS.
- **Explicitly deferred beyond the MVP**
  - App-managed WireGuard.
  - CRDT editing.
  - Scheduler failover.
  - Dynamic workload routing optimization.
  - Unrestricted remote execution.
- **Documentation completeness**
  - Every new file or feature must be documented at the closest useful documentation layer.
  - New or changed code, script, tool, test, and build directories require local `README.md` coverage.
  - New or changed scripts require useful `--help` output and `scripts/README.md` inventory coverage.
  - Build and runtime behavior that normal users or contributors run, configure, observe, or troubleshoot must be surfaced in the root `README.md`.
  - `make verify` includes automated documentation-completeness checks where the requirement can be mechanically validated.

## Salvagecore Reference Baseline

The ignored local checkout at `third_party/salvagecore` is an older implementation of the same broad personal-cluster idea. This section is the durable inheritance record for that temporary material. It explains what existed, what was only proposed, why selected concepts matter to Apparat, and what must change for the HUD-first MVP. Apparat must remain understandable and implementable after the local checkout is deleted.

- **Reference status and handling**
  - **Repository relationship**
    - Salvagecore is a predecessor and design reference, not an upstream dependency or a component of the new application.
    - The local checkout is intentionally ignored by Git and absent from `.gitmodules`.
    - No Apparat build, test, generator, documentation link, or runtime path may require the checkout to exist.
    - The checkout must never be committed accidentally, including its nested submodules, generated files, release artifacts, or copied dependency trees.
  - **Permitted use**
    - Inspect its architecture, implementation details, tests, design documents, dependency choices, and failure modes.
    - Extract contracts, invariants, and narrowly reusable implementation patterns.
    - Compare a proposed Apparat design against a concrete earlier attempt at the same problem.
    - Use its source paths as temporary provenance while deciding whether a particular implementation deserves a focused port.
  - **Prohibited assumptions**
    - Do not treat code presence as evidence that a feature was complete, integrated, or production-ready.
    - Do not treat a roadmap statement or database-design document as evidence that the corresponding runtime behavior existed.
    - Do not copy the repository wholesale or preserve old package boundaries merely because they already exist.
    - Do not inherit qTox, Tor, RPG, desktop-input, dependency-version, or platform-support decisions without new evidence.
    - Do not modify the local predecessor to make it resemble this project; all new product work belongs in Apparat.
  - **Reuse authorization**
    - Documentation may retain conclusions directly because this section records their meaning and limits.
    - Copying or adapting source code requires a focused implementation plan that identifies:
      - The exact predecessor files being studied.
      - The behavior and invariant being retained.
      - The assumptions being removed or renamed.
      - The tests that prove the adapted behavior in Apparat.
      - The license and provenance of copied code and any transitive source.

- **What the predecessor actually demonstrated**
  - **Implemented and inspectable infrastructure**
    - A Go application of roughly nine thousand lines with an Ebitengine and EbitenUI desktop shell.
    - A shared runtime constructor that initialized configuration, runtime directories, append-only logging, SQLite, identity state, modules, and the application store before choosing a GUI or headless adapter.
    - Explicit `auto`, `gui`, and `headless` runtime modes.
      - GUI mode constructed Ebitengine and rendered the shell.
      - Headless mode avoided constructing Ebitengine, ran diagnostics, required identity setup, and then waited for shutdown.
      - Auto mode selected the GUI only when the platform check considered it available.
    - A feature-module interface with navigation metadata, declared command kinds, reducer names, list projections, and detail projections.
    - A typed command dispatcher and store that returned copied application state to the renderer.
    - A resizable three-pane mock HUD with section navigation, list selection, detail rendering, Settings routes, a developer overlay, and a read-only database-inspection surface.
    - Local SQLite lifecycle code with forward migration tracking, schema checksums, repositories, device snapshots, identity persistence, and read-only query restrictions.
    - User and device identity setup using:
      - Ed25519 key pairs and signatures.
      - SHA-256 public-key fingerprints.
      - App-native signed canonical JSON certificates.
      - Argon2id passphrase derivation.
      - XChaCha20-Poly1305 encrypted private-key files.
      - Public identity manifests and consistency checks between files and SQLite.
      - CLI and GUI flows for create, import, status, doctor, repair, recovery, and archived reset.
    - Append-only JSON Lines logging with severity levels, component and event names, command-transition records, and basic sensitive-field redaction.
    - Tests around database startup and migration behavior, database inspection, encrypted keys, signed certificates, identity manifests, identity recovery, and first-run identity UI.
  - **Scaffolding rather than completed product behavior**
    - The Comrades, Projects, Devices, Services, Models, Queues, Tasks, and Settings packages were mostly static modules projecting mock snapshot data.
    - The command model handled shell selection, layout resizing, Settings navigation, and database-inspector state; it did not implement the full durable command and event system envisioned by its design.
    - The store was primarily an in-memory mock store and did not yet coordinate real project, queue, service, or transport effects.
    - The headless runtime proved that GUI initialization could be avoided, but it did not yet host real queues, schedulers, inference workers, synchronization, or an API server.
    - The three-pane shell demonstrated reusable layout and state boundaries, but not controller navigation, accessibility, touch interaction, or the final Apparat tab model.
    - Identity and recovery were substantially deeper than the networking and compute features that would have used them.
  - **Documented architecture that was not yet proven end to end**
    - Device-owned projects and queues.
    - Durable inbound and outbound transport records.
    - Change feeds, sequence cursors, tombstones, and synchronization retries.
    - Project transactions submitted to an owner and retained as editable drafts after conflicts.
    - Shared inference pools, queue permissions, comrade resource sharing, and authoritative remote job results.
    - qTox and Tor transport adapters and transport-key binding.
    - Real model execution, project workspace operations, chat delivery, scheduled tasks, event-driven automation, and multi-device orchestration.
  - **Platform and dependency limits discovered by inspection**
    - The build helper only enabled the current host target.
      - Windows, additional Linux architectures, macOS, and Android appeared in a target list but were explicitly marked as planned and rejected by the script.
      - Therefore the predecessor is evidence of an intended cross-platform architecture, not evidence that those target builds worked.
    - The UI was mouse-first.
      - Pane resizing and row selection used pointer hit testing.
      - Keyboard support was limited and did not provide the controller focus graph required by Apparat.
      - No implemented Steam Deck mapping established `L1`, `R1`, `R2`, directional navigation, or controller-driven form editing.
    - The Go module pinned a pre-release Ebitengine 2.10 revision and local source replacements.
      - This demonstrates how local engine references can be wired.
      - It does not justify carrying an alpha engine into the new MVP.
    - Python OpenAI Whisper appeared as a speech direction.
      - That approach does not satisfy the desired portable Go application and service packaging boundary as cleanly as whisper.cpp or an explicitly managed speech service.

- **Canonical architecture inherited and adapted for Apparat**
  - **Ports-and-adapters separation**
    - **Meaning**
      - Product rules and durable state transitions must not depend directly on Ebitengine, SQLite, HTTP, WireGuard, llama.cpp, whisper.cpp, BOINC, Signal, Meshtastic, or any other external system.
      - External systems are reached through narrow interfaces implemented by adapters.
      - Replacing a transport, database implementation, inference provider, or visual toolkit should not require rewriting queue policy or project rules.
    - **How it maps to Apparat**
      - The Ebitengine HUD is an input and presentation adapter.
      - HTTPS client and server code are transport adapters for the canonical signed application messages.
      - SQLite repositories are persistence adapters.
      - OpenAI-compatible endpoints, Ollama, llama.cpp, whisper.cpp, TTS services, and BOINC are workload adapters.
      - WireGuard is an externally managed network substrate, not the application protocol or identity model.
      - Signal and Meshtastic will be constrained message adapters over the same durable application semantics.
    - **Required boundary**
      - UI packages may render view models and dispatch commands.
      - UI packages must not issue SQL, mutate durable state directly, perform blocking network calls, or contain queue-authorization policy.
      - Domain and application packages must be testable without initializing a window, GPU, network interface, model runtime, or real database.
      - Adapter failures return typed outcomes that the application layer can retry, reject, display, or persist.
  - **Command, effect, event, reducer, and snapshot flow**
    - **Meaning**
      - User intent enters the application as a typed command rather than an arbitrary UI callback mutating shared state.
      - The application validates the command against identity, authorization, current state, workload capabilities, and ownership.
      - Pure state changes can emit events immediately.
      - External work is represented as an effect executed by an adapter.
      - The adapter returns typed progress, success, cancellation, timeout, or failure events.
      - Reducers apply events to produce a new immutable or copy-safe snapshot consumed by the UI.
    - **Canonical flow**
      1. A controller, keyboard, pointer, API request, scheduler, or transport adapter creates a typed command.
      2. The application assigns or propagates a stable command ID and correlation ID.
      3. Command handling validates preconditions and determines whether the operation is local, durable, remote, or denied.
      4. Durable intent is committed before an external side effect when retry safety requires it.
      5. An adapter performs the external operation outside the render/update path.
      6. Progress and terminal outcomes return as typed events.
      7. Reducers update durable records and the current read snapshot.
      8. The HUD redraws from the new snapshot and never from adapter-owned mutable state.
    - **Adaptation from the predecessor**
      - Preserve typed command kinds and copied snapshots.
      - Replace the predecessor's single broad command struct as the system grows.
        - Use command-specific payload types or interfaces so invalid field combinations cannot be constructed casually.
        - Keep serialization versions explicit for commands that cross devices or survive restarts.
      - Separate commands that express intent from events that record facts.
      - Do not use a reducer name string as the final registration contract; register concrete handlers with compile-time types where practical.
  - **Feature modules**
    - **Meaning**
      - A feature should join the application through a narrow registration surface rather than editing a central switch in every layer.
      - Modules own feature behavior while sharing application-wide identity, authorization, persistence, event, and rendering conventions.
    - **Apparat module responsibilities**
      - Declare stable module and route identifiers.
      - Contribute top-level or nested navigation metadata without changing the canonical tab order.
      - Register supported commands, event handlers, reducers, effects, repositories, and background services.
      - Produce view models rather than exposing database rows to the HUD.
      - Declare required permissions, transport capabilities, workload classes, and device capabilities.
      - Expose diagnostics and health state for the Cluster and Settings tabs.
    - **Module boundaries for the canonical tabs**
      - **Comrades**
        - Own trusted-person relationships, conversation projections, sharing grants, and comrade-queue policy.
        - Must not own transport-specific Signal or qTox identity as the canonical friend identity.
      - **Projects**
        - Own project metadata, repository/worktree references, chats, artifacts, workflow state, and owner-authoritative project transactions.
        - Keep ordinary project files in the filesystem and Git rather than embedding source trees in SQLite.
      - **Research**
        - Own BOINC project catalog state, validation status, delegation budgets, contribution accounting, and later gameplay validation.
        - Use the common workload and resource-policy framework while remaining semantically distinct from model inference.
      - **Cluster**
        - Own device inventory projections, reachability, service health, capability advertisements, utilization, and diagnostics.
      - **Routing**
        - Own workload-typed queues, route eligibility, priority, budgets, provider preferences, and device assignments.
      - **Tasks**
        - Own durable scheduled, webhook, signal-driven, event-driven, and manually submitted task definitions and run history.
      - **Settings**
        - Own local configuration, identity management, trust administration, diagnostics, developer tools, and platform options.
  - **Shared GUI and headless runtime**
    - **Meaning**
      - GUI and service operation are two adapters around the same application runtime, not separate products with diverging state models.
      - A headless device must be capable of acting as an authoritative project, queue, task, or service owner.
    - **Shared initialization**
      - Resolve runtime paths and configuration.
      - Initialize structured logging and diagnostics.
      - Open SQLite and apply verified migrations.
      - Load identity and trust state.
      - Register modules, repositories, services, and effect handlers.
      - Recover durable pending work before accepting new commands.
    - **Mode-specific behavior**
      - GUI mode initializes Ebitengine, input adapters, view models, and visual resources after the shared runtime is healthy.
      - Headless mode never imports through a path that initializes Ebitengine, creates a window, or assumes a display server.
      - Headless mode hosts the HTTPS API, queues, workers, schedulers, webhooks, synchronization, and health endpoints selected for that device.
      - Auto mode may choose a GUI only when availability is positively detected; otherwise it must fail clearly or select headless according to documented policy.
      - Doctor mode performs read-mostly checks and reports actionable remediation without starting normal work.
    - **Why it matters here**
      - Steam Deck can run the HUD and selected local workers.
      - Debian servers, desktops without an active display, and compute nodes can participate fully without carrying a visual lifecycle.
      - Windows, macOS, and Android can share product behavior while supplying platform-specific visual, audio, filesystem, and lifecycle adapters.

- **Persistence and synchronization inheritance**
  - **SQLite role**
    - **Store in SQLite**
      - Identity metadata, signed certificates, device records, trust and authorization grants.
      - Project metadata, chats, message metadata, artifact indexes, transaction records, and workflow state.
      - Workload capability advertisements, service instances, queue definitions, routes, jobs, attempts, results, and status projections.
      - Task definitions, schedules, triggers, runs, retries, and audit events.
      - Durable transport inboxes, outboxes, acknowledgements, delivery attempts, and synchronization cursors.
      - BOINC project metadata, validation state, resource budgets, work records, and contribution summaries.
    - **Keep outside SQLite**
      - Project source trees and ordinary working files.
      - Git objects and repositories.
      - Large model files, generated media, video, audio, and bulky task artifacts.
      - Decrypted private keys, passphrases, raw tokens, and transient model memory.
    - **External-file records**
      - SQLite stores stable IDs, paths or content-addressed references, media types, hashes, sizes, ownership, retention state, and authorization metadata.
      - Moving or deleting a file must become an explicit transaction so indexes and durable state cannot silently diverge.
  - **Database implementation contract**
    - Use a per-device SQLite database under the configured runtime directory.
    - Prefer `modernc.org/sqlite` initially to avoid cgo across desktop and Android targets.
      - Validate actual build and runtime behavior on each supported target.
      - Reconsider the driver only when measured compatibility or performance evidence requires it.
    - Hide SQL behind repositories consumed by application services and stores.
    - Use forward-only migrations from the first schema.
      - Record migration version, name, checksum, and application timestamp.
      - Test migration from an empty database and every supported prior schema.
      - Treat rollback as restore from a verified backup or rebuild of disposable development state.
    - Use ULID text identifiers for durable records unless a focused design establishes another externally stable identity.
    - Store timestamps as UTC integer milliseconds and preserve explicit timezone information only where the human schedule requires it.
    - Keep query-critical fields relational.
      - Versioned JSON is appropriate for extensible envelopes, snapshots, and provider-specific payloads.
      - JSON must not become an excuse to avoid indexes, constraints, or documented state transitions.
    - Enable foreign keys on every connection.
    - Evaluate WAL and busy-timeout behavior separately for desktop and mobile lifecycle, backup, and suspension constraints.
  - **Read models and inspection**
    - The HUD reads application snapshots and projections, not live mutable repository objects.
    - Feature repositories own writes; projection builders own display-oriented reads.
    - A developer database inspector may be retained under Settings.
      - It must open through a read-only path or enforce read-only statements.
      - It must reject mutation, attachment, unsafe pragmas, and multi-statement escape paths.
      - It is a diagnostic tool, not an application API.
      - Sensitive columns and payloads require redaction or explicit privileged access.
  - **Authority model**
    - **Core rule**
      - During the MVP, every Project and queue has exactly one authoritative owner device; Project Tasks are owned with their Project.
      - Other devices may cache authorized projections and submit transactions, but they do not silently become co-authoritative.
    - **Projects**
      - A Project is a Git repository owned by the device where its working tree lives and runs.
      - Every device builds a cluster-wide Projects projection from owner-local records plus signed/cached summaries from all authorized owners.
      - Remote reads and writes use authenticated REST to the owner; no remote device reads the owner's filesystem or SQLite directly.
      - The owner serializes accepted project transactions and publishes the resulting version.
      - A non-owner submits a transaction with its base version and stable transaction ID.
      - Accepted changes advance the authoritative project version.
      - Rejected or conflicting changes remain local editable drafts with a machine-readable reason and enough context for the user to revise or rebase them.
    - **Queues**
      - The queue owner stores authoritative ordering, admission, attempts, cancellation state, and results.
      - Requesters submit through authenticated REST and the owner validates authorization, schema, workload requirements, policy, limits, quota, and idempotency before admission.
      - Authorized inference workers poll or long-poll the owner for compatible leases, then post signed results back through REST.
      - The owner validates active lease/fencing identity and result integrity before authoritative completion.
      - Requesters retain their submission record and authorized status/result projections.
      - Requesters do not mirror the complete queue unless a later availability design explicitly requires it.
      - A direct queue targets one eligible inference device; that device still pulls an owner-issued lease through REST and may happen to be the owner.
      - A pool queue is coordinated by the owner across eligible member devices, each of which pulls compatible leased work.
    - **Tasks**
      - A Pipeline is a Project with at least one Apparat Task entrypoint; it is not a separate ownership object.
      - The Project owner defines and runs its Tasks and evaluates their optional schedules and durable trigger bindings.
      - A Task with no trigger remains manually executable through Apparat.
      - Offline replicas may retain definitions and observations but do not execute an owner-only trigger independently.
      - Scheduler failover is deferred until leases, fencing, clock behavior, and duplicate execution are designed explicitly.
  - **Durable delivery and synchronization**
    - Use an outbox for committed local intent awaiting delivery.
    - Use an inbox for received envelopes and their validation/application outcome.
    - Assign stable message, command, transaction, job, and correlation identifiers before first delivery.
    - Assume at-least-once delivery.
      - Retries may duplicate envelopes.
      - Application handlers must return the prior result or a semantically equivalent acknowledgement when the same idempotency key is replayed.
    - Maintain owner-scoped monotonically increasing change sequence numbers where ordered replay is required.
    - Maintain per-peer or per-scope cursors recording the last durably applied sequence.
    - Represent deletion with tombstones retained long enough for authorized replicas to observe it.
    - Detect gaps and request bounded replay instead of guessing that missing state never existed.
    - Keep transport delivery state separate from domain application state.
      - Delivered does not mean authorized or applied.
      - Applied does not necessarily mean the sender received the acknowledgement.
    - Compact feeds only after retention, acknowledgement, backup, and offline-device policy permit it.

- **Identity, trust, and cryptographic inheritance**
  - **Identity layers**
    - **User identity**
      - Represents the person or administrative principal who owns a personal cluster.
      - Authorizes devices and grants access to comrades.
      - Must remain stable when a device, transport address, TLS certificate, or WireGuard key rotates.
    - **Device identity**
      - Represents one Apparat installation or managed physical/virtual device.
      - Is authorized by a signed user-identity statement.
      - Signs application envelopes and binds current transport credentials.
    - **Transport identity**
      - Represents credentials needed by one carrier or network substrate.
      - HTTPS may use a TLS certificate or public-key fingerprint.
      - WireGuard has its own public key and peer configuration.
      - Future Signal and Meshtastic adapters have their own account, node, or gateway identifiers.
      - Transport identity must never replace the canonical Apparat user or device identity.
  - **App-native authorization documents**
    - Use versioned canonical documents signed with Ed25519 for application identity and authorization facts.
    - Initial document families should cover:
      - User identity self-description.
      - Device authorization by a user.
      - Binding of a TLS, WireGuard, Signal, Meshtastic, or other transport credential to a device.
      - Comrade trust and resource-sharing grants.
      - Revocation and replacement of devices, bindings, and grants.
    - Signed fields include stable type and version, document ID, issuer, subject, issue time, optional expiry, payload, algorithm, and signature.
    - Verification must reject:
      - Unknown mandatory versions.
      - Changed canonical payload bytes.
      - Expired or not-yet-valid documents.
      - Revoked issuers, subjects, devices, or bindings.
      - A valid signature whose issuer lacks authority for the requested statement.
    - App-native signatures complement rather than replace TLS.
      - TLS authenticates and encrypts the online connection.
      - Signed envelopes preserve application provenance, authorization context, replay protection, and transport independence.
  - **Private-key storage**
    - Generate user and device identity keys using Go's cryptographic random source and `crypto/ed25519`.
    - Derive public fingerprints with SHA-256 over canonical public-key bytes.
    - Encrypt private-key files with:
      - Argon2id for passphrase-based key derivation.
      - Per-file salts and stored KDF parameters so future defaults can increase safely.
      - XChaCha20-Poly1305 for authenticated encryption.
      - A unique nonce for every encryption operation.
    - Never place decrypted private keys in SQLite or logs.
    - Evaluate OS key stores and hardware-backed keys later as optional adapters rather than changing the canonical identity model.
  - **Identity evidence and recovery**
    - Keep a public manifest containing expected fingerprints, certificate IDs, filenames, and timestamps.
      - It is a recovery aid, not an authority.
      - It contains no secrets, passphrases, prompts, model outputs, messages, or raw private material.
    - Classify startup identity state explicitly:
      - **Configured**
        - Required SQLite rows, public files, encrypted-key metadata, key fingerprints, and certificate signatures agree.
      - **Needs setup**
        - Neither database identity state nor recoverable identity files exist.
      - **Needs recovery**
        - One evidence source exists but another is absent, incomplete, corrupted, or safely reconstructable.
      - **Invalid**
        - Cryptographic facts conflict or an automatic repair could bind the wrong user, device, or key.
    - Doctor reports evidence and safe next steps without mutating identity.
    - Repair requires a declared authority source, such as verified encrypted files or an explicitly imported backup.
    - Reset archives recoverable state before creating a new identity and requires deliberate confirmation.
    - Revocation and key rotation must be durable signed operations, not deletion of inconvenient rows.
  - **MVP sequencing correction**
    - Preserve the identity model and minimum secure setup needed for mutual authentication and signed envelopes.
    - Do not reproduce the predecessor's sequencing mistake of polishing every repair and recovery surface before proving the product loop.
    - The first secure two-device echo job should precede advanced recovery UX.
    - Deep recovery, import/export, rotation, and cross-platform secure storage follow as explicit hardening phases.

- **Observability and operational inheritance**
  - **Structured append-only logs**
    - Write one JSON object per line so logs remain streamable, grep-friendly, and machine-readable after an abnormal shutdown.
    - Include UTC timestamp, severity, component, event name, device ID where safe, stable operation ID, correlation ID, and non-sensitive outcome metadata.
    - Record:
      - Runtime startup, selected mode, configuration sources, and orderly shutdown.
      - Database open, migration, backup, integrity, and close events.
      - Command acceptance or rejection and event-application outcomes.
      - Adapter requests, retries, timeouts, cancellations, and terminal status.
      - Queue admission, ownership decisions, dispatch, attempts, and completion.
      - Synchronization cursor movement, replay gaps, duplicate suppression, and authorization failures.
    - Do not record by default:
      - Raw prompts or model responses.
      - Chat or Signal message bodies.
      - Captured voice, transcripts, generated media, or project-file contents.
      - Tokens, cookies, passphrases, private keys, recovery material, or unredacted authorization headers.
    - Redaction must be schema-aware; substring matching alone is only a last defensive layer.
  - **Correlation and audit**
    - Carry one correlation ID across the originating UI/API command, durable outbox record, transport attempts, remote inbox, job attempts, result, and acknowledgements.
    - Keep security-relevant audit records durable and queryable independently from disposable debug logs.
    - Distinguish user intent, automated task action, remote request, and adapter retry as separate actors or causes.
  - **Doctor and health reporting**
    - Doctor checks runtime paths, permissions, database integrity and migration state, identity consistency, certificate validity, network prerequisites, configured services, and adapter availability.
    - Doctor must distinguish:
      - Healthy and ready.
      - Healthy but optional capability unavailable.
      - Misconfigured with an actionable remediation.
      - Unsafe or inconsistent, requiring manual recovery.
    - Cluster health views consume the same structured health model rather than parsing log strings.
  - **Runtime-directory convention**
    - Keep device-local durable state under one configurable runtime root with documented subdirectories for:
      - SQLite and verified backups.
      - Append-only logs and audit exports.
      - Public and encrypted identity material.
      - Cached artifacts and generated outputs.
      - Model or service metadata, without forcing large model binaries into the same backup policy.
      - Temporary files that are safe to remove.
    - Every path must be overrideable for packaged desktop, headless service, and mobile lifecycle requirements.

- **HUD and interaction lessons**
  - **Retain as reusable primitives**
    - Deterministic layout calculations independent of drawing calls.
    - List and detail view models derived from snapshots.
    - Clear headings and non-selectable explanatory rows.
    - Settings routes and developer-only diagnostic surfaces.
    - Mock-data-first construction of navigation, empty, loading, error, and populated states.
    - EbitenUI for conventional controls where it provides reliable focus, forms, text, lists, and layout.
    - Raw Ebitengine rendering for dense custom visualizations where retained widgets would be awkward.
    - Debug UI only for development diagnostics, never as the primary user interface.
  - **Do not preserve the old shell as the product structure**
    - The predecessor's nav/list/detail arrangement is one useful responsive composition, not the mandatory shape of every tab.
    - Each top-level tab may use a layout appropriate to its work:
      - Comrades may combine a roster, conversation, and sharing-policy surface.
      - Projects may combine a project tree, chat/editor workspace, artifacts, and Git operations.
      - Research may combine validated project discovery, budgets, work status, and contribution evidence.
      - Cluster may emphasize device cards, topology, capability matrices, health, and utilization.
      - Routing may emphasize queues, workload classes, eligibility, priorities, and assignments.
      - Tasks may emphasize definitions, triggers, schedules, run history, and event traces.
      - Settings may use ordinary forms and nested diagnostic routes.
    - Layout must adapt to Steam Deck resolution, desktop resizing, touch targets, software keyboard use, and later mobile constraints.
  - **Shared controller and Debian GUI focus model**
    - `L1` and `R1` cycle through top-level tabs in canonical order.
    - Holding `R2` starts push-to-talk capture; releasing `R2` stops capture and submits audio to speech-to-text.
    - Debian GUI uses `Ctrl+PageUp` and `Ctrl+PageDown` for tab cycling, `Alt+1` through `Alt+5` for the five canonical tabs, and right `Ctrl` for hold-to-talk.
    - Debian GUI uses standard keyboard focus, activation, cancellation, contextual-action, scrolling, text-editing, and mouse conventions without creating separate application behavior.
    - Directional controls move through an explicit focus graph, not pointer-coordinate emulation.
    - Confirm activates the focused control.
    - Cancel returns to the previous focus scope, closes a transient surface, or aborts a safe in-progress interaction according to context.
    - Every screen defines:
      - Initial focus.
      - Directional neighbors.
      - Disabled-control behavior.
      - Scroll-container entry and exit.
      - Modal focus trapping and restoration.
      - A visible focus indicator.
      - A non-controller equivalent for keyboard, pointer, and touch.
    - Text entry, file selection, drag-like reordering, and complex editors require explicit controller interaction designs rather than being deferred to mouse behavior.

- **Product and delivery lessons**
  - **What the earlier direction taught**
    - The personal-area-network goal is coherent: a person's devices can collectively host projects, workflows, services, and heterogeneous compute.
    - A portable game engine is useful for a controller-friendly cross-platform console even when the MVP is not itself an RPG.
    - Social resource sharing, research contribution, and gameplay can eventually reinforce participation, but they multiply trust, abuse, networking, and UX complexity.
    - Transport choice must not define the product's identity or durable data model.
    - Local-first durability and explicit authority are more important than pretending every device is always online.
  - **Corrected MVP sequence**
    1. Prove the five-tab shell, controller focus, and push-to-talk interaction with mock data.
    2. Prove shared GUI/headless startup, logging, SQLite migrations, and diagnostics.
    3. Establish the minimum user/device identity and TLS trust needed for two authorized devices.
    4. Send one signed, idempotent HTTPS echo job over a trusted LAN or externally configured WireGuard.
    5. Advertise typed device capabilities and route one text-generation job to one eligible provider.
    6. Make queue state, retries, cancellation, results, and restart recovery durable.
    7. Add project workspace and owner-authoritative project transaction behavior.
    8. Add task schedules, webhooks, and event-driven execution without scheduler failover.
    9. Validate packaged builds and lifecycle behavior on Steam Deck/Linux before claiming broader platform support.
    10. Add Windows, macOS, and Android through evidence-producing build and runtime checkpoints.
    11. Deepen identity recovery, rotation, backup, and secure platform storage after the core loop exists.
    12. Activate Comrades sharing, Research/BOINC, Signal, Meshtastic, and gameplay through separate threat-modeled phases.
  - **Why this order differs**
    - It preserves the predecessor's useful foundation-first discipline.
    - It avoids spending the earliest milestones on security UX that no working cluster behavior consumes yet.
    - It forces architecture claims to survive a real two-device path.
    - It makes platform support an evidence requirement rather than a list of desired build targets.

- **Concepts explicitly rejected or deferred**
  - **RPG-first framing**
    - Reject a simulated game world as the primary navigation and implementation dependency.
    - Retain the option for later gameplay mechanics around:
      - BOINC project validation and contribution evidence.
      - Cooperative milestones and cluster achievements.
      - Understandable visualization of useful work.
    - Gameplay must sit above real durable operations and may not obscure resource use, authorization, cost, or failure state.
  - **qTox and Tor as the primary connection layer**
    - Do not import the qTox/Tor adapters or model Comrades around Tox friend records.
    - The canonical online connection is authenticated HTTPS REST over external WireGuard or a trusted LAN.
    - Keep transport-independent identity and signed envelopes so Signal, Meshtastic, or a future anonymity transport can be added without replacing project, queue, task, or trust models.
    - Reconsider anonymity or store-and-forward networks only against a specific requirement and threat model.
  - **Transport-specific domain fields**
    - Do not put Tox IDs, onion addresses, WireGuard keys, Signal accounts, or Meshtastic node IDs directly into canonical user/device records as if they were identity.
    - Store them as versioned bindings with capabilities, lifecycle state, verification evidence, and revocation.
  - **Fixed three-column layout**
    - Do not require every tab to implement navigation, list, and detail panes.
    - Retain only layout primitives that remain useful under the controller-first responsive HUD.
  - **Mouse-first input**
    - Pointer support remains valuable, but no MVP workflow is complete until it is operable with the target Steam Deck controls.
    - Pointer hit testing must not become the underlying controller model.
  - **Unverified pre-release dependencies**
    - Start from stable Ebitengine, EbitenUI, SQLite, and cryptographic releases.
    - Adopt an alpha or source replacement only for a documented feature or fix, with a rollback path and target validation.
  - **Unproven cross-platform claims**
    - A target enters the supported matrix only after reproducible build, startup, persistence, networking, input, audio, suspend/resume, and packaging checks appropriate to that platform.
    - Listing a `GOOS/GOARCH` pair is not validation.
  - **Python Whisper as the default embedded ASR**
    - Prefer whisper.cpp as the first portable local ASR reference.
    - Keep speech behind an adapter so OS-native, remote, or later Qwen speech services can be selected by capability and policy.
    - Never run model inference synchronously in the Ebitengine update loop.
  - **Complete identity recovery before product proof**
    - Implement enough identity setup, verification, revocation, and diagnostics to secure the vertical slice.
    - Defer polished edge-case recovery workflows until real multi-device operations establish their requirements.
  - **Wholesale code migration**
    - Package names, mock modules, schemas, and old transport assumptions are not a template to clone.
    - Reuse is behavior-by-behavior and test-by-test.

- **Focused reuse procedure**
  - **Before selecting code**
    1. State the Apparat requirement without referring to a predecessor package.
    2. Identify the contract or invariant that satisfies the requirement.
    3. Inspect the predecessor implementation and tests for hidden assumptions.
    4. Classify the source as proven implementation, scaffolding, design-only intent, or obsolete experiment.
    5. Compare it with current stable upstream APIs and Apparat's selected dependency versions.
  - **While adapting code**
    1. Copy the smallest cohesive behavior rather than a directory tree.
    2. Rename or remove qTox, Tor, RPG, three-pane, desktop-path, and host-build assumptions.
    3. Replace predecessor domain types with Apparat's canonical identity, workload, route, queue, task, and envelope types.
    4. Keep external dependencies behind Apparat interfaces.
    5. Preserve useful tests or write new tests that express the behavior independently of the predecessor.
    6. Record source provenance and license obligations in the focused plan and resulting code inventory.
  - **Before accepting the port**
    1. Verify the behavior without the Salvagecore checkout present.
    2. Run focused unit, integration, race, and target-build checks appropriate to the change.
    3. Confirm no import, replace directive, script, generated path, or documentation dependency points at `third_party/salvagecore`.
    4. Confirm logs and tests contain no secrets or sensitive payloads.
    5. Update this roadmap only if implementation evidence changes a retained conclusion.

- **Temporary provenance map**
  - **Runtime composition**
    - `internal/runtime/` demonstrates shared configuration, mode selection, runtime paths, identity commands, diagnostics, and common initialization.
    - `internal/app/gui/` demonstrates a thin GUI adapter that creates the Ebitengine shell after runtime initialization.
    - `internal/app/headless/` demonstrates avoiding Ebitengine initialization in headless mode.
  - **State and modularity**
    - `internal/events/` demonstrates the initial typed UI-command vocabulary.
    - `internal/store/` demonstrates copied state snapshots and command dispatch.
    - `internal/modules/` demonstrates a narrow additive feature registration surface and static mock modules.
    - `internal/domain/` demonstrates UI-independent state and identity types.
  - **Persistence and synchronization design**
    - `internal/database/` demonstrates SQLite opening, migrations, repositories, identity records, snapshots, and constrained inspection.
    - `docs/database_design.md` contains the predecessor's broader owner-authority, queue, transaction, outbox/inbox, and synchronization design; much of this was not implemented.
  - **Identity and diagnostics**
    - `internal/identity/` demonstrates keys, encrypted-key envelopes, manifests, and signed certificate documents.
    - `internal/logging/` demonstrates append-only JSONL logging and basic field redaction.
    - `internal/ui/identity_*` demonstrates first-run, recovery, repair, and reset surfaces that should be sequenced later in Apparat.
  - **HUD**
    - `internal/ui/` demonstrates theme, shell, list/detail projection, deterministic layout calculations, database inspection, Settings routes, and development overlays.
    - Its three-column and pointer-first interaction choices are reference material, not inherited requirements.
  - **Build and dependency evidence**
    - `go.mod` records the predecessor's exact Ebitengine alpha, EbitenUI, debug UI, `modernc.org/sqlite`, and Go cryptography choices.
    - `build.py` proves only host builds were enabled and that all other listed targets remained planned.
    - `.gitmodules` and `third_party/` show how source replacements and reference checkouts were arranged, not which dependencies Apparat must adopt.

- **Reference-removal completion criteria**
  - Every retained architectural concept has a self-contained meaning, Apparat mapping, and constraint in tracked documentation.
  - Every rejected concept has a recorded reason so it is not reintroduced accidentally after the local context disappears.
  - Every copied or adapted implementation has tracked provenance, license review, Apparat-native tests, and no runtime dependency on the predecessor.
  - Apparat's source, tests, build files, scripts, plans, README, and ROADMAP contain no required link into the ignored checkout.
  - The third-party inventory distinguishes actual Apparat source references from this temporary predecessor.
  - Deleting `third_party/salvagecore` changes no build, test, runtime, or documented design meaning.
  - The recursive ignore rule remains in place so a developer may keep or remove the local checkout without affecting repository state.

## Phase 0: Repository And Dependency Foundation

**Goal:** Establish the tracked source references, module boundaries, governance, and reproducible tooling required for implementation.

**Dependencies:** None.

- [x] Add third-party grouping directories and inventories.
  - [x] Create `third_party/README.md`.
  - [x] Create `third_party/game/README.md`.
  - [x] Create `third_party/database/README.md`.
  - [x] Create `third_party/networking/README.md`.
  - [x] Create `third_party/inference/README.md`.
  - [x] Create `third_party/speech/README.md`.
  - [x] Document each source tree's path, upstream URL, license, purpose, pin, and build/reference status.
- [x] Add stable game/HUD source references.
  - [x] Add `https://github.com/hajimehoshi/ebiten.git` at `third_party/game/ebiten`.
  - [x] Pin Ebitengine to a stable 2.9.x revision unless a focused plan proves a later feature is required.
  - [x] Add `https://github.com/ebitenui/ebitenui.git` at `third_party/game/ebitenui`.
  - [x] Add `https://github.com/ebitengine/debugui.git` at `third_party/game/debugui`.
  - [x] Record which sources are referenced through `replace` directives and which remain audit/reference checkouts.
- [x] Add persistence source reference.
  - [x] Add `https://gitlab.com/cznic/sqlite` at `third_party/database/modernc-sqlite`.
  - [x] Pin the actual `modernc.org/sqlite` dependency through `go.mod`.
  - [x] Document why the cgo-free driver is preferred for the initial desktop and Android strategy.
- [x] Add WireGuard source references.
  - [x] Add `https://git.zx2c4.com/wireguard-go` at `third_party/networking/wireguard-go`.
  - [x] Add `https://github.com/WireGuard/wgctrl-go.git` at `third_party/networking/wgctrl-go`.
  - [x] Add `https://git.zx2c4.com/wireguard-tools` at `third_party/networking/wireguard-tools`.
  - [x] Document that these are references for detection, control, platform behavior, and later integration—not proof of an embedded cross-platform tunnel.
- [x] Add inference and speech source references.
  - [x] Add `https://github.com/ggml-org/llama.cpp.git` at `third_party/inference/llama.cpp`.
  - [x] Mark llama.cpp as a future service adapter rather than an initial HUD binary dependency.
  - [x] Add `https://github.com/ggml-org/whisper.cpp.git` at `third_party/speech/whisper.cpp`.
  - [x] Mark whisper.cpp as the first portable local ASR reference.
- [x] Define deferred source admission gates.
  - [x] Research Qwen3-TTS runtime, packaging, hardware, licensing, and service boundaries before adding its source.
  - [x] Select a Meshtastic language/client/protobuf integration before adding Meshtastic source.
  - [x] Establish Signal gateway feasibility and maintenance constraints before adding an implementation.
  - [x] Select BOINC client, RPC, manager, or integration sources only after the Research architecture defines the required boundary.
  - [x] Require an approved use case before adding any alternative model, speech, artifact, or networking runtime.
- [x] Record MVP exclusions.
  - [x] Exclude qTox, TokTok qTox, and go-toxcore-c.
  - [x] Exclude Tor.
  - [x] Exclude WebRTC until a requirement cannot be met by HTTPS and event cursors.
  - [x] Exclude curl because the Go HTTP stack covers the first API.
  - [x] Exclude OpenSSL and libsodium from the MVP build.
  - [x] Record that OpenSSL does not supply PGP semantics.
  - [x] Exclude Qwen3-ASR while whisper.cpp is the selected local ASR reference.
  - [x] Exclude `golang/mobile` as a source checkout while using pinned Ebitengine mobile tooling.
  - [x] Exclude termframe until an interactive headless TUI is approved.
- [x] Establish the Go application workspace.
  - [x] Create and pin the root Go module.
  - [x] Define supported Go and Ebitengine versions.
  - [x] Separate application code from third-party reference modules.
  - [x] Establish formatting, linting, unit-test, race-test, and dependency-audit commands.
  - [x] Pin build tools independently from source-reference submodules.
- [x] Establish application governance.
  - [x] Define module/package boundaries.
  - [x] Define file-size and decomposition expectations.
  - [x] Require README inventories for application and third-party grouping directories.
  - [x] Define structured logging and sensitive-data redaction requirements.
  - [x] Define documentation synchronization requirements.

**Exit criteria**

- Required source submodules are pinned and documented.
- Deferred and excluded sources are explicitly recorded.
- Go dependencies and tool versions are reproducible.
- The repository can run baseline formatting, linting, and test commands.

## Phase 1: Architecture And Protocol Contracts

**Goal:** Resolve security- and interoperability-critical questions before production networking or persistence code.

**Dependencies:** Phase 0.

- [x] Create `docs/architecture.md`.
  - [x] Define ports-and-adapters package boundaries.
  - [x] Define GUI, headless, service-host, queue-owner, project-owner with owner-local Task scheduling, and enrollment-authority roles.
  - [x] Define module registration and command/event/store boundaries.
  - [x] Define the versioned workload-class taxonomy and extension rules.
  - [x] Define typed device, service, queue, route, and job capability contracts.
  - [x] Define which Salvagecore components are copied, adapted, rewritten, or rejected.
- [x] Create the shared input and focus contract in `docs/controller-map.md`.
  - [x] Define `L1`, `R1`, D-pad, sticks, `A`, `B`, `X`, Menu, and `R2`.
  - [x] Define Debian `Ctrl+PageUp`, `Ctrl+PageDown`, `Alt+1` through `Alt+5`, focus traversal, activation, cancellation, contextual actions, command palette, scrolling, and right-`Ctrl` push-to-talk.
  - [x] Define Debian mouse activation, context menu, scrolling, back, drag alternatives, and optional configurable push-to-talk buttons.
  - [x] Define Debian headless CLI, API, service-manager, health-check, and process-signal controls.
  - [x] Define focus traversal, disabled controls, modal focus, scrolling, and pane transitions.
  - [x] Define keyboard, mouse, touch, and controller equivalence.
  - [x] Define binding precedence while text controls or editors own focus.
  - [x] Define configurable bindings, conflict reporting, and platform-reserved shortcut handling.
  - [x] Define `Steam+X` and visible on-screen keyboard entry points.
  - [x] Define push-to-talk recording, cancellation, release-to-submit, and feedback states.
- [x] Create `docs/security.md`.
  - [x] Produce a threat model for local LAN, WireGuard, stolen devices, malicious peers, replay, queue abuse, and compromised services.
  - [x] Choose the X.509 hierarchy.
  - [x] Decide whether TLS keys reuse app device keys or are separately bound.
  - [x] Define user, device, cluster, WireGuard, TLS, and future transport identities.
  - [x] Define enrollment invite, QR code, token expiration, confirmation, and authorization.
  - [x] Define certificate issuance, expiration, rotation, revocation, and lost-device recovery.
  - [x] Define authorization scopes for projects, queues, services, tasks, settings, and external transports.
  - [x] Define audit events and secret-redaction rules.
  - [x] Disable TLS 0-RTT for mutating operations.
- [x] Create `docs/api.md` and an OpenAPI source.
  - [x] Define `/v1/health`.
  - [x] Define `/v1/device`.
  - [x] Define `/v1/capabilities`.
  - [x] Define `POST /v1/jobs`.
  - [x] Define `GET /v1/jobs/{id}`.
  - [x] Define `POST /v1/jobs/{id}/cancel`.
  - [x] Define cursor-based `/v1/events`.
  - [x] Define `/v1/project-transactions`.
  - [x] Define `202 Accepted`, resource locations, and asynchronous error bodies.
  - [x] Define `Idempotency-Key`, body limits, deadlines, bounded concurrency, and content types.
  - [x] Define authentication and authorization errors without leaking sensitive state.
  - [x] Define typed capability descriptors for text, image, video, STT, TTS, BOINC, and future workload classes.
  - [x] Define job workload-class and capability-requirement validation errors.
  - [x] Prohibit generic remote execution endpoints.
- [x] Define the signed-envelope contract.
  - [x] Define envelope version and message type.
  - [x] Define message ID, idempotency key, and correlation ID.
  - [x] Define sender identity and recipient target.
  - [x] Define timestamps, expiration, and deadline.
  - [x] Define payload type, schema version, length, hash, and artifact references.
  - [x] Select canonical signed encoding.
  - [x] Define signature verification, replay rejection, duplicate handling, and version negotiation.
- [x] Create `docs/database.md`.
  - [x] Define identity, directory, project, chat, artifact, typed service capability, queue, job, event, transaction, task, research, and audit layers.
  - [x] Define workload classes independently from runtime/provider names and model/project IDs.
  - [x] Define one authoritative owner per project, queue, and task.
  - [x] Define forward-only migrations and compatibility policy.
  - [x] Define ULID identifiers and UTC millisecond timestamps.
  - [x] Define SQLite backup, repair, restore, and optional at-rest encryption decisions.
- [x] Create `docs/transport-adapters.md`.
  - [x] Define transport capability descriptors.
  - [x] Define payload-size and fragmentation rules.
  - [x] Define online, delayed, direct, broadcast, acknowledgement, attachment, and store-forward capabilities.
  - [x] Define how REST JSON and future compact binary encodings carry the same logical envelope.
- [x] Create `docs/platform-matrix.md`.
  - [x] Define Steam Deck/Linux GUI requirements.
  - [x] Define Linux headless requirements.
  - [x] Define Windows and macOS packaging and external-WireGuard assumptions.
  - [x] Define Android native wrapper, lifecycle, storage, keyboard, microphone, audio, and background constraints.
  - [x] Define evidence required before claiming support.
- [x] Create the canonical build artifact contract.
  - [x] Define `./releases/[os]/[architecture]/apparat/latest[.exe]` for the GUI console.
  - [x] Define `./releases/[os]/[architecture]/apparatd/latest[.exe]` for the headless worker/service.
  - [x] Use Go `GOOS` and `GOARCH` naming for release directories.
  - [x] Use `.exe` for Windows artifacts and no suffix for Unix-like artifacts.
  - [x] Implement a Python build pipeline that detects host OS and architecture.
  - [x] Build both canonical binaries by default while preserving single-target builds.
  - [x] Add build-pipeline tests.
  - [x] Track generated binary artifacts in Git so other devices can pull latest builds directly.
- [x] Create deferred-feature design stubs.
  - [x] Create `docs/comrades.md` with the accepted social identity, chat, comrade queue, permission, priority, quota, revocation, and audit goals.
  - [x] Create `docs/research.md` with the accepted BOINC, resource-budget, isolation, validation, provenance, and gameplay goals.
  - [x] Keep both documents clearly marked post-MVP until their implementation phases begin.

**Exit criteria**

- The package architecture, security model, OpenAPI contract, signed envelope, database boundaries, transport interface, controller map, and platform matrix are approved.
- The canonical release artifact path and Python build pipeline are implemented and tested.
- The two-device proof can be implemented without inventing protocol or identity semantics mid-build.

## Phase 2: Steam Deck HUD Prototype

**Goal:** Build the controller-first visual shell against mock data before networking or durable domain features.

**Dependencies:** Phase 1 controller and architecture contracts.

**Historical evidence note:** The completed checklist records the prototype/mockup milestone and its intended input contract. It does not prove that every focus, activation, context, scrolling, accessibility, or input-equivalence behavior is present in the current executable. New Phases 6–7 reconcile and implement that remaining evidence without rewriting the historical checkpoint.

- [x] Create the Ebitengine application shell.
  - [x] Add a small executable entrypoint.
  - [x] Add a thin Ebitengine game adapter.
  - [x] Add resize and display-scale handling.
  - [x] Add a stable theme and readable Steam Deck sizing.
- [x] Implement the shared input/focus model.
  - [x] Detect and normalize standard gamepad input.
  - [x] Implement directional focus movement.
  - [x] Implement activation, back, contextual action, and scrolling.
  - [x] Implement Debian GUI keyboard controls.
    - [x] Implement `Ctrl+PageUp` and `Ctrl+PageDown` tab cycling.
    - [x] Implement the then-documented direct-tab selection; the current canonical contract is `Alt+1` through `Alt+5`.
    - [x] Implement `Tab`, `Shift+Tab`, arrows, `Enter`, `Space`, `Escape`, Menu or `Shift+F10`, `Ctrl+Shift+P`, and collection-navigation keys.
    - [x] Preserve ordinary text editing and clipboard behavior while text controls own focus.
  - [x] Implement mouse/touch focus, activation, context actions, scrolling, and non-pointer drag alternatives without separate application behavior.
  - [x] Add deterministic focus-navigation tests.
- [x] Implement top-level tabs.
  - [x] Add Comrades as the first tab.
  - [x] Add Projects.
  - [x] Add Research after Cluster.
  - [x] Add Cluster.
  - [x] Add Routing as Cluster selector/content-panel content.
  - [x] Add Tasks as Cluster selector/content-panel content.
  - [x] Add Settings as the final tab.
  - [x] Implement `L1`/`R1` tab switching and wrap behavior.
- [x] Add mock product views.
  - [x] Show a clearly labeled future Comrades placeholder describing chat and shared compute.
  - [x] Show mock device health, capabilities, queue depth, and activity.
  - [x] Show mock project chats, files, artifacts, and Git state.
  - [x] Show mock text-generation, image-generation, video-generation, STT, TTS, and BOINC capabilities on different devices.
  - [x] Show mock typed services, models/projects, queues, routing profiles, compatibility filtering, and fallbacks.
  - [x] Show mock schedules, webhooks, events, approvals, and run history.
  - [x] Show mock identity, networking, database, audio, and diagnostics settings.
  - [x] Show a clearly labeled future Research placeholder describing BOINC and validated research compute.
- [x] Prototype voice input states without real ASR.
  - [x] Start capture state while `R2` is held.
  - [x] End and submit capture state when `R2` is released.
  - [x] Start capture state while right `Ctrl` is held in Debian GUI mode.
  - [x] End and submit capture state when right `Ctrl` is released.
  - [x] Cancel a held right-`Ctrl` recording with `Escape` without submitting on release.
  - [x] Support cancellation.
  - [x] Render recording, queued, transcribing, failed, and complete states.
- [x] Add developer diagnostics.
  - [x] Frame time and memory.
  - [x] Active route and focused control.
  - [x] Controller identity and current input.
  - [x] Store/event queue depth.
  - [x] Layout bounds and clipping.

**Exit criteria**

- Every primary HUD operation works from a Steam Deck controller.
- Every primary HUD operation works in Debian GUI mode with the documented keyboard controls.
- Keyboard, mouse, touch, and controller input follow the same focus and command semantics.
- `R2` and right-`Ctrl` push-to-talk states are testable without a real speech model.
- No networking, database, or model runtime is required to demonstrate the shell.
- Comrades and Research are navigable placeholders without active backend implementation.

## Phase 3: Local Runtime, Identity, Persistence, And Diagnostics

**Goal:** Establish durable local behavior shared by GUI and headless devices.

**Dependencies:** Phase 1 contracts; Phase 2 may proceed partly in parallel.

- [?] Split GUI and headless runtime adapters.
  - [?] Make the default GUI binary enter a real Ebitengine run loop instead of exiting after runtime initialization.
    - The Ebitengine loop is implemented behind the `gui` build tag and the release pipeline builds the `apparat` artifact with that tag.
    - Native desktop-library and display-server validation remains target-specific evidence.
  - [x] Build GUI and headless artifacts into separate binary-specific release directories.
    - [x] Use `releases/<goos>/<goarch>/apparat/latest[.exe]` for the GUI artifact.
    - [x] Use `releases/<goos>/<goarch>/apparatd/latest[.exe]` for the headless artifact.
    - [x] Track generated artifacts in Git as the current latest build surface.
  - [x] Keep GUI and headless default runtime roots separate unless `--runtime-dir` explicitly overrides them.
  - [x] Keep `--smoke-test` as the non-window build and CI verification path.
  - [x] Keep Ebitengine initialization out of headless mode.
  - [x] Add explicit GUI, headless, and auto modes.
  - [x] Add startup diagnostics and doctor mode.
  - [x] Add clean shutdown and context cancellation.
- [x] Add runtime configuration and directories.
  - [x] Define config precedence.
  - [x] Define database, logs, identity, cache, artifacts, backups, and recovery paths.
  - [x] Avoid storing durable runtime state inside project source directories by default.
- [x] Add structured logging.
  - [x] Write append-only JSONL.
  - [x] Recreate runtime-root `last_run.log` at each process start for immediate verbose diagnostics.
  - [x] Record binary name, mode, runtime root, OS, architecture, Go version, process ID, flags, component startup, doctor, smoke-test, errors, panics, and shutdown state in `last_run.log`.
  - [x] Include component, event, device, project, job, task, and correlation IDs where relevant.
  - [x] Redact secrets, tokens, private keys, raw prompts, raw model outputs, and raw voice data by default.
  - [x] Add safe log rotation and retention.
- [x] Add source-size governance.
  - [x] Require included code files to stay at or below 400 physical lines.
  - [x] Exclude generated, vendored/reference, `third_party/`, `.tools/` and `.tmp/`, release, plan, journal, downtime, and prose documentation files.
  - [x] Add `scripts/check_code_file_lines.py`.
  - [x] Add `make check-code-size`.
  - [x] Include the check in `make verify`.
- [x] Add documentation completeness governance.
  - [x] Require tracked source directories to carry local `README.md` documentation.
  - [x] Require tracked scripts to be inventoried in `scripts/README.md`.
  - [x] Require executable Python scripts to provide `--help` usage.
  - [x] Add `scripts/check_directory_docs.py`.
  - [x] Add `make check-docs`.
  - [x] Include the check in `make verify`.
- [x] Add SQLite lifecycle.
  - [x] Open, close, ping, and configure connections.
  - [x] Enable foreign keys.
  - [-] Keep WAL opt-in until platform validation is complete.
  - [x] Add forward migrations with checksums.
  - [x] Add ULID and timestamp helpers.
  - [x] Add repository interfaces that do not leak SQL into the HUD.
  - [x] Add read-only database diagnostics.
- [x] Add user and device identity.
  - [x] Generate/import user identity.
  - [x] Generate device identity.
  - [x] Sign device authorization.
  - [x] Encrypt private-key files with Argon2id and XChaCha20-Poly1305.
  - [x] Create public manifests and identity metadata.
  - [x] Add startup consistency classification.
  - [x] Add doctor, repair, rotation, revocation, and archived reset.
- [x] Add the local cluster directory.
  - [x] Store signed device profiles.
  - [x] Store roles, permissions, endpoints, certificate fingerprints, WireGuard keys, and typed workload capabilities.
  - [x] Store capability runtime/provider, models or research projects, modalities, limits, hardware, queue eligibility, health, and policy constraints.
  - [x] Store last-seen and reachability state.
  - [x] Add change feeds and sync cursors.
- [x] Add durable local messaging primitives.
  - [x] Add outbox.
  - [x] Add inbox.
  - [x] Add replay and duplicate tracking.
  - [x] Add event cursor state.
  - [x] Add bounded retry scheduling.

**Exit criteria**

- GUI and headless modes share one durable runtime.
- Identity can be created, recovered, rotated, and diagnosed.
- SQLite survives restart and migrations.
- Logs explain state transitions without exposing sensitive payloads.

## Phase 4: Basic HUD Tabs And Content

**Goal:** Turn the mock HUD foundation into the first usable local UI before adding more backend networking and service complexity.

**Dependencies:** Phases 1–3.

- [x] Establish the tab shell as the next implementation focus.
  - [x] Keep the canonical tab order: Comrades, Projects, Cluster, Research, Settings; render Routing and Tasks as Cluster selector/content-panel content.
  - [x] Represent tabs as data from a tab-view model rather than hard-coding a single visual strip.
  - [x] Store the tab list as ordered tab descriptors with stable IDs, labels, icons or glyph slots, accessibility labels, visibility state, and future badge/status metadata.
  - [x] Default to a top tab bar for the MVP.
  - [x] Design the tab-view model so a later setting can realign tabs from the top edge to a side rail without changing tab content implementations.
  - [x] Keep tab content independent from tab placement so top, left, right, compact, and future responsive layouts can share the same selected-tab state.
  - [x] Keep `L1` and `R1` tab switching for Steam Deck/controller input.
  - [x] Keep Debian/Linux keyboard tab switching through `Ctrl+PageUp`, `Ctrl+PageDown`, and `Alt+1` through `Alt+5`.
  - [x] Represent input actions as named bindings from the configuration manager rather than scattering hard-coded key checks through tab code.
  - [x] Keep default bindings hard-coded for now while preserving a path to user-editable bindings later.
  - [x] Preserve mouse/touch activation without making essential workflows pointer-only.
  - [x] Keep `R2` and right `Ctrl` push-to-talk visible as UI state even before real ASR is integrated.
- [x] Add a temporary HUD configuration manager.
  - [x] Provide hard-coded default values through one configuration manager boundary during Phase 4.
  - [x] Keep the boundary shaped so a later implementation can load and save the same values through SQLite-backed user configuration tables.
  - [x] Include tab-view defaults.
    - [x] Canonical tab order.
    - [x] Default tab placement: top.
    - [x] Future allowed placements: top, left side rail, right side rail, compact/sidebar-responsive.
    - [x] Tab density: comfortable by default, with compact and expanded options planned.
    - [x] Tab label mode: icon plus text by default, with icon-only and text-only options planned where usable.
    - [x] Default selected tab.
  - [x] Include key-binding defaults.
    - [x] Previous tab: `L1` and `Ctrl+PageUp`.
    - [x] Next tab: `R1` and `Ctrl+PageDown`.
    - [x] Direct tab selection: `Alt+1` through `Alt+5`.
    - [x] Push-to-talk: `R2` and right `Ctrl`.
    - [x] Cancel recording: `Escape`.
    - [x] Focus, activation, back, context menu, command palette, scroll, and collection-navigation actions.
  - [x] Include display and accessibility defaults.
    - [x] Theme: dark by default, with light and high-contrast variants planned.
    - [x] Accent color: Apparat default now, user-selectable later.
    - [x] UI scale/zoom: default `1.0`.
    - [x] Font size: Steam Deck readable default, with future small/medium/large/custom choices.
    - [x] Font family: bundled/default UI font now, user-selectable later where platform packaging permits.
    - [x] Motion/reduced-animation preference.
    - [x] Contrast and focus-ring strength.
    - [x] Panel density, list row height, and card spacing.
    - [x] Text wrapping and truncation preference for long project, queue, device, and task names.
  - [x] Include interaction defaults.
    - [x] Controller sensitivity and repeat delay.
    - [x] Keyboard repeat delay for held navigation.
    - [x] Mouse/touch scroll speed.
    - [x] Push-to-talk mode: hold by default, with future toggle mode if accessibility testing supports it.
    - [x] Confirmation requirements for destructive actions.
    - [x] Default command-palette visibility and shortcut.
    - [x] Default landing tab on startup.
    - [x] Remember last selected tab, panel, project, route, and task when enabled later.
    - [x] Default sort and filter preferences for devices, projects, queues, tasks, comrades, and research projects.
  - [x] Include notification defaults.
    - [x] Notification visibility: important local events only by default.
    - [x] Notification sound volume and mute state.
    - [x] Toast duration and whether controller focus should move to urgent notifications.
    - [x] Categories for job completion, device offline/online, task failure, comrade request, research milestone, and security warning.
    - [x] Quiet-hours schedule placeholder.
  - [x] Include diagnostic defaults.
    - [x] Developer overlay visibility.
    - [x] Log detail level for local UI diagnostics.
    - [x] Whether runtime paths and build artifacts are shown in Settings by default.
    - [x] Whether frame timing, memory, input events, focus path, and layout bounds are shown in diagnostics.
  - [x] Include default view preferences.
    - [x] Projects default view: recent projects first.
    - [x] Cluster default view: device health summary first.
    - [x] Cluster includes Routing detail: workload-class overview first.
    - [x] Cluster Tasks selector default view: active and failed runs first.
    - [x] Comrades default view: placeholder relationship list first until chat exists.
    - [x] Research default view: placeholder validated-project catalog first until BOINC integration exists.
  - [x] Include privacy and safety defaults.
    - [x] Hide sensitive paths and identifiers by default in presentation surfaces where practical.
    - [x] Require explicit reveal for secrets, tokens, private keys, raw prompts, model outputs, and raw voice diagnostics.
    - [x] Default sharing posture: no comrade or research resource sharing until explicitly enabled.
  - [x] Do not expose user editing UI yet unless it is clearly marked non-persistent or future.
- [x] Implement reusable HUD layout primitives.
  - [x] Add a consistent top tab bar.
  - [x] Build the tab bar through the tab-view model so the same content can later render as a side rail.
  - [x] Add focusable panels, lists, cards, empty states, status pills, and action rows.
  - [x] Add a shared detail-pane pattern for selected items.
  - [x] Add loading, offline, warning, and disabled states.
  - [x] Add controller/keyboard focus styling that is visible at Steam Deck scale.
  - [x] Keep rendering driven by view models rather than direct database or adapter calls.
- [x] Implement the Comrades tab as a visible placeholder.
  - [x] Explain real-friend chat as a future capability.
  - [x] Explain comrade queues for low-priority shared inference access.
  - [x] Show placeholder sharing grants, queue access, quota, revocation, and audit concepts.
  - [x] Keep all controls disabled or clearly marked future until backend support exists.
- [x] Implement the Projects tab basic content.
  - [x] Show project list, selected project summary, chat preview, file tree placeholder, artifact list placeholder, and Git status placeholder.
  - [x] Add local-only mock actions for selecting projects, opening files, viewing chat entries, and inspecting Git state.
  - [x] Show offline draft and transaction concepts without applying real file changes yet.
- [x] Implement the Research tab as a visible placeholder.
  - [x] Explain BOINC delegation as validated public-interest compute.
  - [x] Show placeholder project catalog, validation state, budget, schedule, contribution, and gameplay-validation concepts.
  - [x] Keep BOINC execution disabled until the later Research phase.
- [x] Implement the Cluster tab basic content.
  - [x] Show local device identity status, runtime mode, runtime root, database path, and `last_run.log` status.
  - [x] Show mock device cards with roles, reachability, health, typed capabilities, and queue/service ownership.
  - [x] Surface doctor status and recent diagnostics in a human-readable panel.
- [x] Implement Routing as Cluster detail content.
  - [x] Show workload classes: text generation, image generation, video generation, STT, TTS, and BOINC research compute.
  - [x] Show mock queues, priorities, device assignments, compatibility filtering, fallback routes, and policy constraints.
  - [x] Make it clear that BOINC is schedulable research compute, not model inference.
- [x] Implement Tasks as Cluster selector/content-panel content.
  - [x] Show placeholder scheduled tasks, webhooks, event-driven tasks, Signal-driven tasks, manual approvals, and run history.
  - [x] Show disabled create/edit controls until durable task storage and execution exist.
- [x] Implement the Settings tab basic content.
  - [x] Show local runtime paths, build artifact paths, mode, identity status, documentation/check status, and developer diagnostics.
  - [x] Show the current temporary HUD configuration values, including tab placement, theme, scale, font size, and key-binding defaults.
  - [x] Label configuration values as hard-coded Phase 4 defaults that will later load from and save to SQLite-backed user settings.
  - [x] Show controls or command hints for `--doctor`, `--smoke-test`, `last_run.log`, and verification commands.
  - [x] Keep destructive identity/runtime operations disabled until explicit backend support exists.
- [x] Add UI verification and documentation.
  - [x] Add deterministic tests for tab order, tab content models, input actions, focus transitions, and placeholder disabled states.
  - [x] Document each tab's current MVP behavior and future backend boundary in `internal/hud/README.md` or tab-specific docs.
  - [x] Update screenshots or text walkthroughs when the visual shell is stable enough to show.

**Exit criteria**

- The GUI opens into a usable five-tab HUD with readable basic content for every canonical tab, including Routing and Tasks under Cluster and Pipelines under Projects.
- The tab system is data-driven enough to support future top/side realignment without rewriting tab contents.
- Key bindings and user-facing display defaults come from a temporary configuration manager rather than scattered literals.
- Controller, keyboard, mouse, and touch can navigate the basic tab content without backend services.
- All backend-dependent controls are clearly disabled, mocked, or labeled future.
- Cluster and Settings expose enough local diagnostics to debug startup, runtime paths, and `last_run.log`.
- The next backend phase can wire real data into established view-model boundaries instead of inventing UI structure.

## Phase 5: Android GUI APK Build Pipeline

**Goal:** Produce a working Android GUI APK artifact for Apparat in the canonical release directory without adding or claiming an Android headless target.

**Current status:** Build pipeline, package inspection, Pixel install, process-liveness validation, app-private runtime storage, Android `last_run.log` creation, wrapper HUD rendering, full-screen Android view sizing, touch tab selection, and screenshot evidence are implemented for Linux-hosted `android/arm64` GUI APK builds. Direct `GoNativeActivity` paths can initialize Apparat runtime state but stay on the Android splash/default icon instead of attaching Ebitengine's view. The current APK now uses tracked Apparat wrapper sources under `android/apparat`, `cmd/apparatmobile`, and Ebitengine's generated mobile view classes. A temporary Settings `Updates` fieldset and native Android button bridge are being added during Phase 5 for the tracked GitHub `latest.apk`; production update manifests, installed/latest version display, and release signing remain future work. Phase 5 remains incomplete until additional device testing, Android safe-area and density hardening, runtime-path validation depth, release-hardening deferrals, and any local Ebitengine patch/submodule reproducibility work are resolved or explicitly deferred.

**Dependencies:** Phase 4 HUD shell and Apparat-owned tracked dependencies. Salvagecore may be inspected temporarily, but the final build pipeline must not require it.

**Scope boundary:** This phase builds only the GUI Android APK. It does not build `apparatd` for Android; users who want headless behavior on Android can later run Linux headless builds through Termux-like environments or a separate approved worker strategy.

**Salvagecore retirement boundary:** `third_party/salvagecore` is a temporary ignored reference and will eventually be removed. Android source, scripts, manifests, wrapper code, and tooling required for Apparat must live in Apparat-owned tracked paths or documented external tool prerequisites. The implemented pipeline does not reference `third_party/salvagecore`.

**Build host policy:** Android APK builds are not inherently Linux-only. Phase 5 uses Linux as the first evidence-producing build host because it is the current development and verification baseline, but the Python pipeline remains host-agnostic where practical. macOS and Windows Android build hosts are planned only after explicit validation on those hosts.

- [x] Confirm Android source and reference baseline.
  - [x] Treat `third_party/game/ebiten` as the admitted Ebitengine source containing `cmd/ebitenmobile` and mobile runtime packages.
  - [x] Inspect Ebitengine mobile behavior and confirm direct `gomobile build` can emit a GUI APK for `cmd/apparat`.
  - [x] Reject a durable dependency on salvagecore's ignored `third_party/cicd/mobile` checkout.
  - [x] Document that salvagecore remains temporary reference material only and is not an Android build input.
  - [x] Keep salvagecore ignored and unstaged; do not add any file under `third_party/salvagecore` to this repository.
- [x] Convert temporary reference lessons into Apparat-owned inputs.
  - [x] Use tracked Apparat-owned files for Android behavior: `cmd/apparat/AndroidManifest.xml`, `cmd/apparat/gomobile_app.go`, `scripts/build.py`, Makefile targets, tests, and documentation.
  - [x] Keep `golang/mobile` absent as a source checkout; the first APK uses Ebitengine's pinned `github.com/ebitengine/gomobile` Go module instead.
  - [x] Keep `ebitenmobile bind` deferred as future wrapper/AAR reference material rather than the Phase 5 APK path.
  - [x] Add a unit guard that fails if the Android build pipeline references `third_party/salvagecore`.
  - [ ] Prove the Android build works after temporarily moving or hiding `third_party/salvagecore`; the implemented script has no reference to it, but this destructive local proof still needs an explicit checkpoint if desired.
- [x] Choose the first APK architecture.
  - [x] Select `android/arm64` as the first APK architecture.
  - [x] Use direct Ebitengine `gomobile build` as the shortest diagnostic path for producing and installing an APK.
  - [x] Record that the first Android-only `mobile.SetGame` runner was not sufficient because it reached process startup without attaching a visible Ebitengine HUD surface on Android.
  - [x] Test the shared `ebiten.RunGame` adapter and record that it still fails under direct `GoNativeActivity` with `internal/ui: Run is not implemented for GOOS=android`.
  - [x] Promote the host-owned wrapper/AAR-style path to required Phase 5 work.
  - [x] Record the decision in `README.md`, `ROADMAP.md`, `scripts/README.md`, `cmd/apparat/README.md`, and `docs/platform-matrix.md`.
- [x] Pin Android build prerequisites.
  - [x] Define required JDK version: JDK 21.
  - [x] Define Android SDK command-line tools as an external prerequisite under `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or ignored `.tools/android`.
  - [x] Define Android platform/API level: `android-35`.
  - [x] Define Android build-tools version: `35.0.0`.
  - [x] Define Android NDK version: `27.2.12479018`.
  - [x] Define Gradle/Android Gradle Plugin versions as not applicable because the wrapper is assembled with Android SDK/JDK tools in the Python pipeline.
  - [x] Define tool discovery through `ANDROID_HOME`, `ANDROID_SDK_ROOT`, `ANDROID_NDK_HOME`, `JAVA_HOME`, PATH, and ignored repo-local `.tools/` and `.tmp/` fallbacks.
  - [x] Keep downloaded tools outside tracked source.
- [x] Define Android build host support.
  - [x] Treat Linux as the first evidence-producing host for Phase 5.
  - [x] Keep `scripts/build.py` path handling, environment detection, and command construction portable with `pathlib` and environment variables.
  - [x] Avoid Linux-only shell assumptions inside the Python build logic.
  - [x] Report macOS/Windows hosts as unvalidated rather than unsupported when running Android preflight.
  - [x] Document that macOS and Windows APK build hosts are not claimed supported until their preflight, build, and install/launch validation pass.
  - [x] Allow APK install/launch validation from any host with working `adb`.
- [x] Add Android build preflight checks.
  - [x] Add `python3 scripts/build.py --check-android-env`.
  - [x] Add `make check-android-build-env`.
  - [x] Report missing Java, SDK command-line tools, platform, build-tools, NDK, gomobile module source, and optional `adb`.
  - [x] Detect unsupported Android ABI values before build time.
  - [x] Keep `python3 scripts/build.py --os android --target apparatd` rejected with a headless-out-of-scope explanation.
  - [x] Prepare ignored `.tools/bin/gomobile-apparat` when the pinned Ebitengine gomobile helper is missing.
- [x] Define canonical Android GUI artifact paths.
  - [x] Use `releases/android/arm64/apparat/latest.apk` for the primary Android GUI APK.
  - [x] Keep additional ABI paths deferred until individually validated.
  - [x] Do not create `releases/android/<arch>/apparatd/` in this phase.
  - [x] Track generated APKs in Git as the current latest Android build surface.
- [x] Integrate APK output into the Python build pipeline.
  - [x] Extend `scripts/build.py --os android --arch arm64 --target apparat` to produce `latest.apk` after preflight succeeds.
  - [x] Keep default host desktop builds unchanged when no Android target is requested.
  - [x] Ensure `--print-path` reports the APK path without building.
  - [x] Ensure `--target all --os android` builds only Android-supported targets, currently `apparat`.
  - [x] Add `make build-android`.
  - [x] Add `make check-android-build-env`.
- [x] Add Android application metadata.
  - [x] Define app package/application ID: `com.cjtrowbridge.apparat`.
  - [x] Define app label and launcher metadata: `Apparat`, `com.cjtrowbridge.apparat.MainActivity`.
  - [x] Define current version metadata: `versionCode=1`, `versionName=0.1.0`.
  - [x] Define orientation behavior: the wrapper no longer forces portrait and should fill the available screen in supported orientations.
  - [x] Define debug signing behavior: direct `gomobile build` emits a debug APK.
  - [ ] Define release signing, store packaging, and automated version generation in a later release-hardening phase.
  - [x] Define explicit SDK metadata through the patched helper: `minSdkVersion=23`, `targetSdkVersion=30`, and platform build version `35`.
- [x] Add Android permissions and platform behavior.
  - [x] Request `android.permission.INTERNET` for HTTPS over external WireGuard/local network.
  - [x] Avoid broad storage permissions; runtime data remains app-scoped by default.
  - [x] Request microphone permission for the existing push-to-talk state path while keeping real Android audio capture validation as future Phase 10 work.
  - [x] Defer VPN-service permissions and app-managed WireGuard to the later transport/platform phase.
  - [x] Validate touch tab selection on Android while keeping keyboard, controller, and text-input coverage as follow-up testing.
  - [x] Defer Android audio/TTS behavior beyond package startup.
- [ ] Adapt runtime paths for Android GUI.
  - [x] Reuse the existing runtime startup path so `last_run.log` creation is part of Android app startup.
  - [ ] Verify the actual Android app-scoped runtime root on device/emulator.
  - [ ] Verify `last_run.log` is recreated on every Android GUI launch.
  - [ ] Verify SQLite, logs, identity, cache, artifacts, backups, and recovery directories are Android-safe.
  - [x] Document Android runtime-path assumptions and validation gap in `docs/platform-matrix.md`.
- [ ] Validate the Android GUI smoke path.
  - [x] Build a debug APK locally.
  - [x] Inspect package metadata with Android build-tools: package ID, label, `INTERNET` permission, launcher activity, and `arm64-v8a` native code.
  - [x] Install the APK on a physical Pixel device with `adb`.
  - [x] Launch the app and verify the process remains alive after startup.
  - [x] Verify Android package/startup fixes discovered during Pixel testing: modern SDK metadata, v2/v3 signing, 16 KB native page alignment, and app-private runtime storage.
  - [x] Record that the first wrapper APK showed only the Android splash/default icon even though the process remained alive.
  - [x] Replace the direct `GoNativeActivity` startup path with an Apparat-owned wrapper activity that attaches Ebitengine's generated `EbitenView`.
  - [x] Verify the app opens to the Apparat HUD instead of the Android splash/default icon.
  - [x] Verify the seven tabs render and remain clickable/touchable on Android.
  - [ ] Verify keyboard/controller navigation where the device supports it.
  - [x] Verify `last_run.log` exists in the Android app-private runtime root after launch.
  - [x] Capture `adb logcat` and `last_run.log` evidence for install/startup failures and fixes.
- [ ] Implement Android GUI parity.
  - [x] Decide the minimal Apparat-owned Android integration path: tracked wrapper sources plus generated Ebitengine mobile view classes assembled by the Python pipeline.
  - [x] Keep durable Android wrapper sources, manifests, generated-AAR instructions, and scripts in tracked Apparat-owned paths.
  - [ ] Keep `third_party/salvagecore` as reference-only and prove the final Android GUI path does not depend on it.
  - [x] Ensure the Android entrypoint starts the same `internal/app` runtime initialization path as Debian GUI startup.
  - [x] Ensure the Android UI uses the same HUD tab model, tab order, clickable tabs, disabled-placeholder states, runtime diagnostics, and logging behavior as Debian.
  - [ ] Add Android safe-area/status-bar/navigation-bar layout handling so the HUD is readable on phone screens.
  - [ ] Add Android scale/density handling so tab buttons and body text remain usable on Pixel-class devices.
  - [x] Remove fixed portrait orientation and allow the wrapper view to fill the Android screen in supported orientations.
  - [ ] Validate additional phone, tablet, portrait, landscape, keyboard, controller, process-liveness, and `last_run.log` behavior on real Android devices.
  - [x] Capture visual evidence that the Phase 4 HUD renders on Android like it does on Debian.
- [x] Add Android build tests.
  - [x] Unit-test Android artifact path selection.
  - [x] Unit-test that Android supports only the `apparat` target in this phase.
  - [x] Unit-test Android `apparatd` rejection.
  - [x] Unit-test Android unsupported ABI rejection.
  - [x] Unit-test preflight failure reporting for missing SDK/JDK/NDK/tooling.
  - [x] Unit-test `--print-path` for Android APK output.
  - [x] Unit-test that the Android pipeline does not reference `third_party/salvagecore`.
  - [ ] Add an optional integration test target that installs and launches the APK when Android tools and a device/emulator are available.
- [x] Document Android build and troubleshooting.
  - [x] Update root `README.md` with Android APK build commands, prerequisites, artifact path, and support caveats.
  - [x] Update `scripts/README.md` with Android build options, preflight, outputs, side effects, and common failures.
  - [x] Update `cmd/apparat/README.md` for the Android manifest and gomobile hook.
  - [x] Update `docs/platform-matrix.md` with the exact Android evidence collected.
  - [x] Record the Android APK phase checkpoint in the journal and regenerate plan indexes.

**Exit criteria**

- [x] `python3 scripts/build.py --os android --arch arm64 --target apparat` produces `releases/android/arm64/apparat/latest.apk`.
- [x] `python3 scripts/build.py --os android --arch arm64 --target apparatd` fails clearly because Android headless is out of scope.
- [x] `--print-path` reports the Android APK path without building.
- [x] The APK installs and launches on at least one physical Android device.
- [x] The APK no longer fails modern Pixel install gates for obsolete SDK metadata, missing v2/v3 signing, or 16 KB page-size compatibility.
- [x] The Android wrapper GUI displays the Phase 4 tab HUD and supports touch/click tab selection on Android.
- [x] The Android wrapper app opens to the Apparat HUD instead of remaining on the Android splash/default icon.
- [x] The Android wrapper no longer forces portrait orientation and can fill the available screen in supported orientations.
- [ ] Android safe-area, density, and touch handling make the HUD readable and usable on a Pixel-class device.
- [x] Android startup creates a fresh `last_run.log` in the runtime root and exposes enough diagnostics to debug failures.
- [x] The Android build, tests, and documentation do not require `third_party/salvagecore`.
- [x] Documentation explains prerequisites, commands, artifact paths, validation evidence, and known limitations.

## Remaining Program Execution Contract

Phases 0–5 above are the completed foundation and historical implementation record. From Phase 6 onward, each phase is a concrete implementation program rather than a broad theme.

Every focused execution plan bound to a remaining phase must identify:

- User-visible outcome and explicit deferrals.
- Dependencies and must-resolve-before decisions.
- Domain authority and core-versus-presentation state ownership.
- SQLite migrations, compatibility, backup, and restart behavior.
- REST/OpenAPI, signed-envelope, authorization, limit, and audit changes.
- External, platform, provider, and transport adapters.
- GUI read models, commands, loading/empty/stale/offline/error states, focus, and input behavior.
- Failure, cancellation, retry, idempotency, recovery, and rollback behavior.
- Headless, display-free GUI, integration, target-platform, performance, and security evidence.
- Documentation updates and phase exit evidence.

Every functional phase ends in a reviewable vertical slice:

`core state -> SQLite -> command/query or REST boundary -> GUI projection -> failure/restart evidence`

Shared abstractions grow from completed slices. A phase may not claim support from compilation alone, replace durable state with channels or widget state, expose a provider's localhost endpoint, grant authority through cached records, or bypass Project-owner and queue-owner validation.

## Phase 6: Documentation, Evidence Reconciliation, And Decision Gates

**User outcome:** Contributors can begin backend implementation from one accurate architecture and roadmap without mistaking the GUI mockup or planned contracts for completed behavior.

**Dependencies:** Completed Phases 0–5 and the 2026-07-19 architecture review.

**Scope and deferrals:** This phase changes contracts, evidence labels, and executable planning only. It does not implement the backend features described by later phases.

### Documentation truth

- [ ] Audit every completed Phase 0–5 claim against executable evidence.
  - [ ] Verify the five canonical tab IDs and order against code and tests.
  - [ ] Reclassify unfinished focus, activation, back, context, command-palette, scrolling, input-equivalence, accessibility, configuration, offline, and recovery behavior as planned or validation-pending.
  - [ ] Distinguish implemented startup and persistence primitives from the planned shared-core package structure.
  - [ ] Preserve journal history; record corrections in current evidence rather than rewriting past checkpoints.
- [x] Keep one canonical product and contract vocabulary.
  - [x] GUI state never becomes headless-core state.
  - [x] Projects are owner-local Git repositories projected into one authorized cluster-wide catalog.
  - [x] Pipelines are Projects with Task entrypoints.
  - [x] Tasks may run manually with no trigger.
  - [x] Queue owners validate REST submissions; workers pull leases and return outcomes; only owners complete jobs authoritatively.
  - [x] Providers are static drivers, endpoints are stable service instances, models/features are capabilities, and remote peers use the Apparat gateway.
- [x] Mark every affected document section as implemented, partially validated, planned, deferred, or removed where ambiguity could mislead implementation.
- [ ] Add or retain automated consistency checks for tabs, shortcuts, artifact paths, package boundaries, service cardinality, OpenAPI links, and platform-support claims.

### Resolved architecture decisions

- [x] Record and test-plan the selected one-node process model.
  - [x] `apparat` and `apparatd` embed the same headless-capable core.
  - [x] They are alternative process forms of one node with one default runtime root, identity, SQLite database, service inventory, and artifact store.
  - [x] One exclusive runtime lock permits one authoritative process and SQLite writer.
  - [x] Simultaneous GUI/daemon use requires a later approved daemon-client mode.
  - [x] Independent nodes on one host require explicit roots, identities, ports, and service ownership.
- [x] Record the selected identity and TLS model.
  - [x] One cluster-local X.509 root fingerprint is verified out of band.
  - [x] One enrollment authority controls MVP issuance under that root.
  - [x] TLS leaf keys remain separate from Apparat Ed25519 device-signing keys.
  - [x] Signed device records bind both keys, certificate serial/fingerprint, WireGuard key, roles, scopes, validity, and status.
  - [x] Rotation and revocation update the binding rather than trusting a still-chain-valid leaf.
- [x] Record the signed-envelope model.
  - [x] RFC 8785 canonical UTF-8 JSON.
  - [x] Integer UTC millisecond timestamps.
  - [x] SHA-256 payload or canonical artifact-metadata hashes.
  - [x] Ed25519 signature by the Apparat device key over the envelope with the signature value omitted.
  - [x] Strict recipient, expiry, deadline, replay, idempotency, size, authorization, and schema validation.
- [x] Record owner authority.
  - [x] Project owners own Project Task definitions, trigger bindings, scheduling evaluation, and runs during the MVP.
  - [x] Queue owners own admission, ordering, attempts, leases, cancellation, results, retention, and audit.
  - [x] Cached directory, Project, service, queue, and result projections grant no authority.
- [x] Record service advertisement behavior.
  - [x] Owner-scoped monotonic revisions.
  - [x] Default 120-second lifetime and refresh by 60 seconds.
  - [x] Immediate routing exclusion on expiry.
  - [x] Stale diagnostic visibility for up to 24 hours.
  - [x] Fresh observation and newer revision required to become routable again.
- [x] Record credential and artifact boundaries.
  - [x] SQLite stores provider credential references, not provider secrets.
  - [x] Secret adapters use an OS credential store or Apparat-managed encrypted local secret file.
  - [x] Artifact bytes live in the owner runtime artifact store; SQLite holds metadata and lifecycle.
  - [x] HTTPS transfer is bounded, resumable, authenticated, and SHA-256 verified before atomic finalization.
  - [x] Queue completion waits for owner validation of the active lease and every artifact.

### Decision gates retained for later phases

- [x] Assign every unresolved decision to the phase that needs it.
  - [x] Phase 7: final migration/backup/restore procedure, lock recovery, clock/ID ports, WAL platform evidence, and whether optional at-rest database encryption is admitted.
  - [x] Phase 8: enrollment authorization vocabulary, invite recovery, endpoint-discovery seed, rate/size limits, and certificate rotation/revocation operations.
  - [x] Phase 9: directory conflict resolution and Project-summary authorization/freshness policy.
  - [x] Phase 10: supported file/binary limits, Task sandbox per platform, transaction conflicts, and artifact retention defaults.
  - [x] Phase 11: provider auto-discovery defaults, verified-service auto-promotion, credentials per platform, probe/admission defaults, and approved image providers.
  - [x] Phase 12: lease/heartbeat/retention defaults, artifact quotas, retry classes, and worker trust policy.
  - [x] Phase 13: routing score inputs, fallback policy, load freshness, and route-explanation schema.
  - [x] Phase 14: webhook authentication, schedule/timezone behavior, approval policy, and post-MVP scheduler failover.
  - [x] Phase 15: capture formats, device permissions, privacy retention, ASR/TTS adapter selection, and streaming behavior.
  - [x] Phase 16: release artifact hosting, tracked-binary replacement, signing custody, reproducibility threshold, update manifests, and rollback support.
  - [x] Post-MVP Track A: Meshtastic, Signal, app-managed WireGuard, ownership migration, replication, CRDT, and dynamic optimization.
  - [x] Post-MVP Track B: Comrades transport/privacy, grants, quotas, moderation, visibility, and abuse defaults.
  - [x] Post-MVP Track C: BOINC boundary, source dependencies, isolation, validation governance, gameplay, reputation, and anti-gaming.

**Exit criteria**

- [ ] No contradiction can change the identity, authority, cardinality, persistence, security, or process model of Phase 7 or 8.
- [ ] Current evidence and future requirements are visibly distinct.
- [ ] README and detailed contracts agree with this roadmap.
- [ ] `make check-docs`, plan-index checks, link/OpenAPI checks, and whitespace validation pass.

## Phase 7: Shared Core, SQLite, Lifecycle, And First Local-Service Slice

**User outcome:** GUI and headless artifacts operate the same durable backend state without putting presentation state into the core, and users can inspect a real core-backed local service inventory.

**Dependencies:** Phase 6.

**Deferred:** Enrollment, remote advertisement, real provider calls, distributed queues, routing, and Task triggers.

### Shared-core ownership and package seams

- [ ] Establish the real shared-core composition.
  - [ ] Keep executable entry points in `cmd/apparat` and `cmd/apparatd`.
  - [ ] Put mode-neutral orchestration in `internal/app`.
  - [ ] Put durable product rules and value types in `internal/domain`.
  - [ ] Put SQLite, provider, Git, filesystem, HTTP, and other external integrations in `internal/adapters`.
  - [ ] Put OS paths, locks, signals, credential stores, services, and lifecycle in `internal/platform`.
  - [ ] Keep Ebitengine/EbitenUI imports below the GUI adapter boundary.
- [ ] Remove duplicate backend ownership.
  - [ ] The core owns no HUD shell.
  - [ ] The GUI owns navigation, focus, selection, layout, scrolling, gestures, animation, modal state, and unsaved widget text.
  - [ ] Core state includes identity, backend configuration, Projects, durable drafts, Tasks, queues, jobs, services, capabilities, artifacts, retries, and cached remote state.
  - [ ] Boundary operations convert GUI/API actions into core commands and core read models/change notifications into projections.
- [ ] Implement the smallest command/query seam needed by the first slice.
  - [ ] Typed commands validate intent and durable invariants.
  - [ ] Queries return immutable read models.
  - [ ] Change notifications are hints to re-query, not durable truth.
  - [ ] Inject clock and ID sources into deterministic application logic.
  - [ ] Add abstractions only after a second slice proves a shared contract.

### One-node runtime and transactional lifecycle

- [ ] Migrate to one logical default node root for both artifacts.
  - [ ] Preserve or explicitly migrate current binary-specific roots.
  - [ ] Acquire an exclusive runtime lock before opening writable state.
  - [ ] Report the owning PID/artifact and safe recovery guidance on contention.
  - [ ] Require explicit roots and identities for independent same-host nodes.
- [ ] Make startup transactional.
  - [ ] Resolve configuration and paths.
  - [ ] Acquire the lock.
  - [ ] Start structured logging and panic capture.
  - [ ] Open and inspect SQLite.
  - [ ] Apply checksumed forward migrations.
  - [ ] Load/validate identity and secret adapters.
  - [ ] Construct repositories and managers.
  - [ ] Start supervised background work only after durable dependencies are ready.
  - [ ] Publish readiness last.
- [ ] Make shutdown idempotent and ordered.
  - [ ] Stop admission.
  - [ ] Cancel supervised work.
  - [ ] Flush durable state and logs.
  - [ ] Close managers, repositories, SQLite, and lock.
  - [ ] Preserve actionable `last_run.log` diagnostics on partial startup or shutdown failure.

### SQLite and recovery hardening

- [ ] Centralize database ownership behind repositories and transactions.
  - [ ] Define connection limits, busy timeout, foreign-key behavior, and retry classes.
  - [ ] Keep WAL opt-in until Linux, Windows, macOS, and Android behavior is evidenced.
  - [ ] Add cancellation-aware transactions and read models.
  - [ ] Preserve migration checksum mismatch diagnostics.
- [ ] Define and test backup, integrity, repair, restore, and rollback.
  - [ ] Never treat file copy during an active uncoordinated write as a valid backup.
  - [ ] Record schema/app version and integrity evidence with backups.
  - [ ] Test restore into a disposable root before replacing authoritative state.
  - [ ] Keep optional database encryption gated on key storage and recovery design.

### First durable local-service vertical slice

- [ ] Implement a statically registered mock provider driver.
  - [ ] Explicit factory registration at both composition roots.
  - [ ] Stable `ServiceID` and `CapabilityID` types.
  - [ ] Two configured instances of the same driver and workload class plus one different mock provider.
  - [ ] Independent desired enablement, observed health, inventory, concurrency, failure, and shutdown.
- [ ] Persist desired and observed service state separately.
  - [ ] Service-instance desired configuration table.
  - [ ] Observation table with safe errors and probe timestamps.
  - [ ] Capability table keyed by service and capability ID.
  - [ ] No remote advertisement in this phase.
- [ ] Expose one shared read model.
  - [ ] Headless query/diagnostic output lists every instance distinctly.
  - [ ] GUI Routing detail renders the same core projection.
  - [ ] GUI filter/selection/expansion remains GUI-owned.
  - [ ] Loading, empty, disabled, unhealthy, and error states are explicit.
- [ ] Prove restart and failure isolation.
  - [ ] Stable identities survive restart.
  - [ ] One failed instance does not hide or stop another same-provider instance.
  - [ ] Removing one instance cannot remove another by driver or workload key.

### GUI, test, and build foundation

- [ ] Route controller, keyboard, mouse, and touch through one application action model.
- [ ] Implement or correctly defer focus, activation, back, context, scrolling, modal, disabled-control, and accessibility semantics.
- [ ] Replace unconditional widget-tree rebuilds with reconciliation where measured evidence shows churn or state loss.
- [ ] Standardize loading, empty, stale, offline, unauthorized, retrying, cancelled, and failed data states.
- [ ] Persist only honest configuration with validation, save state, defaults, and secret references.
- [ ] Separate headless core tests, display-free GUI state/projection tests, and optional native display integration tests.
- [ ] Remove placeholder/sentinel packages only when real owners replace them; do not perform a broad speculative rewrite.
- [ ] Avoid a service-locator, generic event bus, or global registry that hides ownership; use explicit composition roots and narrow interfaces.
- [ ] Keep Ebitengine update/draw work bounded; move SQLite, network, provider probes, artifact I/O, and other blocking effects outside render/update paths.
- [ ] Harden existing sensitive foundations before exposing the network.
  - [ ] Strict identity parsing, validation, permissions, and atomic file replacement.
  - [ ] Recursive structured-log redaction for secrets and sensitive nested fields.
  - [ ] Updater trust, hash/signature, temporary-file, cancellation, and rollback boundaries.
- [ ] Separate fast host builds/tests from release orchestration.
- [ ] Stop mutating tracked or submodule source during builds; use ignored worktrees/caches or upstreamable patches.
- [ ] Establish startup, memory, SQLite, frame-time, and probe-concurrency budgets with measurements.

### Continuous target evidence

- [ ] Keep Linux GUI and headless builds passing through every later phase.
- [ ] Prove headless startup imports/initializes no GUI dependency.
- [ ] Continue Android validation from Phase 5.
  - [ ] Safe-area and density/readability on a Pixel-class device.
  - [ ] Touch, keyboard, controller, portrait, landscape, process-liveness, and runtime-path evidence.
  - [ ] Optional install/launch integration target when tools/device are available.
  - [ ] No second backend/service registry in the Android GUI.
- [ ] Carry unresolved input, accessibility, and platform behavior into the exact phase that supplies its real backend effect.

**Exit criteria**

- [ ] One shared core runs headlessly and behind the GUI.
- [ ] One runtime lock prevents competing writers/advertisements for one node.
- [ ] Core tests require no GUI, SQLite, network, real clock, or platform service unless explicitly integration-scoped.
- [ ] Two same-provider mock instances and another provider remain distinct through restart and failure.
- [ ] The GUI renders core service state without moving GUI state into the core.
- [ ] Lifecycle, backup/restore, diagnostics, and continuous target checks have reproducible evidence.

## Phase 8: Identity, Trusted Device Directory, Secure REST, And Reusable Mock Queue

**User outcome:** Two devices can enroll, authenticate, exchange durable work through the owner-authoritative REST protocol, disconnect or restart, and recover one logical result.

**Dependencies:** Phase 7 and Phase 6 identity/envelope contracts.

**Deferred:** Automatic endpoint discovery, Project workspaces, real inference providers, pools, and route selection.

### Network configuration and trusted directory

- [ ] Support externally managed WireGuard and trusted LAN through the same HTTPS API.
  - [ ] Detect expected WireGuard interfaces where possible.
  - [ ] Support explicit peer endpoint configuration for the first proof.
  - [ ] Treat discovery as advisory and never as trust.
- [ ] Add an authoritative durable trusted-device directory.
  - [ ] Device identity, TLS binding, WireGuard key, roles, scopes, endpoints, revision, validity, and status.
  - [ ] Signed/cached peer records for offline degradation.
  - [ ] Explicit revoked, expired, stale, unavailable, and key-changed states.
  - [ ] No authority from a cached record after revocation/expiry.

### Enrollment and mTLS

- [ ] Generate a short-lived QR/invite.
  - [ ] Cluster root fingerprint, one-time token, intended role/scopes, endpoint hints, and expiration.
  - [ ] Mutual human confirmation before trust is recorded.
  - [ ] Device profile, signing-key proof, TLS key/CSR proof, and WireGuard binding.
  - [ ] Token single use, expiration, revocation, and audit.
- [ ] Issue and validate TLS certificates.
  - [ ] Chain to the verified cluster root.
  - [ ] Require current signed device-record binding.
  - [ ] Separate leaf and Apparat signing keys.
  - [ ] Test issuance, rotation, revocation, lost-device recovery, and key mismatch.
  - [ ] Disable mutating TLS 0-RTT.

### Signed REST foundation

- [ ] Implement versioned authenticated REST resources.
  - [ ] Health, version, readiness, and clock state.
  - [ ] Device profile and trusted-directory projection.
  - [ ] Safe aggregate capabilities.
  - [ ] Submit/read/cancel mock jobs.
  - [ ] Queue-owner submit, worker claim/long-poll, heartbeat, and complete.
  - [ ] Cursor-based event polling.
  - [ ] Owner-local Project and Task resources remain placeholders until Phases 9–10.
- [ ] Enforce OpenAPI schemas, content types, body/artifact limits, deadlines, bounded concurrency, scopes, audit, and redaction-safe errors.
- [ ] Implement RFC 8785/SHA-256/Ed25519 envelopes.
  - [ ] Sign outgoing durable operations.
  - [ ] Validate canonical form, key binding, signature, recipient, hash, expiry, deadline, replay, authorization, size, and schema.
  - [ ] Apply duplicate messages idempotently and return prior durable outcomes where possible.

### Reusable mock-queue proof

- [ ] Persist requester outbound submission before network delivery.
- [ ] Have the queue owner authenticate, authorize, validate, and durably accept or reject.
- [ ] Return `202 Accepted` only after durable acceptance with resource location.
- [ ] Let an authorized worker poll/long-poll with current mock capabilities.
- [ ] Issue one bounded owner-created attempt, lease, fencing token, and deadline.
- [ ] Execute only under the active lease.
- [ ] Accept bounded heartbeat/progress according to policy.
- [ ] Post a signed terminal result/failure to the owner.
- [ ] Let the owner validate worker, lease, fencing, schema, idempotency, and result before authoritative completion.
- [ ] Support cancellation, timeout, retry, lease expiry, and clear incompatibility rejection.
- [ ] Reuse these identities and transitions in Phase 12; do not build a disposable echo-only protocol.

### GUI and recovery proof

- [ ] Show enrollment, fingerprint, trust, directory, connection, queue/job, attempt, and failure states in existing HUD surfaces.
- [ ] Keep invite text, selected peer, forms, focus, and navigation in GUI state.
- [ ] Use stable device, message, job, attempt, lease, and correlation IDs in API, SQLite, logs, events, and GUI.
- [ ] Demonstrate Steam Deck requester and headless worker.
  - [ ] Disconnect requester, owner, or worker.
  - [ ] Restart requester and owner independently.
  - [ ] Reconnect and retrieve the one authoritative result.
  - [ ] Reject duplicate delivery, stale completion, replay, and unauthorized work.

**Exit criteria**

- [ ] Two devices repeatedly complete the proof across restart and temporary disconnection.
- [ ] LAN/WireGuard presence alone grants no trust.
- [ ] Duplicate delivery cannot duplicate the logical job.
- [ ] Only the queue owner records authoritative acceptance and completion.
- [ ] The mock queue is a reusable foundation for Phase 12.

## Phase 9: Discovery, Presence, Project Registry, And Cluster-Wide Project Catalog

**User outcome:** Every enrolled device shows every Project it is authorized to discover across the cluster, with correct owner, freshness, and offline state.

**Dependencies:** Phase 8 secure REST and trusted directory.

**Deferred:** File/Git operations, Project mutation, Pipeline execution, and full artifacts until Phase 10.

### Endpoint discovery and presence

- [ ] Add endpoint discovery after the explicit Phase 8 seed.
  - [ ] Discovery records are suggestions until matched to an authorized device record and mTLS identity.
  - [ ] Preserve explicit endpoint configuration as fallback.
  - [ ] Record last success, failure, clock state, revision, and availability safely.
- [ ] Define and implement directory conflict behavior.
  - [ ] Newer valid signed revision supersedes older state.
  - [ ] Revocation and key-change conflicts fail closed.
  - [ ] Cached offline records remain visible but cannot authorize.
  - [ ] Surface actionable conflict and recovery information.

### Owner-local Project registry

- [ ] Register existing filesystem/Git folders.
  - [ ] Assign stable Project ID and one owner device.
  - [ ] Validate canonical safe roots and path traversal protection.
  - [ ] Store owner-local metadata and revision in SQLite.
  - [ ] Treat the device holding/running the Git working tree as authoritative.
  - [ ] Do not expose owner-local filesystem paths remotely.
- [ ] Produce signed authorization-filtered Project summaries.
  - [ ] Project and owner IDs, display metadata, Pipeline/task presence summary, capabilities, revision, observation/expiry, and availability.
  - [ ] Separate permission to discover from file, Git, Task, mutation, artifact, chat, and secret permissions.

### Cluster-wide catalog projection

- [ ] Merge local Projects with authorized remote summaries from every owner.
- [ ] Cache remote summaries with signer, revision, freshness, and expiry.
- [ ] Keep offline Projects visible as stale/unavailable according to policy.
- [ ] Never relabel a cached remote Project as local or authoritative.
- [ ] Route Project detail reads to the owner through REST.
- [ ] Prevent one device from republishing another owner's summary as its own.
- [ ] Reconcile additions, updates, authorization loss, owner revocation, and deletion idempotently.

### GUI and evidence

- [ ] Replace mock Project list data with a core read model.
- [ ] Show local/remote owner, online/stale/unavailable, authorization, revision, and last observation.
- [ ] Keep filter, selection, expansion, sorting, focus, and unsaved search text in GUI state.
- [ ] Provide loading, empty, partial-cluster, offline, unauthorized, conflict, and retry states.
- [ ] Test at least two owners, duplicate display names, offline cache, revoked access, stale revision, and restart recovery.

**Exit criteria**

- [ ] Every device presents the same authorized cluster-wide Project set modulo current authorization and freshness.
- [ ] Each Git repository remains authoritative only on its owner.
- [ ] Remote reads go to the owner; no remote filesystem or SQLite access exists.
- [ ] Cached summaries remain useful offline without granting authority.

## Phase 10: Project Workspaces, Git, Pipelines, And Constrained Manual Tasks

**User outcome:** A user can open a real local or remote Project, perform safe repository work, manage durable drafts/artifacts, and run a Project entrypoint manually without a trigger or unrestricted shell.

**Dependencies:** Phase 9 Project ownership/catalog and Phase 8 secure REST.

**Deferred:** Queue-backed Task steps until Phase 12, route selection until Phase 13, and schedules/webhooks/events until Phase 14.

### Files and Git

- [ ] Add constrained owner-served filesystem resources.
  - [ ] Browse directories and view supported text.
  - [ ] Create/edit/rename/move/delete with explicit scope and confirmation.
  - [ ] Protect roots, symlinks, traversal, binary files, large files, and races.
  - [ ] Preserve unsaved/offline drafts independently from authoritative files.
- [ ] Add safe Git operations.
  - [ ] Status, diff, stage, unstage, scoped commit.
  - [ ] Branch list/switch, history, and commit detail.
  - [ ] Conflict visualization and explicit resolution transactions.
  - [ ] No shell escape or arbitrary command construction.

### Chats, transactions, drafts, and artifacts

- [ ] Add Project chats.
  - [ ] Chat list/history and prompt editor.
  - [ ] Message, job, route, result, and artifact relationships.
  - [ ] Offline pending messages with durable state.
- [ ] Add owner-authoritative idempotent Project transactions.
  - [ ] Stable transaction and base-revision identity.
  - [ ] Authorization, validation, conflict checks, and owner apply.
  - [ ] Durable rejection reason and editable content.
  - [ ] Revise, discard, retry, and duplicate replay.
- [ ] Implement artifact lifecycle.
  - [ ] Owner metadata, SHA-256, size, MIME type, provenance, authorization, and retention.
  - [ ] Bounded authenticated upload/download with ranges/chunks.
  - [ ] Temporary partial state, resume, digest verification, and atomic finalization.
  - [ ] Explicit expired/deleted/corrupt/unavailable states.
  - [ ] Cleanup without silently invalidating successful authoritative results.

### Pipelines and manual Tasks

- [ ] Derive Pipeline status from one or more enabled Apparat Task entrypoints.
  - [ ] No separate Pipeline owner, repository, scheduler, or identity.
- [ ] Define versioned Task entrypoints.
  - [ ] Stable Task and Project IDs.
  - [ ] Typed inputs/outputs and schema version.
  - [ ] Project-owner authority and revision.
  - [ ] Permissions, secret references, timeouts, retries, approvals, and enabled state.
  - [ ] Durable run/correlation/idempotency identity and history.
- [ ] Permit manual execution with zero trigger bindings.
  - [ ] Owner-local allowlisted application actions.
  - [ ] Project-scoped filesystem/Git operations.
  - [ ] Explicit mock/local executor calls.
  - [ ] Configured human approval for consequential actions.
  - [ ] No unrestricted remote shell, generic process endpoint, or arbitrary tool execution.
- [ ] Keep trigger bindings as separate planned records; a manual action is not a persistent trigger.
- [ ] Expose authorized Task summaries and manual run requests through the Project owner's REST API.

### GUI, recovery, and safety

- [ ] Wire workspace buttons to real routes/commands or visibly disable them with a reason.
- [ ] Preserve editor focus, selection, scroll, panes, unsaved buffers, and conflict UI as presentation state.
- [ ] Show transaction pending/accepted/rejected/conflicted, Task approval/running/cancelled/failed/completed, and artifact transfer states.
- [ ] Recover drafts, transactions, Task runs, and transfers across restart.
- [ ] Redact file contents, prompts, chat bodies, secrets, and artifact bytes from default logs.
- [ ] Test authorization distinctions among discovery, read, Git, Task invocation, mutation, artifact, chat, and secret scopes.

**Exit criteria**

- [ ] A Steam Deck can operate a real Project locally or through its owner without arbitrary shell access.
- [ ] Offline drafts and rejected/conflicted transactions remain recoverable and editable.
- [ ] A Project with an entrypoint appears as a Pipeline.
- [ ] The same owner-defined Task runs manually with no trigger.
- [ ] Artifact transfers resume and verify integrity.

## Phase 11: Multi-Instance Local Inference Drivers, Health, Capabilities, And Advertisements

**User outcome:** One Apparat node can discover, configure, supervise, inspect, and safely advertise any number of local inference services, including duplicates of the same provider.

**Dependencies:** Phase 7 shared core/service slice, Phase 8 identity/gateway, and Phase 6 service contracts.

**Deferred:** Distributed queue execution to Phase 12 and route/pool selection to Phase 13.

### Workload and driver contracts

- [ ] Establish the versioned workload-class registry.
  - [ ] `text_generation`, `image_generation`, `video_generation`, `speech_to_text`, `text_to_speech`, and `research_boinc`.
  - [ ] Extension rules for embeddings, reranking, classification, vision analysis, audio generation, simulation, compilation, and future types.
  - [ ] Keep workload class independent from driver, service, model, Project, and queue identity.
- [ ] Implement static provider driver registration.
  - [ ] Typed factory, configuration validation, instance, inspection, executor, progress, result, and error contracts.
  - [ ] Explicit composition-root registration in GUI and headless artifacts.
  - [ ] Workload-specific or deliberately tagged/versioned requests and results, never unbounded generic maps.
  - [ ] No Go dynamic `plugin.Open` or hidden package-global registration.
  - [ ] Defer authenticated out-of-process extension IPC until a concrete third-party requirement exists.

### Arbitrary service instances and persistence

- [ ] Generalize the Phase 7 manager.
  - [ ] Stable `ServiceID` independent from driver/workload.
  - [ ] Arbitrary same-driver and same-workload instances.
  - [ ] Primary index by ServiceID; secondary driver/class/model/health/policy indexes.
  - [ ] Independent probing, health, admission semaphore, concurrency, cancellation, retry classification, failure isolation, refresh, and shutdown.
- [ ] Implement desired/observed/capability persistence.
  - [ ] Desired endpoint, enablement, advertise policy, provider configuration, credential reference, admission, concurrency, and revision.
  - [ ] Observed lifecycle, health, safe failure, availability, load, inventory hash, and probe time.
  - [ ] Capabilities with stable IDs, workload/schema, model, modality, format, features, limits, artifact/progress/cancellation support, and observation time.
  - [ ] Provider secrets never enter general JSON, advertisements, logs, or GUI read models.
- [ ] Implement secret resolution.
  - [ ] OS credential store when supported.
  - [ ] Apparat-managed encrypted local secret file fallback.
  - [ ] Missing/locked/rotated/revoked secret states without disclosing values.

### Discovery, verification, and supervision

- [ ] Probe approved provider defaults and explicit endpoints with bounded concurrency and deadlines.
- [ ] Validate provider identity/protocol before creating a verified service.
- [ ] Implement `discovered -> verified -> enabled -> advertised`.
- [ ] Require explicit enablement/advertisement unless an approved policy auto-promotes a verified known service.
- [ ] Refresh health/inventory without blocking unrelated instances.
- [ ] Recover desired state after restart without duplicate instances or advertisements.
- [ ] Keep providers as separately supervised processes/services; do not load model runtimes into the HUD.

### Safe advertisements and gateway

- [ ] Derive signed service/capability advertisements from desired policy and safe observed state.
  - [ ] Owner/device, ServiceID, DriverKind, display name, safe health/availability/concurrency/policy.
  - [ ] CapabilityID, workload/schema, model, modalities, formats, features, and limits.
  - [ ] Monotonic revision, observed time, 120-second expiry, and refresh by 60 seconds.
  - [ ] No provider endpoint, secret reference, token, prompt, result, or raw failure.
- [ ] Cache remote advertisements safely.
  - [ ] Reject older revisions.
  - [ ] Make expired services immediately non-routable.
  - [ ] Show stale diagnostics for up to 24 hours.
  - [ ] Require fresh newer revision to restore eligibility.
- [ ] Expose logical service/capability resources through authenticated Apparat REST.
- [ ] Apply gateway authorization, policy, limits, audit, and queue requirements before provider invocation.

### Provider sequence

- [ ] Retain the mock driver as conformance coverage.
- [ ] Add OpenAI-compatible text.
  - [ ] Chat/completion, streaming/non-streaming, timeouts, cancellation, usage/error normalization, and results/artifacts.
- [ ] Add Ollama.
  - [ ] Model inventory, generation, authorized pull/install state, health, and cancellation.
- [ ] Add llama.cpp.
  - [ ] External service process, inventory, generation normalization, health, cancellation, and platform acceleration evidence.
- [ ] Add approved image providers through the same registry.
  - [ ] Automatic1111 and/or ComfyUI.
  - [ ] Multiple simultaneous providers and same-provider instances.
  - [ ] Text-to-image/image-to-image, dimensions, formats, model, sampler, progress, cancellation, previews, artifacts, and metadata.
- [ ] Define video contract without claiming a concrete adapter.
  - [ ] Text/image-to-video, duration, dimensions, frame rate, format, model, progress, cancellation, checkpoint, artifacts, and storage limits.
- [ ] Define STT/TTS contracts; concrete adapters remain Phase 15.
- [ ] Define BOINC capability contract; execution remains Post-MVP Track C.

### GUI and evidence

- [ ] Render every service instance distinctly with stable name/ID, driver, health, capabilities, desired enablement, advertise policy, revision, expiry, and safe failure.
- [ ] Support independent enable/disable/advertise actions through core commands.
- [ ] Keep filtering, selection, expansion, form text, and layout in GUI state.
- [ ] Test two same-provider instances plus another provider through restart, independent failure, removal, secret failure, inventory change, expiry, and shutdown.
- [ ] Measure bounded probe and refresh resource use on supported platforms.

**Exit criteria**

- [ ] One node manages and advertises several services, including two same-provider/same-workload instances, without identity collision or shared failure.
- [ ] GUI and headless surfaces use the same manager and read models.
- [ ] Remote peers see only safe logical identities and cannot invoke localhost providers directly.
- [ ] Desired, observed, capability, and advertisement state remain separate and restart-safe.

## Phase 12: Authoritative Queue Protocol, Worker Leasing, Results, Artifacts, And Recovery

**User outcome:** Typed jobs can be submitted to their queue owner, leased by an eligible worker, completed with verified results/artifacts, and recovered without duplicate authoritative completion.

**Dependencies:** Phase 8 reusable mock queue, Phase 10 artifacts, and Phase 11 capabilities.

**Deferred:** Pool choice, route profiles, and fallback policy to Phase 13.

### Queue definitions and admission

- [ ] Implement direct-device, pool-owner, single-class, and explicit multi-class queue definitions.
- [ ] Persist owner, membership policy, workload allowlist, priorities, deadlines, concurrency, quotas, retries, retention, and enabled state.
- [ ] Direct every remote submission to the queue owner by authenticated REST.
- [ ] Validate authentication, scope, idempotency, schema, workload/capability requirements, policy, quota, size, deadline, retention, and queue state before durable accept/reject.
- [ ] Keep requester outbound state and authorized cached status/result separate from owner authority.

### Worker pull and lease lifecycle

- [ ] Let authorized workers claim/long-poll with current device/service/capability identity, accepted classes, availability, and bounded wait.
- [ ] Select only compatible work and issue owner-created attempt, lease, fencing token, deadline, and bounded payload/artifact references.
- [ ] Accept heartbeat/progress only under active policy, worker, and fencing token.
- [ ] Expire leases safely and permit reassignment without erasing attempt history.
- [ ] Reject stale, late, replayed, duplicated, mismatched, superseded, or unauthorized worker operations.
- [ ] Do not push unleased work into worker memory/databases or let workers read replicated queue rows as assignments.

### Execution, results, and artifacts

- [ ] Resolve leased logical service/capability identity through the local manager.
- [ ] Apply per-service admission, cancellation, deadline, failure isolation, and safe error normalization.
- [ ] Post signed terminal success/failure to the owner.
- [ ] Bind worker, queue, job, attempt, lease, fencing, schema, idempotency, and artifact hashes.
- [ ] Transfer bounded artifacts through authenticated resumable HTTPS.
- [ ] Verify authorization, provenance, size, MIME policy, and SHA-256 before finalization.
- [ ] Record authoritative completion exactly once only after every required validation succeeds.
- [ ] Preserve explicit partial, corrupt, expired, missing, rejected, cancelled, timed-out, and retryable states.

### Recovery and observability

- [ ] Recover requester outbox, owner queue, leases, attempts, cancellation intent, results, and transfers after restart.
- [ ] Reconcile worker completion after an ambiguous network failure idempotently.
- [ ] Bound retries/backoff and honor deadlines/cancellation.
- [ ] Retain attempt history and user-readable safe failure reasons.
- [ ] Trace stable job/correlation/attempt/lease/service/capability/artifact identity across SQLite, REST, events, logs, and GUI.
- [ ] Provide health and diagnostics without exposing prompts, results, secrets, endpoints, or unrelated jobs.

### GUI and tests

- [ ] Replace mock queue/job data with core read models.
- [ ] Show owner, requester, workload, route requirement, queue, attempts, lease, progress, cancellation, retry, result/artifacts, retention, and failure.
- [ ] Support loading, empty, offline, unauthorized, incompatible, waiting, leased, retrying, cancelled, failed, expired, and completed states.
- [ ] Test duplicate submission, owner/worker/requester restart, lease expiry, reassignment, late result, corrupted artifact, authorization loss, and cancellation races.
- [ ] Run concurrency, race, bounded-resource, and database-recovery tests.

**Exit criteria**

- [ ] Requesters submit only to owners and workers pull only owner-issued leases.
- [ ] A worker result becomes authoritative only after owner validation.
- [ ] Restart, reassignment, duplicate delivery, and ambiguous completion never double-complete a logical job.
- [ ] Artifacts resume and validate before accepted completion.
- [ ] The Phase 8 mock flow uses the same production queue primitives.

## Phase 13: Pools, Routing Profiles, Deterministic Fallback, And Real Text Generation

**User outcome:** A Project or Task can request typed work, understand why a destination was chosen or excluded, fall back deterministically, and retrieve a real text-generation result.

**Dependencies:** Phases 9–12.

### Pools

- [ ] Implement pool membership and one owner per pool queue.
- [ ] Support heterogeneous members while leasing only to currently compatible services.
- [ ] Persist membership revision, authorization, workload allowlist, priorities, availability, and owner policy.
- [ ] Have members pull leases and return signed results to the pool owner.
- [ ] Preserve owner validation and authoritative completion.
- [ ] Handle member offline, stale advertisement, removal, revocation, and partial capability change.

### Routing profiles

- [ ] Implement profiles for Project defaults, chats, Task/workflow steps, and explicit user overrides.
- [ ] Match:
  - [ ] Required workload class and schema.
  - [ ] Optional/required driver, ServiceID, CapabilityID, model/Project, modality, format, hardware, and features.
  - [ ] Privacy and authorization boundary.
  - [ ] Deadline, timeout, priority, cost/resource policy, and artifact requirements.
  - [ ] Current advertisement revision/expiry, health, admission, availability, and queue policy.
- [ ] Keep explicit service targeting distinct from provider kind.
- [ ] Apply ordered deterministic fallback.
- [ ] Reject unsupported or expired destinations before execution.
- [ ] Return a stable route explanation.
  - [ ] Selected queue/device/service/capability/model and matched requirements.
  - [ ] Every excluded candidate and safe reason.
  - [ ] Fallback order and transition reason.

### First real inference slice

- [ ] Submit a real OpenAI-compatible text-generation job from a Project/chat/Task through an explicit profile.
- [ ] Route through an owner-authoritative direct or pool queue.
- [ ] Lease to an eligible worker and resolve a local service through the gateway.
- [ ] Stream or poll bounded progress according to the typed contract.
- [ ] Persist normalized result, usage, errors, and artifacts.
- [ ] Survive owner, requester, or worker restart and deterministic fallback.
- [ ] Retrieve the one authoritative result through Project/job UI.
- [ ] Route mock image, video, STT, TTS, and BOINC jobs only to matching advertised capabilities.
- [ ] Fail unsupported workload classes and requirements clearly before execution.

### GUI, observability, and budgets

- [ ] Implement route/profile editor with honest validation and save states.
- [ ] Show selected destination, exclusions, fallback, health/expiry, queue state, and result linkage.
- [ ] Keep editor focus, filter, selected candidate, expanded rationale, and unsaved changes in GUI state.
- [ ] Redact prompts and model output from default logs while retaining correlation and safe metrics.
- [ ] Measure route-query latency, SQLite query cost, memory, network payloads, service admission, and fallback behavior.

**Exit criteria**

- [ ] A real text request completes through an explicit owner-authoritative route and survives retry/restart.
- [ ] Fallback is deterministic and explainable.
- [ ] No job reaches an incompatible, expired, unauthorized, disabled, or unavailable service.
- [ ] Remote execution never discloses or bypasses provider-local endpoints/credentials.

## Phase 14: Task Triggers, Automation, Webhooks, Approvals, And Durable Workflows

**User outcome:** The same Project Task entrypoint can run manually or through authorized triggers, survive restart, await queued work, and produce one auditable outcome.

**Dependencies:** Phase 10 Task entrypoints and Phases 12–13 queue/routing.

**Deferred:** Scheduler failover/election to Post-MVP Track A.

### Trigger bindings

- [ ] Add separate bindings for:
  - [ ] Interval and hourly/daily/weekly/monthly or cron-like schedules.
  - [ ] Authenticated webhooks.
  - [ ] Internal application events.
  - [ ] Cluster device/service/queue events.
- [ ] Keep manual execution valid without a binding.
- [ ] Let trigger delivery request a run at the Project owner; do not move Task authority.
- [ ] Persist binding identity, Task/Project owner, revision, timezone/schedule, authentication policy, enabled/paused state, and last/next evaluation.
- [ ] Bound, authenticate, authorize, replay-protect, rate-limit, and audit webhooks.

### Durable workflow execution

- [ ] Extend Task definitions with versioned steps, typed inputs/outputs, permissions, retries, timeouts, approvals, and route references.
- [ ] Persist run/correlation/idempotency ID, initiating actor/trigger, definition revision, current step, checkpoints, await state, submitted jobs, approvals, retry history, outputs, and terminal outcome.
- [ ] Resume idempotently after application/device restart.
- [ ] Reconcile ambiguous job submission and completion through durable IDs.
- [ ] Support cancellation, pause, timeout, retry, and explicit failure.
- [ ] Keep one authoritative scheduler/evaluator at the Project owner during the MVP.

### Safe actions and approvals

- [ ] Permit only allowlisted application actions, Project-scoped file/Git operations, and explicit typed service calls.
- [ ] Resolve secrets from references at execution time.
- [ ] Require configured human approval before consequential steps.
- [ ] Persist approval request, actor, scope, decision, expiry, and resulting transition.
- [ ] Expose no unrestricted remote shell, arbitrary process execution, or generic tool endpoint.
- [ ] Redact sensitive inputs/outputs while preserving safe audit evidence.

### GUI and diagnostics

- [ ] Show Task definition revision, trigger bindings, next/last run, current step, waits, approvals, jobs, retries, result, and failure.
- [ ] Support manual run independently from trigger management.
- [ ] Show offline owner, paused binding, missed schedule policy, webhook denial, approval expiry, recovery, and cancellation states.
- [ ] Preserve forms, filters, focus, selection, and unsaved edits as GUI state.
- [ ] Provide a redacted run timeline linking trigger, command, queue, job, attempt, artifact, approval, and result.

**Exit criteria**

- [ ] One Project-owner Task runs manually or through every supported authorized binding without changing identity/ownership.
- [ ] A triggered Task submits inference, awaits the result, survives restart, resumes idempotently, and produces an auditable outcome.
- [ ] Consequential actions honor approval and sandbox boundaries.
- [ ] Duplicate triggers and webhooks do not duplicate a logical run.

## Phase 15: ASR, TTS, Push-To-Talk, Audio Lifecycle, And Privacy

**User outcome:** Holding and releasing the configured control produces editable transcribed text through a selected local or remote route, and spoken output can be controlled independently.

**Dependencies:** Display-free input contract and Phases 11–14.

### Audio capture and presentation boundary

- [ ] Start capture while `R2` or configured Debian right `Ctrl` is held.
- [ ] Stop and submit on release.
- [ ] Let `Escape` cancel a held right-`Ctrl` recording without submitting on later release.
- [ ] Bound duration, memory, sample format, and temporary storage.
- [ ] Delete temporary audio according to explicit privacy/retention policy.
- [ ] Keep capture device, buffer, held/cancelled state, focused field, and playback controls in GUI/platform state.
- [ ] Create durable core state only on explicit transcription or synthesis submission.

### Speech services and routing

- [ ] Add speech-to-text adapters.
  - [ ] Local whisper.cpp first.
  - [ ] Remote STT through the same service/queue/gateway model.
  - [ ] Language, model, timestamps, streaming policy, progress, timeout, retry, and cancellation.
- [ ] Route by explicit global, Project, chat, or focused-context profile.
- [ ] Add transcription outcomes.
  - [ ] Populate focused text field.
  - [ ] Open command-palette intent.
  - [ ] Submit a prompt only when explicitly configured.
  - [ ] Require review/edit before consequential actions unless policy explicitly approves direct behavior.
- [ ] Add text-to-speech independently.
  - [ ] OS-native or lightweight service adapter first.
  - [ ] Local/remote typed routes independent from ASR.
  - [ ] Voice, language, streaming, format, play, pause, stop, and interruption.
  - [ ] Qwen3-TTS only after focused research and adapter approval.

### Privacy, GUI, platform, and recovery

- [ ] Show recording, cancellation, upload/queue, route/device, transcription, retry, failure, completion, playback, and retention states.
- [ ] Never log raw recordings, transcripts, prompts, generated speech, provider responses, or credentials by default.
- [ ] Authorize audio artifacts and delete them according to policy.
- [ ] Recover submitted jobs/results after restart without pretending an in-memory capture can resume.
- [ ] Validate microphone/audio permissions, interruption, device change, background/lifecycle behavior, and cancellation separately on Linux/Steam Deck, Windows, macOS, and Android where supported.
- [ ] Test the state machine with fake audio before requiring real-device integration.

**Exit criteria**

- [ ] `R2` and right `Ctrl` produce editable transcription through a selected local or remote route.
- [ ] Spoken output routes and controls independently.
- [ ] Capture and focus remain presentation/platform state until explicit submission.
- [ ] Voice behavior is visible, cancellable, bounded, privacy-preserving, and restart-honest.

## Phase 16: Packaging, Release Hardening, And Platform Support Evidence

**User outcome:** Users can install, run, diagnose, upgrade, and roll back honestly supported GUI/headless artifacts on each declared platform.

**Dependencies:** Stable secure vertical slice, shared-core lifecycle, and continuous platform evidence from Phases 7–15.

**Scope:** This phase culminates platform validation; it is not the first platform test.

### Steam Deck and Linux GUI

- [ ] Controller database/mappings and complete application-action equivalence.
- [ ] Debian keyboard text precedence, mouse/touchpad, right-`Ctrl` PTT, and configurable binding conflicts.
- [ ] Settings for bindings and HUD aesthetics, including fonts, glyphs, and panel/button colors.
- [ ] Gamescope/fullscreen/window, 1280x800, Hi-DPI, `Steam+X` keyboard, microphone, audio, storage, HTTPS, and external WireGuard.
- [ ] Packaging, desktop integration, diagnostics, update, and rollback.

### Linux headless

- [ ] User/system service installation choice.
- [ ] CLI, authenticated API, health, service-manager operations, signals, restart, logs, doctor, permissions, and no display dependency.
- [ ] Exclusive node ownership or approved daemon-client behavior.
- [ ] Local inference provider supervision and same-provider failure isolation.
- [ ] Upgrade/rollback without corrupting migrations, identity, queues, services, artifacts, or active work.

### Windows

- [ ] Build/package GUI and headless artifacts.
- [ ] Signing, runtime paths, service/tray choice, certificate/secret store, firewall/HTTPS, controller, audio, provider supervision, and external WireGuard.
- [ ] Install, upgrade, rollback, diagnostics, and target-specific behavior evidence.

### macOS

- [ ] Build/package GUI and headless where supported.
- [ ] Signing, notarization, app lifecycle/sandbox, keychain/certificates, controller, microphone/audio, provider supervision, and external WireGuard.
- [ ] Install, upgrade, rollback, diagnostics, and target-specific behavior evidence.

### Android

- [ ] Continue from Phase 5 and continuous Phase 7+ validation.
- [ ] Release signing, versioning, provenance, icons, store/distribution, signed update manifest, installed-versus-latest display, and rollback policy.
- [ ] Additional ABI only after arm64 proof.
- [ ] Phone/tablet, portrait/landscape, safe-area, density, keyboard, controller, touch, microphone, audio, scoped storage, background, battery, and network evidence.
- [ ] Keep Android headless out of scope unless a Termux/service-worker plan is approved.
- [ ] Keep external WireGuard first; app-managed VPN remains Track A.
- [ ] Do not imply local provider supervision from GUI APK support without provider-specific evidence.

### Release engineering

- [ ] Keep fast host build/test separate from the release matrix.
- [ ] Stop mutating tracked/submodule source during builds.
- [ ] Decide whether Git-tracked large binaries move to immutable release assets plus a signed latest manifest.
- [ ] Record version, source commit, dependency/toolchain locks, target, hash, size, signer, provenance, SBOM where required, and reproducibility evidence.
- [ ] Define signing-key custody, rotation, compromise response, and platform verification.
- [ ] Define compatible schema/protocol upgrade order and rollback limits.
- [ ] Test clean install, upgrade, failed upgrade, rollback, backup/restore, and offline diagnostics.
- [ ] Publish an honest support matrix; compilation is not support.

**Exit criteria**

- [ ] Steam Deck/Linux GUI and headless packages pass installation and lifecycle evidence.
- [ ] Every other platform is marked supported only after its independent build, packaging, input, storage, networking, audio, security, provider, and lifecycle checks pass.
- [ ] Release artifacts are reproducible or carry a documented variance, provenance, signer, hashes, and rollback policy.
- [ ] Upgrades preserve or safely migrate authoritative state.

## Post-MVP Track A: Alternative Transports And Long-Term Resilience

**Independence:** This track depends on stable identity, envelopes, queues, authorization, advertisements, and transport conformance. Tracks B and C do not depend on its completion unless they explicitly select one of its transports.

- [ ] Add transport conformance tests.
  - [ ] Identity/authorization, envelope integrity, replay/duplicate, expiry, acknowledgement, fragmentation, store-and-forward, ordering, and payload/attachment limits.
- [ ] Research and implement Meshtastic.
  - [ ] Select maintained protobuf/client source.
  - [ ] Define compact application port/types, allowlisted commands/status, fragmentation/reassembly, acknowledgement/retry, authorization, replay protection, and size rejection.
- [ ] Research and implement a maintainable Signal gateway.
  - [ ] Define account/device operation and map Signal identity to Apparat authorization.
  - [ ] Restrict to notifications, approvals, compact commands, and selected triggers.
  - [ ] Do not make Signal authoritative for queues or Projects.
- [ ] Add optional platform WireGuard management while preserving external tunnels.
  - [ ] Linux kernel/tools, Windows supported embedding, Apple Network Extension/WireGuardKit, and Android VPN-service adapters.
- [ ] Add resilience only with explicit ownership protocols.
  - [ ] Project-owner and queue-owner migration.
  - [ ] Task scheduler failover tied to Project authority.
  - [ ] Cluster-directory conflict handling and advanced replication.
  - [ ] Optional CRDT research for explicitly selected data.
  - [ ] Dynamic routing optimization after deterministic routing remains explainable.
- [ ] Preserve service expiry, gateway policy, workload authorization, owner authority, idempotency, and artifact limits on every transport.

**Exit criteria**

- [ ] Each adapter passes the common conformance suite for every operation it claims.
- [ ] Constrained transports reject unsupported/oversized work safely.
- [ ] Ownership migration never creates two authoritative owners.

## Post-MVP Track B: Comrades, Chat, And Shared Inference

**Independence:** This track requires stable identity, authorization, queues, multi-instance routing, audit, and one suitable authenticated transport. It does not require Track A in full.

- [ ] Define Comrades trust and social identity.
  - [ ] Invite, accept, reject, block, remove, reauthorize, key change, compromise recovery, direct/group membership, and topology privacy.
- [ ] Add direct and group chat.
  - [ ] Durable outbound/inbound identity, signatures, offline retry, attachments/artifact references, delivery/failure/read state where supported, and separation from Project chat.
- [ ] Define owner-controlled shared-compute grants.
  - [ ] Eligible comrades/groups, devices/pools, workloads, services/models, schedule, idle rules, priority, preemption, concurrency/rates, quota, visibility, retention, expiration, pause, revocation, and emergency stop.
- [ ] Implement owner-authoritative comrade queues.
  - [ ] Personal work higher priority by default.
  - [ ] Inference-only default.
  - [ ] Deny Projects, files, chats, secrets, tools, shell, administration, provider endpoints/credentials, and unrelated cluster state.
  - [ ] Persist admission, execution, result, quota, rejection, and revocation audit.
- [ ] Add Comrades HUD.
  - [ ] Contacts/groups, conversations, trust, shared queues, grant editor, usage/quota/priority/availability/audit, and emergency pause.
- [ ] Add safety and abuse controls.
  - [ ] Consent, authentication, allowlists, size/rate limits, malware/content boundaries, denial-of-service controls, moderation, and privacy-preserving dispute evidence.

**Exit criteria**

- [ ] A verified comrade submits authorized inference without receiving unrelated access.
- [ ] Owner work retains priority and the owner can inspect, pause, or revoke immediately.
- [ ] Shared inference exposes neither provider-local connection data nor personal cluster data.

## Post-MVP Track C: Research, BOINC, And Validation Gameplay

**Independence:** This track requires capabilities, scheduling, resource budgets, Task execution, audit, packaging, and safe workload isolation. It does not require Tracks A or B.

- [ ] Define Research trust and validation governance.
  - [ ] Project identity and authoritative metadata.
  - [ ] Technical, scientific, operator, security, legal, and reputation evidence.
  - [ ] Proposal, review, challenge, approval, suspension, removal, and revalidation.
  - [ ] Gameplay scores/popularity never replace legitimacy or security evidence.
- [ ] Define validation gameplay.
  - [ ] Discovery, evidence review/actions, progression, reputation, achievements, collective goals, anti-gaming, collusion, Sybil, misinformation, moderation, and evidence inspection outside gameplay.
- [ ] Select the BOINC boundary and sources.
  - [ ] Evaluate client control/RPC.
  - [ ] Decide external supervision versus selected embedding.
  - [ ] Select submodules only after approval.
  - [ ] Define attach/auth, fetch, pause/resume/detach/update, version, and Project compatibility.
- [ ] Define resource policy.
  - [ ] Opt-in devices/pools; CPU/GPU, memory, storage, bandwidth, power, battery, thermal, fan, schedule, quota, and metered-network limits.
  - [ ] Personal, Task, and comrade work outrank Research by default.
  - [ ] Pause on user activity, low battery, thermal pressure, metered network, or policy violation.
- [ ] Isolate BOINC.
  - [ ] Separate runtime/credentials.
  - [ ] Restrict Apparat data, identity, Projects, secrets, and networks.
  - [ ] Platform process/container/sandbox boundary.
  - [ ] Verify application signatures/provenance where supported.
  - [ ] Record Project/application/work-unit/result provenance.
- [ ] Add Research HUD and recovery.
  - [ ] Candidate/validated catalog, evidence, assignment, budgets, work units/progress, contributions/impact, failure/suspension/audit, and later gameplay.
  - [ ] Start/pause/resume/stop, restart recovery, outage/revocation, malicious/compromised state, and immediate owner shutdown.

**Exit criteria**

- [ ] A user opts a device into a validated BOINC Project with explicit budgets.
- [ ] Higher-priority owner/comrade work preempts Research by policy.
- [ ] Research is isolated from identities, Projects, queues, and secrets.
- [ ] Validation evidence and provenance remain auditable independently from gameplay.

## Cross-Cutting Requirements

Every focused implementation plan includes the applicable items below and its phase cannot exit without evidence.

- [ ] Security
  - [ ] Least privilege, explicit authorization, safe defaults, key rotation/revocation, secret redaction, no unrestricted remote execution.
  - [ ] Shared-compute grants never imply Project/file/secret/tool/shell/admin access.
  - [ ] Research remains isolated.
  - [ ] Provider-local endpoints and credentials remain local.
  - [ ] Task and artifact boundaries are explicit.
- [ ] Reliability
  - [ ] Stable IDs, idempotent operations, durable transitions, bounded retries, cancellation, deadlines, restart recovery.
  - [ ] Independent failure/admission for every service instance.
  - [ ] One authoritative owner/writer for each node root, Project, queue, Task, result, and artifact.
- [ ] Observability
  - [ ] Structured logs with stable component/event/command/correlation/error fields.
  - [ ] Queue/job/attempt/lease/service/artifact traces, health endpoints, safe diagnostics, and user-readable failures.
- [ ] Privacy
  - [ ] No raw prompts, model output, voice, private keys, tokens, chat bodies, Project contents, provider responses, or secrets in default logs.
  - [ ] Explicit ownership, visibility, retention, expiry, and deletion.
- [ ] Performance
  - [ ] Frame-time, memory, startup, SQLite, API, network, artifact, and constrained-transport budgets.
  - [ ] Bounded discovery/probe and per-service execution concurrency.
- [ ] Recovery
  - [ ] Database backup/integrity/restore, identity repair, migration compatibility, upgrade rollback, artifact verification, desired-service recovery, and no duplicate advertisement/execution.
- [ ] GUI
  - [ ] Core read models and commands; no durable widget state.
  - [ ] Loading, empty, stale, offline, unauthorized, pending, retrying, cancelled, failed, and recovery states.
  - [ ] Controller/keyboard/mouse/touch parity, visible focus, accessibility, and no enabled no-op controls.
- [ ] Platform
  - [ ] Continuous Linux GUI/headless and relevant Android evidence.
  - [ ] Target-specific storage, locks, credentials, networking, provider supervision, input, audio, lifecycle, packaging, and update evidence before support claims.
- [ ] Documentation
  - [ ] README for product/operations changes.
  - [ ] API/OpenAPI, architecture, database, security, signed-envelope, transport, controller, and platform contracts when affected.
  - [ ] Closest directory README for changed code/scripts/tools/build/tests.
  - [ ] Useful script `--help`, prerequisites, side effects, outputs, and failures.
  - [ ] Focused plan binding, compatibility, failure, rollback, verification, and acceptance evidence.

## Architecture Decision Ledger

### Resolved for the active implementation path

- [x] Shared headless-capable core embedded in GUI and headless artifacts.
- [x] GUI state excluded from durable core state.
- [x] One logical node root and exclusive authoritative process/SQLite writer as the target; daemon-client mode is separate future work.
- [x] Project owner owns Task definitions, trigger evaluation, runs, and authoritative Project operations.
- [x] Queue owner validates REST submission; workers pull leases and return outcomes; owner completes.
- [x] Providers are statically registered drivers; arbitrary endpoints are stable service instances; capabilities are subordinate observations.
- [x] Remote inference uses authenticated Apparat logical IDs; provider endpoints/credentials remain local.
- [x] One cluster-local X.509 root for MVP, separate TLS leaf and Apparat signing keys, signed device-record binding.
- [x] RFC 8785 JSON, SHA-256 payload hash, and Apparat Ed25519 envelope signature.
- [x] Service ads use monotonic revisions, 120-second expiry, refresh by 60 seconds, immediate routing exclusion, and 24-hour stale diagnostic retention.
- [x] Provider credentials use opaque local secret references.
- [x] Artifact bytes live outside SQLite and transfer through bounded resumable authenticated SHA-256-verified HTTPS.

### Must resolve before dependent implementation

- [ ] Phase 7: database encryption/restore, WAL/platform policy, lock recovery, and binary-root migration details.
- [ ] Phase 8: authorization vocabulary, endpoint seed/discovery transition, enrollment recovery, and concrete API limits.
- [ ] Phase 9: directory conflicts and Project-summary policy.
- [ ] Phase 10: file/binary limits, platform Task sandboxing, conflicts, and artifact retention defaults.
- [ ] Phase 11: discovery/auto-promotion, provider secrets per platform, admission defaults, and approved image drivers.
- [ ] Phase 12: lease/heartbeat/retry/retention/artifact quota defaults.
- [ ] Phase 13: routing score/fallback/explanation and freshness policy.
- [ ] Phase 14: webhook/schedule/approval rules; scheduler failover remains post-MVP.
- [ ] Phase 15: audio formats, adapters, retention, streaming, and permissions.
- [ ] Phase 16: binary hosting/tracking, signing custody, manifests, reproducibility, and rollback.
- [ ] Track A: Meshtastic, Signal, app-managed WireGuard, migration, replication, CRDT, dynamic routing.
- [ ] Track B: Comrades transport/privacy, grants, visibility, quotas, moderation, and abuse.
- [ ] Track C: BOINC boundary/sources/isolation, validation governance, gameplay, reputation, and anti-gaming.
- [ ] Optional headless TUI: reconsider termframe only after an approved TUI input contract exists.

## MVP Completion Definition

The MVP is complete only when the applicable cross-cutting requirements and all outcomes below have evidence.

- [ ] Steam Deck HUD and input
  - [ ] Controller navigation works across all five primary tabs.
  - [x] Comrades is the first navigable future-facing tab.
  - [x] Research follows Cluster as a navigable future-facing tab.
  - [x] Settings is final.
  - [ ] Focus, activation, back, context, scrolling, text, pointer/touch, accessibility, and enabled actions behave honestly.
  - [ ] `R2` PTT routes to ASR and can be cancelled.
- [ ] Shared runtime and secure connectivity
  - [ ] GUI and headless use the same core/node state without double ownership.
  - [ ] Headless Linux initializes no Ebitengine dependency.
  - [ ] Two devices enroll and mutually authenticate.
  - [ ] HTTPS REST works over WireGuard and trusted LAN with identical authorization.
  - [ ] Signed envelopes reject canonicalization errors, tampering, key mismatch, replay, expiry, wrong recipient, and unauthorized work.
- [ ] Durable state and Projects
  - [ ] A job survives duplicate delivery, disconnection, lease expiry, and requester/owner/worker restart.
  - [ ] Every device shows the authorized cluster-wide Project catalog while each repository/Task remains authoritative on its owner.
  - [ ] A real Project supports safe files/Git, durable drafts/transactions, and verified artifacts.
  - [ ] A Pipeline is a Project with Tasks and a Task runs manually with no trigger.
  - [ ] A scheduled or webhook Task submits/awaits a job, resumes idempotently, and produces an auditable result.
- [ ] Typed services, queues, and routing
  - [ ] One node manages/advertises multiple services including two same-provider/same-workload instances.
  - [ ] Desired, observed, capability, and advertisement state remain separate.
  - [ ] Capabilities distinguish text, image, video, STT, TTS, and BOINC contracts.
  - [ ] A real OpenAI-compatible text job routes through an authoritative queue.
  - [ ] Incompatible, unauthorized, disabled, unhealthy, or expired destinations are excluded with reasons.
  - [ ] Workers pull owner leases and return signed outcomes; only owners complete.
  - [ ] Remote jobs use logical Apparat IDs and never learn/invoke provider-local endpoints.
- [ ] Voice
  - [ ] PTT creates editable text through local or remote ASR.
  - [ ] TTS routes independently.
  - [ ] Audio lifecycle is bounded, cancellable, visible, private, and restart-honest.
- [ ] Diagnostics and release
  - [ ] Logs/diagnostics explain identity, lifecycle, queue, route, service, artifact, and platform failures without sensitive payloads.
  - [ ] Backup/restore and upgrade/rollback preserve authoritative state.
  - [ ] Steam Deck/Linux GUI and headless releases are packaged and validated.
  - [ ] Other platforms are claimed only to their evidenced level.
