# Proposed Phase 7 Amendment from Reviewed Findings

Status: proposal only. This document does not modify `plans/future/2026-07-20-09-04-03_execute-phase-7.md`; apply only the accepted items after user approval.

## Proposed atomic additions and clarifications

| Proposal | Evidence | Proposed Phase 7 location | Required behavior and verification |
| --- | --- | --- | --- |
| P7-R1: migrate transitional schema ownership | DR-02 / DD-01 | Clarify §4.8 | Move `cluster` and `messaging` table creation into the selected persistence migration boundary. Repositories receive no raw `*sql.DB`; integration tests prove existing transitional operations work after migration and restart. |
| P7-R2: preserve replay failures | DR-03 / DD-02 | Add under §4.8 | `Seen` returns an error for cancellation, locking, corruption, and non-uniqueness storage failures; only the replay uniqueness violation yields “already seen.” Tests inject or reproduce both paths. If the store is removed, its replacement must meet the same contract. |
| P7-R3: prove transactional lifecycle, not just lifecycle types | DR-04 / DD-03 | Clarify §5.4-§5.7 | Failure injection at each initialized stage proves reverse-order cleanup, primary-error retention, truthful readiness, and deterministic repeated close. This is deterministic integration evidence, not an inference from the presence of state names. |
| P7-R4: enforce application/presentation direction | DR-05 / DD-04 | Clarify §1.2 and §13.1 | `internal/app` does not import GUI or HUD presentation packages and does not own `hud.Shell`; composition supplies direct application queries to GUI. Add an import-boundary check and a display-free direct-query test. |
| P7-R5: quarantine local identity readiness | DR-06 / DD-05 | Add under §16.1-§16.2 | Document and test that current identity classification is local bootstrap state only. Strict local validation and atomic writes are required; no Phase 7 API or runtime behavior may infer remote authorization from `StatusReady`. |
| P7-R6: make Android readiness truthful | DR-07 / DD-06 | Add under §16.4 and §19.4-§19.5 | Replace unconditional readiness with state reflecting initialization/lifecycle outcome. Test failure and success paths where feasible; record unavailable physical-device evidence as pending rather than successful. |
| P7-R7: adopt canonical ULID | DR-01 / DD-08 | Add under §6.4 | Make Phase 7 durable IDs canonical ULIDs with injected time/entropy dependencies, lexical/time-order checks, and an explicit compatibility decision for development data. |
| P7-R8: document critical boundaries | DR-08 / DD-09 | Add under §20 | Add human-useful comments to Phase 7-touched lifecycle, persistence, identity, replay, and platform-boundary functions. Require prospective semantic review of new or substantially changed critical-boundary functions; do not require a project-wide backfill. |

## Decisions intentionally withheld

| Candidate | Reason it is not an amendment yet |
| --- | --- |
| Project-wide function-comment backfill | DD-09 selected a targeted Phase 7 scope. A broad backfill remains out of scope unless separately approved. |
| Broad governance remediation | DR-09 remains a useful retrospective hypothesis, not verified enough to impose a new completion bureaucracy. |

## Non-goals preserved

The proposal keeps Phase 8 authentication/authorization, remote commands, enrollment, and replay-protected network envelopes out of Phase 7. It does not introduce a redesign, a full dependency audit, a daemon-client architecture, Android release support, or a broad documentation rewrite.

## Approval gate

The user accepted DD-08 and DD-09 on 2026-08-24. The resulting Phase 7 plan amendment preserves existing scope and turns each approved row into atomic checklist items with identified tests and exit evidence. Separate approval remains required before Phase 7 implementation begins.
