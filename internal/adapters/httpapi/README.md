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
