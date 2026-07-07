# Input And Focus Contract

All GUI targets dispatch the same application actions for tabs, focus, activation, cancellation, context, scrolling, text entry, and push-to-talk.

## Steam Deck

- `L1`: previous top-level tab.
- `R1`: next top-level tab.
- D-pad or left stick: move focus.
- `A`: activate focused control.
- `B`: back or cancel.
- `X`: contextual actions.
- Menu: command palette.
- Right stick: scroll long views.
- `R2`: hold to talk; release submits.
- `Steam+X`: system on-screen keyboard for text fields.

## Debian GUI

- `Ctrl+PageUp` / `Ctrl+PageDown`: previous/next top-level tab.
- `Alt+1` through `Alt+7`: Comrades, Projects, Research, Cluster, Routing, Tasks, Settings.
- `Tab` / `Shift+Tab`: next/previous focusable control.
- Arrows: list, tree, grid, menu, and spatial focus movement.
- `Enter`: activate or submit.
- `Space`: button activation or toggle.
- `Escape`: close modal, leave scope, go back, cancel, or cancel held right-`Ctrl` recording.
- Menu or `Shift+F10`: context actions.
- `Ctrl+Shift+P`: command palette.
- `PageUp`, `PageDown`, `Home`, `End`: collection navigation.
- Right `Ctrl`: hold to talk; release submits unless `Escape` cancelled it.

Text fields preserve ordinary Debian editing and clipboard behavior. Global bindings use explicit modifiers so typing does not navigate the app.

## Mouse And Touch

- Left click/tap: focus and activate.
- Right click/long-press: contextual actions.
- Wheel/touchpad/drag scroll: scroll focused or pointed container.
- Mouse back: back or cancel.
- Drag: only for approved pane resize, sliders, text selection, and explicit reorder operations.

Every essential pointer operation must have keyboard and controller alternatives.

## Headless Debian

Headless mode has no HUD focus map and must not initialize Ebitengine. Control surfaces are CLI, authenticated HTTPS REST, service manager, health checks, `SIGINT`, and `SIGTERM`.

## Focus Rules

- Every focusable control has visible focus.
- Disabled controls are skipped or explain why unavailable.
- Modal surfaces trap focus and restore it on close.
- Scrolling has deterministic entry, movement, boundary, and exit behavior.
- Bindings are configurable; conflicts and platform-reserved shortcuts are surfaced in Settings.
- Push-to-talk renders recording, cancellation, queued, transcribing, failure, and completion states.
