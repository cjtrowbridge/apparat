# HTTPS REST API Contract

The MVP API is authenticated HTTPS REST over externally configured WireGuard or trusted LAN. Authentication and authorization are mandatory on every non-public endpoint.

OpenAPI source: [`openapi/apparat-v1.yaml`](./openapi/apparat-v1.yaml).

## Endpoints

- `GET /v1/health`: health, version, readiness, and clock state.
- `GET /v1/device`: authenticated device identity, roles, and trust state.
- `GET /v1/capabilities`: typed device/service capability descriptors.
- `POST /v1/jobs`: submit an asynchronous workload job.
- `GET /v1/jobs/{id}`: inspect job state.
- `POST /v1/jobs/{id}/cancel`: request cancellation.
- `GET /v1/events?after={cursor}&wait={duration}`: cursor-based event long-polling.
- `POST /v1/project-transactions`: submit owner-authoritative project mutations.

Mutating requests require `Idempotency-Key`. Asynchronous success returns `202 Accepted`, a durable resource ID, and a `Location` header.

Requests enforce schema version, content type, body size, deadlines, bounded concurrency, authentication, authorization, replay protection, and redaction-safe errors.

The API exposes no generic remote shell, arbitrary process execution, or unrestricted tool endpoint.
