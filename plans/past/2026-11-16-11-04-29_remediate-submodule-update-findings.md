---
plan_id: 2026-11-16-11-04-29_remediate-submodule-update-findings
title: Remediate submodule update compatibility and governance findings
summary: Restore a reproducible engine and Android build while integrating the updated framework into host-owned bootstrap and verification.
status: past
created_at: 2026-11-16-11-04-29
---

# Remediate submodule update compatibility and governance findings

Superseded at the user’s request by three independent plans:

- [Repair Android display patch and build compatibility](../past/2026-11-16-11-13-16_repair-android-display-build.md)
- [Adopt framework source and build governance](../past/2026-11-16-11-13-17_adopt-framework-source-build-governance.md)
- [Reconcile Ebitengine and mobile tool dependencies](../past/2026-11-16-11-13-18_reconcile-engine-and-mobile-dependencies.md)

All pending tasks below are closed here because ownership moved to those plans. No implementation was performed under this combined proposal. Retained text records the original proposal, not active execution authority.

Resolve all three findings in the [submodule review](../past/2026-11-16-11-00-00_update-and-review-submodules.md): the Android display patch failure, incomplete framework adoption, and unreconciled active engine dependencies/tooling.

## Authorization and scope

The user requested a plan for each issue. This checkpoint creates the plan only; implementation has not begun. On approval, promote this file to `plans/current/` and execute the scoped tasks below on the existing branch. Approval does not include commit, push, publication, new branches, destructive cleanup, or machine-wide installation.

Keep the nine requested source updates as the starting candidate. Do not silently roll back Ebitengine or the framework, fetch another moving head, upgrade unrelated SQLite/EbitenUI application dependencies, activate reference-only trees, or redesign the product. Any material engine-policy choice must have an explicit disposition before dependent work. Preserve the existing uncommitted update and inventory edits.

## Roadmap and product binding

| Workstream | Governing items | Acceptance boundary |
| --- | --- | --- |
| Engine/tool reconciliation (finding 3) | ROADMAP Phase 1: stable game/HUD source references, recorded local replacement relationship, separately pinned tools | One documented effective engine source and compatible module/tool versions; reference updates remain separate from application upgrades. |
| Android repair (finding 1) | Phase 5: admitted Ebitengine mobile source, tracked host wrapper/AAR path, Android build tests and device evidence | Build and launch the wrapper HUD; no obsolete patch failure and no source-checkout mutations. |
| Source/bootstrap migration (finding 2) | Phase 1: governance/reproducible tooling; Phase 4: code-size and documentation checks; Phase 8: stop mutating tracked/submodule source during builds and continuous Linux/Android evidence | Canonical source gate enforced through host entrypoints while retaining the host 400-physical-line limit and deterministic build ownership. |

Observed mismatch: README still specifies stable Ebitengine 2.9.x unless a required newer feature is demonstrated, and Phase 1 marks that selection complete. The refreshed local replacement is 2.11 development source. The prior review did not resolve this product-contract mismatch. Task 1.2 requires a concrete version decision and corresponding README/ROADMAP correction; merely updating a version table is insufficient.

Phase 5 contains historical checked commands and milestones that do not establish current compatibility after the update. Preserve historical evidence, add a current remediation status/link, and update normal operational instructions to the current no-flag pipeline. Do not mark Phase 5 or Phase 8 complete from this maintenance work. The existing future Phase 8 plan remains separate; cross-link any prerequisite this plan satisfies without marking its broader deliverables complete.

## Order and implementation approach

First capture the baseline and settle the engine version policy. Then integrate bootstrap/source governance so subsequent software builds use the required gate. Reconcile the selected engine's module/tool graph before repairing Android behavior. Finish with integrated Linux and Android evidence. Read-only Android investigation and dependency inspection can proceed before the gate is ready; protected software builds must wait for it.

Prefer existing host scripts and direct function calls. Reuse canonical framework checks rather than implementing a second validator or build engine. The current display-info patch should disappear if the selected upstream already handles the failure it addressed. If a patch remains necessary, prepare verified input in an ignored source copy/cache and leave the tracked submodule unchanged.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [-] 1. Establish the implementation baseline and resolve version policy.
  - [-] 1.1 Record root/submodule revisions, dirty paths, current tool availability, effective Go replacement, and the reproduced Android patch failure; preserve existing user work.
  - [-] 1.2 Review why the 2.11 development revision is needed against the README/Phase 1 stable constraint. Present a concrete recommendation: retain the exact refreshed revision with an explicitly accepted development-version exception, select a justified newer stable revision, or restore the prior stable revision. Record the selected revision and authorization before changing its pin or dependent manifests.
  - [-] 1.3 Record a matching README/ROADMAP current-status amendment, local replacement/build relationship, rollback boundary, and target evidence requirements for the selected revision.

- [-] 2. Integrate framework governance and restore verification prerequisites (finding 2).
  - [-] 2.1 Inventory tracked first-party roots/languages and every compile/package entrypoint, including Makefile, Python builds, Android helper/tool builds, PowerShell wrapper and VS Code actions. Run a read-only gap assessment before source migration.
  - [-] 2.2 Add public `source-governance.yaml` with supported language adapters/comment patterns and exact entrypoint contracts where needed. Use canonical AGENTS/reference guidance over stale playbook text requiring private API files; neither `api.yaml` nor credentials may be a source-check prerequisite.
  - [-] 2.3 Add `scripts/bootstrap.py` and the proposed initial `scripts/bootstrap-v1.py`, with documented probes/authorized repairs for the declared framework, Go, checker, build/test tools and Android prerequisites. Select v1 as part of implementation approval, without automatic future version bumps. Reuse host preflight logic; distinguish required, optional, unavailable and dependency-blocked checks.
  - [-] 2.4 Restore pinned Go/tool availability through the reviewed bootstrap using ignored project-local installation/cache paths where feasible. Require OpenJDK 21 for Android, preferring Temurin; reject Oracle JDK. Report unavailable system/display/SDK prerequisites without silent elevation or system Python changes.
  - [-] 2.5 Migrate only first-party source violations identified by the checker: meaningful responsibility headers, statement layout, cohesive decomposition and exact README inventories. Preserve runtime behavior and the host maximum of 400 total physical lines; do not relax gates or rewrite third-party trees.
  - [-] 2.6 Install/check the framework staged pre-commit hook through bootstrap while preserving existing hooks. Gate every supported host software-build path with the canonical full-tree source checker before compiler/package execution; preserve failure and interruption statuses. Generated first-party code must be checked before compilation when applicable.
  - [-] 2.7 Route CLI/Makefile, native wrappers and VS Code task/launch paths through the same bootstrap and failing gate, invoking the canonical no-flag pipeline once. Preserve target eligibility and artifact semantics; avoid bootstrap recursion or two independent build orchestrators.
  - [-] 2.8 Test read-only `--check`, complete per-probe results, failed repair, Ctrl+C/130, and idempotent second execution. Test that invalid source prevents build invocation and valid source allows exactly one invocation with its status preserved, using controlled fixtures instead of producing release artifacts.
  - [-] 2.9 Verify the canonical full-tree source check, host code-size/docs checks and staged-hook behavior. Update scripts/README, affected directory inventories, root operational docs and necessary host governance instructions. Keep unrelated planning-template modernization outside this repair.

- [-] 3. Reconcile effective engine and mobile tool dependencies (finding 3).
  - [-] 3.1 Compare the selected engine's module graph and mobile APIs with root `go.mod`, `go.sum`, build constants and generated gomobile module text. Inspect behavior relevant to old stable-line fixes rather than assuming ancestry proves preservation.
  - [-] 3.2 Resolve and pin the minimum required compatible dependencies/tool versions; review module/sum changes for unrelated upgrades. Keep the engine gitlink, nominal requirement, local replacement and documented effective version coherent.
  - [-] 3.3 Reassess each patched gomobile transformation against the selected tool version; remove obsolete edits or implement bounded, validated replacements. Key generated tool reuse to source version and patch identity so a stale `.tools/bin/gomobile-apparat` cannot hide incompatibility.
  - [-] 3.4 Run module consistency, headless tests/build and GUI compilation through the governed paths. Record compiler/API failures and fix only compatibility changes required by the selected dependency set; verify source pins and manifests remain unchanged on a repeat build.

- [-] 4. Repair and verify the Android display/build boundary (finding 1).
  - [-] 4.1 Trace the previous zero/invalid display-metric guard and selected upstream display handling. Determine whether the workaround is obsolete or still necessary; record the behavior that must be preserved before changing it.
  - [-] 4.2 Remove the obsolete tracked-tree patch when safe. If still required, apply a narrow validated patch to an ignored copy/cache tied to the exact engine revision and patch identity, wire binding to that input, and fail clearly on unexpected source. Never alter the tracked submodule during success, failure or interruption.
  - [-] 4.3 Add regression coverage for supported upstream handling, invalid/zero metrics behavior where testable, clear unsupported-source failure when applicable, and clean source trees after both successful and failed build preparation. Replace brittle needle-only expectations with checks of the supported behavior.
  - [-] 4.4 Build the android/arm64 wrapper APK through the existing canonical pipeline with OpenJDK 21 and declared SDK/NDK tools. Verify package/signing metadata and existing ABI/SDK/page-size requirements; record output hash and complete dependency/tool provenance.
  - [-] 4.5 On an available test device/emulator, install and launch, inspect process/last_run.log, and check HUD rendering, touch tab selection, orientation and display scaling. Record device/OS and evidence; unavailable hardware remains an explicit pending validation, not a passed test.

- [-] 5. Validate and close the integrated remediation.
  - [-] 5.1 Run `make verify`, the canonical source checker, GUI dependency/compile checks and the governed Linux GUI/headless builds on a provisioned host; retain relevant baseline/final results and blockers.
  - [-] 5.2 Confirm repeated builds preserve gitlinks, submodule worktrees and manifests, retain reference-only boundaries, and require no salvagecore input. Use isolated fixtures/copies to test absence rather than moving or deleting the user's reference checkout.
  - [-] 5.3 Synchronize README, ROADMAP current-status notes, platform matrix, third-party inventory, script/local READMEs and plan evidence. Keep support claims limited to tested targets; update license notices if selected packaged inputs require it.
  - [-] 5.4 Disposition all three original findings, review scope/rollback and roadmap alignment, regenerate/check plan indexes and run `make check-docs` plus diff checks. Archive only when tasks are complete or explicitly deferred with named follow-up and accurate support limitations.

## Expected files and change boundaries

- Governance: `source-governance.yaml`, host `AGENTS.md` as needed for explicit synthesis, `scripts/bootstrap.py`, selected versioned bootstrap, `Makefile`, `.vscode/tasks.json`, `.vscode/launch.json`, `scripts/run_build_pipeline.ps1`, relevant native wrapper and tests.
- Engine/Android: `go.mod`, `go.sum`, selected Ebitengine gitlink only if decided in 1.2, `scripts/build.py`, `scripts/android_wrapper.py`, focused build tests, and minimal GUI/mobile call sites only if selected upstream APIs require changes.
- Mechanical first-party governance migration: declared code/script/test roots, with exact touched-file inventory recorded after 2.1. Broad semantic refactoring or application feature work is excluded.
- Documentation: root README/ROADMAP, `scripts/README.md`, affected directory READMEs, `docs/platform-matrix.md`, `third_party/game/README.md`, plan indexes and an append-only journal checkpoint.
- Reuse the selected framework helpers without modifying upstream source. If a checker adapter cannot handle a host language, record the limitation and propose a supported integration; never omit first-party files solely to obtain a pass.

## Acceptance and unavailable evidence

| Finding | Required evidence to close |
| --- | --- |
| Android patch failure | Binding passes the formerly failing boundary; APK builds; required display behavior is retained; tracked source stays clean; device evidence or explicit remaining device-validation item. |
| Framework adoption | Public policy validates; applicable source conventions pass alongside stricter host limits; bootstrap check is read-only and normal mode idempotent; invalid source blocks every supported build path; hook/task/launch paths are verified. |
| Dependency/tool mismatch | Documented engine choice resolves product-policy mismatch; compatible module graph and pinned tooling; repeatable clean Linux and Android builds; no stale patched-tool reuse. |

The previous 41 passing Python tests do not establish engine compilation or Android compatibility. No Go compiler was available at review time. Hardware/tool availability is an execution prerequisite, not a reason to declare these checks passed. Unit tests should cover consequential failure and recovery behavior; avoid tests that merely duplicate implementation text.

## Rollback and stopping conditions

Before implementation, capture task-owned file diffs and submodule revisions so a failed change can be reverted selectively without erasing the requested update or other work. Preserve tracked release artifacts; build/device validation does not authorize committing or publishing generated binaries. On failure, preserve diagnostics and leave truthful status; do not automatically downgrade the engine, disable a gate, delete caches/user state or uninstall host tools. A version fallback requires the recorded decision in 1.2.

Completion means each finding has evidence of resolution, or a user-accepted deferral with accurate platform limitations. Stop adding mechanisms once those conditions are met. This plan does not expand the product roadmap or select inference, remote execution, signing/publication infrastructure, or additional source submodules.
