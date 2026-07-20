---
plan_id: 2026-07-19-17-19-34_replace-roadmap-phase-6-onward
title: Replace Roadmap Phase 6 Onward
summary: Replace the broad legacy roadmap from Phase 6 onward with a concrete Phase 6–N implementation program derived from the completed GUI foundation and architecture review.
status: past
created_at: 2026-07-19-17-19-34
---

# Replace Roadmap Phase 6 Onward

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Objective And Scope

Phases 0–5 established the repository, contracts, GUI mockup, local runtime foundations, and first Android GUI evidence. The project is now ready to plug durable functionality in behind the scenes and expose it through the existing GUI.

This plan replaces everything in `ROADMAP.md` from legacy Phase 6 onward with a new, concrete Phase 6–N implementation program. The new program will translate the architecture review and current codebase evidence into actionable phases with explicit ownership, state, persistence, REST/API, security, lifecycle, testing, documentation, dependencies, acceptance evidence, and GUI integration requirements.

The pre-Phase-6 product baseline, Salvagecore reference record, and completed Phases 0–5 remain in `ROADMAP.md`. This plan does not rewrite their historical implementation record. Where they leave a real carryover requirement for future work, that requirement is represented in the new Phase 6–N program.

`RECOMMENDATIONS.md` is the source review for this replacement. The completed outcome is a canonical `ROADMAP.md` whose Phase 6 onward content is the detailed execution path for the rest of Apparat, rather than a second competing roadmap.

This is a documentation-synthesis and roadmap-replacement plan. Executing it does not implement the product features described by the resulting Phase 6–N program. `RECOMMENDATIONS.md` remains available as temporary research input while this plan is executed; it is deleted only after its unique content has been incorporated or explicitly retired, all canonical references have been migrated, and the replacement has passed its validation gates.

Immediately before the first non-trivial roadmap-migration edit, move this plan from `plans/future/` to `plans/current/` and regenerate the plan indexes. After the Roadmap replacement, documentation reconciliation, Recommendations retirement, and final verification are complete, mark this plan complete, move it to `plans/past/`, and regenerate the indexes again. The archived plan preserves the migration mapping and retirement rationale; `ROADMAP.md` remains the sole canonical implementation sequence while `README.md` remains the canonical product contract.

- [x] 1. Establish the exact Phase 6–N replacement boundary and source mapping.
  - [x] 1.1 Preserve `ROADMAP.md` content before the legacy Phase 6 heading, including the product baseline, Salvagecore record, and completed Phase 0–5 history.
  - [x] 1.2 Inventory every legacy Phase 6–14 goal, dependency, checklist item, exit criterion, cross-cutting requirement, open decision, and MVP requirement.
  - [x] 1.3 Inventory every concrete recommendation, structural admission gate, state model, ownership rule, protocol rule, testing requirement, and GUI requirement in `RECOMMENDATIONS.md`.
  - [x] 1.4 Record within this plan a traceable mapping from every legacy Phase 6–14 item and every Recommendations item into the new Phase 6–N program or an explicit retirement/defer decision.
  - [x] 1.5 Preserve legacy phase aliases where they make old plans, journals, and historical references easier to understand.
  - [x] 1.6 Record explicit retirement or defer reasons for any legacy future item not retained in the new program.
  - [x] 1.7 Consolidate duplicates, reconcile conflicting formulations, classify current evidence separately from future requirements, and assign each retained result to exactly one primary new phase plus any applicable cross-cutting requirement.

- [x] 2. Resolve the concrete architectural path required before backend functionality is added.
  - [x] 2.1 Define the structural integration and stabilization requirements that Phase 6 documents and Phase 7 implements.
    - [x] 2.1.1 Make the shared headless-capable core authoritative for durable backend state in GUI and headless artifacts.
    - [x] 2.1.2 Keep GUI navigation, focus, selection, layout, scroll, gesture, animation, and transient editor state outside the core.
    - [x] 2.1.3 Define the core command, query/read-model, change-notification, lifecycle, readiness, and health boundary.
    - [x] 2.1.4 Define domain, application, SQLite, external-adapter, platform, and GUI package ownership.
    - [x] 2.1.5 Define startup/shutdown supervision, process ownership, persistence hardening, headless tests, display-free GUI tests, logging, identity, and updater safety requirements.
    - [x] 2.1.6 Define the first implementation proof as one minimal durable local-service inventory slice that exercises the shared core, SQLite, headless query surface, GUI read model, restart recovery, and two same-provider mock instances without introducing remote advertisement or speculative frameworks.
  - [x] 2.2 Resolve decisions that must be concrete before their dependent phase.
    - [x] 2.2.1 Separate decided contracts from still-open decisions.
    - [x] 2.2.2 Define the identity, trust/device-directory, certificate, TLS-key binding, authorization vocabulary, signed-envelope, discovery, process-ownership, SQLite-writer, and artifact-lifecycle decisions required before Phases 7–8.
    - [x] 2.2.3 Bind every later decision—including database encryption/restore, Task sandboxing, scheduler failover, alternate transport, shared compute, BOINC isolation, and gameplay—to an explicit must-resolve-before phase gate instead of attempting speculative closure in Phase 6.
    - [x] 2.2.4 Reconcile and record the known contract conflicts before writing their dependent Roadmap items.
      - [x] 2.2.4.1 Define the device-identity, X.509 hierarchy, certificate rotation, and TLS leaf-key relationship.
      - [x] 2.2.4.2 Define the signed-envelope encoding, canonicalization, hashing, signature, expiry, replay, and compatibility rules.
      - [x] 2.2.4.3 Define GUI/headless process ownership, runtime locking or daemon-client behavior, and the single-authoritative SQLite-writer rule.
      - [x] 2.2.4.4 Make the Project owner authoritative for Task definitions, runs, scheduling, and trigger bindings; remove scheduler-owned wording that contradicts Project ownership.
      - [x] 2.2.4.5 Define service-advertisement revision, observation time, expiry, stale retention/removal, and re-advertisement behavior.
      - [x] 2.2.4.6 Define local inference endpoint and credential-reference storage, redaction, authorization, and non-disclosure rules.
      - [x] 2.2.4.7 Define artifact storage, ownership, references, transfer, integrity validation, authorization, retention, cleanup, and failure recovery.
      - [x] 2.2.4.8 Define Task sandboxing, secret references, allowlisted actions, approval boundaries, and the prohibition on unrestricted remote shell execution.
  - [x] 2.3 Correct stale or overclaimed GUI/runtime completion evidence only where it affects the new Phase 6–N path.
    - [x] 2.3.1 Move unfinished controller, focus, input-equivalence, and accessibility behavior into concrete future acceptance work.
    - [x] 2.3.2 Keep historical journal entries append-only and record corrections as current roadmap evidence.
  - [x] 2.4 Close the canonical documentation gaps before synthesizing the new Roadmap phases.
    - [x] 2.4.1 Update `README.md` where the consolidated research establishes or clarifies a product invariant, without moving transient migration detail into the product contract.
    - [x] 2.4.2 Update architecture, database, API/OpenAPI, security, signed-envelope, transport, controller, and platform contracts with the resolved ownership, state, protocol, persistence, lifecycle, and security rules required by the new sequence.
    - [x] 2.4.3 Mark behavior as implemented, partially validated, planned, deferred, or removed wherever the current documents could otherwise overclaim executable support.
    - [x] 2.4.4 Run the applicable documentation and contract checks before using those reconciled contracts as input to the Roadmap rewrite.

- [x] 3. Define the durable ownership and REST execution model for the new phases.
  - [x] 3.1 Define Projects as Git repositories owned by the device where their working trees live and run.
  - [x] 3.2 Define the Phase 9 minimum owner-local Project registry, stable Project identity, signed authorization-filtered summaries, cluster-wide catalog projection, freshness, and stale/offline behavior separately from the Phase 10 workspace and Git feature set.
  - [x] 3.3 Define Pipelines as Projects with Apparat Task entrypoints rather than independently owned workflow objects.
  - [x] 3.4 Define Task ownership and staged capability: Phase 10 constrained manual owner-local execution, Phase 12 queue-backed job steps, and Phase 14 interval/webhook/application-event/cluster-event bindings with durable approval, retry, and recovery.
  - [x] 3.5 Define the Phase 8 reusable mock-queue proof and its Phase 12 expansion into full queue ownership, requester REST submission, owner validation/admission, worker claim/long-poll, leases/fencing, heartbeats, result return, artifacts, idempotency, cancellation, retry, and owner-authoritative completion.
  - [x] 3.6 Define the API, signed-envelope, SQLite, authorization, audit, and GUI read-model consequences of these ownership rules.

- [x] 4. Define the multi-instance inference-service and routing path.
  - [x] 4.1 Separate workload class, driver kind, concrete local service instance, and discovered capability/model identity.
  - [x] 4.2 Define statically registered provider drivers, factories, instances, typed executors, workload-specific requests/results, and explicit composition-root registration.
  - [x] 4.3 Define arbitrary same-provider and same-workload instance cardinality with stable service/capability IDs and secondary routing indexes.
  - [x] 4.4 Define desired configuration, observed health/inventory, capabilities, local credential references, advertisements, revision, expiry, and stale behavior in SQLite.
  - [x] 4.5 Define discovery, verification, enablement, advertisement, independent supervision, admission, concurrency, cancellation, retry, failure isolation, and restart recovery.
  - [x] 4.6 Define the authenticated Apparat gateway boundary so remote peers use logical service/capability identifiers rather than provider-local endpoints.
  - [x] 4.7 Define the provider order: mock driver and inventory evidence, OpenAI-compatible text, Ollama, llama.cpp, approved image drivers, then independently typed video, speech, and BOINC adapters.

- [x] 5. Rewrite legacy Phase 6 onward as a detailed new Phase 6–N program whose numbering follows real dependencies rather than legacy phase count.
  - [x] 5.1 Write new Phase 6: Documentation, evidence reconciliation, and immediate architecture decisions.
    - [x] 5.1.1 Close documentation gaps, reconcile overclaimed completion evidence, establish the phase template and decision register, and resolve only the contracts required to begin Phases 7–8 safely.
    - [x] 5.1.2 Assign every later open decision an explicit must-resolve-before phase gate.
  - [x] 5.2 Write new Phase 7: Shared core, SQLite, lifecycle, and artifact/process-ownership proof.
    - [x] 5.2.1 Include core-versus-GUI state boundaries, command/query/change seams, migrations, one-writer/process ownership, startup/shutdown supervision, headless and display-free tests, and structured diagnostics.
    - [x] 5.2.2 Use a durable local mock-service inventory with two same-provider instances as the first reusable vertical slice; do not advertise it remotely yet.
    - [x] 5.2.3 Prove both GUI and headless artifacts against the same core and begin continuous Linux/Steam Deck and Android validation rather than deferring platform evidence to the release phase.
  - [x] 5.3 Write new Phase 8: Identity, trust/device directory, secure REST, and reusable mock-queue proof.
    - [x] 5.3.1 Include external-network configuration, enrollment, authoritative trusted-device records, mTLS, signed envelopes, REST resources, authorization, limits, audit, and revocation.
    - [x] 5.3.2 Carry one durable mock workload through requester submission, owner validation, worker pull/lease, result return, owner-authoritative completion, restart, disconnection, idempotency, and correlation; retain these primitives for Phase 12.
  - [x] 5.4 Write new Phase 9: Discovery, presence, owner-local Project registry, and cluster-wide Project catalog.
    - [x] 5.4.1 Include stable owner-local Project identity, signed authorization-filtered summaries, discovery and presence, remote summary caching, revision/freshness, stale/offline display, and owner-directed REST reads.
    - [x] 5.4.2 Keep repository contents, Git state, Task definitions, and mutations owner-authoritative and defer full workspace operations to Phase 10.
  - [x] 5.5 Write new Phase 10: Project workspaces, Git, Pipelines, and constrained manual Tasks.
    - [x] 5.5.1 Include safe files/Git, chats, drafts, transactions, conflicts, artifacts, Pipeline derivation, Task entrypoint schemas, permissions, approvals, and durable run history.
    - [x] 5.5.2 Permit manual Tasks with no trigger, but limit execution to constrained owner-local actions or mock/local executors until queue-backed steps and routing arrive in Phases 12–13.
    - [x] 5.5.3 Resolve Task sandboxing, secret references, artifact handling, and no-unrestricted-remote-shell rules before executable Task entrypoints are admitted.
  - [x] 5.6 Write new Phase 11: Multi-instance local inference drivers, health, capabilities, and advertisements.
    - [x] 5.6.1 Include static provider registration, arbitrary same-provider instances, discovery and verification, desired/observed state, independent supervision, capability inventory, advertisement revision/expiry, safe gateway projection, and GUI/headless read models.
    - [x] 5.6.2 Progress from the Phase 7 mock driver to OpenAI-compatible text, Ollama, llama.cpp, approved image drivers, and typed future workload contracts without coupling services to queue ownership or routing policy.
  - [x] 5.7 Write new Phase 12: Full authoritative queue protocol, leasing, results, artifacts, and recovery.
    - [x] 5.7.1 Extend the Phase 8 mock-queue primitives into typed direct and pool-owner queues, admission, priorities, leasing/fencing, heartbeats, cancellation, retries, retention, artifact validation, worker failure isolation, and restart recovery.
    - [x] 5.7.2 Keep the queue owner authoritative while eligible inference workers pull or long-poll and return signed outcomes over REST.
  - [x] 5.8 Write new Phase 13: Pools, routing profiles, deterministic fallback, and first real text generation.
    - [x] 5.8.1 Include capability matching, pool membership, route explanations, privacy and policy constraints, specific-service targeting, load/admission signals, ordered fallback, and an end-to-end real text-generation path.
  - [x] 5.9 Write new Phase 14: Automation, scheduling, webhooks, event triggers, approvals, and durable workflows.
    - [x] 5.9.1 Add interval/cron, authenticated webhook, application-event, and cluster-event bindings without changing Project/Task ownership or requiring a trigger for manual execution.
    - [x] 5.9.2 Include durable steps, queue/job linkage, checkpoints, retries, timeouts, cancellation, safe actions, redacted diagnostics, restart recovery, and configured human approval.
  - [x] 5.10 Write new Phase 15: ASR, TTS, push-to-talk, audio lifecycle, and privacy.
    - [x] 5.10.1 Include typed local and remote speech routes, bounded capture, cancellation, temporary storage/deletion, editable transcription, independently routed speech output, privacy indicators, and GUI/platform integration.
    - [x] 5.10.2 Keep capture, focus, and pre-submission audio in GUI/platform state; create durable core work only on explicit submission.
  - [x] 5.11 Write new Phase 16: Packaging, release hardening, and platform support evidence.
    - [x] 5.11.1 Treat this as culmination of continuous platform validation, not its beginning: Steam Deck/Linux GUI and headless first, then independently validated Windows, macOS, and Android.
    - [x] 5.11.2 Include reproducibility, signing, provenance, installers/services, upgrade, rollback, storage/network/audio/lifecycle evidence, and honest per-target support declarations.
  - [x] 5.12 Define post-MVP work as independently dependency-gated tracks rather than a false Phase 17–19 dependency chain.
    - [x] 5.12.1 Alternative transports and resilience: conformance, Meshtastic, Signal, optional WireGuard management, failover, ownership migration, replication, CRDT research, and routing optimization.
    - [x] 5.12.2 Comrades, chat, and shared inference: identity/trust, chat, grants, owner-authoritative shared queues, quotas, safety, abuse controls, and HUD requirements; require stable identity, queues, routing, audit, and one suitable authenticated transport, not every alternative transport.
    - [x] 5.12.3 Research, BOINC, and validation gameplay: trust/evidence, BOINC boundary, resource policy, isolation, scheduling, packaging, HUD, recovery, provenance, and owner control; do not make it depend on Comrades or alternative transports where no real dependency exists.
    - [x] 5.12.4 If canonical roadmap formatting requires sequential phase numbers, number these tracks by product priority while stating their independent prerequisites explicitly.
  - [x] 5.13 Require every implementation phase to specify user outcome, scope and deferrals, dependencies, domain/state ownership, SQLite changes, REST/API changes, authorization, adapters, GUI projection, failure/recovery behavior, tests, documentation, target-platform evidence, and exit criteria.
  - [x] 5.14 Require every functional phase to end in a reviewable vertical slice from core state through persistence and command/query or REST surfaces to GUI projection, including failure and restart evidence where applicable.
  - [x] 5.15 Retain and refine cross-cutting security, reliability, observability, privacy, performance, recovery, platform-validation, and documentation requirements.
  - [x] 5.16 Retain and refine the MVP completion definition against the new dependency order and concrete acceptance evidence.

- [x] 6. Keep detailed contracts consistent with the new concrete roadmap.
  - [x] 6.1 After rewriting Phase 6 onward, perform a final consistency sweep across README, architecture, database, API/OpenAPI, security, signed-envelope, transport, controller, and platform contracts for any invariant clarified during synthesis.
  - [x] 6.2 Make Project/Pipeline/Task ownership, queue-owner worker-pull REST execution, shared-core state, and multi-instance inference rules explicit and consistent everywhere.
  - [x] 6.3 Mark contract, implemented, partially validated, planned, and deferred behavior where ambiguity would mislead implementation.
  - [x] 6.4 Update every canonical link or status statement that still treats `RECOMMENDATIONS.md` as a continuing planning authority so it points to the consolidated Roadmap or an appropriate detailed contract.
  - [x] 6.5 Keep changed source, script, build, test, and code directories documented at the closest useful layer when later feature implementation begins.

- [x] 7. Validate the Phase 6 onward replacement before completing the plan.
  - [x] 7.1 Confirm every legacy Phase 6–14 item, carryover dependency, cross-cutting requirement, open decision, and MVP outcome maps to the new Phase 6–N program or has an explicit retirement/defer rationale.
  - [x] 7.2 Confirm the new phase order is internally consistent from documentation gates through shared core, secure clustering, Project catalog/workspaces, local services, authoritative queues, routing, automation, voice, and release proof, with post-MVP tracks ordered only by their actual prerequisites.
  - [x] 7.3 Confirm the new roadmap adds concrete architecture, state, protocol, persistence, lifecycle, failure, and acceptance detail rather than merely restating broad goals.
  - [x] 7.4 Run documentation, plan-index, whitespace, link, OpenAPI, and repository verification appropriate to the changed files.
  - [x] 7.5 Review the final diff to confirm that content before legacy Phase 6 remains preserved and the new Phase 6–N program is the sole canonical plan for the remaining project.
  - [x] 7.6 Confirm `RECOMMENDATIONS.md` contains no unique retained requirement, decision, dependency, acceptance criterion, or future plan that is absent from the new Roadmap or the appropriate canonical detailed contract.
  - [x] 7.7 Confirm every item excluded from the consolidated Roadmap has a retirement or defer rationale preserved in this plan for archival evidence.
  - [x] 7.8 Request approval before appending the roadmap-replacement checkpoint to the journal.

- [x] 8. Retire the temporary research document and complete the plan lifecycle.
  - [x] 8.1 Confirm the Recommendations deletion gate: incorporation and retirement mappings are complete, no canonical document depends on the file, no unique retained content remains in it, and the pre-deletion validation suite passes.
  - [x] 8.2 Remove remaining links and status text that refer to `RECOMMENDATIONS.md`, then delete `RECOMMENDATIONS.md`.
  - [x] 8.3 Re-run documentation, OpenAPI, plan-index, whitespace, link, and repository verification after deletion and review the final diff for semantic completeness.
  - [x] 8.4 Mark every completed or intentionally closed checklist item accurately, set this plan to `past`, move it from `plans/current/` to `plans/past/`, and regenerate all plan indexes.
  - [x] 8.5 Confirm `ROADMAP.md` is the sole canonical implementation sequence for the remaining project and the archived plan is historical migration evidence rather than competing authority.

## Consolidation Record

This record is migration evidence retained with the archived plan. It is not a second roadmap.

### Legacy Roadmap Mapping

| Source | Consolidated destination |
| --- | --- |
| Legacy Phase 3 validation-pending shared-runtime, GUI loop, headless, logging, identity, persistence, and Android-safe runtime evidence | New Phases 6–7, with release-only evidence also carried into Phase 16 |
| Legacy Phase 5 validation-pending Android safe-area, density, device/input coverage, runtime-path, and optional install/launch integration evidence | New Phases 6–7 for continuous validation and Phase 16 for release hardening |
| Legacy Phase 6 secure two-device REST/WireGuard slice | New Phase 8; its Project-summary seed moves to Phase 9 and reusable mock-queue primitives expand in Phase 12 |
| Legacy Phase 7 Project workspace and Git | New Phase 9 for owner-local registration/catalog projection and New Phase 10 for workspace, Git, Pipeline, Task, transaction, chat, draft, and artifact behavior |
| Legacy Phase 8 typed compute services, queues, and routing | New Phase 7 mock local-service inventory; Phase 11 service drivers/inventory/advertisements; Phase 12 authoritative queues; Phase 13 pools/routing/real text generation |
| Legacy Phase 9 automation, scheduling, and webhooks | New Phase 10 for Task identity and constrained manual execution; New Phase 14 for triggers, durable workflows, approvals, and recovery |
| Legacy Phase 10 ASR, TTS, and voice | New Phase 15 |
| Legacy Phase 11 packaging and release | Continuous target evidence in Phases 7–15 plus release hardening in New Phase 16 |
| Legacy Phase 12 alternative transports and resilience | Post-MVP Track A |
| Legacy Phase 13 Comrades and shared inference | Post-MVP Track B; depends on one suitable authenticated transport, not completion of Track A |
| Legacy Phase 14 Research, BOINC, and gameplay | Post-MVP Track C; depends on capabilities, scheduling, budgets, audit, packaging, and isolation, not Tracks A or B |
| Legacy cross-cutting requirements | New cross-cutting requirements plus the mandatory template inside every implementation phase |
| Legacy open decisions | New Phase 6 decision ledger with explicit must-resolve-before gates in dependent phases |
| Legacy MVP completion definition | Consolidated MVP evidence definition after the new program |

### Recommendations Mapping

| Research area | Consolidated destination |
| --- | --- |
| Documentation gaps, truth reconciliation, implemented/planned labels, and admission rules | New Phase 6 and the roadmap-wide planning rules |
| Shared headless core, GUI-state exclusion, package seams, command/query/read-model boundary, SQLite lifecycle, supervision, tests, identity/logging/updater hardening | New Phase 7 |
| Multi-instance mock service inventory as first real core slice | New Phase 7, then generalized in Phase 11 |
| Project/Pipeline/Task ownership and cluster-wide catalog | New Phases 9–10 |
| Queue-owner REST validation, worker pull/lease, returned result, and owner-authoritative completion | Reusable proof in Phase 8 and full protocol in Phase 12 |
| Static inference drivers, stable Service/Capability IDs, desired/observed state, gateway-only remote access, discovery, health, and advertisements | New Phase 11 |
| Pool execution, route profiles, fallbacks, exclusions, and first real text generation | New Phase 13 |
| Optional Task triggers and durable automation | New Phase 14 |
| Voice capture boundary, ASR/TTS routes, and privacy | New Phase 15 |
| Input/focus/accessibility/live-data/configuration/performance UX gaps | Phase 6 evidence reconciliation, Phase 7 common GUI foundation, and acceptance work in every affected feature phase |
| Build simplification, non-mutating builds, binary-artifact policy, reproducibility, provenance, signing, update, and rollback | New Phases 7 and 16 |
| Platform-specific inference supervision and honest support evidence | Continuous evidence in Phases 7–15 and release declaration in Phase 16 |
| Alternative transports, Comrades, and Research | Independent post-MVP Tracks A, B, and C |

### Reconciled Or Retired Formulations

- Retire six/seven-tab and `Alt+1` through `Alt+6`/`Alt+7` claims. The contract has five top-level tabs and `Alt+1` through `Alt+5`; Routing and Tasks remain Cluster details.
- Retire `scheduler-owned Task` wording. A Task, its trigger bindings, scheduler evaluation, and authoritative runs belong to the Task's Project owner during the MVP.
- Retire binary-specific default runtime roots as the target ownership design while preserving them as current Phase 3 evidence. The target uses one logical node root and exclusive writer lock; daemon-client mode remains a later explicit decision.
- Retire provider name or workload class as service-instance identity. Stable `ServiceID` is primary and permits arbitrary same-provider/same-workload instances.
- Reject Go dynamic `.so` plugins and hidden global registration for the planned provider system. Retain statically compiled, explicitly registered drivers; defer authenticated out-of-process extension IPC until a real requirement exists.
- Retire direct remote provider access, endpoint disclosure, credential replication, worker-authoritative completion, shared-database queue access, and unrestricted remote shell as incompatible with the selected security/ownership model.
- Retire the legacy all-in-one services/queues/routing phase grouping while retaining every feature across New Phases 7 and 11–13.
- Retire a false sequential dependency among alternative transports, Comrades, and Research. Preserve them as independent post-MVP tracks with explicit prerequisites.
- Keep Android headless, automatic scheduler election, dynamic routing optimization, CRDT multi-writer editing, app-managed WireGuard, Signal, Meshtastic, Comrades, and Research outside the MVP unless their named later gates are approved.
- No retained product goal from legacy Phases 6–14 or the Recommendations research is otherwise retired; duplicated formulations are consolidated into the mapped destination above.

## Completion Criteria

- `ROADMAP.md` retains its pre-Phase-6 baseline and completed-history content.
- Everything from legacy Phase 6 onward is replaced by a new detailed Phase 6–N implementation program.
- The new program converts the completed GUI/mockup foundation into a concrete path for durable core functionality, REST cluster operation, Projects, Pipelines, Tasks, queues, multi-instance inference, automation, platform release, and independently dependency-gated resilience, Comrades, and Research tracks.
- The new roadmap is concrete enough to bind focused implementation plans without re-solving ownership, state, protocol, persistence, lifecycle, or acceptance criteria each time.
- Every Recommendations and legacy future-Roadmap item is mapped to the new sequence or has an explicit retirement/defer rationale preserved in the archived plan.
- `RECOMMENDATIONS.md` has been deleted after successful incorporation and reference migration.
- This plan has been completed and archived under `plans/past/` as migration evidence.
- `ROADMAP.md` is the sole canonical implementation sequence for the remaining project; `README.md` remains the canonical product contract and detailed documents remain authoritative for their named contracts.
