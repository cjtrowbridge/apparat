---
plan_id: 2026-11-16-11-13-18_reconcile-engine-and-mobile-dependencies
title: Reconcile Ebitengine and mobile tool dependencies
summary: Resolve the engine version policy and align the effective source, Go dependency graph, and patched mobile tooling.
status: past
created_at: 2026-11-16-11-13-18
---

# Reconcile Ebitengine and mobile tool dependencies

## Authorization and protected scope

The user requested three separate plans for the three submodule-review findings. The user explicitly authorized execution of all three plans and Pixel 4a testing. Implementation is active; initial bootstrap v1 is selected. Work on the existing branch and preserve the uncommitted submodule updates and inventory edits. This plan does not authorize commit, push, publication, new branches, destructive cleanup or machine-wide installation.

Use the requested source refresh as the starting candidate. Do not silently downgrade a submodule, fetch a different moving head, upgrade unrelated SQLite/EbitenUI application dependencies, activate reference-only trees, or redesign the product. Keep the host's deterministic no-flag build orchestration, target eligibility, artifact semantics and 400-physical-line limit. Do not add a second build engine or inference/remote execution concerns.

## Roadmap binding and dependencies

Finding 3 in the [original review](../past/2026-11-16-11-00-00_update-and-review-submodules.md) is manifest-level incompatibility risk: the local Ebitengine replacement selects development source with newer dependencies while root/module tool pins still describe the old baseline.

Bind to ROADMAP Phase 1 stable game/HUD selection, recorded local replacement relationship and separately pinned tools, plus Phase 5 mobile-tool reproducibility. README and ROADMAP still select stable 2.9.x unless a required newer feature is demonstrated. The refreshed checkout is 2.11 development source. This plan owns resolution of that mismatch and matching current-status documentation; no other plan may silently decide the engine version.

Version-policy analysis can begin immediately. Compilation/tool generation requires the governed prerequisites from [Adopt framework source and build governance](../past/2026-11-16-11-13-17_adopt-framework-source-build-governance.md). Hand off the selected engine and compatible mobile-tool inputs to [Repair Android display patch and build compatibility](../past/2026-11-16-11-13-16_repair-android-display-build.md) after module/tool/Linux validation. Android APK/device verification is owned there and linked back here; do not postpone this handoff until Android closure.

## Intended approach

Recommend and record a concrete version choice before changing pins: retain the refreshed revision with an explicit development-version exception, justify another stable revision, or restore the prior stable baseline. Align the effective engine source, module graph and mobile build tool with that choice. Review minimal dependency changes and stale generated-tool reuse; avoid unrelated application upgrades.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Establish the implementation baseline and resolve version policy.
  - [x] 1.1 Record root/submodule revisions, dirty paths, current tool availability, effective Go replacement, and the reproduced Android patch failure; preserve existing user work.
  - [x] 1.2 Review why the 2.11 development revision is needed against the README/Phase 1 stable constraint. Present a concrete recommendation: retain the exact refreshed revision with an explicitly accepted development-version exception, select a justified newer stable revision, or restore the prior stable revision. Record the selected revision and authorization before changing its pin or dependent manifests.
  - [x] 1.3 Record a matching README/ROADMAP current-status amendment, local replacement/build relationship, rollback boundary, and target evidence requirements for the selected revision.

- [x] 2. Reconcile effective engine and mobile tool dependencies (finding 3).
  - [x] 2.1 Compare the selected engine's module graph and mobile APIs with root `go.mod`, `go.sum`, build constants and generated gomobile module text. Inspect behavior relevant to old stable-line fixes rather than assuming ancestry proves preservation.
  - [x] 2.2 Resolve and pin the minimum required compatible dependencies/tool versions; review module/sum changes for unrelated upgrades. Keep the engine gitlink, nominal requirement, local replacement and documented effective version coherent.
  - [x] 2.3 Reassess each patched gomobile transformation against the selected tool version; remove obsolete edits or implement bounded, validated replacements. Key generated tool reuse to source version and patch identity so a stale `.tools/bin/gomobile-apparat` cannot hide incompatibility.
  - [x] 2.4 Run module consistency, headless tests/build and GUI compilation through the governed paths. Record compiler/API failures and fix only compatibility changes required by the selected dependency set; verify source pins and manifests remain unchanged on a repeat build.

- [x] 3. Complete dependency acceptance and documentation.
  - [x] 3.1 Run `make verify`, GUI dependency/compile checks and governed Linux GUI/headless builds on the provisioned host. Record module consistency and repeat-build pin/manifest cleanliness.
  - [x] 3.2 Document selected engine/tool/dependency revisions and the resolved README/ROADMAP policy mismatch, retain reference-only boundaries, and update required packaging notices for changed inputs.
  - [x] 3.3 Hand off validated source/tool inputs to the Android plan, then reference its APK/device evidence before claiming mobile compatibility. If that evidence remains unavailable, explicitly retain the mobile validation item or record an accepted deferral.

## Expected files and ownership

`go.mod`, `go.sum`, the engine gitlink only after the version decision, gomobile pin/patch/cache logic in `scripts/build.py`, minimal affected GUI/mobile call sites, focused tests, root README/ROADMAP, `third_party/game/README.md` and script/local READMEs. The Android source patch itself belongs to the Android plan.

## Evidence, rollback and closure

Capture task-owned diffs and exact source/tool revisions before implementation; revert only this plan's changes if needed, preserving other work and diagnostic evidence. Do not disable gates, discard user state, uninstall host tools, or publish generated artifacts as a fallback. Preserve tracked release artifacts during validation. Any engine fallback belongs to the engine-version decision in the dependency plan.

Run relevant regressions, `make check-docs`, code-size/source checks for changed source, plan-index validation and diff checks. Record results and unavailable checks truthfully. The earlier 41 Python tests do not establish Go compilation or device compatibility; Go was unavailable during the original review. Require OpenJDK 21 for Android, preferably Temurin, and prohibit Oracle JDK.

Update the closest README inventories and script help/prerequisite/output/failure documentation as behavior changes. Record a concise append-only journal checkpoint and findings disposition. Archive only after all work is complete or explicitly deferred with a named follow-up and accurate support limitations. Update cross-plan links on lifecycle moves. Do not claim Phase 5 or Phase 8 completion from this maintenance work.

## Version decision — 2026-11-16

User asked to investigate why the 2.9 requirement exists and evaluate changing it while executing this plan. `git log -S` traces it to `59ce0a7` (initial architecture/roadmap); the July source-admission plan chose the then-stable 2.9.9, contrasting it with 2.10 alpha. No product-specific ABI, hardware or API requirement fixes Apparat to 2.9. The stability intent remains useful; the fixed minor-version restriction does not.

Select stable v2.10.6, already present in the fetched upstream tags, instead of development head d34b8066. The 2.10 line adds Android context-loss/suspend recovery, input and scale fixes and pure-Go desktop support; Go 1.25 minimum is satisfied by host Go 1.26.4. Upstream release notes: https://ebitengine.org/en/documents/2.10.html . No required 2.11-only feature was established. Update policy to prefer a validated exact stable release; require a concrete unmet need and separate evidence for development revisions. This preserves the requested update while reducing unnecessary prerelease risk.

## Completion evidence — 2026-11-16

Selected stable Ebitengine v2.10.6 at 5f363d17271b029d43e17d17dc00da34f43e2a86; root nominal requirement, local replacement and inventory now agree. Gomobile is v0.0.0-20260820040257-d11f821a26a6. Required graph changes include purego 0.11.3, typesetting 0.3.5, x/image 0.45.0, x/sync 0.22.0, x/sys 0.47.0 and x/text 0.41.0 plus rasterx; unrelated SQLite and EbitenUI pins remain unchanged. Existing third-party license ownership is unchanged (Apache-2.0/MIT/BSD); source license files remain with their inputs. The generated gomobile helper rebuilds from pinned source/current patches each preparation, preventing stale helper reuse.

`make verify` passed (Go unit/race tests, 48 Python tests, source/size/docs checks, lint and vulnerability audit). The audit found zero reachable vulnerabilities; it also reported vulnerabilities in unused imported-package/module code, so this is not a claim of vulnerability-free dependencies. GUI tests under Xvfb, Linux GUI/headless canonical builds and smoke checks, `go mod verify`, and `go mod tidy -diff` passed. Repeat builds leave go.mod/go.sum hashes unchanged and the selected engine checkout clean. Logs are under `.tmp/remediation/`.

The Android plan records the signed target-35 APK and Pixel 4a Android 13 rendering/touch/orientation/resume evidence. Windows/macOS, Android 15 and older Android hardware are not newly validated. The current stable policy supersedes the historical fixed 2.9 minor without adopting unneeded 2.11 development source. Finding 3 is resolved for the tested Linux/Android paths.
