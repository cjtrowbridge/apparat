---
plan_id: 2026-08-09-20-36-57_add-agentic-pipelines-alongside-agents
title: Add Agentic Pipelines Alongside Agents
summary: Add the published Agentic Pipelines repository as a second submodule, redirect live governance and plan tooling to it, and configure the host-owned local build entrypoint without removing the legacy agents submodule.
status: past
created_at: 2026-08-09-20-36-57
---

# Add Agentic Pipelines Alongside Agents

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## 1. Preserve a reversible transition

- [x] Reset Apparat to published baseline `995b653` and the temporary review checkout to published Agentic Pipelines baseline `e8ce8e7`.
- [x] Preserve discarded experiments through the existing `backup/pre-concern-migration-reset-2026-08-09` branches and reflogs.
- [x] Keep `./agents` initialized and unchanged until the replacement has been exercised and reviewed.
- [x] Do not rewrite historical journals or archived plans merely to replace old framework paths.

## 2. Add the second framework submodule

- [x] Add `https://github.com/cjtrowbridge/agentic-pipelines.git` at `./agentic-pipelines`.
- [x] Pin the submodule to a revision available from its upstream remote.
- [x] Initialize its recursive dependencies and verify clean submodule status.
- [x] Update clone/bootstrap documentation to initialize both framework submodules during the transition.

## 3. Redirect live Apparat governance

- [x] Replace the root `AGENTS.md` shared baseline from `./agents/RULES.md` with `./agentic-pipelines/AGENTS.md` while preserving every Apparat-specific override.
- [x] Prefer host-owned playbooks, references, templates, and scripts, then fall back to `./agentic-pipelines` equivalents.
- [x] Redirect active plan-index commands and other live framework links to `./agentic-pipelines`.
- [x] Retain an explicit transition note that `./agents` remains present but is no longer the selected governance baseline.
- [x] Confirm no live instruction requires the obsolete `RULES.md` path.

## 4. Configure the local build path

- [x] Read the new framework's bootstrap, software-operation, and VS Code entrypoint guidance from the pinned submodule.
- [x] Map Apparat build targets to entities and the existing no-flag `scripts/build.py` process to the main deterministic pipeline without adding a second build engine.
- [x] Add the smallest host-owned prerequisite/bootstrap entrypoint required by the selected local build path.
- [x] Add host-owned VS Code task and primary play-action configuration for the ordinary Apparat build when required by the pinned guidance.
- [x] Keep inference, PDF, diagnostic rendering, signing, publication, and remote execution out of this build path unless separately selected.
- [x] Document prerequisites, side effects, outputs, and failure behavior at the closest useful layer.

## 5. Verify and checkpoint

- [x] Run both framework plan-index checks, Apparat documentation checks, build tests, and relevant full verification.
- [x] Verify the build entrypoint reports targets and prerequisites truthfully without requiring model configuration.
- [x] Verify `.gitmodules`, submodule status, ignored local framework state, and fresh-clone commands.
- [x] Review the diff for accidental historical rewrites, credentials, generated state, and unrelated changes.
- [x] Record results in the append-only journal and mark this checklist from evidence.
- [x] Commit the completed local migration checkpoint without pushing.
