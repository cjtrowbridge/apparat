---
plan_id: 2026-11-17-20-34-03_local-project-repos
title: Open local checkouts and clone GitHub repos on this device
summary: Open local checkouts and clone to a chosen location on Linux and Android, with private GitHub access on Linux.
status: current
created_at: 2026-11-17-20-34-03
---

# Phase 9 · Local Project Repos

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

Project Repos opens real working trees offline on a single device. Both platforms support public clones into a default or selected location. Linux also supports private clones with its system GitHub CLI. Both paths use the same registration. Phase 9 proves checkout acquisition and registry recovery. [Phase 9A](../../plans/current/2026-11-18-10-27-42_local-project-workspaces.md) delivers local workspace use before Phase 10 enrollment. Phase 11 adds remote listings, 11A repeats the local workflows in that shared catalog, and 11C adds remote workspace operations. [Phase 11B](../../ROADMAP.md#phase-11b-android-github-authentication-and-private-cloning) owns Android sign-in and authenticated private clones. Validation used cjtrowbridge/vibe-modeling as the public fixture and cjtrowbridge/apparat as the private Linux fixture. [Roadmap](../../ROADMAP.md#phase-9-local-project-repos) · [Product contract](../../README.md#projects)

## Work and acceptance

- [x] **Open existing repo.** Replace mock rows with an initially empty Projects list. Validate the selected working tree in place. In SQLite persist a stable project ID independent of path, local owner, reopening information and optional account reference. Read Git state and action configuration from the checkout. Preserve existing edits, blue/purple outlines and the 🧠 marker. Exercise duplicate registration, move, lost access, reconnect and unregister. Unregistering preserves files and stays effective after restart.
- [x] **Choose the clone destination.** Derive `<node-root>/data/repos/<owner>/<repo>` from the same runtime root as SQLite. Let the user choose another parent and preview the exact resulting destination and, for an authenticated Linux clone, the account before confirming. Offer Open existing repo at an existing checkout, reject other occupied destinations, and register only a complete validated clone. Test cancellation, concurrent attempts at the same destination, interrupted transfer and retry without overwriting existing files.
- [x] **Prove folder selection on each platform.** The inspected EbitenUI and DebugUI revisions supply UI controls but no filesystem chooser. Keep the form in EbitenUI. On Linux use the desktop portal where available and support path entry. On Android use the system tree picker and persist the URI grant. On the Pixel 4a separately prove reading a selected external checkout including Git metadata and outside edits after restart, and cloning into a selected external parent. Describe supported providers and offer the app-private default if the selected provider lacks required Git capabilities. Never silently copy an existing working tree.
- [?] **Use per-repo accounts on Linux.** Offer a device default and an editable per-project override. Separate the repository owner from the authenticated account. On Linux use the system PATH GitHub CLI, reuse its sign-in, and scope the chosen identity to the actual Git transfer. Preserve the terminal’s active account. Persist only host and account references in project metadata. On missing CLI, expiry, revocation or denied access, identify the remedy and retain offline access to registered repos. Keep credentials out of arguments, remote URLs, SQLite and logs.

- [x] Show the full checkout location beneath each project name in smaller muted text. Verify narrow-pane wrapping and the existing action marker.

- [x] Reduce the Project Repos selector inset and display the branch and checked-out revision after the location, including detached or unavailable checkouts.

- [x] On Android, abbreviate the runtime data directory as `apparat` and the primary shared-storage volume as `~` in project list labels. Preserve other volumes and registration locations.

## Evidence and stopping condition

Use the existing internal commands/queries, SQLite migrations and Projects UI. Keep platform storage and authentication at their respective boundaries. Test migration with existing app data. Exercise Linux and the Pixel 4a through the canonical no-flag build pipeline. For each platform verify existing checkout, default clone and custom-location clone, then reopen offline after restart with the same ID. Change action configuration and confirm the marker follows it. On Linux exercise two accounts with disjoint private repos and verify a wrong account fails visibly. Run overlapping Linux clones with different accounts and destinations and verify the active CLI identity before and after. On Android use a public repo and test lost picker permission and folder reselection. Record evidence and remaining platform limits before closing Phase 9.

## Validation record

The Linux portal registered the existing public checkout. Linux integration tests cloned the public and private fixtures concurrently, reopened their registrations and preserved the active system CLI account. Local tests cover cancelled transfers, occupied destinations, conflicting publication and concurrent per-user credential lookup. The Pixel 4a reopened three distinct registrations offline after a cold start and read an external task-label edit. Moving the selected external checkout made it unavailable until folder reselection. Explicit revocation of an Android system grant remains untested. Live concurrent private clones with two accounts and disjoint access remain open pending a second login and private fixture.

- [x] Create a named local project offline on main under the default or selected parent. Protect existing folders, register the empty checkout, and verify reopening and the first commit.

Verified in 0.1.34. The packaged Linux GUI created and opened a named repo. The Pixel 4a created and reopened repos in both app-private and granted Documents storage with HEAD pointing to main. Runtime tests checked the first commit, duplicate creation, invalid names and reopening from SQLite.

### New Project host selection (Phase 9 with a Phase 11 preview)

Consolidate the three pinned entry points into a progressive host, source and template form. Preserve local creation, clone preview, optional system GitHub login and folder registration. Display safe operation progress and open the registered workspace by its canonical location. Verify search, conditional controls, offline template and remote execution boundaries, and the ordinary release matrix. The default Agentic Pipelines template awaits its public repository.

Use a consistent left column for bold project form labels and italic text for descriptions. Verify field and label bounds on narrow and wide layouts for native prompt controls and inline editing.
