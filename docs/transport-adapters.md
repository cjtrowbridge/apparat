# Transport Adapter Contract

Transports carry signed Apparat messages; they do not define product semantics.

## Capability Descriptor

Each transport declares payload size, attachment support, direct addressing, broadcast, online delivery, delayed delivery, acknowledgements, fragmentation, store-and-forward, latency, cost, ordering, and privacy properties.

## REST And Future Encodings

HTTPS REST carries JSON envelopes and artifact references. Future compact transports such as Meshtastic or Signal gateways carry the same logical envelope fields with constrained payloads, fragmentation, and command allowlists.

Large prompts, model data, project files, and artifacts do not belong on constrained transports.
