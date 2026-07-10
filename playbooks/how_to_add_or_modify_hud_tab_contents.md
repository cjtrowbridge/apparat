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

1. **Use Structured Body Elements**
   * A tab body is a layout surface, not a free-form text canvas.
   * Put related content in separate bounded elements such as fieldsets, lists, detail panes, forms, tables, or toolbars.
   * Each bounded element must own its title, explanation, controls, and body content.

2. **Prevent Overlap By Construction**
   * Do not place controls as floating overlays unless they correspond to a reserved body element.
   * Stack vertical elements using measured heights and fixed gaps.
   * Split horizontal elements using explicit columns and a divider.
   * Clip, wrap, scroll, or truncate overflowing content inside its owning element; never let it draw over the next element.

3. **Be Responsive By Default**
   * Every tab body must work down to the current minimum supported body width for the target surface.
   * If a tab uses columns, define a minimum width for each column and collapse or scroll when those minimums cannot fit.
   * Use bounded body dimensions, padding, gaps, and scroll areas instead of viewport-scaled fonts.

4. **Keep Visual Vocabulary Consistent**
   * Fieldsets use a visible border, compact title, short explanatory text when helpful, and a clear content area.
   * Rows use consistent label/detail spacing and disabled/future markers.
   * List items, rows, buttons, and form controls must meet the same touch-first minimum target sizing as tab buttons.
   * Buttons and form controls appear inside the fieldset or toolbar that explains their purpose.

5. **Preserve The HUD Action Model**
   * Controller, keyboard, mouse, touch, and native-platform controls must dispatch the same named actions whenever possible.
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

3. **Render With Measured Layout**
   * Compute body, element, pane, row, and control rectangles before drawing.
   * Keep padding and gaps as named constants.
   * Clip or truncate text to the element width.
   * Reserve space for native bridge controls before placing them.

4. **Handle Responsiveness**
   * Test narrow body widths in unit tests where possible.
   * Confirm every fieldset or pane has a minimum width.
   * Confirm interactive rows and buttons keep minimum touch-target height.
   * For master-detail layouts, define divider bounds and fallback behavior when the body is too narrow.

5. **Document The Change**
   * Update `internal/hud/README.md` for view-model or tab-content rules.
   * Update `internal/adapters/gui/README.md` for rendering or native bridge rules.
   * Update product/platform docs if users or contributors need to understand behavior.

6. **Verify**
   * Run focused HUD/GUI tests.
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
