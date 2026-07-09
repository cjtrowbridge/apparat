# GUI Adapter

This package contains the Ebitengine and EbitenUI-facing HUD adapter.

The GUI adapter renders HUD view models, maps controller/keyboard/mouse/touch input to named HUD actions, and reports presentation diagnostics.

It must not own durable state transitions, SQL, network calls, queue policy, or project authorization.

## Ebitengine Boundary

Ebitengine owns the window, game loop, input polling, and final drawing surface. `Run` starts the Ebitengine loop when the binary is built with the `gui` tag.

Default non-GUI builds use the stub runner so headless validation can run on systems without native desktop headers.

## EbitenUI Boundary

EbitenUI is an active dependency and should be used for standard GUI widgets where it fits controller-first focus and layout requirements: panels, buttons, lists, forms, tab bars or rails, and focusable controls.

Phase 4 starts with an Ebitengine-rendered shell driven by `internal/hud` snapshots, plus an EbitenUI root container for standard widget integration. The current tab bar is custom-rendered so it can share the controller-first tab descriptor model and support pointer-click tab selection immediately. As widgets become interactive, prefer EbitenUI when it preserves the HUD action/focus model. If additional custom Ebitengine widgets are required for Steam Deck/controller behavior, document the reason in this README and keep their state in HUD view models.

The current custom shell uses taller tab buttons with compact outer margins and a small tab-to-body gap so the HUD wastes less screen area on Steam Deck and Debian desktop windows. Keep layout constants named in the adapter until they move into the user-editable HUD configuration manager.

## Native GUI Validation

GUI-specific validation must use:

```bash
go test -tags gui ./internal/adapters/gui
make build
make run-built
```

Linux GUI builds require native desktop development headers used by Ebitengine/GLFW, including X11, cursor, randr, xinerama, xi, OpenGL, xxf86vm, and ALSA development packages. If those are missing, record the exact missing package/header and do not mark GUI validation complete.

## Android Mobile Runner

Android GUI builds use the same `ebiten.RunGame(NewGame())` adapter as Debian. Ebitengine selects its mobile UI backend under `GOOS=android`, while `GoNativeActivity` owns the Android activity lifecycle. A previous `mobile.SetGame` runner initialized Apparat runtime state but left Pixel devices on the Android splash/default icon, so the shared run loop is the current Phase 5 rendering fix.
