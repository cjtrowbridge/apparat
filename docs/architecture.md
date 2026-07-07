# Architecture Contract

Apparat is a ports-and-adapters Go application with one shared runtime and several presentation, transport, persistence, platform, and workload adapters.

## Package Boundaries

- `cmd/apparat`: GUI console entry point.
- `cmd/apparatd`: headless worker/service entry point.
- `internal/app`: shared runtime orchestration, command dispatch, module registration, service lifecycle, and mode selection.
- `internal/domain`: product rules, identity concepts, project ownership, queue ownership, workload classes, commands, events, policies, and validation.
- `internal/hud`: mockable HUD state, tab order, focus, input actions, voice states, mock projections, and diagnostics.
- `internal/adapters`: external-system integrations such as GUI, persistence, API, transport, model, speech, BOINC, Git, filesystem, and service adapters.
- `internal/platform`: OS lifecycle, paths, service-manager hooks, signals, and platform capability discovery.
- `third_party`: source-reference checkouts isolated from root Go package discovery unless a later approved plan activates an adapter.

Domain packages must not import GUI, SQLite, HTTP, WireGuard, model runtime, BOINC, Signal, Meshtastic, or source-reference packages.

## Roles

- GUI console: renders HUD state and dispatches user commands.
- Headless worker: runs without Ebitengine initialization and hosts queues, services, tasks, or API endpoints.
- Service host: advertises typed workload capabilities and executes leased work.
- Queue owner: authoritatively admits, schedules, leases, cancels, audits, and records jobs for one queue.
- Project owner: authoritatively applies project transactions and keeps rejected edits durable.
- Scheduler owner: authoritatively runs a task schedule and persists run state.
- Enrollment authority: issues short-lived invites and signs directory updates.

One device may hold several roles.

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

Devices, services, queues, routes, and jobs declare:

- Workload class and schema version.
- Runtime/provider and endpoint.
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
