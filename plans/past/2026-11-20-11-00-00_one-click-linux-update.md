---
plan_id: 2026-11-20-11-00-00_one-click-linux-update
title: One-click Linux update
summary: Update and restart from a single Settings action with bounded status text.
status: past
created_at: 2026-11-20-11-00-00
---

# One-click Linux update

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Bind the requested Settings change to ROADMAP Phase 17 development updates. Keep the existing verified channel, atomic replacement, backup and graceful restart. Keep Android's system installer flow.

- [x] Make one Linux action check, download, verify, install and restart only on success. Keep progress and retry information below a stable button label.
- [x] Verify one activation, repeated input, already-current and failed checks, layout, and the existing updater regression suite. Update contributor and Deck instructions.

The governed GUI, self-update adapter and HUD race suites passed. A disposable Linux executable was replaced after one Settings click with its previous bytes retained. The GUI requested graceful restart. Current and failed checks kept the process open. A 320-pixel rendered layout verified the button and wrapped detail bounds. Source, size and documentation checks passed.
