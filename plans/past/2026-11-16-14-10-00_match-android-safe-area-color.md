---
plan_id: 2026-11-16-14-10-00_match-android-safe-area-color
title: Match Android safe areas to HUD background
summary: Use the GUI background color for Android cutout and navigation areas.
status: past
created_at: 2026-11-16-14-10-00
---

# Match Android safe areas to HUD background

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

User explicitly authorized the color change. Bind to Phase 8 GUI presentation and Phase 17 continuous Android evidence. Preserve inset geometry, touch behavior, runtime data and minimum SDK. The renderer currently uses a fixed dark palette; do not implement speculative theme switching. Expose that actual color through the existing mobile bridge, apply it to native background/navigation, and select light system buttons without a contrast scrim.

- [x] Implement shared color access and native styling; document ownership and limitations.
- [x] Build through the canonical pipeline, install preserving data, compare screenshot pixels in portrait/landscape, review and archive.

Stop when padding matches rendered background and navigation buttons remain readable. No commit or publication. Keep evidence under `.tmp/pixel10-color/`, restore tracked release artifacts. Other Android versions/navigation modes remain untested.

Acceptance: full governed Linux/Android build, source/size/docs checks passed. APK `e6fd8915e9b8656fcb45581bff462a0141d3644ceabffc2d3f9d5804a9fd58a7` installed with -r preserving data. Pixel 10 Android 17 screenshots `.tmp/pixel10-color/portrait.png` and `landscape.png` verified visually and by pixel sampling: cutout, HUD and navigation background all RGB(15,18,28); home-button center RGB(255,255,255). Rotation restored to 0. Review: color comes from the same renderer variable, no duplicate palette or mutable cross-thread state; new APIs guarded for older OS. No geometry changes, new dependencies or further issues found. Runtime theme switching remains unimplemented, as before. Tracked releases restored; no commit/push.
