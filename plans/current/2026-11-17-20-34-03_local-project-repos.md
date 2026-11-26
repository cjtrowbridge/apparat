---
plan_id: 2026-11-17-20-34-03_local-project-repos
title: Register local projects and clone GitHub repositories on confirmed workers
summary: Register local working trees, choose clone locations and use a confirmed worker’s provider accounts to clone and inspect a remote checkout.
status: current
created_at: 2026-11-17-20-34-03
---

# Phase 9 · Local Project Repos

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

Project Repos opens real working trees offline on a single device. Both platforms support public clones into a default or selected location. Linux also supports private clones with its system GitHub CLI. Both paths use the same registration. Phase 9 proves checkout acquisition and registry recovery. [Phase 9A](../../plans/current/2026-11-18-10-27-42_local-project-workspaces.md) delivers local workspace use before Phase 10 enrollment. Confirmed workers already return live admin-only checkout summaries. Phase 11 generalizes remote listings, 11A repeats the local workflows in that shared catalog, and 11C adds remote workspace operations. [Phase 11B](../../ROADMAP.md#phase-11b-android-github-authentication-and-private-cloning) owns Android sign-in and authenticated private clones. Validation used cjtrowbridge/vibe-modeling as the public fixture and cjtrowbridge/apparat as the private Linux fixture. [Roadmap](../../ROADMAP.md#phase-9-local-project-repos) · [Product contract](../../README.md#projects)

## Work and acceptance

- [x] **Open existing repo.** Replace mock rows with an initially empty Projects list. Validate the selected working tree in place. In SQLite persist a stable project ID independent of path, local owner, reopening information and optional account reference. Read Git state and action configuration from the checkout. Preserve existing edits, blue/purple outlines and the 🧠 marker. Exercise duplicate registration, move, lost access, reconnect and unregister. Unregistering preserves files and stays effective after restart.
- [x] **Choose the clone destination.** Derive `<node-root>/data/repos/<owner>/<repo>` from the same runtime root as SQLite. Let the user choose another parent and preview the exact resulting destination and, for an authenticated Linux clone, the account before confirming. Offer Open existing repo at an existing checkout, reject other occupied destinations, and register only a complete validated clone. Test cancellation, concurrent attempts at the same destination, interrupted transfer and retry without overwriting existing files.
- [x] **Prove folder selection on each platform.** The inspected EbitenUI and DebugUI revisions supply UI controls but no filesystem chooser. Keep the form in EbitenUI. On Linux use Zenity with portal preference when available and support path entry. On Android use the system tree picker and persist the URI grant. On the Pixel 4a separately prove reading a selected external checkout including Git metadata and outside edits after restart, and cloning into a selected external parent. Describe supported providers and offer the app-private default if the selected provider lacks required Git capabilities. Never silently copy an existing working tree.
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

End visible field and choice-group labels with a colon. Pad each field row above and below by half the selected font size. Verify spacing and punctuation in the rendered layout.

Right-align field labels against their controls. Keep device search compact and single-line for inline entry, pastes and Android's native search prompt, while retaining multiline editing elsewhere.

- [x] Focus device search on New Project entry and advance to the next enabled group after a selection. A nonempty query with one matching host selects it. Verify keyboard and controller traversal, empty and ambiguous searches, and Android prompt behavior. Record the shared form contract in the HUD playbook.

### Host credential providers and worker clones (Phases 9, 10 and 11)
- [x] **Chooser and installation.** Start with Use no credentials, fetch capabilities and usernames for the selected host, and expose fixed install or add-account actions where eligible. A failed probe offers refresh. Keep local clone preview and folder choice. Remote init, import and the unpublished starter remain previews.
- [x] **Trust and job ownership.** Require the installed worker’s locally confirmed owner, a pinned TLS 1.3 peer, a device-bound session and a current admin grant. Recheck authority during jobs. Test anonymous and wrong-account denial, unauthorized requests, another device attempting to poll, account-specific transfer and idempotent retry.
- [x] **Provider and clone behavior.** Exercise fixed installer selection and already-installed behavior, staged device authorization, secure storage failure and cancellation/revocation paths with fixtures. Keep tokens on the host. Register only a completed clone under the worker’s default repo root. Refresh the list and select the host-scoped summary.
- [x] **Installed acceptance.** In the Linux GUI an authorized admin selected the worker’s existing account and cloned the designated private repository. Git HEAD on the worker matched the revision in the selected summary. A separate SQLite read confirmed the registered name and chosen account. Fixed and regression-tested Clone visibility while typing. The 0.1.72 pipeline installed/tested the worker and Pixel 4a and verified the Linux GUI, worker and Android GUI artifacts from GitHub main at `5e7dfb2`. [Journal](../../journal/2026-11-23.md) · [API contract](../../docs/api.md#worker-provider-operations-and-project-registration)
- [?] **Hardware coverage.** The worker lacked an unlocked default Secret Service collection. On an eligible host, add an account using the actual GitHub approval page, verify the new chooser entry and previous active CLI account, then clone using the new account. The real private-clone initiator was Linux. Pixel release acceptance covered install, launch and Projects. Qualify other worker operating systems separately.

The broader Phase 9 plan stays open for its real two-account test and Android grant-revocation check. Phase 11 owns signed, cached summaries and offline reconciliation. Phase 11A repeats local workspace acceptance with mixed owners. Phase 11B owns credentials for cloning onto Android itself and Phase 11C owns remote workspace operations.

### Worker clone credential guidance

For the Phase 9 host credential chooser, keep public access and saved-account selection independent of remote enrollment. Scope enrollment diagnostics to users asking about a new account, explain the worker's observed blocker, and preserve compatibility with workers returning the previous combined message. Verify the locked-keyring chooser and canonical delivery.

- [x] Verify public and saved-account choices remain available with a locked keyring and the reason appears only in enrollment help. The chooser tests covered 390- and 1280-pixel layouts and older worker replies. On trashcan0, the worker adapter detected the locked default keyring and cloned cjtrowbridge/trashcan anonymously and as the saved cjtrowbridge account into separate temporary directories.
