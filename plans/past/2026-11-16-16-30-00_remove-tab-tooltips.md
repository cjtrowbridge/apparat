---
plan_id: 2026-11-16-16-30-00_remove-tab-tooltips
title: Remove icon tab tooltips
summary: Remove popup tab labels while keeping page headings and navigation.
status: past
created_at: 2026-11-16-16-30-00
---

# Remove icon tab tooltips

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

User authorizes removing confusing emoji tab tooltips. Bound to ROADMAP Phase 8 HUD navigation and continuous Phase 17 UI evidence. Delete the shared tooltip attachment; retain icon sizes, actions, metadata and prominent active-page headings. Update existing tab test and docs. Verify GUI/HUD tests, source/docs/size gates and diff. No new abstractions or release publication needed for this source change.

- [x] Remove shared tooltip attachment and reconcile current documentation.
- [x] Verify the existing tab tests require no tooltips and retain navigation behavior; review and archive.

Verification: native GUI/HUD tests passed under Xvfb, including all five buttons having zero tooltips and existing navigation/heading/fit checks. Source governance, physical-size, documentation and diff checks pass. Review confirms one shared tooltip option removed; no geometry, actions or heading changes. Existing test assertion updated; no extra test framework. No APK rebuild, device installation, commit or publication in this checkpoint.

User explicitly requested GitHub publication. Reopened for release preparation: increment to 0.1.3/code 4, canonical build, pinned signer/newer-version gate, approved commit/push on main and fetched-byte verification. Include the task plan and journal in the publication checkpoint.

- [x] Publish version 0.1.3 and verify the actual updater URL.

Release preparation: canonical build and publication gate pass for 0.1.3/code 4 above published code 3, retaining the pinned signer. APK SHA-256: `a0f75afb8d831355ff016494423ac2a3c593d129117faecad03eaf9ad6e4752a`. Linux artifacts restored to HEAD; only the requested Android release is promoted. Existing native GUI/HUD tests and docs/source/size gates pass. Publication verification follows push.

Published checkpoint `bf6a47d` to origin/main. Downloaded the exact updater URL and verified byte equality with the validated 0.1.3 candidate, SHA-256 `a0f75afb8d831355ff016494423ac2a3c593d129117faecad03eaf9ad6e4752a`. All outcomes complete; archived. No phone installation in this publication checkpoint.
