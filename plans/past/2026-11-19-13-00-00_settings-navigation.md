---
plan_id: 2026-11-19-13-00-00_settings-navigation
title: Settings and navigation usability
summary: Verify portrait and startup defaults, bounded scrolling, release discovery and speech model state.
status: past
created_at: 2026-11-19-13-00-00
---

# Settings and navigation

Implements the five approved inbox items under the HUD, Settings and Phase 9B milestones. Preserve saved preferences, private Git delivery and explicit installation. Verify defaults, equal-distance gestures and bounds, Android version discovery and model activation states through tests and canonical Linux and Pixel qualification.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] Implement and verify the five inbox outcomes and deliver the release.

## Verification

The canonical pipeline delivered 0.1.35 (Android version code 36) at `ddce52d`. Linux GUI, Linux worker and Android arm64 artifacts passed qualification and GitHub SHA-256 readback. The GUI/native speech suite and make verify passed. GitHub reported PUBLIC after the push. The operator subsequently confirmed public distribution for future deliveries.

The Pixel 4a retained its previous Top and Comrades preferences on upgrade. Responsive portrait, landscape rotation, Settings startup, and return to the original preferences were exercised through the UI. A 200-physical-pixel drag moved the measured Settings heading 194 pixels. The fresh Linux portrait Projects viewport remained pixel-identical after scrolling both directions.

Android Settings read release metadata automatically. Controlled responses covered equal and newer versions, inaccessible or invalid metadata, offline failure, re-entry caching, and explicit download. On the Pixel the multilingual Tiny confirmation was cancelled, reopened, confirmed and downloaded. Restart displayed selected/unloaded. PTT inserted the known phrase into a disposable draft. Original English Tiny selection and media volume were restored. A simulated failed selection write retained a usable previous native decoder.
