# Inference Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `llama.cpp` | `https://github.com/ggml-org/llama.cpp.git` | `fdb1db877c526ec90f668eca1b858da5dba85560` (`b9860`) | MIT | Local text-generation runtime and service-adapter reference | Future external or separately packaged service; not linked into the initial HUD binary |

## Selection Notes

- The first routing vertical slice begins with an OpenAI-compatible text-generation adapter and can target a separately running llama.cpp server.
- Keeping llama.cpp behind a service adapter avoids coupling the HUD lifecycle to model loading, accelerator backends, native compilation, and large model files.
- Models, context limits, quantization, acceleration, and device eligibility are advertised as typed capabilities rather than inferred from the presence of this source.
- Adding this checkout does not select a model, download weights, validate GPU support, or grant unrestricted remote execution.
- Image generation, video generation, TTS, STT, and BOINC remain distinct workload classes with their own adapters and capability requirements.
