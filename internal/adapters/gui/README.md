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

The current custom shell uses taller tab buttons with compact outer margins and a small tab-to-body gap so the HUD wastes less screen area on Steam Deck and Debian desktop windows. The top tab strip sizes every tab from the widest measured label plus balanced horizontal padding, and it supports horizontal mouse and touchscreen drag scrolling when the tabs exceed the viewport. Keep layout constants named in the adapter until they move into the user-editable HUD configuration manager.

## Tab Body Layout

Tab bodies are rendered from measured rectangles. Do not draw new content as unconstrained text or floating overlays.

Settings renders as a vertical stack of fieldsets. Each fieldset owns its title, explanation, and content rows. The temporary Android update button is a native bridge only for the platform install action; the HUD still reserves and draws the owning `Updates` fieldset.

Comrades, Projects, Cluster, Routing, and Tasks render as master-detail bodies. The left pane lists selectable objects, and the right pane owns placeholder/detail content until real selection data exists. Future adjustable dividers must preserve the same minimum-width and no-overlap rules.

Rows, list items, and buttons must remain large enough for touchscreens. Keep minimum touch target constants named and covered by tests when changing tab body layout.

## Native GUI Validation

GUI-specific validation must use:

```bash
go test -tags gui ./internal/adapters/gui
make build
make run-built
```

Linux GUI builds require native desktop development headers used by Ebitengine/GLFW, including X11, cursor, randr, xinerama, xi, OpenGL, xxf86vm, and ALSA development packages. If those are missing, record the exact missing package/header and do not mark GUI validation complete.

## Android Mobile Runner

Android GUI APKs use the `cmd/apparatmobile` binding package plus the tracked `android/apparat` wrapper. Direct `GoNativeActivity` paths can initialize Apparat runtime state but do not attach Ebitengine's `EbitenView` on the Pixel, so the wrapper activity owns Android view lifecycle and this adapter owns the shared HUD game model.
