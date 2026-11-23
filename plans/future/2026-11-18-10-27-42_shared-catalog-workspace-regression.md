---
plan_id: 2026-11-18-10-27-42_shared-catalog-workspace-regression
title: Verify local workspaces in the shared catalog
summary: Repeat the local workspace workflows with duplicate names, stale owners and changing catalog permissions.
status: future
created_at: 2026-11-18-10-27-42
---

# Verify local workspaces in the shared catalog

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

Phase 11A verifies the shipped Phase 9A workspace against the Phase 11 catalog. Use Phase 10 enrollment and two owners with distinct project IDs. Exercise ordinary GUI controls and compare checkout files, indexes, commits, run records and artifacts on their respective owners. [Roadmap](../../ROADMAP.md#phase-11a-verify-local-workspaces-in-the-shared-catalog)

The 0.1.72 worker-clone flow is an existing regression case. Start with a live admin-visible worker checkout sharing a name with a local checkout. Verify the hostname, branch and revision, select its summary, and confirm that local edit, commit and task controls target only the local checkout. Refresh with the worker disconnected. Today a live-only row can disappear. Once the signed summary cache ships, the offline row should instead show its last observation and unavailable operations. The full Phase 11A exit still requires the two-owner and offline tests below.

## Work and acceptance

- [ ] Give Linux and Pixel checkouts of the same remote identical display names. Select local and remote rows, refresh and reorder the list, and restart. Verify owner, selection, draft and command destination by ID.
- [ ] Repeat the 9A offline edit-to-commit and task-to-artifact scenarios on each local checkout while the peer is online, then disconnected. Recover a draft and check the peer checkout and records independently for unintended changes.
- [ ] Expire a summary, revoke access, reconnect and reconcile. Verify freshness, visibility and unavailable controls. A selected remote row must never redirect an operation to a same-named local checkout. Discovery alone grants no mutation authority.

## Evidence and stopping condition

Record the two owner/project IDs, initial and final revisions and artifact digests, and observed GUI states. Compare with the Phase 9A baseline using the canonical published artifacts on Linux and Pixel 4a. Resolve regressions or record the failing condition as open. Phase 11C owns implementation and acceptance of remote workspace commands.
