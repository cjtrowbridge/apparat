---
plan_id: 2026-11-16-14-30-00_compact-icon-tabs
title: Compact symbol-only HUD tabs
summary: Fit all five tabs on phone screens using the existing Comrades symbol and bundled emoji.
status: past
created_at: 2026-11-16-14-30-00
---

# Compact symbol-only HUD tabs

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and authority

User authorized icon-only tabs, retaining ☭ for Comrades and choosing 🔨 for Projects; other agreed symbols are 🖥️ Cluster, 🧪 Research, ⚙️ Settings. Bind to Phase 8 GUI/input and Phase 17 responsive-platform evidence. Preserve IDs, names/accessibility metadata, actions, focus, selection, overflow handling, and tab bodies. Use existing EbitenUI graphic buttons with bundled Twemoji v16.0.1 PNGs like STMPCK; retain the existing font for Comrades. No OS emoji dependency or runtime download. Pixel 4a is the authorized test device; do not install to Pixel 10 or publish.

## Work and acceptance

- [x] Render five compact equal-width 54x54 buttons with 28px symbols, retain names as tooltips, bundle attributed artwork and document behavior.
- [x] Verify five tabs fit at 320 logical pixels, retain overflow behavior below that size, run GUI/HUD tests and canonical build, inspect/tap all five tabs on Pixel 4a in portrait and landscape, review and archive.

Baseline: title-plus-glyph tabs have >=130px minimum and overflow phone widths. New row needs 294px plus 24px outer margins; five tabs should fit a 320px viewport without reducing touch targets. Asset decode/resize occurs once per icon, cached on the GUI thread for process lifetime (four images). Missing/corrupt embedded assets fail visibly; no network fallback. Native accessibility tree remains a known platform gap despite preserved descriptor labels.

Evidence: `.tmp/emoji-tabs/`; restore pre-task tracked releases. Stop when all five symbols render and select correctly, focus/selection/overflow regressions pass and docs/source checks pass. No commit/push.

## Acceptance and review

Retained. Full governed Linux GUI/headless and Android build passed; 117-source-file governance and 400-line gates pass. Native GUI/HUD tests under Xvfb pass, including 320px fit, touch sizes, artwork/names, 280px overflow, drag arbitration and single-selection/navigation. Docs and diff checks pass. Pixel 4a serial 08051JEC218861 accepted install -r preserving data; final cold launch 679ms (instrumented observation). All five 150ms held taps select their intended tab, observed in final screenshots and exclusive selected-button pixel comparisons. Portrait/landscape render all five icons without scrolling; tooltip displayed in landscape. Rotation settings restored to 0/0. Evidence and final APK `.tmp/emoji-tabs/`; SHA-256 `242ae188f88c52016ac27a5bcf9cb500070485f02f4e9fac33e7fce754fc69ca`.

Review caught inherited empty text/spacing shifting graphic icons right; tab buttons now use their complete explicit style with an isolated empty theme. Final screenshots confirm centered images. STMPCK rendering approach reused without a runtime source dependency; four attributed PNGs bundled, existing font glyph preserved. Labels/IDs remain in HUD descriptors; this does not implement native screen-reader support. Physical controllers not newly tested. No new dependency, network runtime, engine mutation, data clear or publication. Pre-task tracked release artifacts restored. No remaining delta-specific finding.
