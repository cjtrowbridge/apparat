---
plan_id: 2026-11-20-09-00-00_add-pi-inference-submodule
title: Add Pi as an inference source submodule
summary: Pin earendil-works/pi under third_party/inference and document its MIT license and source-reference role.
status: past
created_at: 2026-11-20-09-00-00
---

# Add Pi as an inference source submodule

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Outcome and scope

The user explicitly requests `https://github.com/earendil-works/pi` as a submodule
under inference. Bind to ROADMAP Phase 0 inference source admission and README's
pinned source-reference contract. Add `third_party/inference/pi`, select the
observed upstream main HEAD `cb1205e4661723c7a8c9fb01acd38b7691f97797` and leave
the checkout detached and unmodified. Document provider API and agent-toolkit
source study, not an active runtime integration. Its monorepo/CLI/RPC interfaces
justify a source tree for review rather than a new npm dependency in Apparat.

Update `.gitmodules`, the source inventories, README and roadmap; preserve all
existing submodule pins and user TODO edits. No upstream installer, npm lifecycle
script, model, credential, application build, deployment, commit or push is
needed or authorized by this narrow source-admission request.

## Work and acceptance

- [x] 1. Add the requested submodule at the exact selected revision and review its actual license and nested-submodule metadata.
- [x] 2. Document its purpose, MIT license, revision, update policy and reference-only boundary; verify gitlink/config/clean detached checkout, docs, source/size, index and whitespace checks.

## Evidence and stopping condition

Checkout identity and read-only bootstrap passed before edits. Upstream default
branch is main; `git ls-remote --symref` resolved the selected commit above.
Upstream README and repository metadata identify an MIT-licensed multi-provider
LLM API and agent toolkit. Verify LICENSE in the pinned tree before completion.
Stop after local submodule/inventory admission is complete and reviewable; no
build/runtime support claim follows from a source checkout. Archive this plan,
append the journal and refresh host indexes when the checks pass.

### Admission verification — 2026-11-20

Added `.gitmodules` URL/path and mode-160000 gitlink at exactly
`cb1205e4661723c7a8c9fb01acd38b7691f97797`. The detached checkout is clean;
recursive initialization of this path passed and no nested `.gitmodules` exists.
Reviewed pinned LICENSE: MIT, copyright 2025 Mario Zechner, with notice/license
preservation and standard warranty disclaimer. No upstream code or installer ran.

Updated root README, ROADMAP Phase 0 source admission and both third-party
inventories with the pin, license, source-study purpose and update procedure.
`go.mod`/`go.sum` and all existing submodule pins remain unchanged. Documentation,
physical-size, source governance, plan-index and whitespace checks pass. Preserve
the user's TODO bytes. Changes are local and staged for review; no application
build, commit, push or deployment was requested or performed. Archive and stop.
