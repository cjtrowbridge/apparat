---
plan_id: 2026-11-16-15-00-00_tab-page-headings
title: Name each icon tab with a page heading
summary: Add a prominent shared page-title row beneath the icon navigation.
status: past
created_at: 2026-11-16-15-00-00
---

# Name each icon tab with a page heading

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

User authorizes a heading containing the former tab-button text. Bind to Phase 8 GUI presentation and Phase 17 platform evidence. Add one non-scrolling 28px heading derived from the active tab title between navigation and content, shared by all tabs and master/detail views. Preserve icon-only navigation, scroll bounds, input and runtime state. Existing EbitenUI grid owns geometry; no separate per-tab implementations or new dependency. H1-type describes visual hierarchy, not native accessibility semantics.

- [x] Add the shared heading and document its placement and ownership.
- [x] Run GUI/HUD and source/docs checks, build/install on Pixel 4a preserving data, inspect all titles and landscape/scroll bounds, review and archive.

Evidence `.tmp/tab-headings/`; stop when names are visible and body/footer remain bounded. Restore tracked release artifacts and phone rotation. No Pixel 10 install or GitHub publication. Updater release/signing issue from the previous diagnosis remains separate and unresolved.

Acceptance: canonical Linux/Android build and native GUI/HUD tests under Xvfb pass, as do source/size/docs and diff checks. All five titles viewed in Pixel 4a portrait screenshots, Settings scrolled with fixed heading and bounded content, and landscape checked. APK SHA-256 `fa08e1d64ac8a8db656b103e5eb9776207a791c59184f3d5616418ed2692e22f` installed with -r preserving data. Rotation restored; tracked releases restored. Review confirms a single shared title row and one font face per Game, no per-frame loading or extra input owner. No remaining delta-specific finding. Existing native accessibility and updater signing/version/publication gaps remain unchanged. Evidence `.tmp/tab-headings/`.
