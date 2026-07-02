# Speech Sources

| Path | Upstream | Revision | License | Apparat Role | Build Status |
| --- | --- | --- | --- | --- | --- |
| `whisper.cpp` | `https://github.com/ggml-org/whisper.cpp.git` | `6fc7c33b4c3a2cec83e4b65abd5e96a890480375` (`v1.9.1-81-g6fc7c33b`) | MIT | First portable local speech-to-text runtime reference for push-to-talk | Future service adapter; not linked into the initial HUD binary |

## Selection Notes

- Holding `R2` captures push-to-talk audio; releasing it submits the captured audio to a speech-to-text adapter.
- Audio capture, model execution, transcription, cancellation, and failure reporting must remain outside the Ebitengine update loop.
- The adapter contract permits whisper.cpp, an OS-native service, or a remote service without changing the controller interaction.
- Model files and captured audio are not stored in Git or embedded in SQLite.
- Qwen3-ASR remains excluded from the MVP source set.
- Qwen3-TTS remains research-before-adding; initial TTS should use a lightweight OS-native or service-backed adapter.
