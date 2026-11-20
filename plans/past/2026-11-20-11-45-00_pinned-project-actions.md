---
plan_id: 2026-11-20-11-45-00_pinned-project-actions
title: Pin Projects repository actions
summary: Keep the three plus-prefixed entry actions visible above scrolling checkouts.
status: past
created_at: 2026-11-20-11-45-00
---

# Pin Projects repository actions

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Phase 9 local repository UI. Preserve form entry points and selection indices.

- [x] Pin clone, init and import above the list and prepend the requested ➕ artwork.
- [x] Exercise scrolling and directional focus on wide and collapsed layouts.

The governed GUI race suite verified fixed button bounds and directional entry into a scrolled 24-checkout list at widths of 360 and 1280 pixels.
