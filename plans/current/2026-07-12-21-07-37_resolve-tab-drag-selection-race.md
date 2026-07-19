---
plan_id: 2026-07-12-21-07-37_resolve-tab-drag-selection-race
title: Resolve Tab Drag Selection Race
summary: Prevent tab-strip drag gestures from selecting or highlighting a release-target tab by combining persistent gesture arbitration with EbitenUI radio-group exclusivity.
status: current
created_at: 2026-07-12-21-07-37
---

# Resolve Tab Drag Selection Race

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 HUD tab shell expectations and Phase 5 Android GUI parity validation.
- Product contract: preserve the six canonical HUD tabs, left-align the strip whenever all tabs fit, preserve user-driven drag position only when tabs overflow, minimally reveal an off-screen selected tab without centering it, treat a drag as scrolling rather than tab activation, and display exactly one selected tab at all times.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Parent recovery plan: `plans/current/2026-07-11-09-33-26_fix-mobile-overflow-and-tab-scroll.md`.
- User validation: on 2026-07-12, the user confirmed that the mobile HUD is much better but dragging across tab buttons can still leave both the active tab and the drag-release tab highlighted.
- Confirmed mechanism: EbitenUI recognizes a click when a pointer press and release occur inside a button without canceling for scroll-distance movement. Button click and state-change events are deferred, while Apparat currently clears `tabStripDragMoved` immediately after `game.ui.Update()`. The deferred toggle therefore runs after suppression and immediate state reconciliation have ended.

## Checklist

- [x] 1. Establish persistent tap-versus-drag gesture arbitration.
  - [x] 1.1 Replace release-frame-only suppression with a gesture generation or equivalent state that survives until EbitenUI has processed deferred button events.
  - [x] 1.2 Record whether each completed pointer or touch gesture exceeded the tab drag threshold.
  - [x] 1.3 Reject tab activation associated with a completed drag gesture while preserving ordinary tap activation.
  - [x] 1.4 Clear completed gesture state only after deferred click and state-change processing and final tab-state reconciliation.
  - [x] 1.5 Apply identical tap-versus-drag semantics to mouse and touch input paths.

- [x] 2. Enforce single-selection behavior through EbitenUI `RadioGroup`.
  - [x] 2.1 Construct one `widget.RadioGroup` from the tab buttons with the HUD snapshot's active tab as its initial element.
  - [x] 2.2 Route accepted radio-group changes through the existing HUD tab-selection command path.
  - [x] 2.3 Restore the radio group to the HUD snapshot's active tab when a deferred change belongs to a rejected drag gesture.
  - [x] 2.4 Prevent radio-group callbacks and programmatic state reconciliation from recursively rebuilding or selecting tabs.
  - [x] 2.5 Keep the HUD snapshot authoritative after pointer, touch, keyboard, controller, rebuild, and resize paths.

- [x] 3. Add regression coverage for EbitenUI event ordering and selection invariants.
  - [x] 3.1 Add a test that models press, threshold-crossing drag, release over another tab, deferred event processing, and final reconciliation.
  - [x] 3.2 Prove a drag changes horizontal scroll position without changing the active tab.
  - [x] 3.3 Prove a tap selects its tab and leaves exactly one tab checked.
  - [x] 3.4 Prove repeated drags and taps cannot leave two checked tab buttons.
  - [x] 3.5 Prove keyboard and controller tab changes still select and reveal exactly one tab.

- [x] 4. Update local GUI documentation for the durable interaction rule.
  - [x] 4.1 Document EbitenUI's deferred button-event ordering and the required gesture-lifetime boundary in `internal/adapters/gui/README.md`.
  - [x] 4.2 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` if its tab-strip guidance does not explicitly require tap-versus-drag arbitration through deferred event completion.
  - [x] 4.3 Keep the parent recovery plan and exact user-authored `TODO.md` items aligned with the validation outcome without rewriting TODO text.

- [?] 5. Verify the corrective implementation.
  - [x] 5.1 Run focused GUI compile and regression tests using the documented GUI-tag procedure.
  - [?] 5.2 Run the repository test, documentation, code-size, and diff checks required by the parent recovery plan.
  - [?] 5.3 Rebuild and install the Android APK on the connected phone.
  - [ ] 5.4 Capture phone evidence showing that tapping selects a tab, dragging scrolls without selecting the release target, and exactly one tab remains highlighted.
  - [ ] 5.5 Mark the related TODO and parent-plan validation items complete only after phone evidence demonstrates the intended behavior.

- [ ] 6. Publish the verified checkpoint.
  - [ ] 6.1 Update and archive this plan and its parent recovery plan according to their verified outcomes.
  - [ ] 6.2 Regenerate and validate affected plan indexes.
  - [ ] 6.3 Update the current-day journal with implementation and validation evidence.
  - [ ] 6.4 Confirm no files beneath `third_party/salvagecore/` are staged.
  - [ ] 6.5 Review pending downtime reports before the final summary.
  - [ ] 6.6 Commit and push after the user-approved checkpoint summary.

## Approved Plan Revision: 2026-07-18 Tab Strip Position And Visual State

- [x] 7. Repair the tab-strip position and visual-state behaviors found in the Android screenshots.
  - [x] 7.1 Preserve the tab-strip's left-aligned position when all tab buttons fit in the viewport; do not apply normalized active-index centering during ordinary rebuilds.
  - [x] 7.2 When the buttons overflow, preserve the user's drag position across rebuilds and move only enough to make a newly selected, clipped tab completely visible.
  - [x] 7.3 Give an unselected hovered tab a distinct visual state from the selected/pressed tab so a pointer cannot look like a second active selection.
  - [x] 7.4 Keep the HUD snapshot and radio group authoritative so a drag/release cannot create a second checked tab.
- [x] 8. Prove and document the corrected contract.
  - [x] 8.1 Add focused GUI regression tests for left alignment when tabs fit, persisted overflow scroll, minimal selected-tab reveal, and distinct hover/selected appearance.
  - [x] 8.2 Update the local GUI README and HUD playbook with the precise strip-position and visual-state rules.
  - [x] 8.3 Run focused GUI tests, repository tests, documentation/code-size checks, and `git diff --check`.
  - [x] 8.4 Rebuild/install the Android APK and capture phone screenshots showing one selected visual state and the corrected strip position. On Pixel 10 Pro XL, Comrades and Cluster both rendered left-aligned on the wide surface, and only Cluster used selected blue after touch selection.
- [ ] 9. Publish the follow-up checkpoint.
  - [x] 9.1 Update/archival status, indexes, and the append-only journal; confirm no Salvagecore files or pending downtime reports are included.
  - [x] 9.2 Commit and push directly to `main` after the user-approved checkpoint summary; do not open a pull request.

## Scope Boundaries

- Do not patch or fork EbitenUI unless implementation evidence shows that persistent gesture arbitration plus `RadioGroup` cannot enforce the required behavior through public APIs.
- Do not replace the controller-first tab action model or the horizontally scrolling narrow-screen tab strip.
- Do not mark the three in-progress user-authored TODO items complete before on-device visual validation.

## Execution Notes

- 2026-07-12: Implemented persistent post-drag cancellation, a tab-button `RadioGroup`, HUD-snapshot reconciliation, focused GUI regression coverage, and documentation updates. `go test -c -tags gui` and the focused tab-strip suite passed; the ordinary repository suite passed. The complete GUI-tagged suite still has pre-existing Settings/master-detail widget-discovery failures, and `check_directory_docs.py` still reports pre-existing missing `--help` output in `scripts/build.py` and `scripts/build_orchestrator.py`.
- 2026-07-12: `python -m scripts.build` rebuilt feasible Windows GUI and headless artifacts. Android build/install is blocked on this host because ADB and the configured Android SDK/NDK command-line tools are unavailable. The documented direct invocation `python scripts/build.py` currently fails before detection because it cannot import `scripts.android_wrapper`; no build-system change is included in this plan.
