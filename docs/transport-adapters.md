# Transport Adapter Contract

Transports carry signed Apparat messages; they do not define product semantics.

## Capability Descriptor

Each transport declares payload size, attachment support, direct addressing, broadcast, online delivery, delayed delivery, acknowledgements, fragmentation, store-and-forward, latency, cost, ordering, and privacy properties.

## REST And Future Encodings

HTTPS REST is the authoritative MVP request transport between devices. Project discovery/access, Task invocation, queue submission, worker claim/long-poll, lease heartbeat, and result completion are REST operations directed to the owning device. No transport adapter creates shared-database or direct-filesystem access.

HTTPS REST carries JSON envelopes and artifact references. Future compact transports such as Meshtastic or Signal gateways carry the same logical envelope fields with constrained payloads, fragmentation, and command allowlists. They must preserve Project-owner and queue-owner authority, idempotency, lease/fencing, and result-validation semantics; a constrained adapter may omit an unsupported operation but may not invent a second authoritative queue or project copy.

Large prompts, model data, project files, and artifacts do not belong on constrained transports.
