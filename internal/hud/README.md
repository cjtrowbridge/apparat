# HUD Model

This package owns the mock-data HUD model used by the Phase 2 prototype.

It is deliberately independent of Ebitengine so focus, tab order, command semantics, mock views, and voice-state transitions are testable without a display server. The Ebitengine adapter renders this model and maps platform input into the same actions.
