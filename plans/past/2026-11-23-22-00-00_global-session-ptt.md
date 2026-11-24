---
plan_id: 2026-11-23-22-00-00_global-session-ptt
title: Global session prompt and PTT drag controls
summary: Add the global draft tab and pointer gestures for locking and prompt dictation.
status: past
created_at: 2026-11-23-22-00-00
---

# Global session prompt and PTT drag controls

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The user's request authorizes Phase 9B and Phase 16 HUD work. Use the existing tab catalog, form field and local utterance lifecycle. Preserve editor ownership except on explicit global handoff. Sending, agent execution, TTS and automatic listening remain future work.

- [x] Add the discoverable brain tab and editable session draft after STMPCK.
- [x] Show PTT throughout the HUD and route upper lock, return and left dictation through shared pointer capture.
- [x] Verify tab order, rendered layouts, gestures, cancellation, action routing and governance.

The user also approved including existing checkout edits in the canonical build and publishing to origin/main. The HUD and GUI race suites passed with the new tab and gesture regressions. Rendered the empty field and drag targets at 320 × 740 and the populated field with a side rail at 1100 × 740. The lock target clears the prompt. Screenshots are in the ignored .tmp/global-session-review directory. The Android native editor compiles with the repository’s OpenJDK 21 toolchain. The canonical build records artifact and Pixel qualification in releases/verified.json.
