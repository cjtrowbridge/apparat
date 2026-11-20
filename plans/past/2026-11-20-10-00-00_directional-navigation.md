---
plan_id: 2026-11-20-10-00-00_directional-navigation
title: Directional HUD navigation
summary: Share arrow and D-pad focus across the tab strip, live controls and embedded gallery.
status: past
created_at: 2026-11-20-10-00-00
---

# Directional HUD navigation

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Approved in conversation. Bind to ROADMAP Phase 17 input equivalence and the README keyboard and Steam Deck contracts. Preserve pointer editing, L2 PTT, gameplay input and responsive placement.

- [x] Implement shared spatial traversal, edit ownership, focus restoration, scrolling and embedded menu controls.
- [x] Document the interaction and HUD change requirements. Exercise rendered widgets, responsive bars and embedded input boundaries under the governed GUI test entrypoint.

## Verification

The governed GUI and HUD race suites passed. Coverage includes all six tabs
across four strip placements, directional reveal and restoration, disabled
controls, divider adjustment, checkbox and button activation, editor draft and
selection across refresh, gallery toolbar and grid, and embedded settings.
The 66 build-pipeline tests passed.

An isolated Xvfb GUI package window received X11 key events. Captures showed
the outline on two different Settings controls, the active tab rail button and
the first STMPCK tile. Enter opened Merge and a directional move reached the
saved game. The local captures are in `.tmp/directional-smoke/`. The package
window uses fixture data, so release qualification of the Projects registry
remains the responsibility of the ordinary application build.

Phase 17 physical-device acceptance remains open. This checkpoint covers
implementation and local verification. No release publication or Deck deployment
was requested.
