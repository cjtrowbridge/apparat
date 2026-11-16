---
plan_id: 2026-11-16-11-13-17_adopt-framework-source-build-governance
title: Adopt framework source and build governance
summary: Integrate public source policy, portable bootstrap, and mandatory failing checks into host-owned build paths.
status: past
created_at: 2026-11-16-11-13-17
---

# Adopt framework source and build governance

## Authorization and protected scope

The user requested three separate plans for the three submodule-review findings. The user explicitly authorized execution of all three plans and Pixel 4a testing. Implementation is active; initial bootstrap v1 is selected. Work on the existing branch and preserve the uncommitted submodule updates and inventory edits. This plan does not authorize commit, push, publication, new branches, destructive cleanup or machine-wide installation.

Use the requested source refresh as the starting candidate. Do not silently downgrade a submodule, fetch a different moving head, upgrade unrelated SQLite/EbitenUI application dependencies, activate reference-only trees, or redesign the product. Keep the host's deterministic no-flag build orchestration, target eligibility, artifact semantics and 400-physical-line limit. Do not add a second build engine or inference/remote execution concerns.

## Roadmap binding and dependencies

Finding 2 in the [original review](../past/2026-11-16-11-00-00_update-and-review-submodules.md) is reproduced: the canonical source checker fails because Apparat lacks public source policy. Portable bootstrap and required build gates are also absent.

Bind to ROADMAP Phase 1 governance/reproducible tooling and Phase 4 code-size/documentation checks. Support Phase 8's continuous target evidence without claiming its broader runtime deliverables. Preserve Apparat build ownership and its stricter physical-line limit. Current host checks do not implement the new header/layout/inventory requirements.

This plan can proceed independently of the engine-version decision. It owns prerequisite availability and the build gate needed by [Reconcile Ebitengine and mobile tool dependencies](../past/2026-11-16-11-13-18_reconcile-engine-and-mobile-dependencies.md) and [Repair Android display patch and build compatibility](../past/2026-11-16-11-13-16_repair-android-display-build.md). It must not require a successful real Android build to close its gate/bootstrap integration checks: use controlled build-command fixtures and truthful prerequisite probes. The Android plan owns APK/device evidence.

## Intended approach

Reuse canonical framework checks and existing host preflight code. Add a small standard-library bootstrap dispatcher and initial versioned implementation. Declare public source rules, migrate only demonstrated first-party violations, and invoke the gate through each actual host build path. No private API configuration, credentials, new hosted workflow or generic build framework is required.

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Integrate framework governance and restore verification prerequisites (finding 2).
  - [x] 1.1 Inventory tracked first-party roots/languages and every compile/package entrypoint, including Makefile, Python builds, Android helper/tool builds, PowerShell wrapper and VS Code actions. Run a read-only gap assessment before source migration.
  - [x] 1.2 Add public `source-governance.yaml` with supported language adapters/comment patterns and exact entrypoint contracts where needed. Use canonical AGENTS/reference guidance over stale playbook text requiring private API files; neither `api.yaml` nor credentials may be a source-check prerequisite.
  - [x] 1.3 Add `scripts/bootstrap.py` and the proposed initial `scripts/bootstrap-v1.py`, with documented probes/authorized repairs for the declared framework, Go, checker, build/test tools and Android prerequisites. Select v1 as part of implementation approval, without automatic future version bumps. Reuse host preflight logic; distinguish required, optional, unavailable and dependency-blocked checks.
  - [x] 1.4 Restore pinned Go/tool availability through the reviewed bootstrap using ignored project-local installation/cache paths where feasible. Require OpenJDK 21 for Android, preferring Temurin; reject Oracle JDK. Report unavailable system/display/SDK prerequisites without silent elevation or system Python changes.
  - [x] 1.5 Migrate only first-party source violations identified by the checker: meaningful responsibility headers, statement layout, cohesive decomposition and exact README inventories. Preserve runtime behavior and the host maximum of 400 total physical lines; do not relax gates or rewrite third-party trees.
  - [x] 1.6 Install/check the framework staged pre-commit hook through bootstrap while preserving existing hooks. Gate every supported host software-build path with the canonical full-tree source checker before compiler/package execution; preserve failure and interruption statuses. Generated first-party code must be checked before compilation when applicable.
  - [x] 1.7 Route CLI/Makefile, native wrappers and VS Code task/launch paths through the same bootstrap and failing gate, invoking the canonical no-flag pipeline once. Preserve target eligibility and artifact semantics; avoid bootstrap recursion or two independent build orchestrators.
  - [x] 1.8 Test read-only `--check`, complete per-probe results, failed repair, Ctrl+C/130, and idempotent second execution. Test that invalid source prevents build invocation and valid source allows exactly one invocation with its status preserved, using controlled fixtures instead of producing release artifacts.
  - [x] 1.9 Verify the canonical full-tree source check, host code-size/docs checks and staged-hook behavior. Update scripts/README, affected directory inventories, root operational docs and necessary host governance instructions. Keep unrelated planning-template modernization outside this repair.

- [x] 2. Complete governance acceptance and documentation.
  - [x] 2.1 Verify public source policy and canonical checks pass alongside the host's stricter size checks; required source and generated-code paths are covered.
  - [x] 2.2 Demonstrate read-only check mode, idempotent normal mode, failure/interruption propagation, preserved hooks and fail-before-build behavior across supported entrypoints using controlled fixtures.
  - [x] 2.3 Document prerequisite availability and hand off the governed build path to the dependency/Android plans; close finding 2 without claiming runtime/platform compatibility from source conformance alone.

## Expected files and ownership

`source-governance.yaml`, `scripts/bootstrap.py`, proposed `scripts/bootstrap-v1.py`, `Makefile`, native wrappers, `.vscode/tasks.json`, `.vscode/launch.json`, focused tests, necessary host `AGENTS.md` synthesis, and affected READMEs. Record the exact first-party migration file inventory after assessment; preserve runtime semantics. Do not modernize unrelated planning templates. If a checker lacks a required language adapter, propose supported integration rather than excluding first-party files.

## Evidence, rollback and closure

Capture task-owned diffs and exact source/tool revisions before implementation; revert only this plan's changes if needed, preserving other work and diagnostic evidence. Do not disable gates, discard user state, uninstall host tools, or publish generated artifacts as a fallback. Preserve tracked release artifacts during validation. Any engine fallback belongs to the engine-version decision in the dependency plan.

Run relevant regressions, `make check-docs`, code-size/source checks for changed source, plan-index validation and diff checks. Record results and unavailable checks truthfully. The earlier 41 Python tests do not establish Go compilation or device compatibility; Go was unavailable during the original review. Require OpenJDK 21 for Android, preferably Temurin, and prohibit Oracle JDK.

Update the closest README inventories and script help/prerequisite/output/failure documentation as behavior changes. Record a concise append-only journal checkpoint and findings disposition. Archive only after all work is complete or explicitly deferred with a named follow-up and accurate support limitations. Update cross-plan links on lifecycle moves. Do not claim Phase 5 or Phase 8 completion from this maintenance work.

Verification exposed five pre-existing unused mock-content builders rejected by the pinned lint tool. Removed only those unreachable declarations (routingSections, pipelineSections, tasksSection, taskScenarioSections, routingScenarioSections); the active selector content is unchanged. This is bounded cleanup needed for the plan's required verification.

## Completion evidence — 2026-11-16

Public source policy, bootstrap v1, the shared failing build gate, native wrappers, Make and VS Code build routing are implemented. The source migration covers the first-party roots declared by source-governance.yaml, with an exact change inventory in the working diff and pre-migration files in `.tmp/remediation/before`. The settings tests were split to retain the 400-physical-line maximum. Third-party source was not reformatted.

Bootstrap copied Go/native inputs into independent ignored local installations; Android archives and Temurin were installed through the reviewed bootstrap. The system PATH ADB is shared with other projects. Read-only Android prerequisite check reports 8 passed/0 failed. Hook installation/check preserves the framework's existing-hook chain; no commit was made. Controlled tests cover gate failure before invocation, exactly-once invocation/status, read-only bootstrap, idempotent existing installation, and interruption. Isolated subprocess fixtures additionally verified repair failure exit 1 and controlled interruption exit 130 (`.tmp/remediation/bootstrap-boundaries.txt`).

Canonical source governance reports 85 source files/0 failures; physical size and documentation checks pass. All 48 Python tests pass. Linux CLI/build entrypoints executed through the gate; Windows/macOS routing was inspected but not executed on those hosts. Generated Java comes from pinned third-party templates and checked first-party wrapper source; generated disposable tool/engine copies retain the documented upstream-only scope. This closes finding 2's missing public policy/bootstrap/gate; it does not establish cross-platform runtime compatibility.
