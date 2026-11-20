---
plan_id: 2026-11-20-10-30-00_compact-gallery-header
title: Compact embedded gallery header
summary: Place STMPCK and its gallery controls together at the top of the available page.
status: past
created_at: 2026-11-20-10-30-00
---

# Compact embedded gallery header

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The requested layout refines the optional STMPCK edition in ROADMAP and the README gallery contract.

- [x] Remove both obsolete rows and put the title before the star. Share toolbar geometry between drawing, pointer and directional focus.
- [x] Verify the Deck-sized and narrow layouts, tile scrolling and activation. Update the host documentation.

The Linux GUI window rendered the single toolbar at 1280×800 and 320×720. At Deck size the first focused tile moved from y=211 to y=73. Enter opened Merge and gameplay saved a move. The GUI/HUD race suite passed, including the top-row gear and first and last tile hit tests. Source governance, size and documentation checks passed. Local evidence is in `.tmp/directional-smoke/compact-deck.png` and `compact-narrow.png`.
