# Deep-Research Decision Register

These are proposed dispositions, not execution approval. “Accepted” means the finding should be represented in the Phase 7 amendment proposal; it does not alter the existing plan.

| ID | Claim(s) | Disposition | Decision and rationale | Prerequisite / verification consequence |
| --- | --- | --- | --- | --- |
| DD-01 | DR-02 | Accepted | Centralize persistence ownership, migrations, transactions, and schema creation as Phase 7 already intends. This prevents raw SQL and independent runtime DDL from becoming a broader dependency. | Amend Phase 7 only to make existing §4.8 explicit about migrating transitional DDL and testing it. |
| DD-02 | DR-03 | Accepted | A storage failure must not be converted into a successful duplicate result. This is a small, high-leverage correction before any replay semantics become remotely visible. | Add an error-classification test; decide whether the transitional store is corrected or removed during §4.8. |
| DD-03 | DR-04 | Accepted | The report confirms that the current lifecycle differs from the already approved target. The remedy belongs in the existing §5 scope. | Add explicit failure-injection coverage that proves reverse cleanup and safe repeated close. |
| DD-04 | DR-05 | Accepted | GUI/HUD dependencies must leave the application runtime to satisfy the existing shared-core boundary. | Add compile-level/import-boundary evidence and direct-GUI-query evidence to §13 closure. |
| DD-05 | DR-06 | Accepted | `StatusReady` currently proves only local file presence. Phase 7 should prohibit it from implying authorization and harden local parsing/writes before Phase 8 trusts identity. | Add tests and documentation around local-bootstrap semantics; Phase 8 owns actual remote authorization. |
| DD-06 | DR-07 | Accepted | An unconditional Android readiness result is misleading and blocks honest lifecycle evidence. | Add failure/readiness/lifecycle tests or explicitly record device validation pending; preserve Phase 16 release work as deferred. |
| DD-07 | DR-10 | Accepted, already covered | Per-connection foreign-key behavior belongs in existing Phase 7 §4.2. | Retain existing requirement; no new amendment item. |
| DD-08 | DR-01 | Accepted | User selected canonical ULID as the Phase 7 durable ID contract before new durable or external consumers proliferate. | Add canonical format, injection, ordering, and compatibility checks to Phase 7 §6. |
| DD-09 | DR-08 | Accepted | User selected targeted comments on critical lifecycle, persistence, identity, and boundary code during Phase 7, with prospective review rather than a repository-wide backfill. | Add focused comment scope and semantic review sampling to Phase 7 §20; do not treat presence counting as quality proof. |
| DD-10 | DR-09 | Deferred | The report's historical governance diagnosis is valuable, but this narrow review did not audit enough historical evidence to turn it into a mandatory Phase 7 process change. | Revisit after Phase 7 with a bounded retrospective if recurring evidence gaps appear. |

## Resolved user decisions

1. **ID contract:** adopt canonical ULID now.
2. **Comment policy:** require targeted high-risk boundary comments during Phase 7 and prospective semantic review, not a project-wide backfill.

All accepted decisions are represented as bounded clarifications or additions to the existing Phase 7 plan.
