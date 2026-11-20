# Input And Focus Contract

All GUI targets dispatch the same application actions for tabs, focus, activation, cancellation, context, scrolling, text entry, and push-to-talk.

This is the target input contract. The live HUD implements directional navigation on all six tabs. Phase 17 tracks broader device and accessibility qualification.

The canonical top-level tab order is Comrades, Projects, Cluster, Research, STMPCK, Settings. Routing and Tasks are details within Cluster, not top-level tabs.

## Steam Deck

- `L1`: previous top-level tab.
- `R1`: next top-level tab.
- D-pad or left stick: move focus.
- `A`: activate focused control.
- `B`: back or cancel.
- `X`: contextual actions.
- Menu: command palette.
- Right stick: scroll long views.
- `L2`: hold to talk; release submits.
- `Steam+X`: system on-screen keyboard for text fields.

Holding L2 highlights the on-screen PTT button exactly like pressing it with a
pointer; release restores idle. Right Ctrl uses the same feedback. With no
editing target, the press is visible but capture stays silent. Controller loss
cancels a held capture rather than submitting it.

## Debian GUI

- `Ctrl+PageUp` / `Ctrl+PageDown`: previous/next top-level tab.
- `Alt+1` through `Alt+6`: Comrades, Projects, Cluster, Research, STMPCK, Settings. Routing and Tasks are selected within Cluster.
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


## Directional navigation

Arrow keys and the D-pad move the gold focus outline. Along the tab strip they
select the adjacent tab immediately, wrapping at the ends. Move toward the page
to enter its last focused control. On a top strip that means Down, on a bottom
strip Up, on a left rail Right, and on a right rail Left. At the page edge move
back toward the strip to return. Escape or B leaves the current edit, detail or
workspace scope first.

Within a page, move toward an adjacent button, field or checkbox. The outline
scrolls into view. Enter or A activates it, and Space activates buttons and
toggles checkboxes. Tab and Shift+Tab step through HUD controls in layout order.
Activate the pane divider to resize it with Left and Right. Enter, A, Escape or B finishes resizing. Activate a text field to edit. Its caret then owns the arrows. Escape or B
leaves the field with its draft intact.

In STMPCK the centered toolbar, intensity controls and three-column game grid
share directional focus. Enter or A opens the focused game or activates the
focused toolbar control. Embedded settings keep navigation inside the panel.
Gameplay receives the game's own inputs. Back returns to the gallery.
