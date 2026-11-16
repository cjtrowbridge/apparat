# Inference Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `llama.cpp` | `https://github.com/ggml-org/llama.cpp.git` | `ad5cbfa911b975b2bf987d3bf7d8f2b3f36671ab` (`b12444-13-gad5cbfa91`) | MIT | Local text-generation runtime and service-adapter reference | Future external or separately packaged service; not linked into the initial HUD binary |

## Selection Notes

- The first routing vertical slice begins with an OpenAI-compatible text-generation adapter and can target a separately running llama.cpp server.
- Keeping llama.cpp behind a service adapter avoids coupling the HUD lifecycle to model loading, accelerator backends, native compilation, and large model files.
- Models, context limits, quantization, acceleration, and device eligibility are advertised as typed capabilities rather than inferred from the presence of this source.
- Adding this checkout does not select a model, download weights, validate GPU support, or grant unrestricted remote execution.
- Image generation, video generation, TTS, STT, and BOINC remain distinct workload classes with their own adapters and capability requirements.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Source-reference updates do not change root Go module pins or activate new integrations.
