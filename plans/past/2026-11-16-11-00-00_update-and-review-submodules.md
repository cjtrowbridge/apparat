---
plan_id: 2026-11-16-11-00-00_update-and-review-submodules
title: Update and review source submodules
summary: Advance admitted submodules to upstream heads and review host compatibility.
status: past
created_at: 2026-11-16-11-00-00
---

# Update and review source submodules

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

Authorization: user requested “update the submodules and then review the updates”. Scope is the ten admitted submodules, their revision inventories, and review evidence; no commit or push.

Roadmap binding: Phase 1 source dependency admission and recorded build relationships; Phase 5 Ebitengine mobile dependency reproducibility. README source dependencies remain authoritative: only Ebitengine is a local Go replacement; other source updates do not upgrade Go module pins.

- [x] 1. Update admitted source revisions.
  - [x] 1.1 Capture HEAD gitlinks and clean worktree baseline; initialize missing framework instructions.
  - [x] 1.2 Fetch configured upstream heads and update recursively, preserving local work.
  - [x] 1.3 Synchronize exact revision inventories for successfully updated sources.
- [x] 2. Review and verify updates.
  - [x] 2.1 Inspect commit deltas, active dependency compatibility, and framework three-way host synthesis needs.
  - [x] 2.2 Run plan index, documentation, and relevant build/test checks; record blockers truthfully.
  - [x] 2.3 Record findings and rollback information and archive completed work.

Rollback: restore each changed submodule to its gitlink in host HEAD with git submodule update --init --recursive (without --remote); restore only task-owned inventory changes. No source edits inside submodules are intended. Android device and full upstream test matrices are outside this source-update review.

## Review findings and dispositions

1. **High — Android builds stop before binding.** `scripts/android_wrapper.py:120` requires `scale := float64(cScale)` in Ebitengine `internal/ui/ui_android.go`. The old pin contains it; the new source constructs a display-info struct instead. Calling `patched_ebiten_android_display_info()` against the new checkout reproduces `RuntimeError: expected Android display-info patch target not found`. Update/remove the obsolete patch based on upstream display handling and validate Android binding/device behavior, or restore the old Ebitengine pin. Disposition: reported, not repaired during this requested update-and-review checkpoint.
2. **High — framework adoption is incomplete.** New `agentic-pipelines/AGENTS.md` requires public source policy, a canonical failing source gate before software builds, hook installation through bootstrap, and portable `scripts/bootstrap.py` delegation. Apparat lacks the policy and portable bootstrap; its existing build entrypoints do not invoke the new gate. Running `check_source_governance.py --repo-root .` fails with `source changes require source-governance.yaml; no API configuration is required`. Existing 400-physical-line checks remain a stronger host size limit but do not implement the added header/layout/inventory contracts. Adopt these through a focused host migration, or retain the prior framework baseline. Disposition: migration recommendation; no automatic host-policy rewrite.
3. **Medium — active engine dependency/tool pins are unreconciled.** Root `go.mod:35` selects the local engine, so this advances actual GUI source from stable 2.9.9 to 2.11 development code. Upstream now requires newer purego, typesetting, x/image, x/sync, x/sys, x/text and gomobile. Several new module versions have no root `go.sum` entries; `scripts/build.py:28` still pins 2025 gomobile while upstream requires the 2026 revision. Reconcile the module graph and Android tool patch together, then compile/test GUI and mobile paths. This is a compatibility risk supported by manifest differences, not a claimed observed compiler error: Go is absent. Disposition: reported for follow-up; root Go pins and tooling were not changed.

## Verification evidence

- `git submodule update --init --recursive --remote`: completed successfully; nine advanced, wgctrl-go unchanged. All ten submodule worktrees are clean and no nested submodules remain uninitialized. Parent gitlinks are uncommitted.
- `make check-docs test-build check-code-size`: passed; 41 Python build-script tests. The test suite's temporary `tool.py` over-limit output is an expected test fixture, not a repository failure.
- Direct Android patch compatibility probe: failed as described above; the old-source needle check passes. No submodule source was edited by the failed probe.
- New framework full-tree source check: failed as described above.
- `make test`: blocked (`go: not found`). GUI compilation, Go race/lint/vulnerability checks, Android SDK/build/device checks and full upstream native test matrices were not run.
- Scope: integration-focused review of manifests, incoming commit/file summaries, framework policy and host overlaps, small EbitenUI/WireGuard diffs and Android boundary. Thousands of upstream runtime changes were not exhaustively audited; no comprehensive security or platform-support certification is implied.
- No primary top-level license changes were found. Ebitengine adds consolidated third-party notices including Unicode data; SQLite adds third-party/SQLite/vector notices; llama.cpp adds/moves vendor notices. Packaging any newly selected source requires carrying the applicable notices. Reference-only trees remain reference-only.

## Framework three-way synthesis

Old upstream: `bdf758c21319b3a2f99d70a9f95417b55e636398`; new upstream: `51a17b0cff0bde9baf9473adacb1c0c11a50d4a7`; third input: current host-owned files.

| Managed files | Incoming change and host disposition |
| --- | --- |
| `AGENTS.md` | New branch/authentication, concise context, coherent deliverable, roadmap review, bootstrap and source governance rules. Preserve host routing, physical-line rule, build ownership, product constraints; source/bootstrap migration remains open above. |
| `playbooks/how_to_create_and_maintain_task_execution_plans.md`, `templates/task_execution_plan.md`, `templates/change_plan.md` | Upstream replaces exhaustive atomization with bounded deliverables and recognizes existing authorization. Host explicitly selects its planning procedure; preserve host artifacts and recommend a separate deliberate synthesis. |
| `playbooks/how_to_bootstrap_framework_submodule_into_host_repo.md`, `playbooks/how_to_update_submodule_and_synthesize_host_overrides.md` | New portable bootstrap preservation and mandatory source/build adoption guidance. Preserve host update procedure; migration needs its own implementation scope. |
| `templates/submodule_update_synthesis_report.md` | Upstream path correction from pipelines to agentic-pipelines is already present in host; no merge needed. |
| `scripts/regenerate_plan_indexes.py` | Upstream only expands the responsibility docstring; CLI and behavior unchanged. Canonical submodule invocation remains compatible. Host copy preserved. |
| `scripts/README.md` | Framework helper inventory changes belong to framework; preserve host script inventory. |

No host procedure/template/script merges were applied. New LLM/PDF/Ollama concerns are outside Apparat's deterministic build scope. Framework update does not by itself establish host conformance. This is a review report, not approval for the separate migration or policy synthesis.

## Source delta inventory

Revision descriptions are informational; exact hashes are the pin. Ancestry checks pass for nine sources (including the unchanged source). Ebitengine switches from its 2.9 stable lineage to upstream development: 24 old-side commits are not ancestors of the selected head. This is a branch-lineage transition, not a simple fast-forward; commit identity alone does not establish whether those fixes were independently carried forward. All ten worktrees are clean and match the documented revisions.

| Submodule | Previous | Selected | Incoming commits | Changed files |
| --- | --- | --- | ---: | ---: |
| `agentic-pipelines` | `bdf758c21319b3a2f99d70a9f95417b55e636398` | `51a17b0cff0bde9baf9473adacb1c0c11a50d4a7` | 46 | 217 |
| `third_party/database/modernc-sqlite` | `5d243466fa05e8c49a870ead00b79bd4a423f712` | `92d72a45bcb65ab33bbe79c8667f2b8b529ae902` | 121 | 774 |
| `third_party/game/debugui` | `19edc7c03832136c85f3b44f1b05e8c997b4836f` | `c0db3047c9344d406475b7ef9c97bfa0a01db79d` | 18 | 23 |
| `third_party/game/ebiten` | `f65118d0bf2d4cdc35b18d661e0a9f2bf9f3e81e` | `d34b806601a001778800e25e3929f33186618127` | 1589 | 961 |
| `third_party/game/ebitenui` | `b1c31d852489cc8b4bc1d9581eaa75686875e5a7` | `a9fa18fc48cda29447f392a4afca584359b9f81f` | 1 | 1 |
| `third_party/inference/llama.cpp` | `fdb1db877c526ec90f668eca1b858da5dba85560` | `ad5cbfa911b975b2bf987d3bf7d8f2b3f36671ab` | 2597 | 3373 |
| `third_party/networking/wgctrl-go` | `a9ab2273dd1075ea74b88c76f8757f8b4003fcbf` | `a9ab2273dd1075ea74b88c76f8757f8b4003fcbf` | 0 | 0 |
| `third_party/networking/wireguard-go` | `ecfc5a8d54462e18e13c72173e2623d16d8e25a0` | `b3bc704ef0ee21a36298601a9e1e5ff90377ac52` | 2 | 2 |
| `third_party/networking/wireguard-tools` | `a998407747005ea7e4e0258d96f105c97241e1d3` | `ba8556ce45c230906041a1a04450b7b638057917` | 4 | 2 |
| `third_party/speech/whisper.cpp` | `6fc7c33b4c3a2cec83e4b65abd5e96a890480375` | `aba02defd6bb6ff5de63721c21e3456645c153e2` | 1027 | 1288 |

EbitenUI's single change honors the explicit Text LinkColor option. WireGuard-go fixes a handshake-state race and packet-length overflow; tools move netdevice filtering to the kernel and skip unused statistics. Debugui improves focus/dropdowns/scaling. SQLite advances driver/library behavior, including connection opening and sync options. llama.cpp and whisper.cpp contain broad backend/model/server/build changes; they remain inactive source references. The EbitenUI and SQLite root module requirements remain at their original versions.

## Completion

Items 1–2 are complete as an update and review, with failed/unavailable compatibility checks explicitly recorded. No product roadmap completion claim changed: Phase 1 source inventory remains documented and Phase 5 Android readiness remains incomplete. Follow-up repair/migration is not implemented by this plan. No commit, push, branch creation, TODO edit, release build, or publication occurred. Documentation inventories and generated plan indexes were refreshed; checkout timestamps account for index timestamp changes.

Suggested eventual commit message after disposition of findings: `chore: refresh source submodules and record compatibility review`.
