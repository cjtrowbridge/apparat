# httpapi

Loopback-only read-only JSON mappings over the application query boundary. Numeric loopback addresses, bounded requests/responses/concurrency and safe errors enforce the Phase 8 boundary. No authentication or mutation is implied; Phase 9 owns those. Verify with httptest and listener/lifecycle tests.

## Source inventory

- `server.go`: Serves bounded loopback-only read queries without persistence access.

- `server_test.go`: Tests the loopback wire contract, bounded admission and safe failure mapping.
