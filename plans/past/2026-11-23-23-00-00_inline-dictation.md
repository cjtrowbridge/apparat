---
plan_id: 2026-11-23-23-00-00_inline-dictation
title: Inline dictation and multiline prompts
summary: Stack multiline labels, replace live recognition inside editors and latch global PTT until dragged home.
status: past
created_at: 2026-11-23-23-00-00
---

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The requested HUD correction follows roadmap Phases 9B and 16. Reuse the existing editable document surface, speech capture and directional lock lifecycle. Keep ordinary command capture, cancel, native editor recovery and user-controlled submission.

- [x] Full-width multiline fields with labels above.
- [x] Revise speech in the editor with rollback and one final edit, including Android.
- [x] Keep global capture locked at the left until the return drag. Verify gestures, inline editing and layout using focused, rendered and device checks.

The rendered 320 × 740 and 1100 × 740 screens show a full-width multiline field, in-place speech and a green left lock. HUD regression coverage exercises revised and final recognition, selection restoration, undo and slide-back. Screenshots are retained under .tmp/inline-dictation. The user approved publishing with the concurrent provider and CLI edits. The canonical release records device and artifact qualification in releases/verified.json.
