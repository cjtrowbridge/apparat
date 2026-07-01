---
plan_id: 2026-07-01-15-03-27_integrate-agents-framework
title: Integrate Agents Framework
summary: Bootstrap the agents submodule and its host-managed operational framework into apparat.
status: past
created_at: 2026-07-01-15-03-27
---

# Integrate Agents Framework

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

- [x] 1. Establish the agents framework integration.
  - [x] 1.1 Record the upstream framework dependency.
    - [x] 1.1.1 Verify `./agents` is configured as the requested Git submodule.
  - [x] 1.2 Bootstrap host operational state.
    - [x] 1.2.1 Create required plan lifecycle directories and indexes.
    - [x] 1.2.2 Create required journal and kanban directories with upstream starter artifacts.
    - [x] 1.2.3 Create required downtime report directories and upstream starter artifacts.
  - [x] 1.3 Bootstrap host framework files.
    - [x] 1.3.1 Copy upstream runtime shims into the host root.
    - [x] 1.3.2 Copy upstream playbooks into the host-managed `./playbooks/` directory.
    - [x] 1.3.3 Copy upstream references into the host-managed `./references/` directory.
    - [x] 1.3.4 Copy upstream templates into the host-managed `./templates/` directory.
    - [x] 1.3.5 Copy upstream scripts into the host-managed `./scripts/` directory.
  - [x] 1.4 Document host integration.
    - [x] 1.4.1 Update `README.md` with framework structure, ownership, and core commands.
    - [x] 1.4.2 Record the integration checkpoint in `journal/2026-07-01.md`.

- [x] 2. Verify the integrated framework.
  - [x] 2.1 Validate host policy and paths.
    - [x] 2.1.1 Confirm root shims resolve canonical policy and host-first framework paths.
    - [x] 2.1.2 Confirm required host directories and copied framework files exist.
  - [x] 2.2 Validate plan governance.
    - [x] 2.2.1 Regenerate host plan indexes with the submodule script.
    - [x] 2.2.2 Run the host plan-index consistency check.
  - [x] 2.3 Review repository changes.
    - [x] 2.3.1 Review Git status, submodule state, and diff integrity.
    - [x] 2.3.2 Report checklist deltas and propose a checkpoint commit message.

- [x] 3. Separate human and agent documentation.
  - [x] 3.1 Keep project documentation human-facing.
    - [x] 3.1.1 Remove agent framework operating instructions from `README.md`.
  - [x] 3.2 Consolidate agent operating guidance.
    - [x] 3.2.1 Correct `AGENTS.md` to resolve canonical policy from the host root.
    - [x] 3.2.2 Add framework ownership, initialization, indexing, and update guidance to `AGENTS.md`.
    - [x] 3.2.3 Reduce runtime-specific host shims to pointers to `AGENTS.md`.
  - [x] 3.3 Record and verify the correction.
    - [x] 3.3.1 Append the documentation boundary correction to `journal/2026-07-01.md`.
    - [x] 3.3.2 Regenerate plan indexes and verify diff integrity.
