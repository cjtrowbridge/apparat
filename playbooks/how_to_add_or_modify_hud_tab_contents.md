# Playbook: Add Or Modify HUD Tab Contents

*Status: Draft*

## Objective

Define how to add or modify Apparat HUD tab bodies so each tab remains responsive, non-overlapping, controller-first, and consistent with the product information architecture.

## Prerequisites

* Read `README.md`, `ROADMAP.md`, `AGENTS.md`, and `agentic-pipelines/AGENTS.md`.
* Identify the active execution plan item that authorizes the tab-content change.
* Read `internal/hud/README.md` and `internal/adapters/gui/README.md`.
* Inspect the current tab model in `internal/hud` and the rendering/input adapter in `internal/adapters/gui`.

Settings must always be the final top-level tab, including optional editions. Add or reorganize every other tab before Settings. STMPCK precedes Global Agentic Session, which sits immediately before Settings.

## Layout Rules

1. **Use Structured Body Elements via EbitenUI**
   * EbitenUI is the mandatory UI toolkit for the Apparat HUD. All layout and widgets must use `github.com/ebitenui/ebitenui`.
   * Do not use raw imperative coordinates (`ebitenutil.DebugPrintAt`, `ebitenutil.DrawRect`), custom `SubImage` clipping, or custom drag-scroll loops for ordinary tab body content.
   * Shell-level development overlays, tab-strip overflow input, and divider dragging may use small Ebitengine input/rendering shims only when an active plan explicitly authorizes them and the user-facing controls remain EbitenUI widgets.
   * Put related content in EbitenUI layout containers (like `widget.RowLayout` or `widget.GridLayout`) and `widget.ScrollContainer`.
   * Each bounded element must own its title, explanation, controls, and body content.
   * Tab content must remain strictly data-driven (dynamically iterating through `tab.Sections` and `section.Rows`); do not hardcode EbitenUI containers or widgets outside of this loop pattern.
   * For a data table, use the reusable GUI `borderedTable` routine in `internal/adapters/gui/table.go`. Supply one header `tableRow` and same-width data rows, using `tableText` for padded text cells. The routine owns the shared column calculation, outer border, and per-cell borders. Do not build headers and rows as separate grids because they calculate different column widths and will not align.

2. **Prevent Overlap By Construction**
   * Do not place controls as floating overlays unless they correspond to a reserved body element.
   * Stack vertical elements using measured heights and fixed gaps.
   * Split horizontal elements using explicit columns and a visible divider.
   * Let EbitenUI handle clipping, wrapping, and scrolling for overflowing content inside its owning element container.

3. **Be Responsive By Default**
   * Every tab body must work down to the current minimum supported body width for the target surface.
   * If a tab uses columns, define a minimum width for each column and collapse or scroll when those minimums cannot fit.
   * Use bounded body dimensions, padding, gaps, and scroll areas instead of viewport-scaled fonts.
   * HUD `ScrollContainer` widgets and master-detail panes must cap their reported preferred width before parent layout measurement when their content can be wider than the viewport. `StretchContentWidth()` is not enough by itself because clipping happens after parent layout.
   * Descriptive HUD text, summaries, section descriptions, and row details must set a nonzero `widget.TextOpts.MaxWidth(...)` before preferred-size measurement so long strings wrap inside their owner instead of expanding the whole tree.
   * When all tab buttons fit, the tab strip stays left-aligned. When they overflow, rebuilds preserve user-driven scrolling; pointer/touch selection must not reposition it. Keyboard, controller, resize, and other non-pointer selection paths may move the strip only enough to make a clipped active tab fully visible, never by normalized-index centering. Hover styling must remain visually distinct from selected/pressed styling.
   * Dragging or swiping the tab strip must be treated as scrolling, not selection. A drag gesture must not leave a non-selected tab checked, pressed, or visually selected after release. When a widget library defers click or state-change events, drag cancellation must remain active until that deferred event cycle has completed; do not clear it merely because the physical release frame ended.
   * Vertical body, master-list, and detail scroll containers must receive explicit wheel and drag/touch input that changes their `ScrollTop`; EbitenUI clipping alone does not make a container scroll. Target the innermost viewport, preserve tab-strip horizontal scrolling, and keep every body viewport between the tab row and diagnostics/status clearance.

4. **Keep Visual Vocabulary Consistent**
   * Fieldsets use a visible border, compact title, short explanatory text when helpful, and a clear content area.
   * Rows use consistent label/detail spacing and disabled/future markers.
   * All interactive UI elements (list items, rows, buttons, and form controls) *must* enforce a strict minimum touch target height of 44px by configuring `MinSize: (0, 44)` on the widget options.
   * Buttons must be visually distinct from plain text and backgrounds by using distinct background colors (e.g., via `image.NewNineSliceColor(buttonBgColor)`) instead of solid/transparent colors that match the parent panel.
   * Buttons and form controls appear inside the fieldset or toolbar that explains their purpose.
   * Text fields and input-like placeholders are block-level controls; they keep touch target height and clip or wrap content within their own rectangle.

5. **Preserve The HUD Action Model**
   * Controller, keyboard, mouse, touch, and native-platform controls must dispatch the same named actions whenever possible.
   * Scrollable panes support wheel, pointer drag, touch drag, and keyboard/controller scroll bindings through the HUD action model.
   * Native platform views may be used as temporary bridges only when the engine cannot perform the platform operation directly.
   * Native views must be hidden outside the tab and body element they support.

### Right-hand detail navigation

Use a child page in the same detail pane when a button offers several app-owned choices. Show a touch-sized 🔙 Back button first, then a heading and the available options. Keep the selected object visible in the wide selector. On narrow screens use the full detail area. Back returns exactly one level, restores the launching control and scroll position, and leaves pending edits unapplied. Verify touch Back, Escape, controller B and platform Back where supported, including after a data refresh. Keep OS-owned prompts and short confirmations for consequential actions when their interruption serves a specific purpose. Record the reason when retaining a multi-choice popup during HUD review.

## Tab Patterns

* **Settings**: a single vertical list of fieldsets. Each settings category gets a title, optional explanation, and grouped controls or rows. Platform-specific temporary controls belong in the relevant fieldset; new settings groups are appended as additional fieldsets rather than floating over earlier content.
* **Comrades**: selector/content-panel. The selector panel lists people and group chats. The content panel shows the selected thread. When no real thread is selected yet, current placeholder/explanatory content belongs in this blank content panel. The divider is adjustable, and both panels respect minimum widths and scrolling.
* **Projects and Cluster**: selector/content-panel following the Comrades rules, with the selector panel listing relevant objects and optional non-selectable headings. A heading may include a concise, muted, wrapped description; it never receives focus or activation. The content panel shows context, status, actions, and details for the selected item. Routing and Tasks are Cluster selector items; placeholder content belongs in the content panel until real selectable objects are wired.
* **Research**: use the pattern that matches the current feature state. Placeholder/review content can use fieldsets; future project catalogs should become selector/content-panel when selectable projects exist, with placeholder content in the content panel.

## Step-by-Step Instructions

1. **Bind The Work To A Plan**
   * Add or select atomic checklist items for the specific tab-content changes.
   * Name the affected tabs and target pattern: fieldset stack, master-detail, form, table, or custom.

2. **Model The Content**
   * Prefer structured view-model data in `internal/hud` over hard-coded renderer text.
   * Ensure every section has a meaningful title and rows or controls that can be rendered without overlap.
   * Mark unavailable backend actions as disabled or future until their backing systems exist.

3. **Render Using EbitenUI Widget Trees**
   * Construct the EbitenUI widget hierarchy using appropriate containers and layout managers.
   * Use the central theme configuration for padding, margins, gaps, fonts, and colors.
   * Let EbitenUI handle dynamic text wrapping and layout calculation.
   * Keep sortable table headers as ordinary, touch-sized EbitenUI buttons supplied to `borderedTable`; sorting state belongs in the HUD view model or the adapter state boundary, never in formatted cell text.

4. **Handle Responsiveness**
   * Test narrow body widths in unit tests where possible.
   * Confirm every fieldset or pane has a minimum width.
   * Confirm interactive rows and buttons keep minimum touch-target height.
   * For master-detail layouts, define divider bounds and fallback behavior when the body is too narrow. Narrow layouts should show the list first, switch to detail after selection, and expose a touch-sized Back button at the top of detail.

5. **Document The Change**
   * Update `internal/hud/README.md` for view-model or tab-content rules.
   * Update `internal/adapters/gui/README.md` for rendering or native bridge rules.
   * Update product/platform docs if users or contributors need to understand behavior.

6. **Verify**
   * Run focused HUD/GUI tests.
   * For visual layout changes, capture actual rendered screenshots on the target surface before marking the work complete; geometry/unit tests alone are not sufficient.
   * For phone or narrow-surface HUD changes, capture screenshots proving content is bounded to the visible width and horizontally overflowing tab controls can be swiped.
   * Run `make check-docs`.
   * Run `python3 scripts/check_code_file_lines.py`.
   * Run `git diff --check`.
   * For Android or platform-native controls, rebuild and install the APK when the toolchain/device is available.

## Plan Binding

This playbook constrains implementation but does not authorize it by itself. All tab-content work must be represented by approved active plan checklist items before editing source, docs, or platform wrappers.

## Lifecycle Compliance

Prompt -> Select/Create Plan (using relevant playbook guidance) -> Request approval -> Execute approved plan atoms -> Plan update -> Docs update -> Verification.

If this occurs inside a git repo:
* Review `git status` and relevant diffs.
* Suggest a commit message that summarizes the completed tab-content checkpoint.
* Commit after approved checkpoint completion.


## Directional navigation acceptance

Exercise each changed interactive control from the tab strip using arrows and
again with the equivalent D-pad action. Check both a horizontal strip and a side
rail. Verify the visible focus outline, offscreen reveal, entry and return,
single activation, and focus recovery after resize or refresh. A disabled
control or heading must be skipped. Include fixed controls outside scrolling
panes. Give custom-rendered controls the same directional and activation path.

For text inputs, check caret movement, draft preservation and leaving edit mode.
For overlays and games, verify exclusive input and return to the launching
scope. Extend the rendered-widget regression tests for new control types,
layouts or input boundaries. Run the GUI tests through
`python3 scripts/build_gate.py -- xvfb-run -a go test -race -tags gui ./internal/adapters/gui ./internal/hud`.
Record which layouts and input sources were exercised, including any physical
device checks still pending.

## Action discovery acceptance

Register new destinations with the action catalog using the feature's existing source of truth. Verify the same destination opens from its UI, Settings → Actions and an exact spoken command. Cover duplicate names within and between kinds, provider updates after creation, rename and removal, and a result that disappears while the chooser is open. Every enabled tab, including optional editions, must remain discoverable.

Exercise search with a command-shaped dictation, voice cancellation, spoken and directional clarification, and a retry after no match. Verify partial transcripts and approximate matches cannot activate an action. Changing a navigation verb to execute or mutate state requires a separately designed action and authorization boundary.

On Linux, `make test-hud-navigation` runs the rendered navigation and catalog regressions and is included in `make verify`.

### Form typography and alignment

Place a persistent bold label immediately to the left of each editable field, including search filters and controls that open native text entry. End field and choice-group labels with a colon. Give each field row half an em of clearance above and below, based on the selected font size. Right-align labels in a shared column with a four-unit gap before the field. Render search queries in a compact single-line control. Wrap long labels within that column on narrow screens. Render explanatory descriptions in italics and user-entered values in regular type. Check both an empty field and a populated field at phone and desktop widths.

### Form interaction acceptance

- **Entry and focus.** On explicit form entry, focus its first meaningful control. An inline search field is ready to type, with any existing query selected. Focus a native prompt launcher when entry would otherwise interrupt the page with an OS dialog.
- **Filtering and defaults.** Filter as inline text changes or when a native prompt is accepted. Keep focus in search for zero or multiple matches. A nonempty filter with exactly one eligible result may select it and advance when that choice is reversible, as in New Project. Focus the next required group without activating it. Keep a way back to edit earlier answers.
- **Progressive disclosure.** Reveal controls after their prerequisites are answered. Hide irrelevant branches, clear incompatible dependent selections and preserve compatible drafts. Give each group a visible label and an unmistakable selection state. Show why a preview or unavailable operation cannot run.
- **Traversal.** Keep Tab and Shift+Tab in visual reading order. Move with arrows and the D-pad to adjacent controls through the shared HUD route. Skip headings, hidden controls and disabled controls. Revealed focus scrolls into view. Text editing owns caret movement, clipboard and dictation. Escape or controller B leaves editing before leaving the form.
- **Submission and recovery.** Run changes on explicit activation. Show progress and prevent duplicate submission. On errors keep usable input and a clear way to retry. On success move focus to the resulting destination.
- **Verification.** Reuse shared field rows, typography and focus helpers. Check narrow and wide renders, visible focus, forward and reverse order, zero/one/many search results, changing an earlier answer, busy and error states, and return navigation. Exercise the same form by keyboard, controller and touch where supported. Record exceptions alongside the form.
