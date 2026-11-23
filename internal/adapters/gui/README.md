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

When all tab buttons fit in the viewport, the strip is always left-aligned. When they overflow, rebuilds preserve the user's `ScrollLeft`; pointer/touch selection does not reposition it. Keyboard, controller, resize, and other non-pointer selection paths may move it only enough to make a clipped selected tab fully visible—never by normalized-index centering. Tab hover uses a distinct background from selected/pressed state, so a pointer over an inactive tab cannot look like a second selected tab.

`tab_strip_interaction_test.go` covers drag arbitration and radio-group state; `tab_strip_scroll_test.go` covers fit/overflow position preservation, minimal reveal, and hover-versus-selected visuals. Keep these contracts separate so either interaction path can evolve without exceeding the code-file size limit.

EbitenUI `ScrollContainer` clips and renders from `ScrollTop`, but does not supply wheel or drag behavior itself. Register every rebuilt Settings, master-list, and detail scroll container with the GUI adapter. Wheel and vertical drags target only the innermost body viewport under the pointer/touch, use a threshold to suppress release-target activation, and never expand the viewport beneath the tab strip or diagnostics bar. A body drag snapshots and holds the tab strip's horizontal `ScrollLeft` through the deferred release-event window, so touch scrolling cannot select a release target, rebuild the body, or displace tabs horizontally.

Tab-strip drag and swipe gestures scroll only. They must not select a tab and must not leave the release target visually checked or pressed alongside the selected tab. EbitenUI defers button click and state-change events until `UI.Update()` drains its event queue, so drag cancellation must survive the release frame long enough to reject those deferred events. The tab buttons belong to an EbitenUI `RadioGroup` as a second invariant layer: exactly one button may be checked, and a rejected drag restores that selection from the authoritative `hud.Shell` snapshot.

`selector_panel_test.go` proves Cluster selector headings remain non-selectable text and that its Tasks selector owns the Tasks content panel. Keep it focused on selector/content-panel semantics; ordinary shell coverage remains in `ebiten_shell_test.go`.

`table.go` owns the reusable `borderedTable` routine for HUD data tables. Give it one header `tableRow` and same-width data rows; it applies one shared grid, a table perimeter, and a border around every header/data cell. Use `tableText` for padded plain-text cells. Do not recreate tables as independent header and row grids: each grid calculates columns independently and will misalign. Interactive header widgets remain ordinary EbitenUI controls supplied to the table routine.

If additional custom Ebitengine rendering is required for dense visualizations or platform diagnostics, document the reason here and keep authoritative state in HUD view models. Do not reintroduce custom coordinate layout loops for ordinary tab bodies or controls.

## Tab Body Layout

Tab bodies are rendered as EbitenUI widget trees. Do not draw new content as unconstrained text or floating overlays. Settings uses a scrollable vertical fieldset stack. Non-Settings tabs use an EbitenUI master-detail pattern while their durable data remains placeholder/mock content.

Settings renders each `internal/hud` section as a fieldset. The `Updates` section owns an EbitenUI-rendered `Check for update` button plus a wrapped status explanation. `update_status.go` maps stable update reason codes to safe short labels and actionable details; render only those mappings, never raw exceptions, response bodies, tokens, private paths, or unverified release data. Android reports codes through the mobile bridge. Linux uses the public qualified development channel through `selfupdate`: one activation checks, downloads, verifies, installs and restarts after graceful node shutdown. Production signed manifests and remote node updates remain Phase 17 work. Do not add a native Android overlay button for this path.

The Settings Diagnostics fieldset owns the Debug UI overlay toggle. The overlay is a development-only floating panel rendered by the Ebitengine game loop so it can report live screen size, FPS, UPS, runtime path, working directory, binary path, active tab, route, input, voice state, focus, and queue diagnostics. It is draggable by its title bar and must remain opt-in from Settings.

The floating circular **PTT** button is a global Ebitengine overlay near the bottom-right of the app, above diagnostics. A green fill indicates capture and a second ring indicates locked listening. Inside a clarification dialog, the button sits in the lower-right corner of the panel. It holds the same HUD voice-capture state as right Ctrl and controller L2 while pressed and releases capture when the pointer/touch is released or Escape cancels voice capture.

Comrades, Projects, and Cluster render responsive selector/content-panel bodies. Expanded layouts show a selector panel, a visible draggable divider, and only the selected content panel. The selector panel may include non-selectable, non-focusable headings that group selectable items; a heading may have a small, muted, wrapped description beneath its title. Narrow layouts collapse to one panel: the selector panel appears first, a selected section/thread switches to the content panel, and the content panel starts with an EbitenUI `<- Back` button. Cluster's Routing and Tasks selectors render grouped nested detail sections in that content panel. Research uses the pattern that matches its current placeholder/review state. Its mock contribution fieldset renders `Your Contribution` separately from a two-column friend table; a shared grid aligns all cells, with an outer border and borders around each cell. Both 44 px headers are clickable and select/toggle the numeric view-model sort without parsing displayed units. Future scrolling, tab overflow behavior, narrow-screen collapse, and adjustable dividers must preserve the same minimum-width and no-overlap rules, and must be validated with screenshots before release.

Selector groups use the ordered HUD palette `#0032AB`, `#6028A7`, `#8C159F`, `#AF0093`, `#CB0084`, `#E10072`, `#F10060`, cycling only after the final entry. A group's heading text and every selector button share its palette color: headings use it as text and buttons use it as an outline, while hover and selected states stay visually distinct. Mock chat panels include an editable no-op reply field and Send button. Project and Pipeline panels include presentation-only Git/Chat controls; Pipelines also expose a no-op Run control.

Rows, list items, buttons, and input-like controls must remain large enough for touchscreens. Tab buttons, Back buttons, selectable selector items, and divider handles use the same 54 px touch scale unless a future plan changes the shared constant. Selector headings are text, not controls, and do not need a touch target. Selector items are left-aligned. Projects uses a 12 px inset for names, paths, and the branch and revision line. Keep minimum touch target constants named and covered by tests when changing tab body layout. Text blocks must remain inside their owning EbitenUI container; they must not draw over adjacent fieldsets or panels.

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

## Source inventory

- `app_icon.go`: Embeds the application icon for the GUI shell.
- `debug_overlay.go`: Renders developer diagnostics over the HUD.
- `dependencies.go`: Records the package dependency boundary.
- `ebiten_shell.go`: Updates and draws the Ebitengine HUD shell.
- `ebiten_shell_test.go`: Verifies updates and draws the Ebitengine HUD shell.
- `layout_bounds.go`: Computes bounded tab and content geometry.
- `selector_list.go`: Keeps repository entry actions fixed above the scrollable checkout list in both selector layouts.
- `master_detail.go`: Builds selector and content-panel widgets.
- `research_contributions.go`: Builds and sorts local mock research contribution rows.
- `research_contributions_test.go`: Verifies builds and sorts local mock research contribution rows.
- `run_ebiten.go`: Starts the native Ebitengine GUI loop.
- `run_stub.go`: Reports the GUI-disabled build behavior.
- `selector_panel_test.go`: Verifies checks selector-panel geometry and selection behavior.
- `shell.go`: Coordinates tab selection and HUD shell state.
- `tab_strip_interaction_test.go`: Verifies checks tab-strip pointer and drag arbitration.
- `tab_strip_scroll_test.go`: Verifies checks bounded tab-strip scrolling.
- `table.go`: Builds retained HUD table widgets.
- `theme.go`: Loads fonts and defines GUI presentation defaults.
- `ui_builder.go`: Builds retained tab-body widget layouts.
- `update_status.go`: Maps updater outcomes to safe explanatory HUD status.
- `update_status_test.go`: Verifies maps updater outcomes to safe explanatory HUD status.
- `vertical_scroll.go`: Manages bounded vertical widget scrolling.
- `vertical_scroll_test.go`: Verifies manages bounded vertical widget scrolling.

- `settings_test.go`: Settings diagnostics and update callback/status regressions split from shell layout tests to retain the 400-line bound.

- `live_services.go`: Projects live local services into Cluster Routing without HTTP or SQL.

- `services.go`: Loads immutable service projections outside the render/update goroutine.

- `services_test.go`: Verifies direct query projection and cancellation without a display or HTTP.

- `gesture.go`: Arbitrates one pointer owner across touch and mouse controls.

- `gesture_test.go`: Tests exclusive gesture capture and cancellation across pointer sources.

Phase 8 adds a live Local services detail under Cluster. One context-owned worker calls the core directly, serializes queries, coalesces snapshots and preserves a last-known projection on failure. There is no HTTP/SQL in Update/Draw. Timestamp-only observations do not rebuild widgets. A shared pointer owner arbitrates touch/mouse capture; geometry/focus cancellation cannot submit PTT. Native accessibility trees/screen-reader semantics remain unavailable in the canvas toolkit; names, visible state and keyboard shortcuts are retained, and Phase 17 owns the native accessibility acceptance gate. Real microphone/cancel semantics remain Phase 16.

Android safe-area color comes from the GUI renderer through `BackgroundARGB` in the mobile bridge. The current fixed dark palette is authoritative; runtime theme switching is not implemented. The native parent and navigation background use that color, with light navigation icons and no automatic contrast scrim. Insets and touch bounds are unchanged. Verify matching background pixels and readable buttons in portrait/landscape on an authorized device; system/OEM styling on other devices requires separate evidence.

`tab_icons.go` owns the four embedded Twemoji images in `emoji/`, decoded/scaled once on the GUI thread. Tab buttons are up to 54x54 (width shrinks to 44 before overflow scrolling) with 28px symbols: ☭ 🔨 🖧 🧪 ⚙️. All five fit at 320 logical pixels; smaller windows retain overflow scrolling. Tab buttons have no tooltips; the active page heading supplies the visible name. Descriptor accessibility labels and navigation actions remain. Test phone fit and narrow overflow with the tab-strip GUI tests and rendered Pixel 4a screenshots. This does not add a native accessibility tree.

- `tab_icons.go`: Loads fixed tab artwork and supplies icon-only button contents.

Icon tab buttons set their complete style explicitly and isolate it from the global text-button theme. Otherwise EbitenUI adds an empty text child and spacing beside the graphic, shifting icons off center. Keep screenshot checks alongside width/input tests.

The shell grid contains icon navigation, a fixed 28px page heading derived from `ActiveTab().Title()`, and the remaining scrollable body. The title uses one face per Game and stays visible in collapsed detail and expanded master/detail views. Its measured height reduces the body viewport through grid layout; it does not overlay content. GUI tests and Pixel 4a portrait/landscape screenshots verify titles and scroll bounds.

Updater status text now distinguishes an invalid package, incompatible signing key, and no newer published version. Status tests require safe explanatory text; unknown codes remain redacted.

- `arcade.go`: Lazily initializes the embedded STMPCK gallery, translates the viewport, routes Back and saves when the game is suspended.

The Updates card uses platform-specific HUD copy. Its button and status exist when the Linux runner or Android bridge supplies an updater callback.

The desktop master-detail divider captures the mouse through the shared pointer owner. Resizing preserves both panes' scroll positions while the layout rebuilds.

Selector buttons wrap their labels to the pane's inner width as the divider moves. Projects rows reserve space for the configured-action icon, and wrapped buttons grow vertically within the scrollable list. Project paths use the smaller muted description face and wrap within the same selectable button.

- `preferences.go`: Loads preferences before layout, saves asynchronously, and applies confirmed changes.
- `window_state.go`: Restores desktop placement and coalesces geometry writes until shutdown.

Tab rails share the same buttons and selection handlers as the top strip. The side gutter narrows the available content area, including STMPCK gallery rendering and hit testing. An active game uses the full viewport.

- `projects.go`: shows SQLite registrations, empty and unavailable states, GitHub clone preview and account selection. The Linux folder button uses Zenity with portal preference and supports typed paths. Git and storage work run off the frame loop. The selected account applies to a single transfer.

workspace.go renders Files, Git, Tasks and Chat for the selected registration. Open a text file to edit individual lines, keep a draft, and explicitly apply it. Pending writes and commits require confirmation.

- `workspace.go`: Presents the local workspace through the Projects detail pane.

- `workspace_controller.go`: Routes arrow keys and the D-pad to the same directional actions across tabs. Owns key repeat, activation, back and text-edit input precedence.

- `workspace_layout.go`: Keeps the active file's text input attached across layout rebuilds so its focus, caret and selection survive a window resize. The detail viewport retains its proportional scroll position within the same workspace view.
- `workspace_layout_test.go`: Exercises the split and single-pane layouts with a selected draft and a conflict report, and checks that a different file starts with fresh presentation state.

- `document.go`: Routes scoped edit sessions through the Android or Linux multiline editor, including draft persistence and explicit save.
- `document_gtk.py`: Embedded GTK 4 text editor, using Python 3 with PyGObject on Linux. Exchanges document text over process pipes.
- `workspace_screens.go`: Renders Files, Changes, Tasks and Chat and maps Back to each screen’s parent.

- `workspace_directory.go`: Shows one folder at a time, combines its files with Git status, and opens missing tracked files in deletion review.

Workspace Files publishes folder entries first. A cancellable background check decorates the current directory. Navigating discards obsolete results, and a failed check leaves filenames accessible with an error message.

- `dictation.go`: Owns workspace PTT sessions, targeted text insertion and device-local model selection in Settings.
- `project_creation.go`: Defines the pinned New Project entry, folder and account controls, clone destination preview and registration editing.
- `update_discovery.go`: Checks the Android release manifest on Settings entry, coalesces checks and separates discovery from installation.

- `desktop_updates.go`: Owns one cancellable Linux update worker, candidate cleanup, UI status and restart intent. No network I/O runs on the frame thread.
- `desktop_updates_test.go`: Tests check/install/restart and up-to-date states with a disposable executable, untargeted PTT and native-aspect layout.
- `gamescope.go`: Uses the active monitor geometry and fullscreen under Gamescope while ordinary desktops retain saved placement.
- `ptt_overlay.go`: Draws and captures the existing PTT overlay, split from the shell to preserve its 400-line limit. Pressing PTT without an editable target is silent.
- `ptt_overlay_test.go`: Verifies matching pointer, L2 and keyboard pressed appearance, trigger release, the binding contract and cancellation on controller loss.

The PTT button uses one pressed fill while pointer, L2 or right Ctrl is held.
A held trigger also keeps the button visible inside a workspace even without a
focused text input; this indicates the press and does not imply a microphone
session without an editing target. Releasing L2 returns to idle unless another
source remains held. Losing the standard controller cancels its capture and
clears its visual state. R2 is no longer a PTT binding.

Linux updates use the existing Settings card. The **Check for update** button shows **Up to date** after a matching build is confirmed. It can be selected again to recheck. Progress and errors wrap beneath it. One click downloads and verifies a newer artifact, atomically replaces the executable, keeps `.previous`, and closes GUI/core state before restarting. Steam continues tracking the same process and path. Failures keep installed bytes and show safe retry text. Closing cancels and joins the worker and deletes staged candidates.

The embedded gallery starts with one 48-unit toolbar containing STMPCK, the star, friends, discovery and settings. Side and bottom tab placements release the full top edge to the gallery. Top tabs reserve only the navigation strip and gap. The renderer, pointer and directional focus use the same toolbar rectangles, and tile layout and scroll bounds use the compact header height. The host tab order is unchanged and the pinned standalone game source remains untouched.

- `directional_focus.go`: resolves live widget geometry, viewport reveal and per-tab focus restoration.
- `arcade_focus.go`: maps the same cursor onto embedded gallery controls and settings.
- `directional_focus_test.go`: exercises the rendered tabs, scroll reveal, edit ownership and gallery handoff.

- `action_screen.go`: Fullscreen search, clarification, local spoken prompts and input focus.

- `action_catalog.go`: Enumerates current navigation targets and revalidates the stable address before activation.

- `game_setup_test.go`: Exercises all game entries, fresh setup and saved-session recovery through voice and pointer controls.
- `action_catalog_test.go`: Tests tab coverage, project and thread identity, stale choices, dictation ownership and responsive action browsing.

- `ptt_lock.go`: Shared double-tap timing, capture ownership, lock and single submission for canvas and native dictation.

- `action_dialog.go`: Centers the scrollable voice chooser, reserves its PTT space, and dismisses an outside click or tap on the dimmed page.

- `lan.go`: connects Cluster's Add New Device detail to interface-scoped discovery. Linux users select an adapter and Android uses the active Wi-Fi network. Candidates expire with the browse session.

- `cluster_pages.go`: Renders the gear and cluster detail navigation.
- `cluster_pages_test.go`: Verifies settings navigation and Back on wide and narrow layouts.
- `device_directory.go`: projects synchronized GUI peers above workers, resolves device pool menus against the worker that owns the policy and lists assigned devices on routing pages.

- `pool_tabs.go`: renders Tasks, Devices and Authorized users below the pool heading. It keeps filters and sort direction on the navigation frame. Task and device details use the existing detail stack and authorization paths. Discovery compares advertised hostnames with the saved worker and personal-device directory before offering navigation, pairing or claiming.

- `new_project.go`: Builds the host-first New Project form and runs local checkout setup with progress, registration refresh and workspace selection. Remote hosts and the unpublished template have preview-only controls.

- `form_labels.go`: Project form headings and field labels use the bold face. Descriptions use the italic face. Inline inputs and native prompts share a wrapping left label column that scales with the detail pane.
