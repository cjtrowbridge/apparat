# mockprovider

Deterministic local text/image mock probes, explicitly registered by composition roots. They perform no network or process work; cancellation and per-instance failure are observable. Verify with service-isolation tests.

## Source inventory

- `provider.go`: Supplies deterministic local mock probes for the first durable service slice.
