---
plan_id: 2026-07-19-17-19-34_migrate-recommendations-into-canonical-roadmap
title: Migrate Recommendations Into The Canonical Roadmap
summary: Reconcile documentation and decision conflicts, preserve completed history, and synthesize the recommendations into one canonical ROADMAP.md without losing future or historical fidelity.
status: future
created_at: 2026-07-19-17-19-34
---

# Migrate Recommendations Into The Canonical Roadmap

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding And Intended Outcome

- This plan governs the explicit documentation migration anticipated by `RECOMMENDATIONS.md` under `Roadmap Authority Migration Exit Criteria`.
- `README.md` remains the canonical product contract.
- `ROADMAP.md` remains the canonical filename for implementation sequence, dependencies, checklists, exit criteria, and the decision register so `AGENTS.md`, plan bindings, journals, and contributor workflows do not need a second authority model.
- Detailed completed history and Salvagecore provenance will be preserved in tracked history documents rather than discarded.
- The actionable structural and future-program content in `RECOMMENDATIONS.md` will move into `ROADMAP.md`; the review will then be archived so it cannot drift as a competing roadmap.
- This future plan does not authorize execution until the user explicitly approves it and it is promoted to `plans/current/`.

- [ ] 1. Reconcile product decisions and migration policy before rewriting roadmap authority.
  - [ ] 1.1 Confirm the canonical document topology.
    - [ ] 1.1.1 Keep `ROADMAP.md` as the sole canonical roadmap filename.
    - [ ] 1.1.2 Keep `README.md` as the canonical product and terminology contract.
    - [ ] 1.1.3 Define `docs/history/` as non-authoritative preserved evidence rather than an implementation queue.
    - [ ] 1.1.4 Define the archived recommendations document as a dated review whose accepted actions have moved into `ROADMAP.md`.
  - [ ] 1.2 Reconcile identity and protocol decisions that are currently both decided and open.
    - [ ] 1.2.1 Mark separate TLS leaf keys cryptographically bound to Apparat device identity as decided.
    - [ ] 1.2.2 Mark canonical JSON as the MVP signed-envelope encoding while retaining compatible compact encodings as future transport work.
    - [ ] 1.2.3 Define the exact MVP cluster-local X.509 CA and device-leaf hierarchy, issuance authority, rotation, revocation, and recovery boundary.
    - [ ] 1.2.4 Define a stable authorization vocabulary for Project discovery/read/write/Task execution, queue submission/inspection/cancellation/claim/heartbeat/completion, and service inspection/invocation/advertisement.
    - [ ] 1.2.5 Move genuinely unresolved decisions into stage-bound `required before` or `long-term deferred` registers.
  - [ ] 1.3 Resolve Task and scheduler authority terminology.
    - [ ] 1.3.1 State that every Task is a Project entrypoint owned with its Project.
    - [ ] 1.3.2 State that the Project owner evaluates the Task's optional trigger bindings and owns authoritative run records during the MVP.
    - [ ] 1.3.3 Replace `scheduler-owned Task` language with `Project-owner scheduler` language.
    - [ ] 1.3.4 Keep future scheduler failover behind explicit lease, fencing, clock, and duplicate-execution design.
  - [ ] 1.4 Resolve the first-feature sequence.
    - [ ] 1.4.1 Use a minimal local durable mock job to prove the Stage 0 shared-core boundary.
    - [ ] 1.4.2 Carry that mock job through the Stage 1 two-device REST queue protocol.
    - [ ] 1.4.3 Keep Project ownership and remote Project access in Stage 2.
    - [ ] 1.4.4 Keep the multi-instance inference manager and real provider drivers in Stage 3.
    - [ ] 1.4.5 Remove P0 language that would implement the inference inventory before the secure mock-queue and Project stages.

- [ ] 2. Close and status the documentation-first admission gate.
  - [ ] 2.1 Convert the prose `Canonical Documentation Gaps` list into a status-bearing checklist.
    - [ ] 2.1.1 Mark the Project/Pipeline/Task ownership contract complete after verifying it across README, architecture, API, database, security, transport, and roadmap documents.
    - [ ] 2.1.2 Mark the queue-owner validation and worker-pull REST contract complete after verifying it across the same contracts and OpenAPI.
    - [ ] 2.1.3 Keep incomplete inference, process-ownership, controller-truth, and platform-supervision documentation visibly pending until their atomic updates pass validation.
  - [ ] 2.2 Propagate the arbitrary multi-instance inference contract into canonical documentation.
    - [ ] 2.2.1 Update `README.md` with zero-to-many local service instances, multiple same-provider instances, logical service identity, and authenticated Apparat gateway access.
    - [ ] 2.2.2 Update `docs/architecture.md` with workload class, driver kind, service instance, capability/model, factory/registry/manager boundaries, and shared-core ownership.
    - [ ] 2.2.3 Update `docs/database.md` with desired service configuration, observed state, capabilities, advertisement revision/expiry, stable service/capability IDs, and local secret references.
    - [ ] 2.2.4 Update `docs/api.md` and `docs/openapi/apparat-v1.yaml` with logical service inventory, capability addressing, job targeting, route explanations, cancellation, progress/results, and safe health/error projections.
    - [ ] 2.2.5 Update `docs/security.md` with provider endpoint/credential non-disclosure, provider-compromise boundaries, enablement/advertisement policy, authorization, and audit rules.
    - [ ] 2.2.6 Update `docs/signed-envelope.md` and `docs/transport-adapters.md` with advertisement revision, observation time, expiry, stale behavior, and transport compatibility.
    - [ ] 2.2.7 Update `docs/platform-matrix.md` with node-runtime ownership, exclusive locking or daemon-client operation, provider supervision, and platform limitations.
  - [ ] 2.3 Reconcile controller, GUI, and implemented-versus-planned truth.
    - [ ] 2.3.1 Correct six-/seven-tab and `Alt+1` through `Alt+7` documentation to the canonical five-tab model.
    - [ ] 2.3.2 Audit focus, activation, back, context menu, command palette, scrolling, accessibility, and input-equivalence claims against executable evidence.
    - [ ] 2.3.3 Mark implemented-but-target-unvalidated behavior `[?]` and unimplemented behavior `[ ]` instead of preserving optimistic `[x]` status.
    - [ ] 2.3.4 Preserve corrections as later documentation evidence without rewriting append-only journal history.

- [ ] 3. Preserve completed and predecessor history outside the active implementation program.
  - [ ] 3.1 Establish the history-document boundary.
    - [ ] 3.1.1 Create `docs/history/README.md` explaining that the directory preserves non-authoritative historical evidence.
    - [ ] 3.1.2 Document how historical phase numbers and anchors map to the canonical roadmap's legacy aliases.
  - [ ] 3.2 Preserve the Salvagecore inheritance record.
    - [ ] 3.2.1 Move the full Salvagecore reference baseline into `docs/history/salvagecore-reference.md` without losing implemented/scaffolding/design-only distinctions.
    - [ ] 3.2.2 Preserve retained, adapted, rejected, and deferred concepts; reuse procedure; provenance map; and reference-removal criteria.
    - [ ] 3.2.3 Update README and roadmap links to the new tracked history location.
    - [ ] 3.2.4 Preserve the rule that no build, test, runtime, or product authority depends on `third_party/salvagecore`.
  - [ ] 3.3 Preserve and correct the completed foundation ledger.
    - [ ] 3.3.1 Move detailed completed Phase 0–5 checklists, goals, dependencies, evidence, and exit criteria into `docs/history/roadmap-phases-0-5.md`.
    - [ ] 3.3.2 Include explicit correction notes for reopened or target-unvalidated GUI and Android claims.
    - [ ] 3.3.3 Retain concise Phase 0–5 completion/carryover summaries and history links in the canonical roadmap.
    - [ ] 3.3.4 Preserve legacy headings or compatibility anchors needed by archived plans and journal references.

- [ ] 4. Synthesize one canonical `ROADMAP.md` from the reconciled sources.
  - [ ] 4.1 Rebuild the roadmap preamble and authority contract.
    - [ ] 4.1.1 Retain the status key and concise product baseline.
    - [ ] 4.1.2 Link to completed-history and Salvagecore records.
    - [ ] 4.1.3 State that roadmap checklist items require focused approved execution plans before implementation.
  - [ ] 4.2 Put documentation and decisions first.
    - [ ] 4.2.1 Add the status-bearing documentation gate.
    - [ ] 4.2.2 Add `decided`, `required before stage`, and `long-term deferred` decision registers.
    - [ ] 4.2.3 Add execution-plan admission and roadmap-maintenance rules.
  - [ ] 4.3 Integrate the structural recommendations.
    - [ ] 4.3.1 Add the shared headless-capable core and GUI-state ownership model.
    - [ ] 4.3.2 Add Project/Pipeline/Task and queue-owner/worker-pull invariants.
    - [ ] 4.3.3 Add the multi-instance inference service model.
    - [ ] 4.3.4 Add P0 structural admission requirements for package boundaries, read models, SQLite, lifecycle, tests, identity, logging, and updater safety.
    - [ ] 4.3.5 Add P1 GUI behavior and performance requirements and P2 build/release simplifications without duplicating later stage checklists.
  - [ ] 4.4 Integrate the ordered implementation program.
    - [ ] 4.4.1 Add Stage 0 structural admission and corrected legacy Phase 3/5 validation carryover.
    - [ ] 4.4.2 Add Stages 1–9 with legacy Phase 6–14 aliases, goals, dependencies, atomic checklists, and exit criteria.
    - [ ] 4.4.3 Preserve every unfinished, validation-pending, deferred, cross-cutting, and open-decision legacy item or record an explicit retirement reason.
    - [ ] 4.4.4 Preserve the integrated MVP completion definition.
    - [ ] 4.4.5 Remove duplicated tasks, contradictory ownership language, and conflicting sequence statements.
  - [ ] 4.5 Keep the roadmap operationally maintainable.
    - [ ] 4.5.1 Keep stable headings and legacy phase aliases for archived plan and journal readability.
    - [ ] 4.5.2 Keep historical evidence out of active pending checklists.
    - [ ] 4.5.3 Keep implemented, validated, pending, and deferred status explicit.

- [ ] 5. Retire `RECOMMENDATIONS.md` as a competing implementation authority.
  - [ ] 5.1 Preserve the dated architecture review.
    - [ ] 5.1.1 Move the review and migration rationale to `docs/history/2026-07-19-architecture-recommendations.md` after accepted actions are represented in `ROADMAP.md`.
    - [ ] 5.1.2 Add a prominent archived/non-authoritative status and link to the canonical roadmap.
  - [ ] 5.2 Remove active-document ambiguity.
    - [ ] 5.2.1 Update README to name only `ROADMAP.md` as the canonical sequence and decision register.
    - [ ] 5.2.2 Remove or replace root `RECOMMENDATIONS.md` only after all tracked links are migrated.
    - [ ] 5.2.3 Verify `AGENTS.md` still accurately binds implementation plans to `ROADMAP.md`; change it only if the canonical filename contract changes.
    - [ ] 5.2.4 Verify no plan template, playbook, package README, or contributor documentation treats the archived review as active authority.

- [ ] 6. Validate fidelity, consistency, and repository governance.
  - [ ] 6.1 Validate roadmap-content fidelity before removing legacy sections.
    - [ ] 6.1.1 Compare every incomplete, validation-pending, and deferred legacy checklist item with the synthesized roadmap.
    - [ ] 6.1.2 Compare every future goal, dependency, exit criterion, cross-cutting requirement, open decision, and MVP criterion.
    - [ ] 6.1.3 Confirm every retired item has an explicit reason and replacement or non-goal.
  - [ ] 6.2 Validate document links and authority.
    - [ ] 6.2.1 Search for stale `ROADMAP.md` anchors and update moved Salvagecore/history links.
    - [ ] 6.2.2 Search for active `RECOMMENDATIONS.md` authority references and migrate them.
    - [ ] 6.2.3 Verify legacy Phase 3–14 references in archived plans and journals remain understandable without rewriting historical records.
    - [ ] 6.2.4 Validate the OpenAPI YAML with an available parser or documented repository lint target.
  - [ ] 6.3 Run repository checks.
    - [ ] 6.3.1 Run `make check-docs`.
    - [ ] 6.3.2 Run `python3 agents/scripts/regenerate_plan_indexes.py --check --repo-root .`.
    - [ ] 6.3.3 Run `git diff --check` and distinguish unrelated pre-existing user changes from migration changes.
    - [ ] 6.3.4 Run `make verify` because the migration changes canonical product and implementation-governance documentation.
    - [ ] 6.3.5 Review the final diff for accidental loss of historical, security, protocol, platform, or future-feature detail.
  - [ ] 6.4 Complete the documentation checkpoint only after approval.
    - [ ] 6.4.1 Present the synthesized roadmap, moved history files, archived review, link changes, and validation results for review.
    - [ ] 6.4.2 Request approval before appending the migration checkpoint to the current journal.
    - [ ] 6.4.3 Regenerate affected plan indexes after all approved plan lifecycle changes.

## Completion Criteria

- `ROADMAP.md` is the single canonical implementation sequence and decision register.
- `README.md`, `ROADMAP.md`, detailed contracts, and repository governance instructions agree on authority and terminology.
- Product baseline, completed Phase 0–5 evidence, Salvagecore provenance, every unfinished future item, cross-cutting requirements, open decisions, and MVP criteria remain tracked.
- Project/Pipeline/Task ownership, queue-owner REST leasing, and multi-instance inference semantics are explicit and non-contradictory.
- Stale tab/input/status claims are corrected or visibly reopened.
- `RECOMMENDATIONS.md` no longer competes with the roadmap.
- Documentation, plan-index, link, OpenAPI, whitespace, and repository verification gates pass.
