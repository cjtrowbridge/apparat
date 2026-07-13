# Playbook: Add Or Modify HUD Tab Contents

*Status: Draft*

## Objective

Define how to add or modify Apparat HUD tab bodies so each tab remains responsive, non-overlapping, controller-first, and consistent with the product information architecture.

## Prerequisites

* Read `README.md`, `ROADMAP.md`, `AGENTS.md`, and `agents/RULES.md`.
* Identify the active execution plan item that authorizes the tab-content change.
* Read `internal/hud/README.md` and `internal/adapters/gui/README.md`.
* Inspect the current tab model in `internal/hud` and the rendering/input adapter in `internal/adapters/gui`.

## Layout Rules

1. **Use Structured Body Elements via EbitenUI**
   * EbitenUI is the mandatory UI toolkit for the Apparat HUD. All layout and widgets must use `github.com/ebitenui/ebitenui`.
   * Do not use raw imperative coordinates (`ebitenutil.DebugPrintAt`, `ebitenutil.DrawRect`), custom `SubImage` clipping, or custom drag-scroll loops for ordinary tab body content.
   * Shell-level development overlays, tab-strip overflow input, and divider dragging may use small Ebitengine input/rendering shims only when an active plan explicitly authorizes them and the user-facing controls remain EbitenUI widgets.
   * Put related content in EbitenUI layout containers (like `widget.RowLayout` or `widget.GridLayout`) and `widget.ScrollContainer`.
   * Each bounded element must own its title, explanation, controls, and body content.
   * Tab content must remain strictly data-driven (dynamically iterating through `tab.Sections` and `section.Rows`); do not hardcode EbitenUI containers or widgets outside of this loop pattern.

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
   * Tab-strip active-tab auto-scroll must be one-shot after rebuilds, resizes, or programmatic tab changes. It must not run every update or overwrite pointer, wheel, mouse-drag, or touch-drag scrolling.
   * Dragging or swiping the tab strip must be treated as scrolling, not selection. A drag gesture must not leave a non-selected tab checked, pressed, or visually selected after release. When a widget library defers click or state-change events, drag cancellation must remain active until that deferred event cycle has completed; do not clear it merely because the physical release frame ended.

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

## Tab Patterns

* **Settings**: a single vertical list of fieldsets. Each settings category gets a title, optional explanation, and grouped controls or rows. Platform-specific temporary controls belong in the relevant fieldset; new settings groups are appended as additional fieldsets rather than floating over earlier content.
* **Comrades**: master-detail. The left pane lists people and group chats. The right pane shows the selected thread. When no real thread is selected yet, current placeholder/explanatory content belongs in this blank right-hand detail pane. The divider is adjustable, and both panes respect minimum widths and scrolling.
* **Projects, Cluster, Routing, and Tasks**: master-detail following the Comrades rules, with the left pane listing the relevant objects and the right pane showing context, status, actions, and details for the selected item. Placeholder content belongs in the right-hand detail pane until real selectable objects are wired.
* **Research**: use the pattern that matches the current feature state. Placeholder/review content can use fieldsets; future project catalogs should become master-detail when selectable projects exist, with placeholder content in the detail pane.

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
