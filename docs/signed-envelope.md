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

The MVP signs canonical JSON with explicit UTF-8, sorted object keys, integer UTC millisecond timestamps, and no insignificant whitespace. Future constrained transports may carry compact binary encodings if they preserve the same logical fields and signature semantics.

## Validation

Receivers verify version, signature, sender authorization, recipient target, expiration, payload hash, size, replay state, idempotency key, and schema compatibility before applying work.

Duplicate messages return the durable prior outcome where possible. Unknown future versions fail closed unless an approved negotiation rule exists.

## Owner And Lease Bindings

HTTPS REST is the MVP transport for these envelopes. The envelope supplements mTLS and request authorization; it does not let a sender bypass the authoritative owner.

- Project operations bind the owning device, `project_id`, operation/transaction ID, and relevant base revision. Cached remote Project summaries cannot authorize or apply a mutation.
- Task-run requests bind the Project owner, `project_id`, `task_id`, run/idempotency identity, initiating actor or trigger, and input schema version.
- Queue submissions bind the queue owner, `queue_id`, `job_id`, requester, workload/schema requirements, and idempotency identity.
- Worker results bind the queue owner, queue, job, attempt, lease, fencing token, worker device/service identity, terminal outcome schema, and artifact hashes.

The receiving Project or queue owner validates those bindings against current durable state before applying the request. Expired or superseded leases, mismatched workers, stale fencing tokens, and duplicate terminal outcomes fail without changing an already authoritative result.
