---
plan_id: 2026-11-17-14-15-00_platform-update-card
title: Match the Settings update card to the client platform
summary: Show desktop manual update instructions and expose the Android installer action only through its registered bridge.
status: past
created_at: 2026-11-17-14-15-00
---

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Phase 4 Settings and Phase 5 Android. Verify platform copy, the absence of a desktop update action and the existing bridge callback. Qualify all possible targets through the canonical build.

- [x] Verified HUD and GUI tests and the canonical Linux GUI, Linux daemon and Android GUI build qualification.
- [x] Opened Settings in the Linux GUI and on the Pixel 4a. Linux displayed manual replacement instructions and Android displayed the native update button.
