---
plan_id: 2026-11-24-00-01-00_inline-prompt-keyboard
title: Type directly in the global Android prompt
summary: Use the canvas editor's IME connection for prompt entry and retain the draft through voice and keyboard transitions.
status: current
created_at: 2026-11-24-00-01-00
---

Phase 9B / Phase 16. The operator rejected opening a separate Android dialog for the global prompt. Use the retained multiline surface with the existing Ebitengine Android IME driver. Manual focus or tap starts inline editing. Voice-only entry keeps the keyboard dismissed. Verify keyboard entry, dismissal and draft retention on the Pixel when the canonical build is available.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] Replace the modal with inline keyboard editing and exercise the HUD regressions.
- [ ] Verify typing and draft retention in the activity window on the designated Pixel.
