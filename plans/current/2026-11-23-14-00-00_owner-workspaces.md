---
plan_id: 2026-11-23-14-00-00_owner-workspaces
title: Operate owner-hosted projects and capture workspace performance
summary: Deliver host-qualified project navigation, file-based performance captures, safe owner edits, live subscriptions, tasks and restricted terminals.
status: current
created_at: 2026-11-23-14-00-00
---

# Owner workspaces and performance capture

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

[Phase 11C](../../ROADMAP.md#phase-11c-operate-a-workspace-on-its-owning-device) builds on Phase 9A and Phase 10. Keep local checkouts usable offline, preserve Changes and verified artifacts, and scope all remote operations by stable owner and registration ID. The chat destination reserves space for future remote chat.

- [x] Show hostname-qualified project titles and five emoji destinations, including explicit availability. Exercise keyboard, controller and narrow layouts.
- [?] Measure bounded operations in memory. Export recent NDJSON and record capped diagnostic sessions off the request path. Provide historical inspection, Android export and correlated authorized peer capture. Exercise write failure, saturation and restart.
- [?] Serve paged directories, reads and conditional saves on the owner with operation-specific grants. Preserve private drafts and retry receipts. Exercise concurrent edits, lost acknowledgement and reconnect on Pixel and Linux.
- [?] Share one optional authenticated connection per active client/owner, replay from cursors, handle compaction, backpressure and revocation, and release idle subscriptions.
- [?] Group Apparat tasks, VS Code tasks and launch configurations. Report unsupported execution requirements and preserve verified artifacts. Page history and measure invocation and completion separately.
- [?] Qualify a restricted Linux PTY and VT emulator, screen checkpoints and chunked scrollback, explicit single-controller transfer and independent read/control permissions. Validate interactive and full-screen programs, slow consumers and restart.
- [?] Run the full canonical pipeline and record eligible platform and real-device acceptance. Use disposable working files in an authorized checkout for destructive test cases. Automated performance analysis and tuning remain deferred.

### Verification checkpoint

The race-enabled app and adapter suites passed on Linux, including the real isolated PTY, authenticated live socket, 2,100-entry directory, stale base, retry receipt, grant revocation and private capture failure tests. The current protocol recovers by snapshot on reconnect. Cross-connection event replay, wider catalog scale and hardware acceptance remain separate checks.

The first published checkpoint qualified all three host targets, including Pixel 4a installation and trashcan0 installation. A physical Pixel inspection verified the five workspace icons and identified a hostname fallback when the keys were locked. The follow-up uses the saved installation hostname. The selector pages registrations by owner, and the terminal list streams session directories. Keep acceptance on existing main checkouts. Hardware tests may use a temporary file inside an existing approved checkout.
