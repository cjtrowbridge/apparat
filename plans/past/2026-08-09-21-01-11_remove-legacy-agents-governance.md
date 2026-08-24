---
plan_id: 2026-08-09-21-01-11_remove-legacy-agents-governance
title: Remove Legacy Agents Governance
summary: Remove deprecated Kanban and downtime systems, consolidate Apparat-specific routing in root AGENTS.md, audit retained host governance, and remove the legacy agents submodule.
status: past
created_at: 2026-08-09-21-01-11
---

# Remove Legacy Agents Governance

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## 1. Remove deprecated host systems

- [x] Delete the live `kanban/` boards, Kanban playbook, reference, and template.
- [x] Remove Kanban fields and procedures from the journal template, kickoff workflow, checkpoint reference, journal commit workflow, bootstrap guidance, and root instructions.
- [x] Delete the live `downtime/` report directories, downtime playbook, and downtime report template.
- [x] Remove downtime bootstrap, routing, code-check exclusions, and live documentation references.
- [x] Preserve historical journal and archived-plan text as immutable history.

## 2. Consolidate host routing and retained governance

- [x] Add an Apparat-specific task-routing table to root `AGENTS.md` for product planning, debugging, review, tool wrappers, HUD work, submodule synthesis, framework assimilation, journal kickoff, and local commit checkpoints.
- [x] Define Apparat build targets as deterministic entities and `scripts/build.py` as the host build pipeline without implicitly selecting model, PDF, semantic-review, rejection-evidence, signing, publication, or remote-execution concerns.
- [x] Replace the journal template and kickoff procedure with TODO/plan/journal-oriented equivalents.
- [x] Require explicit push approval for every host commit, including journal-only checkpoints.
- [x] Remove retained host playbooks, references, and templates that have no remaining routed or supporting role.
- [x] Verify every retained host governance artifact has a root route or a documented supporting owner.

## 3. Remove the legacy submodule

- [x] Confirm no required live behavior remains exclusive to `./agents`.
- [x] Remove the `agents` gitlink and `.gitmodules` entry without modifying `agentic-pipelines`.
- [x] Remove transitional `agents` language from root `AGENTS.md`, `README.md`, build-wrapper help, and clone/bootstrap commands.
- [x] Keep archived plans and journal records unchanged.

## 4. Verify and checkpoint

- [x] Validate root routing and all live path references.
- [x] Regenerate and validate plan indexes using `agentic-pipelines`.
- [x] Run Python tests, Go tests, code-size checks, documentation checks, VS Code JSON validation, build-wrapper help, and a target-discovery/build smoke test proportional to the change.
- [x] Verify clean Agentic Pipelines submodule status and fresh-clone initialization instructions.
- [x] Review diff and Git status for unintended history rewrites, generated artifacts, credentials, and unrelated changes.
- [x] Append the checkpoint to today's journal, archive this completed plan, and commit locally without pushing.
