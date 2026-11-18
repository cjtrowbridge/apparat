# Database Contract

SQLite stores durable application metadata and workflow state; ordinary project files remain filesystem and Git data.

Phase 8 implements SQLite lifecycle, migrated transitional cluster/messaging tables, service intent/observations/capabilities and verified backup/restore. Project, Task, queue, artifact and remote-projection records below remain target contracts.

## Layers

- Identity, certificates, trust, enrollment, and directory records.
- Device and typed capability records.
- Project metadata, owner-local Git repository locations, remote project summaries, chats, artifacts, transactions, drafts, and cursors.
- Typed services, queues, jobs, leases, attempts, results, and artifacts.
- Events, outbox, inbox, idempotency keys, replay records, and audit logs.
- Project Task entrypoints, trigger bindings, schedules, webhooks, run history, and approval state.
- Research project catalog, validation state, budgets, work-unit references, and provenance.

## Ownership

Each MVP project and queue has exactly one authoritative owner device. A Task belongs to one Project, so its authoritative definition, trigger bindings, and run records are owned by that Project's device during the MVP. Remote changes are idempotent REST-submitted transactions or queue protocol requests.

## Projects, Pipelines, Tasks, And Cluster Projection

- An owner-local Project record identifies a Git repository and its validated working-tree root on the owning device. Repository contents remain in Git/the filesystem rather than SQLite.
- A remote Project-summary record is a cached, signed, authorization-filtered projection containing stable `project_id`, `owner_device_id`, display metadata, capabilities, revision, observation/expiry time, and online/stale/unavailable state. It never contains a path that another device should treat as locally accessible.
- Every Apparat device queries local owner records plus authorized remote summaries to build the cluster-wide Projects list. A remote summary cannot change `owner_device_id` or become authoritative because it was cached.
- A Pipeline is derived from `Project has one or more enabled Apparat Task entrypoints`; it does not require a separate ownership table or second project identity.
- A Task record includes stable `task_id`, `project_id`, owner device, entrypoint/schema version, typed inputs/outputs, execution policy, permissions, enabled state, and revision.
- Trigger bindings are separate records keyed to a Task. A Task may have zero bindings and remain manually runnable. Bindings may represent interval/cron schedules, authenticated webhooks, internal application events, or cluster events.
- Task-run records preserve run/correlation/idempotency IDs, initiating actor/trigger, input references, current step, queued jobs, approvals, result/artifact references, retry state, and terminal outcome.

## Queue Owner And Pull-Lease State

- The queue owner stores the only authoritative queue definition, admission order, jobs, attempts, leases, cancellation state, results, retention state, and audit trail.
- A requester stores its durable outbound submission and authorized cached status/result, not a mirrored authoritative queue row.
- Submission records preserve requester, stable job/message/correlation/idempotency IDs, schema/workload requirements, policy decision, and durable acceptance or rejection.
- Workers poll or long-poll the owner through REST. A successful claim creates an attempt and lease containing `lease_id`, worker device/service identity, fencing token, issue/expiry time, heartbeat policy, and attempt number.
- Heartbeat/progress updates and terminal completions are accepted only for the active worker and fencing token. Completion is idempotent; expired, stale, duplicated, or superseded lease results cannot overwrite authoritative state.
- Lease expiry may make work eligible for another attempt. Attempt history is retained so reassignment never erases the earlier execution record.
- Result/artifact records become authoritative only after the owner validates worker authorization, lease state, schema, integrity, and policy.

## Local Inference Service State

The service model must preserve zero-to-many cardinality for every driver and workload class. `driver_kind` and `workload_class` are never primary instance identity.

- `inference_service_instances` stores owner-local desired configuration keyed by stable `service_id`: local device, driver kind, display name, local endpoint, enabled/advertise flags, schema-versioned provider configuration, opaque credential reference, admission/concurrency policy, revision, and timestamps.
- `inference_service_observations` stores last-known observed lifecycle and health independently from desired state: service ID, observation time, reachability, safe failure code/message, inventory hash, availability, load, and probe revision.
- `inference_service_capabilities` stores observed capabilities keyed by service and stable `capability_id`: workload class, schema version, model ID, modalities, formats, features, limits, capability data, and observation time.
- `service_advertisement_state` stores the last derived owner-scoped revision, observation time, expiration, publication status, and safe advertisement digest. Advertisements do not store or expose local endpoints, credential references, tokens, raw provider failures, prompts, or results.
- Cached remote service advertisements preserve signer/owner, revision, observation and expiration times, safe health/capability fields, and stale state. Expiry immediately removes routing eligibility; the default stale diagnostic retention is 24 hours.

The default advertisement lifetime is 120 seconds and an eligible owner refreshes by 60 seconds. A newer owner-scoped revision supersedes older revisions. Re-advertisement after expiry requires a fresh observation and revision.

SQLite stores only opaque credential references. A secret adapter resolves them from an OS credential store when available or an Apparat-managed encrypted secret file under the node runtime root. Secret values never appear in general provider JSON, replicated records, advertisements, logs, backups lacking the corresponding secret policy, or GUI read models.

## Artifacts

Artifact content is stored as files beneath the authoritative node's runtime artifact root; SQLite stores metadata and lifecycle state rather than large blobs.

An artifact record includes stable artifact ID, owner device, SHA-256 digest, byte length, MIME type, provenance, Project/job/result links, authorization policy, creation/expiry times, retention state, and local content location. Transfer records preserve source/recipient, expected digest and size, completed ranges/chunks, retry state, deadlines, and terminal validation outcome.

Authenticated transfers are bounded and resumable. Received bytes remain temporary until size and digest validation succeeds, then become atomically visible. A queue owner accepts a worker result only after validating the active lease/fencing token and every referenced artifact. Retention cleanup records deletion explicitly; it cannot silently leave a successful authoritative result claiming that missing content is available.

## Migrations And Data Types

Migrations are forward-only and checksumed. IDs are ULID-compatible sortable strings. Timestamps are UTC integer milliseconds; human schedule timezone data is stored explicitly when needed.

Phase 3 implements the first SQLite lifecycle:

- Open, close, and ping via the cgo-free `modernc.org/sqlite` driver.
- Enable foreign keys at connection startup.
- Keep WAL opt-in until platform behavior is validated; Phase 3 does not claim universal WAL support.
- Apply forward-only migrations with stored SHA-256 checksums.
- Detect migration checksum mismatches.
- Provide read-only diagnostics for user version, foreign-key status, and migration count.
- Provide sortable local IDs and UTC millisecond helpers.

The node runtime permits exactly one authoritative SQLite writer process. `apparat` and `apparatd` now use one logical default node root and an exclusive runtime lock. Multiple independent nodes on one host require explicit roots and identities rather than concurrent writers to one database.

## Backup, Repair, Restore, And Encryption

The MVP defines backup, repair, restore, and integrity checks before broad replication. Optional at-rest encryption remains a Phase 8 decision gate, not a default until key storage, backup interaction, and recovery are designed and validated.

## Implemented local schema and recovery

`internal/adapters/persistence` privately owns one writable `database/sql` connection. Foreign keys are enabled in the connection URI (including replacement connections); busy timeout is 1,500 ms. WAL is disabled. Only typed SQLite BUSY/LOCKED errors are retryable; cancellation, corruption and foreign-key errors are not replay hits. Callers own any bounded retry decision; no automatic retry subsystem exists.

Forward migrations preserve the original Phase 3 migration bytes/checksum, adopt transitional cluster/messaging tables as version 2 and add `services`, `service_observations`, `service_capabilities` at version 3. Versions must increase and names must be unique; checksum/name mismatches fail visibly. Each migration is transactional. Older development IDs in transitional tables remain opaque and unchanged. New service/capability IDs use canonical 26-character Crockford ULIDs; no legacy identifier is silently reinterpreted.

Back up through the owning runtime with `--backup /new/backup.db`: integrity check, SQLite `VACUUM INTO`, file sync and a SHA-256/schema/app metadata sidecar. Existing snapshot or metadata files cause refusal. Restore with `--restore-backup /backup.db --runtime-dir /new/disposable-root`: verify metadata/digest, create only a new destination, integrity-check it, confirm schema and close cleanly. A checksum proves file integrity, not origin/authenticity. Test the restored root with `--doctor` and `--services` before choosing it as authoritative.

For a full-node backup, stop every GUI/headless/older process and preserve the complete runtime root including encrypted identity files, manifests and application state. A database snapshot alone does not back up keys. Keep the original root and backup untouched during repair. On migration failure do not edit stored checksums or delete rows; retain failure diagnostics and restore a verified compatible snapshot into a new root. Downgrading code against a newer schema is unsupported; rollback means selecting the preserved old root with a matching old binary after all other owners stop. Cross-device/filesystem moves are manual offline recovery when atomic rename is unavailable. Optional database encryption and WAL remain gated on key recovery/platform evidence.

Migration 4 adds singleton preferences and desktop_window records. Updating geometry preserves display settings. Restore accepts snapshots from schema 3 or 4 and startup applies outstanding migrations.

Migration 5 adds tab placement to the existing preference record. Earlier rows receive the top layout and keep their display, startup and overlay choices. Restore accepts schema versions 3 through 5.

Migration 6 adds the local project registry and account references. Restore accepts schemas 3 through 6. Checkout files remain at their registered locations and require separate backup.

Migration 7 adds project-scoped drafts and durable entries for pending messages and manual Task runs. Restore accepts schemas 3 through 7. Artifact bytes reside in the runtime artifact directory; a database snapshot contains their metadata.
