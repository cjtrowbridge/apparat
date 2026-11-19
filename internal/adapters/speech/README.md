# Local dictation

- `catalog.json`: Immutable weight revisions, lengths, SHA-256 digests and language variants. `languages.txt` lists the 99 language tokens for these multilingual checkpoints.
- `ggml-tiny.en-q5_1.bin`: Bundled quantized OpenAI Whisper Tiny English weights distributed by ggerganov/whisper.cpp. See LICENSE-weights and LICENSE-whisper.

- `native.cpp`: Owns bounded device capture, silence suppression and CPU transcription.
- `native.go`: Go bindings and per-target static linkage.
- `native_stub.go`: Reports absence of native support in development builds.
- `models.go`: Verifies and installs bundled or downloaded models using atomic sibling files.
- `session.go`: Serializes model activation and recording, stops capture on release and finalizes bounded audio on a worker.
- `models_test.go`: Exercises cancelled and corrupt transfers and file cleanup.
- `native_test.go`: Loads the bundled weights and checks silence and the upstream speech fixture.
