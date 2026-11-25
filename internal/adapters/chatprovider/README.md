# Chat Completions adapter

- `completions.go`: translates one completed provider-neutral step to a non-streaming Chat Completions POST. Configure the full HTTPS endpoint (or loopback HTTP), model and optional bearer secret on the executing owner. Use a model that supports function tools and the max_tokens field. No redirect following, automatic retries or elapsed-time generation deadline. Wire responses are size limited. A transport interruption can leave an indeterminate job. Inspect its owner receipt before starting another request.

- `ollama.go`: translates loopback /api/chat messages and tool exchanges and streams visible text and phase changes to local preview subscribers. It retains numeric timings and speculative counters from the same completed response. The owner stores these on the step receipt. Other completion endpoints retain the OpenAI-compatible exchange.
- `ollama_test.go`: exercises native tool results, argument objects and numeric-only telemetry extraction.
