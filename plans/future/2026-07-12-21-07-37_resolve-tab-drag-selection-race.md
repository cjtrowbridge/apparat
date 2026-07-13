---
plan_id: 2026-07-12-21-07-37_resolve-tab-drag-selection-race
title: Resolve Tab Drag Selection Race
summary: Prevent tab-strip drag gestures from selecting or highlighting a release-target tab by combining persistent gesture arbitration with EbitenUI radio-group exclusivity.
status: future
created_at: 2026-07-12-21-07-37
---

# Resolve Tab Drag Selection Race

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 HUD tab shell expectations and Phase 5 Android GUI parity validation.
- Product contract: preserve the seven canonical HUD tabs, allow horizontal touch dragging on narrow screens, treat a drag as scrolling rather than tab activation, and display exactly one selected tab at all times.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Parent recovery plan: `plans/current/2026-07-11-09-33-26_fix-mobile-overflow-and-tab-scroll.md`.
- User validation: on 2026-07-12, the user confirmed that the mobile HUD is much better but dragging across tab buttons can still leave both the active tab and the drag-release tab highlighted.
- Confirmed mechanism: EbitenUI recognizes a click when a pointer press and release occur inside a button without canceling for scroll-distance movement. Button click and state-change events are deferred, while Apparat currently clears `tabStripDragMoved` immediately after `game.ui.Update()`. The deferred toggle therefore runs after suppression and immediate state reconciliation have ended.

## Checklist

- [ ] 1. Establish persistent tap-versus-drag gesture arbitration.
  - [ ] 1.1 Replace release-frame-only suppression with a gesture generation or equivalent state that survives until EbitenUI has processed deferred button events.
  - [ ] 1.2 Record whether each completed pointer or touch gesture exceeded the tab drag threshold.
  - [ ] 1.3 Reject tab activation associated with a completed drag gesture while preserving ordinary tap activation.
  - [ ] 1.4 Clear completed gesture state only after deferred click and state-change processing and final tab-state reconciliation.
  - [ ] 1.5 Apply identical tap-versus-drag semantics to mouse and touch input paths.

- [ ] 2. Enforce single-selection behavior through EbitenUI `RadioGroup`.
  - [ ] 2.1 Construct one `widget.RadioGroup` from the tab buttons with the HUD snapshot's active tab as its initial element.
  - [ ] 2.2 Route accepted radio-group changes through the existing HUD tab-selection command path.
  - [ ] 2.3 Restore the radio group to the HUD snapshot's active tab when a deferred change belongs to a rejected drag gesture.
  - [ ] 2.4 Prevent radio-group callbacks and programmatic state reconciliation from recursively rebuilding or selecting tabs.
  - [ ] 2.5 Keep the HUD snapshot authoritative after pointer, touch, keyboard, controller, rebuild, and resize paths.

- [ ] 3. Add regression coverage for EbitenUI event ordering and selection invariants.
  - [ ] 3.1 Add a test that models press, threshold-crossing drag, release over another tab, deferred event processing, and final reconciliation.
  - [ ] 3.2 Prove a drag changes horizontal scroll position without changing the active tab.
  - [ ] 3.3 Prove a tap selects its tab and leaves exactly one tab checked.
  - [ ] 3.4 Prove repeated drags and taps cannot leave two checked tab buttons.
  - [ ] 3.5 Prove keyboard and controller tab changes still select and reveal exactly one tab.

- [ ] 4. Update local GUI documentation for the durable interaction rule.
  - [ ] 4.1 Document EbitenUI's deferred button-event ordering and the required gesture-lifetime boundary in `internal/adapters/gui/README.md`.
  - [ ] 4.2 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` if its tab-strip guidance does not explicitly require tap-versus-drag arbitration through deferred event completion.
  - [ ] 4.3 Keep the parent recovery plan and exact user-authored `TODO.md` items aligned with the validation outcome without rewriting TODO text.

- [ ] 5. Verify the corrective implementation.
  - [ ] 5.1 Run focused GUI compile and regression tests using the documented GUI-tag procedure.
  - [ ] 5.2 Run the repository test, documentation, code-size, and diff checks required by the parent recovery plan.
  - [ ] 5.3 Rebuild and install the Android APK on the connected phone.
  - [ ] 5.4 Capture phone evidence showing that tapping selects a tab, dragging scrolls without selecting the release target, and exactly one tab remains highlighted.
  - [ ] 5.5 Mark the related TODO and parent-plan validation items complete only after phone evidence demonstrates the intended behavior.

- [ ] 6. Publish the verified checkpoint.
  - [ ] 6.1 Update and archive this plan and its parent recovery plan according to their verified outcomes.
  - [ ] 6.2 Regenerate and validate affected plan indexes.
  - [ ] 6.3 Update the current-day journal with implementation and validation evidence.
  - [ ] 6.4 Confirm no files beneath `third_party/salvagecore/` are staged.
  - [ ] 6.5 Review pending downtime reports before the final summary.
  - [ ] 6.6 Commit and push after the user-approved checkpoint summary.

## Scope Boundaries

- Do not patch or fork EbitenUI unless implementation evidence shows that persistent gesture arbitration plus `RadioGroup` cannot enforce the required behavior through public APIs.
- Do not replace the controller-first tab action model or the horizontally scrolling narrow-screen tab strip.
- Do not mark the three in-progress user-authored TODO items complete before on-device visual validation.
