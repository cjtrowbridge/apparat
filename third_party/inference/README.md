# Inference Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `llama.cpp` | `https://github.com/ggml-org/llama.cpp.git` | `ad5cbfa911b975b2bf987d3bf7d8f2b3f36671ab` (`b12444-13-gad5cbfa91`) | MIT | Local text-generation runtime and service-adapter reference | Future external or separately packaged service; not linked into the initial HUD binary |
| `pi` | `https://github.com/earendil-works/pi.git` | `cb1205e4661723c7a8c9fb01acd38b7691f97797` | MIT | Multi-provider LLM API and agent-toolkit source reference | Reference only; no Apparat import, npm installation or runtime integration |

## Selection Notes

- The first routing vertical slice begins with an OpenAI-compatible text-generation adapter and can target a separately running llama.cpp server.
- Keeping llama.cpp behind a service adapter avoids coupling the HUD lifecycle to model loading, accelerator backends, native compilation, and large model files.
- Models, context limits, quantization, acceleration, and device eligibility are advertised as typed capabilities rather than inferred from the presence of this source.
- Adding this checkout does not select a model, download weights, validate GPU support, or grant unrestricted remote execution.
- Image generation, video generation, TTS, STT, and BOINC remain distinct workload classes with their own adapters and capability requirements.

## Pi Source Admission — 2026-11-20

Pi is admitted at the user's request for review of its multi-provider API,
agent loop and coding-agent CLI/SDK/RPC boundaries. It is an agent toolkit rather
than a local model inference engine. The source checkout supports reproducible
study of that monorepo without adding a JavaScript runtime dependency to Apparat.

The selected revision is the observed upstream main HEAD, pinned by the parent
gitlink and checked out detached. Its root `LICENSE` is MIT, copyright 2025 Mario
Zechner; preserve that notice and the license text in any future copied or
distributed portions. No nested `.gitmodules` is present at this revision.
No upstream installer, dependency lifecycle script or application was executed.
Apparat's Go module and build/runtime behavior are unchanged.

Initialize this reference with
`git submodule update --init --recursive -- third_party/inference/pi`. Update it
only through the parent inventory's focused update procedure: inspect upstream
changes and license notices, select an exact revision, stage its gitlink, update
this table and verify the checkout. Do not configure moving branch tracking.
Any active adapter, packaged runtime, credential access or tool execution requires
its own approved implementation scope and Apparat ownership boundaries.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Source-reference updates do not change root Go module pins or activate new integrations.
