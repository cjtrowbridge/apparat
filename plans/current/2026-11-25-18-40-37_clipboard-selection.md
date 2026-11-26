---
plan_id: 2026-11-25-18-40-37_clipboard-selection
title: Exchange field text and selected chat passages with the system clipboard
summary: Restore ordinary editing shortcuts and add pointer and touch selection to shared chat bubbles.
status: current
created_at: 2026-11-25-18-40-37
---

# Clipboard and chat selection

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The user approved implementation and Pixel 4a testing. README's focused editor contract and ROADMAP's “Preserve ordinary text editing and clipboard behavior while text controls own focus” govern this repair. Keep terminal, IME, dictation, full-message copy and controller navigation functional.

- [x] Restore exact selection editing and real clipboard transfer through the GUI and platform bridges.
- [x] Add formatted message selection for both chats, with drag, long press, keyboard and touch copy.
- [ ] Verify Unicode, repeated substrings, failures, focus, wrapping and gesture ownership, qualify through scripts/build.py and exercise the changes on Pixel 4a.

The baseline ordinary widget lacks clipboard commands and the renderer only supports full-message copy. Finish on evidence of clipboard exchange and qualified publication.

The full GUI package passed under Xvfb. Documentation and source gates passed. ADB confirmed the authorized device. Build and on-device feature acceptance are pending.
