# GUI Adapter

This package tree will contain the Ebitengine and EbitenUI HUD adapter.

The GUI adapter renders view models, maps controller/keyboard/mouse/touch input to application commands, and reports presentation diagnostics.

It must not own durable state transitions, SQL, network calls, queue policy, or project authorization.

The Phase 0 dependency anchor uses the `gui` build tag so baseline headless validation can run on systems without native desktop headers. GUI-specific validation must use `go test -tags gui ./internal/adapters/gui` after the required platform libraries are installed.
