---
plan_id: 2026-11-18-10-27-42_local-project-workspaces
title: Use a local project offline
summary: Edit a real checkout, review and commit changes, and run a supported local task through Projects.
status: current
created_at: 2026-11-18-10-27-42
---

# Use a local project offline

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

Phase 9 supplies registered checkouts. Phase 9A turns a selected checkout into an offline workspace before enrollment. Use the existing Projects selector and stable ID, application commands and queries, SQLite, Git adapter, and Android document-tree boundary. Keep filesystem access on the owning device. Preserve Phase 9 registration and reconnect behavior. [Roadmap and full scope](../../ROADMAP.md#phase-9a-local-project-workspaces)

## Work and acceptance

- [x] Open and edit supported files through Projects. Validate scope, symlinks, races and size limits. Prove edits in place on Linux and on the Pixel in both app-private and picker-granted working trees. A provider capability failure identifies the affected operation and preserves the original files.
- [x] Show working-tree status, diff, staging, scoped commit, branches, history and conflicts. Use two registrations of the same remote at the same HEAD with different uncommitted changes. Compare files, index and history in both checkouts before and after operating on one.
- [x] Recover drafts, project chat history and pending messages after restart. Record a local transaction against a base revision, change the file externally, then recover and revise the conflicted draft. Retry a successful transaction and verify a single change.
- [x] Present enabled, versioned local Task entrypoints with their constraints, inputs and outputs. Run a supported entrypoint through the GUI with zero triggers, inspect its result and verified artifact after restart, and exercise failure or cancellation. Explain unsupported discovered task labels in place. Keep later queue, inference routing and trigger work in its assigned phases.

## Evidence and stopping condition

Use the canonical all-target pipeline and disposable checkouts. On Linux and Pixel 4a, disconnect networking, open a registered repo, edit, diff, stage, commit, restart and find the commit in history. Repeat the action-to-result test and record fixture revisions, platform, storage type, visible results and independent Git or digest verification. Exercise the controller workflow on Steam Deck before marking its device criterion complete. Keep a repeatable copy of the scenarios for Phase 11A. Record untested device or provider cases explicitly.

- [x] Exercise a real editor selection and scrolled detail pane across window resizes, including the single-pane breakpoint. Preserve the active file, draft and conflict report.
- [?] Physical Steam Deck controller acceptance is outstanding. The device workflow and independent checkout evidence are recorded in the [platform matrix](../../docs/platform-matrix.md#phase-9a-offline-local-workspace).

## Workspace usability pass

- [x] Implement a multiline document editor and keep storage scope, durable drafts and conflict checks.
- [x] Make repository selection open Files directly, with Changes, Tasks, Chat and contextual details.
- [x] Replace command forms with changed file, commit, branch, run, artifact and conversation views.
- [x] Verify actual editing, scoped commits, Task output recovery, pending chat and adaptive navigation.
- [?] Observe a first-time tester fixing a typo and committing only that file on the Pixel, without navigation instructions.

- [ ] Tighten editor margins. Show one directory at a time as plain filename rows with white clean files and red U/C/D changes, including missing tracked files. Refresh the containing directory immediately after Save.
