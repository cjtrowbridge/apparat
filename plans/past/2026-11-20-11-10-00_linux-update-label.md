---
plan_id: 2026-11-20-11-10-00_linux-update-label
title: Show the Linux update check state
summary: Distinguish the initial action and a confirmed current build in Settings.
status: past
created_at: 2026-11-20-11-10-00
---

# Show the Linux update check state

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The requested correction refines the Phase 17 Linux updater. Preserve one-click installation, restart and wrapped detail.

- [x] Label the check, in-progress, current and retry states. Verify the rendered button and recheck behavior.

GUI, self-update and HUD race suites passed. The 320-pixel Settings regression exercises initial, busy, installed, current, recheck and failed-check states. Documentation and source gates passed.
