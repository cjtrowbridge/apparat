# httpapi

Loopback-only read-only JSON mappings over the application query boundary. Numeric loopback addresses, bounded requests/responses/concurrency and safe errors enforce the Phase 8 boundary. No authentication or mutation is implied; Phase 9 owns those. Verify with httptest and listener/lifecycle tests.

## Source inventory

- `server.go`: Serves bounded loopback-only read queries without persistence access.

- `server_test.go`: Tests the loopback wire contract, bounded admission and safe failure mapping.
- `claim_tls.go`: Creates an ephemeral TLS leaf certified by the unlocked node identity. The claim client verifies TLS 1.3, the pinned root and the leaf name.
- `claim_server.go`: Serves JSON public-key lookup and first-owner claims behind the existing Linux interface-granted listener. Limits request size and lifetime.
- `claim_test.go`: Exercises a real TLS 1.3 handshake, rejects a wrong pin and retries an acknowledged claim without creating another admin.
- `worker_auth.go`: Limits and expires login challenges and bearer credentials in memory. Current SQLite admission and device status gate each protected request.
- `worker_client.go`: Authenticates with the worker pin, a client TLS certificate and a fresh proof from both identities. Credentials stay in the client session. The protected identity and policy-page helpers use the same connection policy.
- `worker_admin.go`: Checks a signed proposal against the logged-in admin. A handoff to the worker requires the owner signature and the current authority's signed receipt. Later proposals can come from any approved admin.
- `worker_tasks.go`: Accepts and leases a local hash-text job, writes and verifies its result in the private pool artifact store, and serves the result by authorized job ID. Polling can resume interrupted work.
- `personal_pair.go`: Implements single-attempt personal-device confirmation. Both devices contribute a four-digit code. RFC 9807 OPAQUE authenticates both ends with its context bound to the TLS 1.3 exporter. The sender opens the protected transfer only after confirming the receiver. Expiry, cancellation, an invalid message or a mismatched code closes the connection.
- `personal_pair_test.go`: Verifies successful mutual authentication and rejects mistyped codes before releasing any payload.
- `replication.go`: Serves the protected batch sync route and catches a reader up using snapshots and per-pool queue cursors.
- `queue_client.go`: Submits, cancels and reads the user’s hash tasks. Result retrieval checks the artifact digest.

Worker clients probe TLS without client credentials and require an approved device fingerprint for login. The authenticated GitHub username directory also requires current admin access and host-local owner confirmation. Exercise the positive and denial paths with TestPinnedWorkerClaim.

- `git_jobs.go`: Executes and tracks device-bound administrator provider and clone operations.
- `git_jobs_test.go`: Runs private HTTP Git clones against a local fixture through the real worker admin API, including anonymous and wrong-account failures.

- `owner_workspace.go`: Adapts the scoped owner workspace command to a bounded authenticated JSON route.

- `live_workspace.go`: Multiplexes project invalidations on one pinned authenticated WebSocket per active owner.

- `live_workspace_test.go`: Exercises real notifications over the pinned TLS worker with two subscriptions.

- `traffic.go`: Counts application-layer HTTP traffic on both ends with a monotonic timer.

- `workspace_denial.go`: Translates recognized authorization codes and scope names into local recovery instructions. Legacy workers receive general investigation guidance.

- `workspace_denial_test.go`: Checks permission guidance through the worker client.

- `workspace_session_acceptance_test.go`: Checks all protected workspace route families for uniform JSON denial and verifies file integrity with anonymous, expired and incorrectly bound sessions.
- `inventory.go`: authenticated worker inventory GET and administrator-only metadata probe POST, with typed responses and failure guidance.
- `inventory_test.go`: checks administrator-only port trials, member redaction and actionable probe errors. Device-bound session tests cover both inventory routes.

- `chat_models.go`: Exchanges scoped model choices over the worker's authenticated, device-bound API.

- `model_live.go`: Emits invalidation hints without project IDs, model names or message contents.

- `inference.go`: transports user-scoped inference submissions and worker claim transitions. User operations use the existing bound login. Peer operations require a verified device certificate and membership in the agreed roster.

- `global_chunks.go`: Verifies offsets, session identity and revision while reading large global transcripts in bounded chunks.

- `preview.go`: Streams private per-job text deltas with live authorization and reconnect snapshots.
