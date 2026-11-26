---
plan_id: 2026-11-25-18-40-37_clipboard-selection
title: Exchange field text and selected chat passages with the system clipboard
summary: Restore ordinary editing shortcuts and add pointer and touch selection to shared chat bubbles.
status: past
created_at: 2026-11-25-18-40-37
---

# Clipboard and chat selection

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

The user approved implementation and Pixel 4a testing. README's focused editor contract and ROADMAP's “Preserve ordinary text editing and clipboard behavior while text controls own focus” govern this repair. Keep terminal, IME, dictation, full-message copy and controller navigation functional.

- [x] Restore exact selection editing and real clipboard transfer through the GUI and platform bridges.
- [x] Add formatted message selection for both chats, with drag, long press, keyboard and touch copy.
- [x] Verify Unicode, repeated substrings, failures, focus, wrapping and gesture ownership, qualify through scripts/build.py and exercise the changes on Pixel 4a.

The baseline ordinary widget lacks clipboard commands and the renderer only supports full-message copy. Finish on evidence of clipboard exchange and qualified publication.

Apparat 0.1.117 passed the canonical build and authenticated GitHub artifact read-back at 8c578bcd9e01ff0430ae45107e8413699cb5a6b3. The Pixel 4a install, signer, package, byte comparison and launch passed. On the device, Android Settings and the global composer exchanged a synthetic string in both directions. The project composer and native search field accepted system clipboard text. A long press in a formatted test reply copied café. Extending the selection across a visual line break retained the initial word, style-free text, spaces and a single logical line. Android Settings exposed the expected exact UTF-8 value. The temporary project transcript was removed.

An independent GTK clipboard client on isolated X11 exchanged Unicode text with the desktop New Project search field. Selecting and cutting the second occurrence of a repeated word preserved the first. The full GUI tests passed, including global bubble actions at phone and desktop sizes and stale-selection coverage. The Windows GUI test binary cross-compiled. Windows and native Wayland interactive clipboard transfers were not exercised.
