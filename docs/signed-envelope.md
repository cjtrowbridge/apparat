# Signed Envelope Contract

Durable cross-device operations use transport-independent signed envelopes.

## Fields

- Envelope version.
- Message type.
- Message ID.
- Idempotency key.
- Correlation ID.
- Sender identity.
- Recipient target.
- Created time, expiration time, and optional deadline.
- Payload type and schema version.
- Payload length and hash.
- Inline payload or artifact reference.
- Signature algorithm and signature.

## Encoding

The MVP uses UTF-8 JSON canonicalized according to RFC 8785 JSON Canonicalization Scheme rules. Timestamps are integer UTC milliseconds. The payload hash is SHA-256 over the exact inline payload bytes; a JSON payload is itself RFC 8785-canonicalized before hashing. An artifact reference hashes its canonical metadata containing at least artifact ID, owner, SHA-256 content digest, byte length, and MIME type.

The sender signs the RFC 8785-canonical envelope with the `signature` value omitted using its Apparat Ed25519 device key. `signature_algorithm` is therefore `ed25519` for the MVP and `payload_hash_algorithm` is `sha256`. The verified signed device record supplies the authorized signing-key binding; the mTLS leaf key is separate and cannot substitute for the envelope key.

Future constrained transports may carry compact binary encodings only after a versioned deterministic mapping to the same logical fields and signature input is defined. An adapter cannot reinterpret JSON numbers, timestamps, omitted fields, payload bytes, or artifact metadata while claiming the same signature.

## Validation

Receivers verify canonical form, version, device-record signing-key binding, signature, sender authorization, recipient target, created/expiration/deadline times, payload hash, size, replay state, idempotency key, and schema compatibility before applying work.

Duplicate messages return the durable prior outcome where possible. Unknown future versions fail closed unless an approved negotiation rule exists.

## Owner And Lease Bindings

HTTPS REST is the MVP transport for these envelopes. The envelope supplements mTLS and request authorization; it does not let a sender bypass the authoritative owner.

- Project operations bind the owning device, `project_id`, operation/transaction ID, and relevant base revision. Cached remote Project summaries cannot authorize or apply a mutation.
- Task-run requests bind the Project owner, `project_id`, `task_id`, run/idempotency identity, initiating actor or trigger, and input schema version.
- Queue submissions bind the queue owner, `queue_id`, `job_id`, requester, workload/schema requirements, and idempotency identity.
- Worker results bind the queue owner, queue, job, attempt, lease, fencing token, worker device/service identity, terminal outcome schema, and artifact hashes.
- Service advertisements bind the owner device, monotonic revision, observation and expiration times, logical service/capability IDs, safe health/policy fields, and advertisement digest. They exclude provider-local endpoints and credential references.

The receiving Project or queue owner validates those bindings against current durable state before applying the request. Expired or superseded leases, mismatched workers, stale fencing tokens, and duplicate terminal outcomes fail without changing an already authoritative result.
