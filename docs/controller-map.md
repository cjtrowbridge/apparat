# Input And Focus Contract

All GUI targets dispatch the same application actions for tabs, focus, activation, cancellation, context, scrolling, text entry, and push-to-talk.

This is the target input contract. The live HUD implements directional navigation on all seven tabs. Phase 17 tracks broader device and accessibility qualification.

The canonical top-level tab order is Comrades, Projects, Cluster, Research, STMPCK, Apparat Global Planning Agent, Settings. Routing and Tasks are details within Cluster, not top-level tabs.

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
- `Alt+1` through `Alt+7`: Comrades, Projects, Cluster, Research, STMPCK, Apparat Global Planning Agent, Settings. Routing and Tasks are selected within Cluster.
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

Opening a discovered worker exposes **Claim this host for your cluster**. Keyboard and D-pad activation authenticate the saved identity through the same claim action as a tap. Settings → Personal identity → Edit screen name opens local text input. The action catalog opens each dialog through the same handler.

The Cluster heading gear opens Cluster settings. Escape, B, Android Back and the detail-page Back button return one level. Selecting a task opens its details. Cancellation and role changes have a separate confirmation.

Cluster pool detail uses three focusable emoji tabs for Tasks, Devices and Authorized users. Directional navigation reaches task status and order, device rows and the administrator's add-user control. Opening a worker from Devices pushes its normal detail page. Back restores Devices for that pool.

In Projects, focus the pinned ➕ New Project and activate it to focus device search. A nonempty search with one matching host selects it and puts focus on repository source. Choosing a source focuses the first template or import option. Choosing an option with additional fields focuses the first field. Tab and Shift+Tab follow the visible form order, and the D-pad moves to adjacent controls. Android focuses the search launcher and opens the system keyboard on activation. A confirmed worker also exposes Clone from GitHub. Local setup opens the workspace and worker setup opens its read-only checkout summary.

In New Project, Use no credentials, Use GitHub as @username, Refresh credential providers and available install or add-account actions follow the same focus order on keyboard and D-pad. Worker identity review is a native modal that owns input. Dismissal keeps the worker disconnected.

New Project exposes the selected host’s provider choices, install, enroll, and remote clone with the normal button focus route. A pending remote operation owns its form; device authorization offers Copy code and open GitHub and Cancel. A completed worker clone opens its host-qualified workspace. Older projects and Newest projects use the normal selector focus route.


### Owner workspaces

Select a host-qualified project, then navigate the five emoji buttons in order: Chat, Files, Editor, Terminal and Tasks. Files opens one directory at a time. Back returns to its parent. Save returns to Files after owner acknowledgement. Terminal input belongs to the text field until Send is activated. Take control transfers input between the same user's devices. Task history and retained terminal output each have explicit continuation controls.

The Projects Editor opens in view mode. Focus Edit to enable text input in the same pane. Within the document, arrow keys move the caret and Shift extends the selection. Tab leaves the document for its controls. Save and Cancel are available to writers.

On the HUD, drag the PTT circle up to lock recording. Drag the locked circle down to finish. Drag left from the resting circle to open Apparat Global Planning Agent and lock listening there. Text appears directly in Prompt. Slide the locked circle right to finish the utterance and stage it in the planning session. Tap Prompt on Android to type with the on-screen keyboard. Keyboard and controller users can open the brain tab, focus Prompt and hold right Ctrl or L2. Back opens history. The adjacent plus starts a session, and the pencil edits its title. Outside the editor, Page Up/Down and the vertical right stick scroll the transcript.

In a planning or project conversation, activate the model name above the editor to open the model picker. Navigate its text buttons with Tab, Shift+Tab or the D-pad. Choose a model, then Save for this chat or the administrator-only scope default action. Cancel, Back or Escape restores the chat. Selection and the saved scope default have separate labels.

Cluster Status occupies the former Local Services selector. Select it with the normal Cluster navigation to inspect agreement, queue totals and provider slots. Administrators open Inference validators, inspect the listed workers and confirm enrollment. Back returns to the status page. Pool task filters, keyboard and D-pad activation open the same inference input and output details as touch. Global and project conversation refreshes preserve the compose draft.

Project Chat opens the message composer directly. Back from the conversation opens its session list. New conversation opens another empty composer. A saved repository session exposes the pencil for title edits.

With the global or project message editor focused, Enter or keypad Enter invokes Send. Shift+Enter inserts a newline. A pending turn keeps the next draft editable.
