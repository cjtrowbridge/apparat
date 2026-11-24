---
plan_id: 2026-11-23-23-30-00_prompt-stop-and-keyboard
title: Preserve stopped prompt dictation and open Android typing
summary: Retain visible dictated words on stop and give manual prompt entry an Android keyboard.
status: past
created_at: 2026-11-23-23-30-00
---

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Phase 9B and Phase 16. On 0.1.83 the Pixel reported mInputShown=false after tapping the prompt. An empty final result calls discardInline and inserts nothing. Preserve the last visible words immediately on stop, permit a final correction, and preserve deliberate cancellation before stop.

- [x] Fix stop/final/interrupt behavior and regression-test selection, late corrections and retained text.
- [x] Open the native multiline editor on deliberate Android typing entry. Physical verification accompanies canonical delivery. Voice entry should continue recording unobstructed.
