---
plan_id: 2026-11-24-03-00-00_chat-model-choice
title: Choose models and synchronize chat defaults
summary: Attach model identity to chat sessions and turns, build the scoped picker, and exchange durable defaults through authenticated workers.
status: past
created_at: 2026-11-24-03-00-00
---

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] Phase 9A and phases 11–14: model inventory, selection, scope revisions, cluster reconciliation and packet binding.
- [x] Phase 9A and 16: shared text-only model chooser, stable composer and session return.
- [x] Test storage, scope isolation, inventory churn, remote permissions, rendered controls.

Global defaults converge by revision and operation ID. Members can relay signed edits. Workers independently check the author’s current administrator role and revocation status. Project reads and writes use the project owner. An offline owner retains its last durable settings until an authenticated client reconnects.
