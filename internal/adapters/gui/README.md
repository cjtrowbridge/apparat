# GUI Adapter

This package contains the Ebitengine and EbitenUI-facing HUD adapter.

The GUI adapter renders HUD view models, maps controller/keyboard/mouse/touch input to named HUD actions, and reports presentation diagnostics.

It must not own durable state transitions, SQL, network calls, queue policy, or project authorization.

## Ebitengine Boundary

Ebitengine owns the window, game loop, input polling, and final drawing surface. `Run` starts the Ebitengine loop when the binary is built with the `gui` tag.

Default non-GUI builds use the stub runner so headless validation can run on systems without native desktop headers.

The desktop window icon is generated in code from the same simple blue gear motif as the tracked root `logo.svg`. `RunWithRuntimeInfo` applies the generated icon images through Ebitengine's `SetWindowIcon` on desktop platforms where Ebitengine supports it.

## EbitenUI Boundary

EbitenUI is the active HUD widget and layout layer for standard GUI surfaces: panels, buttons, lists, forms, tab bars or rails, focusable controls, and scroll containers.

The GUI adapter builds an EbitenUI root container from `internal/hud` snapshots. The root uses an `AnchorLayout` with HUD margins and diagnostics clearance, then a one-column shell with a horizontally scrollable top tab strip and a stretched active-tab body. Tab buttons are ordinary EbitenUI buttons inside a `ScrollContainer`; pointer, wheel, touch, keyboard, and controller tab changes stay synchronized with `hud.Shell`, and non-pointer tab changes move the tab strip enough to keep the active tab visible.

Scrollable HUD regions that can contain viewport-wide or wider content must bound their reported preferred width before parent layout. This adapter uses local bounded preferred-size wrappers for the tab strip, Settings body, collapsed master/detail panes, and expanded master/detail panes because EbitenUI calculates parent layout from child preferred sizes before scroll clipping is applied. `StretchContentWidth()` can still be used for final placement, but it is not a substitute for preferred-width bounds.

HUD body text must wrap before preferred-size measurement. Use the local text helpers for summaries, section descriptions, titles, and row details so every `widget.Text` has a nonzero max width derived from the current viewport and owning pane. Do not add raw body text that can report a single-line width wider than the visible HUD area.

The tab strip only auto-scrolls the active tab into view after a rebuild, resize, or programmatic tab selection. It must not continuously force `ScrollLeft` during the update loop, because that prevents mouse, wheel, and touch swipes from persisting on phone-width surfaces.

EbitenUI `ScrollContainer` clips and renders from `ScrollTop`, but does not supply wheel or drag behavior itself. Register every rebuilt Settings, master-list, and detail scroll container with the GUI adapter. Wheel and vertical drags target only the innermost body viewport under the pointer/touch, use a threshold to suppress release-target activation, and never expand the viewport beneath the tab strip or diagnostics bar. A body drag snapshots and holds the tab strip's horizontal `ScrollLeft` through the deferred release-event window, so touch scrolling cannot select a release target, rebuild the body, or displace tabs horizontally.

Tab-strip drag and swipe gestures scroll only. They must not select a tab and must not leave the release target visually checked or pressed alongside the selected tab. EbitenUI defers button click and state-change events until `UI.Update()` drains its event queue, so drag cancellation must survive the release frame long enough to reject those deferred events. The tab buttons belong to an EbitenUI `RadioGroup` as a second invariant layer: exactly one button may be checked, and a rejected drag restores that selection from the authoritative `hud.Shell` snapshot.

If additional custom Ebitengine rendering is required for dense visualizations or platform diagnostics, document the reason here and keep authoritative state in HUD view models. Do not reintroduce custom coordinate layout loops for ordinary tab bodies or controls.

## Tab Body Layout

Tab bodies are rendered as EbitenUI widget trees. Do not draw new content as unconstrained text or floating overlays. Settings uses a scrollable vertical fieldset stack. Non-Settings tabs use an EbitenUI master-detail pattern while their durable data remains placeholder/mock content.

Settings renders each `internal/hud` section as a fieldset. The `Updates` section owns an EbitenUI-rendered `Check for update` button. The button calls the GUI adapter's update callback and then listens for coarse status text (`Checking...`, `Already current`, `Permission needed`, `Installer opened`, or failure states) through the game model. Android implements the request and status-report path through the mobile wrapper, while desktop builds use local fallback feedback. Do not add a native Android overlay button for this path.

The Settings Diagnostics fieldset owns the Debug UI overlay toggle. The overlay is a development-only floating panel rendered by the Ebitengine game loop so it can report live screen size, FPS, UPS, runtime path, working directory, binary path, active tab, route, input, voice state, focus, and queue diagnostics. It is draggable by its title bar and must remain opt-in from Settings.

The floating circular `PTT` button is a global Ebitengine overlay near the bottom-right of the app, above diagnostics. It holds the same HUD voice-capture state as right Ctrl and controller R2 while pressed and releases capture when the pointer/touch is released or Escape cancels voice capture.

Comrades, Projects, Cluster, Routing, and Tasks render as responsive master-detail bodies. Expanded layouts show the left list, a visible draggable divider, and the right detail pane. Narrow layouts collapse to one pane: list first, selected section/thread switches to detail, and detail starts with an EbitenUI `<- Back` button. Research uses the pattern that matches its current placeholder/review state. Future scrolling, tab overflow behavior, narrow-screen collapse, and adjustable dividers must preserve the same minimum-width and no-overlap rules, and must be validated with screenshots before release.

Rows, list items, buttons, and input-like controls must remain large enough for touchscreens. Tab buttons, Back buttons, list rows, and divider handles use the same 54 px touch scale unless a future plan changes the shared constant. Master-list rows are left-aligned with leading room reserved for later avatars or status glyphs. Keep minimum touch target constants named and covered by tests when changing tab body layout. Text blocks must remain inside their owning EbitenUI container; they must not draw over adjacent fieldsets or panes.

Phone-width layout fixes require rendered evidence before completion. Capture narrow screenshots that show bounded tab contents, swipeable tab overflow, and collapsed master-detail navigation behavior when those surfaces are touched.

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
