---
plan_id: 2026-07-19-17-19-34_replace-roadmap-phase-6-onward
title: Replace Roadmap Phase 6 Onward
summary: Replace the broad legacy roadmap from Phase 6 onward with a concrete Phase 6–N implementation program derived from the completed GUI foundation and architecture review.
status: future
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

- [ ] 1. Establish the exact Phase 6–N replacement boundary and source mapping.
  - [ ] 1.1 Preserve `ROADMAP.md` content before the legacy Phase 6 heading, including the product baseline, Salvagecore record, and completed Phase 0–5 history.
  - [ ] 1.2 Inventory every legacy Phase 6–14 goal, dependency, checklist item, exit criterion, cross-cutting requirement, open decision, and MVP requirement.
  - [ ] 1.3 Inventory every concrete recommendation, structural admission gate, state model, ownership rule, protocol rule, testing requirement, and GUI requirement in `RECOMMENDATIONS.md`.
  - [ ] 1.4 Record within this plan a traceable mapping from every legacy Phase 6–14 item and every Recommendations item into the new Phase 6–N program or an explicit retirement/defer decision.
  - [ ] 1.5 Preserve legacy phase aliases where they make old plans, journals, and historical references easier to understand.
  - [ ] 1.6 Record explicit retirement or defer reasons for any legacy future item not retained in the new program.
  - [ ] 1.7 Consolidate duplicates, reconcile conflicting formulations, classify current evidence separately from future requirements, and assign each retained result to exactly one primary new phase plus any applicable cross-cutting requirement.

- [ ] 2. Resolve the concrete architectural path required before backend functionality is added.
  - [ ] 2.1 Define the structural integration and stabilization requirements that Phase 6 documents and Phase 7 implements.
    - [ ] 2.1.1 Make the shared headless-capable core authoritative for durable backend state in GUI and headless artifacts.
    - [ ] 2.1.2 Keep GUI navigation, focus, selection, layout, scroll, gesture, animation, and transient editor state outside the core.
    - [ ] 2.1.3 Define the core command, query/read-model, change-notification, lifecycle, readiness, and health boundary.
    - [ ] 2.1.4 Define domain, application, SQLite, external-adapter, platform, and GUI package ownership.
    - [ ] 2.1.5 Define startup/shutdown supervision, process ownership, persistence hardening, headless tests, display-free GUI tests, logging, identity, and updater safety requirements.
    - [ ] 2.1.6 Define the first implementation proof as one minimal durable local-service inventory slice that exercises the shared core, SQLite, headless query surface, GUI read model, restart recovery, and two same-provider mock instances without introducing remote advertisement or speculative frameworks.
  - [ ] 2.2 Resolve decisions that must be concrete before their dependent phase.
    - [ ] 2.2.1 Separate decided contracts from still-open decisions.
    - [ ] 2.2.2 Define the identity, trust/device-directory, certificate, TLS-key binding, authorization vocabulary, signed-envelope, discovery, process-ownership, SQLite-writer, and artifact-lifecycle decisions required before Phases 7–8.
    - [ ] 2.2.3 Bind every later decision—including database encryption/restore, Task sandboxing, scheduler failover, alternate transport, shared compute, BOINC isolation, and gameplay—to an explicit must-resolve-before phase gate instead of attempting speculative closure in Phase 6.
    - [ ] 2.2.4 Reconcile and record the known contract conflicts before writing their dependent Roadmap items.
      - [ ] 2.2.4.1 Define the device-identity, X.509 hierarchy, certificate rotation, and TLS leaf-key relationship.
      - [ ] 2.2.4.2 Define the signed-envelope encoding, canonicalization, hashing, signature, expiry, replay, and compatibility rules.
      - [ ] 2.2.4.3 Define GUI/headless process ownership, runtime locking or daemon-client behavior, and the single-authoritative SQLite-writer rule.
      - [ ] 2.2.4.4 Make the Project owner authoritative for Task definitions, runs, scheduling, and trigger bindings; remove scheduler-owned wording that contradicts Project ownership.
      - [ ] 2.2.4.5 Define service-advertisement revision, observation time, expiry, stale retention/removal, and re-advertisement behavior.
      - [ ] 2.2.4.6 Define local inference endpoint and credential-reference storage, redaction, authorization, and non-disclosure rules.
      - [ ] 2.2.4.7 Define artifact storage, ownership, references, transfer, integrity validation, authorization, retention, cleanup, and failure recovery.
      - [ ] 2.2.4.8 Define Task sandboxing, secret references, allowlisted actions, approval boundaries, and the prohibition on unrestricted remote shell execution.
  - [ ] 2.3 Correct stale or overclaimed GUI/runtime completion evidence only where it affects the new Phase 6–N path.
    - [ ] 2.3.1 Move unfinished controller, focus, input-equivalence, and accessibility behavior into concrete future acceptance work.
    - [ ] 2.3.2 Keep historical journal entries append-only and record corrections as current roadmap evidence.
  - [ ] 2.4 Close the canonical documentation gaps before synthesizing the new Roadmap phases.
    - [ ] 2.4.1 Update `README.md` where the consolidated research establishes or clarifies a product invariant, without moving transient migration detail into the product contract.
    - [ ] 2.4.2 Update architecture, database, API/OpenAPI, security, signed-envelope, transport, controller, and platform contracts with the resolved ownership, state, protocol, persistence, lifecycle, and security rules required by the new sequence.
    - [ ] 2.4.3 Mark behavior as implemented, partially validated, planned, deferred, or removed wherever the current documents could otherwise overclaim executable support.
    - [ ] 2.4.4 Run the applicable documentation and contract checks before using those reconciled contracts as input to the Roadmap rewrite.

- [ ] 3. Define the durable ownership and REST execution model for the new phases.
  - [ ] 3.1 Define Projects as Git repositories owned by the device where their working trees live and run.
  - [ ] 3.2 Define the Phase 9 minimum owner-local Project registry, stable Project identity, signed authorization-filtered summaries, cluster-wide catalog projection, freshness, and stale/offline behavior separately from the Phase 10 workspace and Git feature set.
  - [ ] 3.3 Define Pipelines as Projects with Apparat Task entrypoints rather than independently owned workflow objects.
  - [ ] 3.4 Define Task ownership and staged capability: Phase 10 constrained manual owner-local execution, Phase 12 queue-backed job steps, and Phase 14 interval/webhook/application-event/cluster-event bindings with durable approval, retry, and recovery.
  - [ ] 3.5 Define the Phase 8 reusable mock-queue proof and its Phase 12 expansion into full queue ownership, requester REST submission, owner validation/admission, worker claim/long-poll, leases/fencing, heartbeats, result return, artifacts, idempotency, cancellation, retry, and owner-authoritative completion.
  - [ ] 3.6 Define the API, signed-envelope, SQLite, authorization, audit, and GUI read-model consequences of these ownership rules.

- [ ] 4. Define the multi-instance inference-service and routing path.
  - [ ] 4.1 Separate workload class, driver kind, concrete local service instance, and discovered capability/model identity.
  - [ ] 4.2 Define statically registered provider drivers, factories, instances, typed executors, workload-specific requests/results, and explicit composition-root registration.
  - [ ] 4.3 Define arbitrary same-provider and same-workload instance cardinality with stable service/capability IDs and secondary routing indexes.
  - [ ] 4.4 Define desired configuration, observed health/inventory, capabilities, local credential references, advertisements, revision, expiry, and stale behavior in SQLite.
  - [ ] 4.5 Define discovery, verification, enablement, advertisement, independent supervision, admission, concurrency, cancellation, retry, failure isolation, and restart recovery.
  - [ ] 4.6 Define the authenticated Apparat gateway boundary so remote peers use logical service/capability identifiers rather than provider-local endpoints.
  - [ ] 4.7 Define the provider order: mock driver and inventory evidence, OpenAI-compatible text, Ollama, llama.cpp, approved image drivers, then independently typed video, speech, and BOINC adapters.

- [ ] 5. Rewrite legacy Phase 6 onward as a detailed new Phase 6–N program whose numbering follows real dependencies rather than legacy phase count.
  - [ ] 5.1 Write new Phase 6: Documentation, evidence reconciliation, and immediate architecture decisions.
    - [ ] 5.1.1 Close documentation gaps, reconcile overclaimed completion evidence, establish the phase template and decision register, and resolve only the contracts required to begin Phases 7–8 safely.
    - [ ] 5.1.2 Assign every later open decision an explicit must-resolve-before phase gate.
  - [ ] 5.2 Write new Phase 7: Shared core, SQLite, lifecycle, and artifact/process-ownership proof.
    - [ ] 5.2.1 Include core-versus-GUI state boundaries, command/query/change seams, migrations, one-writer/process ownership, startup/shutdown supervision, headless and display-free tests, and structured diagnostics.
    - [ ] 5.2.2 Use a durable local mock-service inventory with two same-provider instances as the first reusable vertical slice; do not advertise it remotely yet.
    - [ ] 5.2.3 Prove both GUI and headless artifacts against the same core and begin continuous Linux/Steam Deck and Android validation rather than deferring platform evidence to the release phase.
  - [ ] 5.3 Write new Phase 8: Identity, trust/device directory, secure REST, and reusable mock-queue proof.
    - [ ] 5.3.1 Include external-network configuration, enrollment, authoritative trusted-device records, mTLS, signed envelopes, REST resources, authorization, limits, audit, and revocation.
    - [ ] 5.3.2 Carry one durable mock workload through requester submission, owner validation, worker pull/lease, result return, owner-authoritative completion, restart, disconnection, idempotency, and correlation; retain these primitives for Phase 12.
  - [ ] 5.4 Write new Phase 9: Discovery, presence, owner-local Project registry, and cluster-wide Project catalog.
    - [ ] 5.4.1 Include stable owner-local Project identity, signed authorization-filtered summaries, discovery and presence, remote summary caching, revision/freshness, stale/offline display, and owner-directed REST reads.
    - [ ] 5.4.2 Keep repository contents, Git state, Task definitions, and mutations owner-authoritative and defer full workspace operations to Phase 10.
  - [ ] 5.5 Write new Phase 10: Project workspaces, Git, Pipelines, and constrained manual Tasks.
    - [ ] 5.5.1 Include safe files/Git, chats, drafts, transactions, conflicts, artifacts, Pipeline derivation, Task entrypoint schemas, permissions, approvals, and durable run history.
    - [ ] 5.5.2 Permit manual Tasks with no trigger, but limit execution to constrained owner-local actions or mock/local executors until queue-backed steps and routing arrive in Phases 12–13.
    - [ ] 5.5.3 Resolve Task sandboxing, secret references, artifact handling, and no-unrestricted-remote-shell rules before executable Task entrypoints are admitted.
  - [ ] 5.6 Write new Phase 11: Multi-instance local inference drivers, health, capabilities, and advertisements.
    - [ ] 5.6.1 Include static provider registration, arbitrary same-provider instances, discovery and verification, desired/observed state, independent supervision, capability inventory, advertisement revision/expiry, safe gateway projection, and GUI/headless read models.
    - [ ] 5.6.2 Progress from the Phase 7 mock driver to OpenAI-compatible text, Ollama, llama.cpp, approved image drivers, and typed future workload contracts without coupling services to queue ownership or routing policy.
  - [ ] 5.7 Write new Phase 12: Full authoritative queue protocol, leasing, results, artifacts, and recovery.
    - [ ] 5.7.1 Extend the Phase 8 mock-queue primitives into typed direct and pool-owner queues, admission, priorities, leasing/fencing, heartbeats, cancellation, retries, retention, artifact validation, worker failure isolation, and restart recovery.
    - [ ] 5.7.2 Keep the queue owner authoritative while eligible inference workers pull or long-poll and return signed outcomes over REST.
  - [ ] 5.8 Write new Phase 13: Pools, routing profiles, deterministic fallback, and first real text generation.
    - [ ] 5.8.1 Include capability matching, pool membership, route explanations, privacy and policy constraints, specific-service targeting, load/admission signals, ordered fallback, and an end-to-end real text-generation path.
  - [ ] 5.9 Write new Phase 14: Automation, scheduling, webhooks, event triggers, approvals, and durable workflows.
    - [ ] 5.9.1 Add interval/cron, authenticated webhook, application-event, and cluster-event bindings without changing Project/Task ownership or requiring a trigger for manual execution.
    - [ ] 5.9.2 Include durable steps, queue/job linkage, checkpoints, retries, timeouts, cancellation, safe actions, redacted diagnostics, restart recovery, and configured human approval.
  - [ ] 5.10 Write new Phase 15: ASR, TTS, push-to-talk, audio lifecycle, and privacy.
    - [ ] 5.10.1 Include typed local and remote speech routes, bounded capture, cancellation, temporary storage/deletion, editable transcription, independently routed speech output, privacy indicators, and GUI/platform integration.
    - [ ] 5.10.2 Keep capture, focus, and pre-submission audio in GUI/platform state; create durable core work only on explicit submission.
  - [ ] 5.11 Write new Phase 16: Packaging, release hardening, and platform support evidence.
    - [ ] 5.11.1 Treat this as culmination of continuous platform validation, not its beginning: Steam Deck/Linux GUI and headless first, then independently validated Windows, macOS, and Android.
    - [ ] 5.11.2 Include reproducibility, signing, provenance, installers/services, upgrade, rollback, storage/network/audio/lifecycle evidence, and honest per-target support declarations.
  - [ ] 5.12 Define post-MVP work as independently dependency-gated tracks rather than a false Phase 17–19 dependency chain.
    - [ ] 5.12.1 Alternative transports and resilience: conformance, Meshtastic, Signal, optional WireGuard management, failover, ownership migration, replication, CRDT research, and routing optimization.
    - [ ] 5.12.2 Comrades, chat, and shared inference: identity/trust, chat, grants, owner-authoritative shared queues, quotas, safety, abuse controls, and HUD requirements; require stable identity, queues, routing, audit, and one suitable authenticated transport, not every alternative transport.
    - [ ] 5.12.3 Research, BOINC, and validation gameplay: trust/evidence, BOINC boundary, resource policy, isolation, scheduling, packaging, HUD, recovery, provenance, and owner control; do not make it depend on Comrades or alternative transports where no real dependency exists.
    - [ ] 5.12.4 If canonical roadmap formatting requires sequential phase numbers, number these tracks by product priority while stating their independent prerequisites explicitly.
  - [ ] 5.13 Require every implementation phase to specify user outcome, scope and deferrals, dependencies, domain/state ownership, SQLite changes, REST/API changes, authorization, adapters, GUI projection, failure/recovery behavior, tests, documentation, target-platform evidence, and exit criteria.
  - [ ] 5.14 Require every functional phase to end in a reviewable vertical slice from core state through persistence and command/query or REST surfaces to GUI projection, including failure and restart evidence where applicable.
  - [ ] 5.15 Retain and refine cross-cutting security, reliability, observability, privacy, performance, recovery, platform-validation, and documentation requirements.
  - [ ] 5.16 Retain and refine the MVP completion definition against the new dependency order and concrete acceptance evidence.

- [ ] 6. Keep detailed contracts consistent with the new concrete roadmap.
  - [ ] 6.1 After rewriting Phase 6 onward, perform a final consistency sweep across README, architecture, database, API/OpenAPI, security, signed-envelope, transport, controller, and platform contracts for any invariant clarified during synthesis.
  - [ ] 6.2 Make Project/Pipeline/Task ownership, queue-owner worker-pull REST execution, shared-core state, and multi-instance inference rules explicit and consistent everywhere.
  - [ ] 6.3 Mark contract, implemented, partially validated, planned, and deferred behavior where ambiguity would mislead implementation.
  - [ ] 6.4 Update every canonical link or status statement that still treats `RECOMMENDATIONS.md` as a continuing planning authority so it points to the consolidated Roadmap or an appropriate detailed contract.
  - [ ] 6.5 Keep changed source, script, build, test, and code directories documented at the closest useful layer when later feature implementation begins.

- [ ] 7. Validate the Phase 6 onward replacement before completing the plan.
  - [ ] 7.1 Confirm every legacy Phase 6–14 item, carryover dependency, cross-cutting requirement, open decision, and MVP outcome maps to the new Phase 6–N program or has an explicit retirement/defer rationale.
  - [ ] 7.2 Confirm the new phase order is internally consistent from documentation gates through shared core, secure clustering, Project catalog/workspaces, local services, authoritative queues, routing, automation, voice, and release proof, with post-MVP tracks ordered only by their actual prerequisites.
  - [ ] 7.3 Confirm the new roadmap adds concrete architecture, state, protocol, persistence, lifecycle, failure, and acceptance detail rather than merely restating broad goals.
  - [ ] 7.4 Run documentation, plan-index, whitespace, link, OpenAPI, and repository verification appropriate to the changed files.
  - [ ] 7.5 Review the final diff to confirm that content before legacy Phase 6 remains preserved and the new Phase 6–N program is the sole canonical plan for the remaining project.
  - [ ] 7.6 Confirm `RECOMMENDATIONS.md` contains no unique retained requirement, decision, dependency, acceptance criterion, or future plan that is absent from the new Roadmap or the appropriate canonical detailed contract.
  - [ ] 7.7 Confirm every item excluded from the consolidated Roadmap has a retirement or defer rationale preserved in this plan for archival evidence.
  - [ ] 7.8 Request approval before appending the roadmap-replacement checkpoint to the journal.

- [ ] 8. Retire the temporary research document and complete the plan lifecycle.
  - [ ] 8.1 Confirm the Recommendations deletion gate: incorporation and retirement mappings are complete, no canonical document depends on the file, no unique retained content remains in it, and the pre-deletion validation suite passes.
  - [ ] 8.2 Remove remaining links and status text that refer to `RECOMMENDATIONS.md`, then delete `RECOMMENDATIONS.md`.
  - [ ] 8.3 Re-run documentation, OpenAPI, plan-index, whitespace, link, and repository verification after deletion and review the final diff for semantic completeness.
  - [ ] 8.4 Mark every completed or intentionally closed checklist item accurately, set this plan to `past`, move it from `plans/current/` to `plans/past/`, and regenerate all plan indexes.
  - [ ] 8.5 Confirm `ROADMAP.md` is the sole canonical implementation sequence for the remaining project and the archived plan is historical migration evidence rather than competing authority.

## Completion Criteria

- `ROADMAP.md` retains its pre-Phase-6 baseline and completed-history content.
- Everything from legacy Phase 6 onward is replaced by a new detailed Phase 6–N implementation program.
- The new program converts the completed GUI/mockup foundation into a concrete path for durable core functionality, REST cluster operation, Projects, Pipelines, Tasks, queues, multi-instance inference, automation, platform release, and independently dependency-gated resilience, Comrades, and Research tracks.
- The new roadmap is concrete enough to bind focused implementation plans without re-solving ownership, state, protocol, persistence, lifecycle, or acceptance criteria each time.
- Every Recommendations and legacy future-Roadmap item is mapped to the new sequence or has an explicit retirement/defer rationale preserved in the archived plan.
- `RECOMMENDATIONS.md` has been deleted after successful incorporation and reference migration.
- This plan has been completed and archived under `plans/past/` as migration evidence.
- `ROADMAP.md` is the sole canonical implementation sequence for the remaining project; `README.md` remains the canonical product contract and detailed documents remain authoritative for their named contracts.
