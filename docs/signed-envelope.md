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
