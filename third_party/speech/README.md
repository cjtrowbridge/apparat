# Speech Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `whisper.cpp` | `https://github.com/ggml-org/whisper.cpp.git` | `aba02defd6bb6ff5de63721c21e3456645c153e2` (`v1.9.5-335-gaba02def`) | MIT | First portable local speech-to-text runtime reference for push-to-talk | Future service adapter; not linked into the initial HUD binary |

## Selection Notes

- Holding `R2` captures push-to-talk audio; releasing it submits the captured audio to a speech-to-text adapter.
- Audio capture, model execution, transcription, cancellation, and failure reporting must remain outside the Ebitengine update loop.
- The adapter contract permits whisper.cpp, an OS-native service, or a remote service without changing the controller interaction.
- Model files and captured audio are not stored in Git or embedded in SQLite.
- Qwen3-ASR remains excluded from the MVP source set.
- Qwen3-TTS remains research-before-adding; initial TTS should use a lightweight OS-native or service-backed adapter.

## 2026-11-16 Source Refresh

Updated to configured upstream heads at the user’s request. Exact revisions above remain pinned by gitlinks; branch tracking was not added. See the [update review](../../plans/past/2026-11-16-11-00-00_update-and-review-submodules.md) for deltas, verification limits, and rollback. Source-reference updates do not change root Go module pins or activate new integrations.
