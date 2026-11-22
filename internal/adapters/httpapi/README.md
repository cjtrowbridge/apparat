# httpapi

Loopback-only read-only JSON mappings over the application query boundary. Numeric loopback addresses, bounded requests/responses/concurrency and safe errors enforce the Phase 8 boundary. No authentication or mutation is implied; Phase 9 owns those. Verify with httptest and listener/lifecycle tests.

## Source inventory

- `server.go`: Serves bounded loopback-only read queries without persistence access.

- `server_test.go`: Tests the loopback wire contract, bounded admission and safe failure mapping.
- `claim_tls.go`: Creates an ephemeral TLS leaf certified by the unlocked node identity. The claim client verifies TLS 1.3, the pinned root and the leaf name.
- `claim_server.go`: Serves JSON public-key lookup and first-owner claims behind the existing Linux interface-granted listener. Limits request size and lifetime.
- `claim_client.go`: Posts the dual-signed JSON claim only to the invited numeric address after verifying the worker's TLS identity. Redirects and environment proxies are disabled.
- `claim_test.go`: Exercises a real TLS 1.3 handshake, rejects a wrong pin and retries an acknowledged claim without creating another admin.
