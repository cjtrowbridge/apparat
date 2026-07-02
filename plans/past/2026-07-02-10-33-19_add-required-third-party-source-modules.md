---
plan_id: 2026-07-02-10-33-19_add-required-third-party-source-modules
title: Add Required Third-Party Source Modules
summary: Add, pin, document, and validate the nine early third-party source references selected by the Apparat roadmap.
status: past
created_at: 2026-07-02-10-33-19
---

# Add Required Third-Party Source Modules

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Establish the tracked third-party source inventory.
  - [x] 1.1 Create grouping inventories.
    - [x] 1.1.1 Create `third_party/README.md` with source-admission and update rules.
    - [x] 1.1.2 Create inventories for `game`, `database`, `networking`, `inference`, and `speech`.
  - [x] 1.2 Record source contracts.
    - [x] 1.2.1 Record each source path, upstream URL, license, purpose, selected revision, and build/reference status.
    - [x] 1.2.2 Record that source presence does not make a module an active application dependency.

- [x] 2. Add the required source submodules.
  - [x] 2.1 Add game and HUD sources.
    - [x] 2.1.1 Add Ebitengine at `third_party/game/ebiten` and pin a stable 2.9.x tag.
    - [x] 2.1.2 Add EbitenUI at `third_party/game/ebitenui`.
    - [x] 2.1.3 Add debugui at `third_party/game/debugui`.
  - [x] 2.2 Add persistence source.
    - [x] 2.2.1 Add modernc SQLite at `third_party/database/modernc-sqlite`.
  - [x] 2.3 Add networking sources.
    - [x] 2.3.1 Add wireguard-go at `third_party/networking/wireguard-go`.
    - [x] 2.3.2 Add wgctrl-go at `third_party/networking/wgctrl-go`.
    - [x] 2.3.3 Add wireguard-tools at `third_party/networking/wireguard-tools`.
  - [x] 2.4 Add inference and speech sources.
    - [x] 2.4.1 Add llama.cpp at `third_party/inference/llama.cpp`.
    - [x] 2.4.2 Add whisper.cpp at `third_party/speech/whisper.cpp`.
  - [x] 2.5 Initialize recursive dependencies.
    - [x] 2.5.1 Run recursive submodule initialization for every added source.
    - [x] 2.5.2 Confirm the ignored Salvagecore checkout remains absent from `.gitmodules` and staging.

- [x] 3. Synchronize project documentation.
  - [x] 3.1 Update the roadmap.
    - [x] 3.1.1 Mark completed grouping, source-addition, pin, and role-documentation items.
    - [x] 3.1.2 Leave dependency wiring, deferred admission gates, and unrelated Phase 0 work pending.
  - [x] 3.2 Record the implementation checkpoint.
    - [x] 3.2.1 Append the source-module addition to `journal/2026-07-02.md`.

- [x] 4. Verify the source foundation.
  - [x] 4.1 Verify repository state.
    - [x] 4.1.1 Confirm all nine gitlinks and URLs match the canonical README and ROADMAP.
    - [x] 4.1.2 Confirm every inventory revision matches its gitlink commit.
    - [x] 4.1.3 Confirm license declarations against each checked-out source.
  - [x] 4.2 Verify recursive and policy boundaries.
    - [x] 4.2.1 Confirm recursive submodule status has no missing required checkout.
    - [x] 4.2.2 Confirm deferred and excluded candidates were not added.
    - [x] 4.2.3 Validate plan indexes, Markdown whitespace, and staged diff integrity.
  - [x] 4.3 Complete the plan lifecycle.
    - [x] 4.3.1 Mark verified checklist items complete and archive the plan under `plans/past/`.
    - [x] 4.3.2 Regenerate future, current, and past plan indexes.
