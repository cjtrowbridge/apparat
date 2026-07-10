---
plan_id: 2026-07-10-16-33-53_fix-ebitenui-regressions
title: Fix EbitenUI Layout Regressions
summary: Restores the canonical layout, master-detail UI, sizing constraints, and dynamic update logic missing after EbitenUI migration.
status: current
created_at: 2026-07-10-16-33-53
---

# Plan: Fix EbitenUI Layout Regressions

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Product Binding
- **ROADMAP.md:** This plan binds to **Phase 5: Android GUI APK Build Pipeline**, specifically resolving regressions in "wrapper HUD rendering" and "touch tab selection" introduced during the recent layout framework migration. It fulfills the phase requirement to correctly render a cross-platform HUD with proper safe-area and density hardening (44px touch targets).
- **README.md:** This plan restores the canonical Apparat tab model and multi-pane information architecture (Comrades, Projects, Research, Cluster, Routing, Tasks, Settings) as defined in the README. It also ensures EbitenUI adheres to the controller-friendly and touch-accessible constraints mandated for the application shell.

## 1. Problem: Deleted UI / Missing Content
**Issue:** During the migration, `ui_builder.go` was written with placeholders for both the master-detail tabs and the dynamic contents of the Settings tab. The original logic that iterated over `tab.Sections` and `section.Rows` was discarded.
**Fix:**
- **Settings Tab:** Rewrite `buildSettingsTab` to dynamically iterate through `tab.Sections`. For each section, generate a `widget.Container` (with the new bordered panel background), add a title `widget.Text`, and iterate through `section.Rows` to generate `widget.Text` (or buttons/inputs based on the row ID).

**Execution Tasks:**
- [ ] 1. Restore dynamic Settings tab.
  - [ ] 1.1 Update `internal/adapters/gui/ui_builder.go` to recreate the dynamic Settings loop.
    - [ ] 1.1.1 Implement a loop over `tab.Sections` in `buildSettingsTab` to generate EbitenUI containers for each fieldset.
    - [ ] 1.1.2 Implement an inner loop over `section.Rows` to generate EbitenUI text widgets for each row.

## 1.5. Problem: Master-Detail Layout Deleted for All Other Tabs
**Issue:** The UI for all other tabs (`Comrades`, `Projects`, `Research`, `Cluster`, `Routing`, `Tasks`) was completely replaced with a generic "Placeholder for other tabs" text widget. The original `drawMasterDetailBody` logic, which rendered a complex split-pane interface with calculated ratios, borders, and dynamic lists, was discarded.
**Fix:**
- **Rebuild `buildMasterDetailTab`:** Recreate the split-pane layout using EbitenUI's `GridLayout` or a horizontal `RowLayout`.
- **Left List Pane:** Implement the list pane to replicate `drawListPane`. It needs a minimum width of 170px (`masterMinListW`), the dark `listPaneColor` background, and the `bodyBorderColor` outline. It should dynamically list all `section.Title`s from `tab.Sections`. Create a `widget.List` or vertical layout of interactive 44px buttons for these items. The first item must default to a selected visual state.
- **Right Detail Pane:** Implement the detail pane to replicate `drawDetailPane`. It must use the `fieldsetColor` background and have a border. It should display a "Placeholder Detail" header, a wrapped text block for `tab.Summary`, and then dynamically loop over `tab.Sections` (just like the Settings pane) to render the fieldsets, titles, descriptions, and rows within the available scrolling area.
- **Responsive Sizing:** Ensure the left pane dynamically respects the `3/10` ratio logic (`masterListRatioNum` / `masterListRatioDen`) while constraining to the `170px` minimum.

**Execution Tasks:**
- [ ] 1.5 Restore Master-Detail Layout for all non-Settings tabs.
  - [ ] 1.5.1 Implement `buildMasterDetailTab` in `ui_builder.go`.
    - [ ] 1.5.1.1 Create a horizontal split-pane EbitenUI Grid or RowLayout.
  - [ ] 1.5.2 Recreate the left list pane.
    - [ ] 1.5.2.1 Apply `masterListRatioNum` (3) / `masterListRatioDen` (10) sizing with a minimum width constraint of 170px.
    - [ ] 1.5.2.2 Loop over `tab.Sections` to populate the left list pane with buttons for each section title.
  - [ ] 1.5.3 Recreate the right detail pane.
    - [ ] 1.5.3.1 Provide a placeholder text block showing `tab.Summary`.
    - [ ] 1.5.3.2 Dynamically loop over `tab.Sections` to render the fieldsets in the right pane exactly as they render in Settings.
    - [ ] 1.5.3.3 Implement an event handler so clicking an item in the left list pane updates the scroll position or visibility of the right detail pane.

## 2. Problem: Button Touch Targets Too Small
**Issue:** The Android UX guidelines in this project enforce a minimum `touchTargetH = 44`. The current EbitenUI buttons auto-size based on padding and text height, making them too small to reliably tap on a tablet or phone.
**Fix:**
- Apply `widget.ButtonOpts.WidgetOpts(widget.WidgetOpts.MinSize(0, 44))` to all generated interactive buttons, including the "Check for update" button and the master list items.
- Update `TabbookTheme` in `theme.go` to ensure `TabButton` instances also respect the minimum height requirements.

**Execution Tasks:**
- [ ] 2. Apply minimum touch target dimensions.
  - [ ] 2.1 Apply `MinSize: (0, 44)` to all EbitenUI button definitions in `ui_builder.go`.
  - [ ] 2.2 Update `TabbookTheme` in `theme.go` to inject minimum sizing constraints.

## 3. Problem: Broken Characters in Tab Names
**Issue:** The old `body_layout.go` did not prefix tab titles with glyphs (it just used `tab.Title()`), but the new `ui_builder.go` explicitly prepends `tabData.Descriptor.Glyph`. The default `GoXFace` font does not contain these Unicode icons, rendering them as broken "tofu" boxes.
**Fix:**
- Revert the tab label construction in `buildTabs` to strictly use `tabData.Title()` without the missing glyph, OR load an external font that supports those glyphs. For parity with the previous implementation, stripping the unsupported glyph is the correct immediate fix.

**Execution Tasks:**
- [ ] 3. Fix broken Unicode tab labels.
  - [ ] 3.1 Revert `ui_builder.go` `buildTabs` logic to strictly use `tabData.Title()`.
  - [ ] 3.2 Drop the unsupported `tabData.Descriptor.Glyph` from the layout configuration.

## 4. Problem: Missing Outlines
**Issue:** The legacy UI explicitly drew borders around panels using `bodyBorderColor`. The new EbitenUI theme relies on `image.NewNineSliceColor(bgColor)`, which paints a flat, borderless rectangle.
**Fix:**
- Create a new helper function in `theme.go`: `createBorderedNineSlice(fill color.Color, border color.Color) *image.NineSlice`.
- This function will generate an actual `ebiten.Image` (e.g., a 3x3 or larger square) where the outer pixels are `bodyBorderColor` and the inner pixels are the `fill` color, then slice it appropriately.
- Apply this bordered NineSlice to `theme.PanelTheme`, `ScrollContainerImage`, and relevant `ButtonTheme` states to restore the crisp outlines.

**Execution Tasks:**
- [ ] 4. Introduce bordered graphics.
  - [ ] 4.1 Introduce `createBorderedNineSlice` in `theme.go`.
    - [ ] 4.1.1 Programmatically create a 3x3 graphic using `bodyBorderColor` for the outer edge and a distinct fill color for the inner square.
  - [ ] 4.2 Apply this bordered NineSlice to `theme.PanelTheme` and `ScrollContainerImage` so layout panes have crisp borders.

## 5. Problem: Update Button No Longer Works and Looks Like Plain Text
**Issue:** 
1. The "Check for update" button blends into the background because `theme.ButtonTheme.Image` uses `bgColor`, making it look like floating, clickable text rather than a distinct button.
2. It is hardcoded at the top of the settings page rather than correctly mapped to the `Updates` section from the snapshot data.
3. If `onCheckForUpdate` is `nil` (e.g., when testing on the Linux desktop environment), the button silently does nothing, appearing broken.
**Fix:**
- **Visuals:** Ensure `theme.ButtonTheme.Image` utilizes `createBorderedNineSlice` with a contrasting fill color (like `fieldsetColor` or `accentColor`) so buttons look like distinct UI elements.
- **Dynamic Placement:** Inside the new dynamic `buildSettingsTab` loop, when encountering a section with `Title: "Updates"`, render the `widget.Button` in that section container.
- **Feedback:** Provide fallback behavior: if `game.onCheckForUpdate == nil`, update an on-screen status text or the button label itself to provide immediate visual feedback that the button was pressed.

**Execution Tasks:**
- [ ] 5. Fix update button logic and aesthetics.
  - [ ] 5.1 Apply the bordered NineSlice to `theme.ButtonTheme` (with a contrasting fill like `accentColor` or `fieldsetColor`) so buttons look like physical buttons.
  - [ ] 5.2 Map the "Check for update" button to the correct dynamic section.
    - [ ] 5.2.1 Add conditional logic in the `tab.Sections` loop to intercept `section.Title == "Updates"`.
    - [ ] 5.2.2 Create the interactive EbitenUI update button at that specific section.
  - [ ] 5.3 Add a local UI fallback for when `game.onCheckForUpdate` is nil (non-Android environments).
    - [ ] 5.3.1 Ensure the button updates its own label or logs to the screen when clicked if the Gomobile bridge is unavailable.

## 6. Playbook Governance
**Issue:** The UI framework migration lacked clear documentation for the specific constraints of the EbitenUI implementation, directly leading to these regressions.
**Fix:**
- Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to formally document these new EbitenUI standards.
- Explicitly dictate that all interactive elements *must* enforce the 44px minimum touch target via `MinSize`.
- Explicitly dictate that buttons must use bordered NineSlice images (not solid transparent/background colors) to ensure they are visually distinct from text.
- Reiterate that all tab content must remain data-driven (rendered by looping over `tab.Sections` and `section.Rows`) rather than hardcoding EbitenUI containers manually in `ui_builder.go`.

**Execution Tasks:**
- [ ] 6. Update playbooks and documentation.
  - [ ] 6.1 Edit `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
    - [ ] 6.1.1 Document the strict 44px minimum touch target requirement (`MinSize: (0, 44)`).
    - [ ] 6.1.2 Document the requirement to use bordered NineSlice graphics for buttons so they are visually distinct.
    - [ ] 6.1.3 Prohibit hardcoding UI layouts outside of the data-driven `tab.Sections` and `section.Rows` loop pattern.

## 7. Verification and Finalization
**Execution Tasks:**
- [ ] 7.1 Verification.
  - [ ] 7.1.1 Verify layout aesthetics and rendering locally on Linux (`make verify`).
  - [ ] 7.1.2 Build the Android APK (`make build-android`).
  - [ ] 7.1.3 Deploy to the Android tablet (`adb install -r releases/android/arm64/apparat/latest.apk`).
  - [ ] 7.1.4 Visually verify layout logic, 44px touch targets, and pane borders on the tablet.
- [ ] 7.2 Finalization.
  - [ ] 7.2.1 Update the daily journal (`journal/2026-07-10.md`) detailing the UI restorations.
  - [ ] 7.2.2 Commit and push the final fixes to `main`.
