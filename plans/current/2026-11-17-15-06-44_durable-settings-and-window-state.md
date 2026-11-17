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

This extends [Phase 4 HUD configuration and Settings](../../ROADMAP.md#phase-4-basic-hud-tabs-and-content) using [Phase 8 SQLite and recovery](../../ROADMAP.md#phase-8-shared-core-sqlite-internal-api-read-only-rest-and-first-local-service-slice). Preserve the documented six-tab order with Settings last, icon-only default, keyboard/controller/touch navigation, platform update behavior and developer diagnostics.

- [x] Persist and apply display theme, scale, font size, startup tab and developer overlay through the node database. Verified UI edits and cold restart on Linux and Pixel 4a.
- [x] Restore normal desktop window size and position and maximized state. Verified a maximize/restart/unmaximize cycle on Linux under KWin.
- [x] Implement top, left, right and responsive placement and preserve existing preference records on upgrade.
- [ ] Finish with Bottom placement. Verify the horizontal strip, push-to-talk clearance, gallery navigation, and persisted choice, then archive this plan.
