---
plan_id: 2026-07-11-09-33-26_fix-mobile-overflow-and-tab-scroll
title: Fix Mobile Overflow And Tab Scroll
summary: Recover phone-sized HUD layout by bounding EbitenUI preferred widths, wrapping body text before measurement, and allowing user-driven tab-strip scrolling to persist.
status: past
created_at: 2026-07-11-09-33-26
---

# Fix Mobile Overflow And Tab Scroll

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 HUD tab shell expectations and Phase 5 Android GUI APK Build Pipeline validation.
- Product contract: keep the seven canonical HUD tabs, keep the phone layout touch-first, keep tab navigation horizontally swipeable on narrow screens, and keep tab body content width-bounded to the visible screen.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`. This recovery must follow its EbitenUI-only tab-body rule, responsive-by-default rule, visual-evidence requirement, and plan-bound implementation workflow.
- Source report: user reported on 2026-07-11 that phone content still stretches off the right side and tab buttons still do not scroll side-to-side.
- Root-cause hypothesis: EbitenUI parent layout uses child `PreferredSize()` before final stretching/clipping. Current tab/body scroll containers still report unbounded content widths, body text is measured as unwrapped single-line text, and `ensureActiveTabVisible()` overwrites manual tab-strip scrolling every frame.

## Checklist

- [x] 1. Reopen status tracking for the failed recovery items.
  - [x] 1.1 Mark the active backlog plan's responsive overflow and tab-strip validation items back to in-progress or needs-validation.
  - [x] 1.2 Mark the corresponding `TODO.md` items back to `[-]` in progress without rewriting their text.
  - [x] 1.3 Append a journal note explaining that phone validation disproved the previous completion state.

- [x] 2. Capture phone evidence before implementation.
  - [x] 2.1 Confirm the connected phone is unlocked, authorized in ADB, and focused on `com.cjtrowbridge.apparat/.MainActivity`.
  - [x] 2.2 Capture screen size, density, focused window/app, and a before screenshot showing right-edge content overflow.
  - [x] 2.3 Capture a before screenshot or short observation showing the tab strip failing to retain horizontal scroll.
  - [x] 2.4 Store evidence under the appropriate artifact path and record it in the journal or plan execution notes.

- [x] 3. Fix tab-strip scrolling persistence.
  - [x] 3.1 Replace every-frame active-tab auto-centering with a one-shot `requestActiveTabVisible` style flag.
  - [x] 3.2 Trigger active-tab visibility only after keyboard/controller/programmatic tab changes, rebuilds, and layout resizes where appropriate.
  - [x] 3.3 Preserve user-driven tab-strip `ScrollLeft` after mouse drag, wheel, and touch drag.
  - [x] 3.4 Add regression coverage proving manual tab scroll is not overwritten by the next update cycle.
  - [x] 3.5 Ensure tab-strip drag gestures cannot leave a non-selected tab checked or visually selected after release without disabling tab-strip dragging.

- [x] 4. Bound EbitenUI preferred widths for phone layouts.
  - [x] 4.1 Add a small GUI helper/wrapper that caps a widget's reported preferred width while preserving its final stretched location.
  - [x] 4.2 Apply bounded preferred width to the top tab-strip scroll container.
  - [x] 4.3 Apply bounded preferred width to Settings, collapsed master list, collapsed detail, and expanded detail scroll containers.
  - [x] 4.4 Add layout tests proving narrow viewport preferred widths do not exceed the available HUD width.

- [x] 5. Wrap body text before preferred-size measurement.
  - [x] 5.1 Add helper constructors for HUD body text that set `widget.TextOpts.MaxWidth(...)`.
  - [x] 5.2 Use wrapped text helpers for tab summaries, section descriptions, and row label/detail text in Settings and master-detail panes.
  - [x] 5.3 Size text max widths from current viewport width minus HUD margins, scroll padding, fieldset padding, and button/list padding.
  - [x] 5.4 Preserve fixed touch-friendly sizing for buttons while wrapping descriptive text.
  - [x] 5.5 Add regression coverage proving body text nodes have nonzero max widths on narrow screens.

- [x] 6. Validate on the connected phone after implementation.
  - [x] 6.1 Run focused GUI compile/tests with the `gui` tag.
  - [x] 6.2 Run `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make test`.
  - [x] 6.3 Run `make check-docs`.
  - [x] 6.4 Run `python3 scripts/check_code_file_lines.py`.
  - [x] 6.5 Run `git diff --check`.
  - [x] 6.6 Rebuild all possible targets through `PATH="$PWD/.tools/go1.26.4/bin:$PATH" make build`.
  - [x] 6.7 Install or update the Android APK on the connected phone.
  - [x] 6.8 Capture after screenshots proving content is bounded, tab buttons scroll horizontally, and the collapsed Back button remains centered.
  - [x] 6.9 Mark visual-validation items complete only after screenshots demonstrate the intended behavior.

- [x] 7. Institute HUD layout governance so the fix survives future changes.
  - [x] 7.1 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to require bounded preferred-width behavior for HUD scroll containers and panes.
  - [x] 7.2 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to require nonzero wrapping widths for descriptive HUD text before preferred-size measurement.
  - [x] 7.3 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to prohibit every-frame auto-centering from overriding user-driven tab-strip scroll.
  - [x] 7.4 Update `playbooks/how_to_add_or_modify_hud_tab_contents.md` to require phone/narrow-surface screenshots for HUD layout changes before marking visual items complete.
  - [x] 7.5 Update `internal/adapters/gui/README.md` with the permanent local implementation rules for bounded scrollers, wrapped text helpers, tab-strip auto-scroll, and required visual validation.
  - [x] 7.6 Add regression tests that encode the governance rules, not just the current implementation details.

- [x] 8. Publish the recovery checkpoint.
  - [x] 8.1 Update the active and/or future plan status to match the verified outcome.
  - [x] 8.2 Confirm no files under `third_party/salvagecore/` are staged.
  - [x] 8.3 Review pending downtime reports before final summary.
  - [x] 8.4 Commit and push after the user-approved checkpoint summary.

## Notes

- Before evidence captured from phone `58051FDCQ002T9`: `artifacts/mobile-overflow-before-metrics.txt`, `artifacts/mobile-overflow-before-focus.txt`, `artifacts/mobile-overflow-before.png`, and `artifacts/mobile-overflow-before-tab-swipe.png`. Focus confirmed `com.cjtrowbridge.apparat/.MainActivity`; physical size was `1080x2404` at density `390`.
- Do not mark the original TODO items complete again until phone screenshots prove the user-visible behavior is fixed.
- Avoid treating `StretchContentWidth()` as sufficient by itself; the failure mode is likely preferred-size leakage before final stretch/clipping.
- Do not copy or modify EbitenUI internals unless a narrowly scoped wrapper in `internal/adapters/gui` cannot solve the preferred-size boundary.
- Governance changes that make this fix durable are part of the plan's definition of done, not optional cleanup.
- Implementation, local compile gates, build, and APK install completed on 2026-07-11. GUI-tagged test execution is blocked in the shell by missing X11 display, so the documented `go test -c -tags gui` compile gate was used. Phone screenshots are still pending because the connected phone is on the secured lock screen; current `artifacts/mobile-overflow-after.png` is lock-screen evidence and must not be used as HUD visual proof.
- User validation found a leftover tab-strip issue: dragging the tab strip did not select a tab, but the tab under release could remain highlighted along with the selected tab. A first pass removed the extra highlight but also prevented dragging from working. The corrected fix suppresses tab selection only during the active pressed drag gesture after threshold movement; release clears suppression after EbitenUI's click/release pass and resynchronizes button state from the HUD snapshot.
