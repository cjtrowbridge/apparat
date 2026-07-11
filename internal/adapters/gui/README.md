# GUI Adapter

This package contains the Ebitengine and EbitenUI-facing HUD adapter.

The GUI adapter renders HUD view models, maps controller/keyboard/mouse/touch input to named HUD actions, and reports presentation diagnostics.

It must not own durable state transitions, SQL, network calls, queue policy, or project authorization.

## Ebitengine Boundary

Ebitengine owns the window, game loop, input polling, and final drawing surface. `Run` starts the Ebitengine loop when the binary is built with the `gui` tag.

Default non-GUI builds use the stub runner so headless validation can run on systems without native desktop headers.

## EbitenUI Boundary

EbitenUI is the active HUD widget and layout layer for standard GUI surfaces: panels, buttons, lists, forms, tab bars or rails, focusable controls, and scroll containers.

The GUI adapter builds an EbitenUI root container from `internal/hud` snapshots. The root uses an `AnchorLayout` with HUD margins and diagnostics clearance, then a one-column shell with a horizontally scrollable top tab strip and a stretched active-tab body. Tab buttons are ordinary EbitenUI buttons inside a `ScrollContainer`; pointer, wheel, touch, keyboard, and controller tab changes stay synchronized with `hud.Shell`, and non-pointer tab changes move the tab strip enough to keep the active tab visible.

If additional custom Ebitengine rendering is required for dense visualizations or platform diagnostics, document the reason here and keep authoritative state in HUD view models. Do not reintroduce custom coordinate layout loops for ordinary tab bodies or controls.

## Tab Body Layout

Tab bodies are rendered as EbitenUI widget trees. Do not draw new content as unconstrained text or floating overlays. Settings uses a scrollable vertical fieldset stack. Non-Settings tabs use an EbitenUI master-detail pattern while their durable data remains placeholder/mock content.

Settings renders each `internal/hud` section as a fieldset. The `Updates` section owns an EbitenUI-rendered `Check for update` button. The button calls the GUI adapter's update callback and then listens for coarse status text (`Checking...`, `Already current`, `Permission needed`, `Installer opened`, or failure states) through the game model. Android implements the request and status-report path through the mobile wrapper, while desktop builds use local fallback feedback. Do not add a native Android overlay button for this path.

The Settings Diagnostics fieldset owns the Debug UI overlay toggle. The overlay is a development-only floating panel rendered by the Ebitengine game loop so it can report live screen size, FPS, UPS, runtime path, working directory, binary path, active tab, route, input, voice state, focus, and queue diagnostics. It is draggable by its title bar and must remain opt-in from Settings.

Comrades, Projects, Cluster, Routing, and Tasks render as responsive master-detail bodies. Expanded layouts show the left list, a visible draggable divider, and the right detail pane. Narrow layouts collapse to one pane: list first, selected section/thread switches to detail, and detail starts with an EbitenUI `<- Back` button. Research uses the pattern that matches its current placeholder/review state. Future scrolling, tab overflow behavior, narrow-screen collapse, and adjustable dividers must preserve the same minimum-width and no-overlap rules, and must be validated with screenshots before release.

Rows, list items, buttons, and input-like controls must remain large enough for touchscreens. Keep minimum touch target constants named and covered by tests when changing tab body layout. Text blocks must remain inside their owning EbitenUI container; they must not draw over adjacent fieldsets or panes.

## Native GUI Validation

GUI-specific validation must use:

```bash
go test -tags gui ./internal/adapters/gui
go test -c -tags gui -o /tmp/apparat-gui.test ./internal/adapters/gui
make build
make run-built
```

Linux GUI builds require native desktop development headers used by Ebitengine/GLFW, including X11, cursor, randr, xinerama, xi, OpenGL, xxf86vm, and ALSA development packages. Running GUI-tagged tests or the built GUI also requires a usable display server. If no display is available, use `go test -c -tags gui` as a compile gate, record the display limitation, and do not mark visual validation complete.

## Android Mobile Runner

Android GUI APKs use the `cmd/apparatmobile` binding package plus the tracked `android/apparat` wrapper. Direct `GoNativeActivity` paths can initialize Apparat runtime state but do not attach Ebitengine's `EbitenView` on the Pixel, so the wrapper activity owns Android view lifecycle and this adapter owns the shared HUD game model.
