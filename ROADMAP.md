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
    - `Alt+1` through `Alt+7` open the seven canonical tabs directly.
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
    3. Research
    4. Cluster
    5. Routing
    6. Tasks
    7. Settings
  - Comrades is visible first but remains a navigable placeholder during the MVP.
    - It eventually supports real-friend chat.
    - It eventually supports revocable low-priority inference sharing through owner-controlled comrade queues.
  - Research is visible third but remains a navigable placeholder during the MVP.
    - It eventually delegates explicitly budgeted compute to validated BOINC projects.
    - Research-project validation eventually participates in gameplay mechanics.
  - Settings remains the final tab.
- **Data and authority**
  - Project files remain ordinary filesystem and Git repositories.
  - SQLite stores identities, metadata, chats, transactions, events, queue state, indexes, and durable workflow state.
  - Projects, queues, and scheduled tasks have one authoritative owner device during the MVP.
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
      - During the MVP, every project, queue, and scheduled task has exactly one authoritative owner device.
      - Other devices may cache authorized projections and submit transactions, but they do not silently become co-authoritative.
    - **Projects**
      - The owner serializes accepted project transactions and publishes the resulting version.
      - A non-owner submits a transaction with its base version and stable transaction ID.
      - Accepted changes advance the authoritative project version.
      - Rejected or conflicting changes remain local editable drafts with a machine-readable reason and enough context for the user to revise or rebase them.
    - **Queues**
      - The queue owner stores authoritative ordering, admission, attempts, cancellation state, and results.
      - Requesters retain their submission record and authorized status/result projections.
      - Requesters do not mirror the complete queue unless a later availability design explicitly requires it.
      - Direct queues execute on the owner; pool queues are coordinated by the owner across eligible member devices.
    - **Tasks**
      - The task owner evaluates schedules and durable triggers.
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
    - Debian GUI uses `Ctrl+PageUp` and `Ctrl+PageDown` for tab cycling, `Alt+1` through `Alt+7` for direct tab selection, and right `Ctrl` for hold-to-talk.
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
    1. Prove the seven-tab shell, controller focus, and push-to-talk interaction with mock data.
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
  - [x] Define GUI, headless, service-host, queue-owner, project-owner, scheduler-owner, and enrollment-authority roles.
  - [x] Define module registration and command/event/store boundaries.
  - [x] Define the versioned workload-class taxonomy and extension rules.
  - [x] Define typed device, service, queue, route, and job capability contracts.
  - [x] Define which Salvagecore components are copied, adapted, rewritten, or rejected.
- [x] Create the shared input and focus contract in `docs/controller-map.md`.
  - [x] Define `L1`, `R1`, D-pad, sticks, `A`, `B`, `X`, Menu, and `R2`.
  - [x] Define Debian `Ctrl+PageUp`, `Ctrl+PageDown`, `Alt+1` through `Alt+7`, focus traversal, activation, cancellation, contextual actions, command palette, scrolling, and right-`Ctrl` push-to-talk.
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
  - [x] Define `./releases/[os]/[architecture]/latest[.exe]`.
  - [x] Use Go `GOOS` and `GOARCH` naming for release directories.
  - [x] Use `.exe` for Windows artifacts and no suffix for Unix-like artifacts.
  - [x] Implement a Python build pipeline that detects host OS and architecture.
  - [x] Add build-pipeline tests.
  - [x] Keep generated binary artifacts ignored by Git.
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
    - [x] Implement `Alt+1` through `Alt+7` direct tab selection.
    - [x] Implement `Tab`, `Shift+Tab`, arrows, `Enter`, `Space`, `Escape`, Menu or `Shift+F10`, `Ctrl+Shift+P`, and collection-navigation keys.
    - [x] Preserve ordinary text editing and clipboard behavior while text controls own focus.
  - [x] Implement mouse/touch focus, activation, context actions, scrolling, and non-pointer drag alternatives without separate application behavior.
  - [x] Add deterministic focus-navigation tests.
- [x] Implement top-level tabs.
  - [x] Add Comrades as the first tab.
  - [x] Add Projects.
  - [x] Add Research as the third tab.
  - [x] Add Cluster.
  - [x] Add Routing.
  - [x] Add Tasks.
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
    - The Ebitengine loop is implemented behind the `gui` build tag and still needs native desktop-library validation; default builds use a non-window blocking runtime path for CI and headless safety.
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
  - [x] Include component, event, device, project, job, task, and correlation IDs where relevant.
  - [x] Redact secrets, tokens, private keys, raw prompts, raw model outputs, and raw voice data by default.
  - [x] Add safe log rotation and retention.
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

## Phase 4: Secure Two-Device HTTPS/WireGuard Vertical Slice

**Goal:** Complete the MVP proof between a Steam Deck and one headless worker.

**Dependencies:** Phases 1–3.

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
  - [ ] Submit job.
  - [ ] Read job.
  - [ ] Cancel job.
  - [ ] Poll events by cursor.
  - [ ] Submit project transaction placeholder.
  - [ ] Enforce schemas, limits, deadlines, authorization, and audit logs.
- [ ] Implement the signed envelope.
  - [ ] Sign outgoing messages.
  - [ ] Verify incoming identity, signature, hash, expiration, and authorization.
  - [ ] Reject replay.
  - [ ] Apply duplicate messages idempotently.
- [ ] Implement the echo/mock queue.
  - [ ] Persist requester outbox submission.
  - [ ] Return `202 Accepted` and a durable job resource.
  - [ ] Persist owner acceptance or rejection.
  - [ ] Execute the mock job.
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

## Phase 5: Project Workspace And Git Operations

**Goal:** Make Apparat useful for real project navigation and controlled repository work.

**Dependencies:** Phase 4 transport and persistence.

- [ ] Add project registration and ownership.
  - [ ] Register existing filesystem/Git folders.
  - [ ] Assign one owner device.
  - [ ] Store metadata and routes in SQLite.
  - [ ] Validate safe project roots and path traversal protection.
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
- [ ] Add artifacts.
  - [ ] Metadata and ownership.
  - [ ] Hash and MIME type.
  - [ ] Bounded upload/download.
  - [ ] Resume and integrity verification.
  - [ ] Retention and cleanup.

**Exit criteria**

- A Steam Deck can open a real project, inspect files and Git state, submit a project chat job, and recover offline drafts without granting arbitrary shell access.

## Phase 6: Typed Compute Services, Queues, And Routing

**Goal:** Route each workload only through authoritative queues and devices that explicitly support its workload class and requirements.

**Dependencies:** Phases 4–5.

- [ ] Establish the workload-class registry.
  - [ ] Add `text_generation`.
  - [ ] Add `image_generation`.
  - [ ] Add `video_generation`.
  - [ ] Add `speech_to_text`.
  - [ ] Add `text_to_speech`.
  - [ ] Add `research_boinc`.
  - [ ] Define versioning and extension rules for future classes such as embeddings, reranking, classification, vision analysis, audio generation, simulation, and compilation.
  - [ ] Keep workload class independent from runtime/provider and model/project identity.
- [ ] Add typed service capability inventory.
  - [ ] Workload class and schema version.
  - [ ] Service runtime/provider.
  - [ ] Endpoint.
  - [ ] Device owner.
  - [ ] Models or BOINC projects.
  - [ ] Input/output modalities and limits.
  - [ ] Hardware and accelerator requirements.
  - [ ] Memory, storage, concurrency, and queue limits.
  - [ ] Streaming, progress, cancellation, and artifact support.
  - [ ] Health, load, availability, and validation timestamp.
  - [ ] Power, thermal, and schedule constraints.
  - [ ] Privacy and authorization scope.
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
  - [ ] Defer concrete STT/TTS adapters to Phase 8 while preserving typed discovery and routing now.
- [ ] Register BOINC workload contract.
  - [ ] Define BOINC project identity, client/runtime, platform, application, resource, schedule, and validation requirements.
  - [ ] Defer concrete BOINC execution to Phase 12 while preserving typed discovery and routing now.
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
  - [ ] Privacy boundary.
  - [ ] Priority and timeout.
  - [ ] Ordered fallback.
  - [ ] Clear explanation of the selected route.
  - [ ] Clear reason each incompatible device or queue was excluded.

**Exit criteria**

- A project can submit real text generation through an explicit route, survive retry/restart, fall back deterministically, and retrieve an authoritative result.
- Mock image, video, STT, TTS, and BOINC jobs route only to matching advertised capabilities.
- Unsupported workload classes and incompatible requirements fail clearly before execution.

## Phase 7: Automation, Scheduling, And Webhooks

**Goal:** Run durable cluster tasks even when some devices are offline.

**Dependencies:** Phase 6 typed queues.

- [ ] Add task definitions.
  - [ ] Owner device.
  - [ ] Trigger.
  - [ ] Steps.
  - [ ] Inputs and outputs.
  - [ ] Permissions.
  - [ ] Retry and timeout.
  - [ ] Approval policy.
  - [ ] Enabled/paused state.
- [ ] Add triggers.
  - [ ] Manual.
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

## Phase 8: ASR, TTS, And Voice Control

**Goal:** Turn controller and Debian GUI push-to-talk into a reliable routed cluster capability.

**Dependencies:** Phases 2, 6, and 7.

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

## Phase 9: Platform Packaging And Release Pipeline

**Goal:** Validate and ship each supported platform honestly and independently.

**Dependencies:** Stable vertical slice and HUD.

- [ ] Steam Deck/Linux GUI.
  - [ ] Controller database and mappings.
  - [ ] Debian keyboard mapping and text-input precedence.
  - [ ] Debian mouse and touchpad behavior.
  - [ ] Debian right-`Ctrl` push-to-talk and cancellation.
  - [ ] Configurable binding persistence and conflict reporting.
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
  - [ ] Native Android application wrapper.
  - [ ] Ebitengine AAR generated with pinned tooling.
  - [ ] Activity lifecycle and suspend/resume.
  - [ ] Touch, controller, keyboard, and text input.
  - [ ] Microphone and audio permissions.
  - [ ] Scoped storage and runtime state.
  - [ ] Background and battery constraints.
  - [ ] External WireGuard first.
  - [ ] VPN-service integration only in a later plan.
- [ ] Release engineering.
  - [ ] Artifact naming and directory layout.
  - [ ] Checksums and provenance.
  - [ ] Version metadata.
  - [ ] Reproducible build inputs.
  - [ ] Platform test matrix.
  - [ ] Upgrade and rollback.

**Exit criteria**

- Each platform is marked supported only after its build, packaging, input, storage, networking, audio, and lifecycle checks pass.

## Phase 10: Alternative Transports And Long-Term Resilience

**Goal:** Carry the same authenticated durable operations across constrained or human-mediated transports.

**Dependencies:** Stable signed envelope, queues, authorization, and transport adapter tests.

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

## Phase 11: Comrades, Chat, And Shared Inference

**Goal:** Add trusted real-friend communication and owner-controlled sharing of otherwise idle inference capacity.

**Dependencies:** Stable identity, authorization, queues, routing, audit, and at least one suitable authenticated transport.

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

## Phase 12: Research, BOINC, And Validation Gameplay

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

## Cross-Cutting Requirements

These requirements apply to every implementation phase:

- [ ] Security
  - [ ] Least privilege.
  - [ ] Explicit authorization.
  - [ ] Safe defaults.
  - [ ] Secret redaction.
  - [ ] Key rotation and revocation.
  - [ ] No unrestricted remote execution.
  - [ ] Shared-compute grants never imply project, file, secret, tool, shell, or administrative access.
  - [ ] Research workloads remain isolated from personal and comrade data.
- [ ] Reliability
  - [ ] Stable IDs.
  - [ ] Idempotent operations.
  - [ ] Durable state transitions.
  - [ ] Bounded retries.
  - [ ] Cancellation and deadlines.
  - [ ] Restart recovery.
- [ ] Observability
  - [ ] Structured logs.
  - [ ] Correlation IDs.
  - [ ] Queue and job traces.
  - [ ] Health endpoints.
  - [ ] User-readable failure reasons.
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
- [ ] Recovery
  - [ ] Database backup.
  - [ ] Identity repair.
  - [ ] Migration compatibility.
  - [ ] Upgrade rollback.
  - [ ] Artifact integrity verification.
- [ ] Documentation
  - [ ] Update README when product behavior changes.
  - [ ] Update API, security, database, transport, controller, and platform contracts with implementation changes.
  - [ ] Keep agent-operation instructions out of the human-facing README.

## Open Architecture Decisions

These remain explicit design tasks:

- [ ] Identity, certificates, and authorization
  - [ ] Select the exact X.509 hierarchy.
  - [ ] Decide how TLS leaf keys relate to Apparat device identity keys.
  - [ ] Finalize authorization vocabulary.
- [ ] Network and protocol
  - [ ] Select canonical signed-envelope encoding.
  - [ ] Select endpoint discovery after temporary static configuration.
  - [ ] Define cluster-directory conflict resolution.
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

## MVP Completion Definition

The MVP is complete only when:

- [ ] Steam Deck HUD and input
  - [ ] Controller navigation works across all primary tabs.
  - [ ] Comrades is present as the first navigable future-facing tab.
  - [ ] Research is present as the third navigable future-facing tab.
  - [ ] Settings is present as the final navigable tab.
  - [ ] `R2` push-to-talk state works and can route to ASR.
- [ ] Shared runtime and secure connectivity
  - [ ] A headless Linux worker runs without Ebitengine initialization.
  - [ ] Two devices enroll and authenticate mutually.
  - [ ] HTTPS REST works over WireGuard and trusted LAN with the same authorization.
  - [ ] Signed envelopes reject tampering, replay, expiration, and unauthorized work.
- [ ] Durable state and project operation
  - [ ] A durable job survives duplicate delivery, temporary disconnection, and application restart.
  - [ ] A real project can be browsed with safe Git operations.
  - [ ] A durable scheduled or webhook task can submit and await a job.
- [ ] Typed compute routing
  - [ ] A real OpenAI-compatible text-generation job can be routed through an authoritative queue.
  - [ ] Device capability records distinguish text generation, image generation, video generation, STT, TTS, and BOINC support.
  - [ ] Jobs cannot route to devices that do not advertise the requested workload class and requirements.
- [ ] Diagnostics and release
  - [ ] Logs and diagnostics explain failures without leaking sensitive payloads.
  - [ ] The Steam Deck/Linux release is packaged and validated.
