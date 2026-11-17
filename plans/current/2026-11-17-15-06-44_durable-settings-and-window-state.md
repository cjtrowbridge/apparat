---
plan_id: 2026-11-17-15-06-44_durable-settings-and-window-state
title: Implement durable Settings controls and desktop window restoration
summary: Add SQLite-backed local preferences and restore desktop size, placement and maximized state.
status: current
created_at: 2026-11-17-15-06-44
---

# Implement durable Settings controls and desktop window restoration

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

A preference changed in Settings takes effect through its real consumer and survives an ordinary restart. On desktop, Apparat reopens with its last usable normal window size and position, and restores maximized state when supported. The per-node `data/apparat.db` owns both sets of values.

This extends [Phase 4 HUD configuration and Settings](../../ROADMAP.md#phase-4-basic-hud-tabs-and-content) using [Phase 8 SQLite and recovery](../../ROADMAP.md#phase-8-shared-core-sqlite-internal-api-read-only-rest-and-first-local-service-slice). Preserve the documented six-tab order with Settings last, icon-only default, keyboard/controller/touch navigation, platform update behavior and developer diagnostics. Phase 4 records icon-plus-text as a historical default. Reconcile that entry and the old five-tab exit criterion with the current README when implementing. Device and network operations remain tied to Phases 9–10, provider/audio preferences to Phases 12 and 16, and platform evidence to Phase 17.

## 1. Establish the durable path

- [ ] Add the next forward migration for local preferences and desktop window state. Use one typed preference snapshot in the core, a SQLite repository, and a separate window-state record so a geometry write cannot overwrite preferences. Reuse current defaults for missing values and validate edits before storing them. Keep config fields that describe supported options or safety invariants in code. Persist local values only and keep credentials and private keys in their existing stores.
- [ ] Load before the first Settings render, with the same core path for desktop and Gomobile. Route edits and reset-to-default through typed application commands. Use an application-owned worker for database I/O, serialize writes, and project confirmed values back to the controls. Show save failures and allow retry while keeping the last confirmed preference visible. Make reset transactional and limited to preferences. Retain migration checksums and verify a version 3 database with existing service, directory and messaging records upgrades in place.

## 2. Remember the desktop window

- [ ] In the desktop entry path, capture normal content width/height and monitor-relative position in device-independent pixels, plus a monitor hint where reliable and the maximized flag. Debounce changes outside the frame loop and flush pending state on normal shutdown before closing the owning runtime. Keep the most recent normal rectangle when maximizing or minimizing. Use the last committed geometry following an abrupt exit.
- [ ] Restore the normal rectangle first and apply maximization once the native window supports it. Check the current monitor setup, resolution, scale and work area where available. Preserve valid offsets on secondary monitors, and clamp or center when a monitor disappears or the saved bounds would leave the window inaccessible. Where the window manager controls placement or cannot report the state, restore what the platform supports and document the observed limits. Android continues to use its native lifecycle and safe insets. Add a desktop Settings action to reset placement without resetting other preferences.

## 3. Replace placeholders in order of working behavior

- [ ] **Display and navigation.** Start with the visible theme, UI scale, font size, landing tab and diagnostics controls as the first end-to-end slice. Continue with top/side/responsive placement, tab density and label modes, accent, bundled font choices, reduced motion, contrast and focus styles, panel/row/card density, and a readable label overflow policy. Add each control alongside its renderer behavior. Preserve wrapping as the default and enforce a safe reset even after extreme layout or scaling changes.
- [ ] **Input and interaction.** Edit keyboard/controller/touch bindings through the existing action model. Validate conflicts, reserved shortcuts and a usable navigation/reset path before committing. Connect repeat delays, scroll behavior, command palette preferences, and remembered tab/panel/selection and sort/filter where the view supports stable IDs. Local selection memory applies only when enabled. Keep hold-to-talk until a toggle implementation passes capture, cancel and focus-loss checks.
- [ ] **Diagnostics, notifications and privacy.** Wire log detail, developer overlay and available diagnostic sections to their real presentation. Keep diagnostics and updates as actions/status. Connect notification categories, volume, duration, quiet hours and focus behavior as their event/sound handlers arrive. Verify privacy display options and retain mandatory reveal and confirmation protections.
- [ ] **Backend-dependent and operational controls.** Track the remaining README Settings categories in this plan through their own phase dependencies. Identity, enrollment/revocation/recovery, WireGuard/LAN, audio devices and ASR/TTS need working backend commands before they become editable. Storage diagnostics, backup and repair should use existing validated operations where available. Add durable preferences only for actual user choices. Keep status as a projection and actions as explicit commands. A backend-blocked item stays pending with its dependency named.

## 4. Verify each slice before private delivery

- [ ] Exercise fresh defaults, upgrade and repeat migration, preference change/restart/reset, invalid input, write failure, backup/restore and preservation of existing rows. Verify immediate consumer effects as well as the reopened control values. Check keyboard/controller access, touch targets, wrapped labels at minimum split width, Settings last and live status after a layout rebuild.
- [ ] On Linux with a window manager, drag and resize, close/reopen, maximize, close/reopen, then restore and compare the normal rectangle. Test keyboard and window-manager close, a missing secondary monitor, DPI/work-area changes and invalid stored geometry. Use deterministic geometry tests and an isolated runtime with a real window manager for native behavior. Record Windows and macOS evidence when available. On the Pixel 4a verify Settings persistence across app restart, safe area behavior and the Android update presentation.
- [ ] Run the canonical no-flag pipeline on main for each implementation checkpoint. Qualify the Linux GUI, Linux daemon and Android APK on this host, verify private delivery, and document current controls, database recovery and platform-specific placement limits. The first checkpoint is complete when a usable Settings slice persists through restart and the desktop restore cycle passes. Keep later control groups pending until their actual consumers and checks pass.

### First implementation checkpoint

Migration 4 stores display preferences and desktop geometry independently. Settings includes dark/high-contrast themes, logical scale, font size, startup tab, persistent overlay and reset actions. Runtime queries supply both GUI entry points. Later controls under section 3 remain pending.

Verified through the real Settings widgets and SQLite on Linux, including immediate scale application and startup-tab selection on reopen. Under KWin/X11 the window reopened maximized and restored to its previous 810 × 620 normal bounds at (72, 90). On the Pixel 4a the high-contrast theme remained active after force-stop and cold launch. Other desktop operating systems and monitor-change behavior await native qualification.

### Tab placement checkpoint

Implement top, left, right and responsive Settings choices through migration 5 and the existing confirmed-save path. Verify the real Settings button, adaptive layout and database reopen, then exercise placement and navigation on Pixel 4a. Earlier preferences and desktop placement survive the upgrade.

Verified Linux right-rail rendering and Pixel 4a left-rail cold restart. Touch navigation opened STMPCK and returned to Settings. Restored the Pixel's original Top choice. Linux visual inspection exposed a push-to-talk overlap with the right rail at short desktop heights. Reserve the rail gutter for both the control and its hit test and cover it in the GUI suite.
