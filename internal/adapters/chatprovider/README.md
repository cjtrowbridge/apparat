# Chat Completions adapter

- `completions.go`: translates one completed provider-neutral step to a non-streaming Chat Completions POST. Configure the full HTTPS endpoint (or loopback HTTP), model and optional bearer secret on the executing owner. Use a model that supports function tools and the max_tokens field. No redirect following, automatic retries or elapsed-time generation deadline. Wire responses are size limited. A transport interruption can leave an indeterminate job. Inspect its owner receipt before starting another request.
