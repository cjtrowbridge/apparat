# Worker inventory

- `monitor.go`: samples at startup and every 45 seconds. The first successful probe persists the default or selected port, and successful changes replace it. A restarted monitor exposes that saved configuration before its first health check. A failed trial on a different port leaves the working configuration intact.
- `ollama.go`: fixed loopback metadata requests verify the server version, list installed and loaded models and check completion support per digest. Rejects redirects, invalid payloads and oversized inventories. No user prompts or model execution.
- `resources_linux.go`: measures CPU use, logical cores, RAM, swap, local filesystem capacity and DRM or NVIDIA VRAM. Reserved disk blocks can make used plus available smaller than total. No mounted paths are advertised.
- `resources_other.go`: exposes portable CPU/OS information when native counters are unavailable.

Snapshots expire after two minutes. Unsupported measurements stay unknown. Verify with the metadata-only fixture and a running Linux worker.
- `monitor_test.go`: uses metadata-only loopback fixtures to verify model capabilities, loaded status, cache reuse, redirection rejection and preservation of a working configuration.

- `compute.go`: Measures parallel scalar CPU FP32 operations per second once per process, using a 50 ms arithmetic loop. Each multiply and add counts as one operation. Scheduling and CPU load affect the result.
