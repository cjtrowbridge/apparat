---
plan_id: 2026-07-10-15-44-00_fix-hud-layout-and-update-button
title: Fix HUD Layout and Android Update Button
summary: Clean up the Ebitengine layout code to prevent text clumping on narrow screens, and replace the native Android update button overlay with a fully Ebitengine-rendered control triggered via JNI.
status: past
created_at: 2026-07-10-15-44-00
---

# Fix HUD Layout and Android Update Button

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## High-Level Issues and Solutions

- ### 1. Native Android Update Button "Chasing" the Scroll Column
  - **Issue Description:** The native Android `Button` is layered on top of Ebitengine's OpenGL canvas using a `FrameLayout`. A background polling loop in Java queries Go for coordinates every 250ms to reposition the button. Because Ebitengine renders at 60 FPS while the polling loop runs at 4 Hz, the button lags behind, jitters, and cannot be clipped by the game viewport when the settings section scrolls off-screen. Additionally, coordinate scale differences between physical and logical pixels cause misalignment.
  - **Proposed Solution:** Eliminate the native Android button overlay entirely and draw the button inside the Ebitengine render loop using EbitenUI.
    - `[ ]` Delete the native `Button` creation, layout parameters, and the 250ms polling thread (`updateButtonVisibility`) in `MainActivity.java`.
    - `[ ]` Replace the layout slot geometry functions (`UpdateButtonX/Y/W/H/Visible`) in Go and Java with a direct JNI trigger hook (e.g., `TriggerCheckForUpdate()`).
    - `[ ]` Add a fully EbitenUI-rendered `widget.Button` inside the `Updates` settings container.
    - `[ ]` Hook the button's click event handler in Go to invoke the `TriggerCheckForUpdate()` JNI method.

- ### 2. Text Clumping and Overlapping on Non-Settings Tabs
  - **Issue Description:** The current custom layout system manually wraps text (`wrapText`) and calculates heights (`fieldsetHeight`) using hardcoded assumptions (e.g., assuming a text wrapping width of `360` pixels). When screens resize or run in portrait aspect ratios, the dynamic width deviates from this assumption, causing text elements to wrap to more lines than their allocated containers. This results in nested layout calculations collapsing, drawing text directly on top of other text.
  - **Proposed Solution:** Migrate the entire HUD rendering codebase to utilize EbitenUI layout managers and widgets.
    - `[ ]` Use EbitenUI container layout managers (like `widget.RowLayout` and `widget.GridLayout`) to dynamically calculate section heights and positions, eliminating manual coordinate calculations.
    - `[ ]` Use EbitenUI's `widget.ScrollContainer` to manage viewport scroll clipping and touch/mouse scroll inputs natively.
    - `[ ]` Declare EbitenUI as the exclusive framework for all current and future HUD UI components.

- ### 3. Unconditional Split-Pane Layout on Narrow Viewports
  - **Issue Description:** Non-settings tabs draw a split-pane Master-Detail layout. On mobile screens with a narrow vertical aspect ratio, splitting the viewport horizontally squeezes both panes to the point where they have negative widths, collapsing all text elements to `(0, 0)`.
  - **Proposed Solution:** Use EbitenUI containers to implement an adaptive, responsive viewport layout strategy in Go.
    - `[ ]` Define a viewport breakpoint (e.g., 640 logical pixels wide).
    - `[ ]` For viewports wider than the breakpoint, construct a horizontal split container (List on left, Detail on right).
    - `[ ]` For viewports narrower than the breakpoint, construct a single-pane viewport that dynamically swaps the active child widget (rendering the list first, and transition to full-screen detail view with a back action button when an item is selected).

---

## Detailed EbitenUI Conversion Plan

- ### 1. Update Playbooks and Governance
  - `[ ]` Modify `playbooks/how_to_add_or_modify_hud_tab_contents.md` to state that EbitenUI is the mandatory UI toolkit for the Apparat HUD.
  - `[ ]` Update playbook layout guidelines to prohibit raw imperative coordinates, custom `SubImage` clipping, and custom drag-scroll loops in favor of EbitenUI widgets.

- ### 2. Define the HUD Theme/Skin Configuration
  - `[ ]` Set up a central theme configuration file (e.g. `internal/adapters/gui/theme.go`) specifying standard color palettes, margins, font faces, and button image assets.
  - `[ ]` Create lightweight 9-slice image resources (using Ebitengine primitive drawing) to style panel backgrounds, buttons (idle, hovered, pressed, disabled), and scrollbars.

- ### 3. Migrate the Top Tab Strip
  - `[ ]` Replace the custom tab render/scroll loop in `ebiten_shell.go` with an EbitenUI container hosting a horizontal bar of `widget.Button` elements or a `widget.TabBook`.
  - `[ ]` Bind the tab button clicked handlers to `game.shell.SelectTab(index)`.

- ### 4. Migrate the Settings Tab
  - `[ ]` Define the Settings view as an EbitenUI container wrapped in a vertical `widget.ScrollContainer`.
  - `[ ]` Rebuild each Settings section as a card-like `widget.Container` containing labels, descriptions, and rows.
  - `[ ]` Place the EbitenUI "Check for Update" button in the `Updates` card.

- ### 5. Migrate Master-Detail Tabs
  - `[ ]` Rebuild the body layout of Comrades, Projects, Research, Cluster, Routing, and Tasks using EbitenUI containers.
  - `[ ]` Render lists using `widget.List` or vertical container stacks of touch-friendly rows.
  - `[ ]` Render details using a scrollable text block/fieldset widget container.
  - `[ ]` Hook the window size layout callback to dynamically rebuild the EbitenUI container structure (switching between horizontal split layout and single-pane detail view based on screen width).

---

## Detailed Implementation of Related TODO.md Backlog Items

We are incorporating and fully solving the following backlog tasks during the EbitenUI migration:

- ### 1. Android Tab Jumping Bug Resolution (Addressing TODO Item 6)
  - **Goal:** Prevent the tab strip from resetting its scroll position or jumping on touch release, allowing users to scroll and tap tabs without coordinate snapping.
  - **Implementation Steps:**
    - `[ ]` Deploy an EbitenUI container with a horizontal `widget.ScrollContainer` to host the tab buttons.
    - `[ ]` Remove the custom tab scroll drag tracking state (`mouseTabDrag`, `touchTabDrag`) and the snap math in `ensureTabVisible` and `dragTabs` in `ebiten_shell.go`.
    - `[ ]` Bind tab-selection actions to EbitenUI button clicked callbacks, ensuring that a tap/click selects the correct index based on standard hit tests, ignoring previous scroll offsets.
    - `[ ]` Verify that manual drag scrolling of the tab strip operates using standard EbitenUI touch physics, preserving position when the user lets go.

- ### 2. View Settings and Draggable Dev Overlay (Addressing TODO Item 8)
  - **Goal:** Add a "View" settings category containing Dev Overlay and Theme options, and implement a draggable debug overlay window.
  - **Implementation Steps:**
    - `[ ]` Update configuration structs (`Config` / `HUDConfig`) to include a `ShowDevOverlay bool` preference.
    - `[ ]` Create a new Settings card-container "View" using EbitenUI.
    - `[ ]` Add a `widget.Checkbox` to the View card, labeled "Show dev overlay", bound to toggle the config's `ShowDevOverlay` setting.
    - `[ ]` Add a `widget.ListCombobutton` dropdown selector labeled "Theme", containing a single placeholder item "Dark Mode" (unwired).
    - `[ ]` Create a floating, draggable EbitenUI `widget.Window` triggered by the `ShowDevOverlay` boolean.
    - `[ ]` Render system diagnostics inside the window: logical/physical resolution (`game.width` x `game.height`), frame rate (FPS), update rate (UPS), system resource usage (memory/CPU metrics), the binary executable path, and the local runtime path.
    - `[ ]` Ensure EbitenUI handles the window dragging events and layout layering over other HUD contents.

- ### 3. Responsive Single-Pane / Split-Pane Adaptive Layout (Addressing TODO Item 9)
  - **Goal:** Implement a screen width threshold that collapses the two-pane Master-Detail layout into a single-pane navigation flow with a "Back" button.
  - **Implementation Steps:**
    - `[ ]` Retrieve the breakpoint threshold (e.g. `640` logical pixels) from the configuration manager.
    - `[ ]` Read the layout dimensions dynamically.
    - `[ ]` If the width is below the threshold:
      - `[ ]` Hide the right Detail pane widget.
      - `[ ]` Display the List pane widget full-screen.
      - `[ ]` When a list item is selected, hide the List pane and transition the viewport to display the Detail pane full-screen.
      - `[ ]` Place a `widget.Button` labeled "<- Back" at the top of the Detail pane.
      - `[ ]` Bind the Back button clicked event, the gamepad `B` button, and keyboard `Escape` to toggle the active pane state back to the List view.

---

## Validation and Verification

To verify that the implementation completely and correctly resolves all layout and functional issues, the following checks must be completed:

- ### 1. Static Analysis & Compilation
  - `[ ]` Verify the Go codebase compiles on the current host by running `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - `[ ]` Verify the Android APK build succeeds by running `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build-android`.
  - `[ ]` Run the test suite using `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test` and `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test-gui-deps` to ensure all UI assertions pass.
  - `[ ]` Run the doc verification check (`make check-docs`) to ensure the updated playbook conforms to style guidelines.
  - `[ ]` Run the file size check (`make check-code-size`) to guarantee that any new package files remain below the 400-line limit.

- ### 2. Visual and Interaction Validation (Debian)
  - `[ ]` Launch the built local binary and test tab switching. Verify that tab selection is responsive and does not cause position jumping.
  - `[ ]` Verify the responsive layout collapses to a single column when the window width is dragged below 640 logical pixels, and expands to dual columns when dragged above.
  - `[ ]` Verify that selecting list items in single-column mode swaps viewports to the detail pane, and clicking the "<- Back" button returns to the list view.
  - `[ ]` Open Settings -> View Settings, check the "Show dev overlay" box, and verify that a draggable window displaying active performance/system stats overlays correctly.
  - `[ ]` Verify that scrolling Settings (via scroll wheel or mouse dragging) operates smoothly at 60 FPS, with all widgets and content clipping correctly at the container boundaries.

- ### 3. Mobile Device Validation (Android)
  - `[ ]` Install the generated APK on an Android device or emulator.
  - `[ ]` Verify touch drag-scrolling behaves smoothly without snapping tabs or column items unexpectedly.
  - `[ ]` Navigate to Settings, verify the EbitenUI-drawn "Check for Update" button is rendered and scaled correctly matching the layout.
  - `[ ]` Tap the update button and confirm the background update download and intent triggers are called via JNI (verifying the JNI method executes without linker errors).

## Archive Note

- 2026-07-10: Archived as superseded by `plans/current/2026-07-10-18-58-38_recover-ebitenui-hud-settings-first.md`. This plan captured the pivot to EbitenUI, but its checklist no longer matches the current blank-tab recovery needs and stale-code/test sweep.
