---
plan_id: 2026-11-22-12-00-00_node-identity-and-rest
title: Node setup, key protection and authenticated REST
summary: Deliver persistent node identities, optional OS unlock and an authenticated local API before cluster enrollment.
status: current
created_at: 2026-11-22-12-00-00
---

# Node setup, key protection and authenticated REST

Bound to [Phase 10 local identity and authenticated loopback REST](../../ROADMAP.md#local-identity-and-authenticated-loopback-rest). Preserve the Phase 8 diagnostic projections, shared runtime lock and offline workspace. Implement the [identity design](../../docs/node-identity.md) and [HTTP contract](../../docs/api.md#planned-phase-10-local-authentication-contract) in the existing identity package, core and REST adapter.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Deliver discovery first

- [x] Qualify and atomically install the Linux amd64 worker on trashcan0 through trusted key-only SSH whenever available. Verify its digest, remote smoke run and discovery on the Pixel and local Linux GUI.

Execute the hosting-default and activation-control tasks below first, limited to the isolated anonymous responder. Use an owner-local administration path for headless configuration. Exercise installation upgrades as well as fresh roots. The first deliverable ends at browsing and a JSON authentication-required response.

- [ ] Move the existing add-new-device selector directly below the Cluster Devices heading. Reuse the Projects plus glyph and keep the stable action ID. Populate the right pane with Local network discovery and provide the usual back navigation on narrow layouts. Verify action catalog, voice, keyboard, controller and pointer navigation by ID after reordering.
- [ ] Add a bounded, per-interface DNS-SD browser and publisher. Publish only after successful listener activation. Test cancellation, goodbye/expiry, duplicates, malformed records, link loss, app suspension and blocked multicast. Display candidates without storing trusted devices or enabling an add/join action.
- [ ] Implement the bootstrap HTTP contract independently of the current Phase 8 router. Verify the exact JSON 401 and absence of health, projects and service data on that listener. Probe that no socket or advertisement is reachable on a denied interface before and after grant changes.
- [ ] Exercise Linux GUI and headless hosting, Pixel foreground browsing with hosting off, and reversed roles wherever network scoping is enforceable. Verify layout on a narrow Pixel and desktop/Deck split pane. On Windows validate adapter/profile and firewall behavior, on macOS Bonjour interface scope and permission denial, and on Android validate network-specific NSD and callbacks across Wi-Fi/VPN changes at supported API levels. Report unavailable target checks explicitly. Never use an all-interface fallback.

Linux and Pixel 4a each displayed the deployed worker in the real Cluster view. The wired HTTP check returned the specified JSON 401 and an unknown resource returned JSON 404. The ungranted Docker address refused the connection. Disabling the grant withdrew DNS-SD and closed its old port. Restoring the wired grant republished the service. Roaming, physical controller and additional OS checks remain open.

- [x] Present live candidates as plus-and-hostname buttons, open an unpaired node detail view, and show claim eligibility. Resolve transient destinations at activation and verify narrow and wide layouts. Trusted enrollment depends on identity setup.

## Standalone Linux worker updates

- [x] Reuse the Linux release validator for the worker artifact, with persisted headless defaults and request-driven scheduling including anonymous responses. Exercise success, failure, suppression, manual override, promotion and restart. Preserve desktop update behavior.
- [x] Document the owner commands, quiet-node semantics and subsequent pool queue, drain and rejoin gates in the roadmap.

## Startup key setup and recovery

- [ ] On startup inspect the current node identity. Give returning users a recovery warning, a private-key file chooser and explicit confirmation before generating a replacement. Stage and verify the selected pair, preserve legacy files and keep the same fingerprint on import.
- [ ] Support masked password setup and unlock, capability-gated system credential protection, unattended fresh Linux workers, and local recovery. Verify actual startup, cancellation, restart, bad-password and mismatched-key paths.

## Subsequent identity and enrollment work

- [ ] **Persist and recover a node identity.** Inventory all startup paths and legacy files. Add explicit storage format/version and distinguish missing, locked, usable and inconsistent states. Perform initial writes under the existing node lock. Test absent, partial, mismatched, revoked, invalid-permissions and legacy fixtures, concurrent startup, injected write failure and interrupted migration. Verify fingerprint and a signing challenge before and after each protection change. Keep recovery operator-driven.
- [ ] **Complete setup through real entry points.** Share the setup operation between GUI and daemon. Add first-run choice, masked passphrase confirmation, Settings controls for lock, unlock and conversion, and a local terminal flow for headless setup. Register eligible navigation and operations with the action catalog and verify keyboard, D-pad, pointer and native modal ownership. Keep passwords outside dictation, command history and logs. Exercise fresh roots with each default and opt-in, wrong password, cancellation, GUI-to-daemon restart while encrypted, and daemon-to-GUI restart while unencrypted.
- [ ] **Make inbound serving opt-in before adding any public route.** In shared runtime startup, decouple outbound clients, internal queries and the local administration channel from HTTP startup. Migrate fresh and existing roots to zero grants unless an explicit hosting policy exists. Persist activation and network/interface pairs, implement effective ingress checks and lifecycle cleanup as specified in the hosting policy. Update legacy listen flags, health probes, examples and smoke tests so an address argument alone cannot authorize serving.
- [ ] **Make activation usable locally.** Implement Settings → API configuration with unchecked pairs, an explicit enable action, Save/Cancel, saved versus effective addresses, unavailable-rule reasons and revocation. Provide equivalent Linux daemon CLI operations over the owner-only channel. Register the modal and eligible actions with the action catalog, retain local confirmation, and validate keyboard, D-pad, touch and narrow-screen focus.
- [ ] **Prove isolation from a second client.** On fresh Windows and Android GUI nodes and a Linux service, verify zero listening sockets (including loopback) and successful outbound calls to an enabled peer. Save a loopback-only grant for local authentication tests. For the later network milestone, approve just a tunnel while home Wi-Fi is also connected, then independently grant and revoke home Wi-Fi. Probe the public-key and protected endpoints over each path on IPv4 and IPv6. Test alternate ingress to an approved local address, same adapter on a different network with reused addresses, VPN loss, restart, resume, Cancel, disable and a live stream during revocation. Require denied paths to return no API response and permitted paths to retain authentication. Record unsupported platform enforcement as inactive, with visible status.
- [ ] **Deliver usable caller authentication.** Implement local token issuance through the unlocked owner UI and account-restricted headless channel. Specify and test finite token lifetime, scope, revocation and restart behavior. Reject unauthorized socket/pipe clients. Wire a single authorization policy across the existing handler. Add the anonymous/authenticated root and public-key resource. Update every local caller and existing smoke check in the same checkpoint. Test the documented status/body/content-type matrix, missing and corrupt manifests, malformed/expired tokens and concurrent requests. Inspect diagnostics for secret leakage using synthetic credentials.
- [ ] **Prove the first local milestone on Linux and Android.** On the Pixel 4a, use a Keystore wrapping key with biometric authorization. Test success, cancellation, unavailable authenticator, invalidation and passphrase recovery. On both platforms show anonymous lookup with a locked key, denied protected access, successful authorized query and failed token reuse after restart. Check that HTTP traffic alone never prompts the owner. Measure unlock latency and memory on the Pixel before settling encryption parameters. Exercise Linux headless unattended and locked reboot paths.
- [ ] **Qualify each desktop unlock option.** Implement the platform adapters behind the shared wrapping operation. Test unlocked, locked and missing Secret Service in Linux and Steam Deck sessions, user-scoped DPAPI and restrictive ACLs under different Windows accounts, and macOS Keychain with user presence after app upgrade. Exercise fallback, OS enrollment removal and device-local credential loss. Record which OS/device/session was actually tested and leave unsupported native options hidden.
- [ ] **Verify cluster handoff with Phase 10 enrollment.** Use the same fingerprint in the trusted invitation, bootstrap HTTPS projection and signed device record. Verify on two nodes that substitution or an expired invite fails, that an enrolled peer can access its authorized scope, and that revoked peers and local bearer tokens cannot access protected network routes. Keep this milestone dependent on the roadmap's certificate, trusted directory, revocation and audit work.

For implementation checkpoints, run checkout identity before edits, package and race tests through the build gate, documentation and navigation checks, then the canonical no-flag build. Record eligible artifacts and GitHub readback. Record physical Pixel validation as pending if unavailable. Promote local functionality independently of later cluster work. Close the full plan when each claimed platform and network behavior has acceptance evidence or an explicitly recorded deferral.
