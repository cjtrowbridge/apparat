# AGENTS Instructions

Read `./agentic-pipelines/AGENTS.md` in its entirety before doing anything in this repository. Follow its shared instructions as though they are written directly in this file. Explicit Apparat-specific overrides below take precedence.

## Framework Resolution

- Treat `./agentic-pipelines/AGENTS.md` as the selected shared-policy baseline; explicit host-specific overrides in this file take precedence for this repository.
- Use host-managed `./playbooks/`, `./references/`, `./templates/`, and `./scripts/` when present.
- Fall back to `./agentic-pipelines/playbooks/`, `./agentic-pipelines/references/`, `./agentic-pipelines/templates/`, and `./agentic-pipelines/scripts/` when host copies are missing.
- Treat `./plans/` and `./journal/` as host-owned operational state.
- Keep agent-facing operating instructions in `AGENTS.md`, not the human-facing `README.md`.
- Agents should be aware of TODO.md but only act on it or pull things out of it to work on when the user directly instructs the agent to do that.
- When the user tells an agent to tackle, pull, accomplish, or otherwise execute a TODO.md item, mark that TODO item `[-]` in progress before implementation begins. After the implementation and appropriate verification are complete, mark the same item `[x]` complete. Keep the TODO item text byte-for-byte unchanged unless the user explicitly asks to rewrite it.

## Apparat Task Routing

Agentic Pipelines routes reusable pipeline concerns. Apparat owns product-development and repository-operation routes:

| Task | Primary host playbook |
| --- | --- |
| Plan or implement Apparat product work | `playbooks/how_to_create_and_maintain_task_execution_plans.md` |
| Debug a change that caused errors | `playbooks/debugging_changes_that_lead_to_errors.md` |
| Review changes for risk or regression | `playbooks/how_to_review_changes_for_risk_and_regression.md` |
| Add or modify a tool wrapper | `playbooks/how_to_add_or_modify_a_tool_wrapper_safely.md` |
| Add or modify HUD tab contents | `playbooks/how_to_add_or_modify_hud_tab_contents.md` |
| Create a host playbook | `playbooks/how_to_create_a_new_playbook.md` |
| Assimilate another agentic framework | `playbooks/how_to_assimilate_another_agentic_framework.md` |
| Migrate README roadmap work into execution plans | `playbooks/how_to_migrate_readme_roadmaps_to_plans_system.md` |
| Bootstrap or update the framework submodule | `playbooks/how_to_bootstrap_framework_submodule_into_host_repo.md` or `playbooks/how_to_update_submodule_and_synthesize_host_overrides.md` |
| Capture the daily journal/TODO snapshot | `playbooks/how_to_run_daily_kickoff_and_capture_snapshot.md` |
| Commit or push a completed checkpoint | `playbooks/how_to_commit_and_push_changes.md` or `playbooks/how_to_commit_and_push_journal_checkpoints.md` |

If both routing tables apply, load the Apparat playbook for host procedure and only the Agentic Pipelines concern artifacts that procedure explicitly selects.

Supporting ownership:

- `references/interaction_checkpoints_and_automation_boundaries.md` supports planning and commit approval boundaries.
- `references/conversation_checkpoint_commits.md` supports journal checkpoint commits.
- `references/verification_patterns_for_docs_and_policy.md` supports playbook creation and regression review.
- `references/how_to_shape_agent_tone_and_timbre.md` supports playbook creation and framework assimilation.
- Templates are loaded only by the host playbook that names or produces their corresponding artifact.

## Apparat Build Pipeline Scope

- Treat each target reported by `scripts/build.py` as a deterministic build entity and the script's existing no-flag orchestration as the main build pipeline.
- The Apparat build path does not select model inference, PDF conversion, semantic review, rejected-candidate evidence, signing, publication, or remote execution unless an approved plan explicitly adds that separate concern.
- Build commands, prerequisites, target eligibility, artifacts, and validation remain host-owned. Do not introduce a second build engine through Agentic Pipelines.

## Product Documentation And Planning

- Host-specific scope override: the plan requirement applies to implementation work, not user-approved product-strategy or documentation refinement by itself.
- Read `./README.md` and `./ROADMAP.md` before proposing or implementing product work.
- Treat `./README.md` as the canonical human-facing product, architecture, scope, terminology, and design-decision contract.
- Treat `./ROADMAP.md` as the canonical high-level implementation sequence, dependency map, phase checklist, exit criteria, and open-decision register.
- Keep `README.md`, `ROADMAP.md`, future design documents, and implementation behavior consistent.
- Reserve files under `./plans/` for approved implementation work involving code, dependencies, schemas, protocols, build/release systems, migrations, or similarly executable repository changes.
- Do not create an execution plan merely to discuss, clarify, or directly refine product goals, strategy, README content, or roadmap content when the user has already approved those documentation changes.
- Before implementation, bind the execution plan to specific `ROADMAP.md` items and verify that its intended behavior agrees with `README.md`.
- Preserve journals as append-only historical records; record corrections in later entries instead of rewriting prior checkpoints.

## Local Salvagecore Reference

- `./third_party/salvagecore/` is an ignored local checkout of an older Apparat implementation used only for temporary source and design reference.
- It is not a tracked submodule, build dependency, or product authority.
- Do not add any file beneath `./third_party/salvagecore/` to the host repository.
- Do not copy it wholesale. Reuse only behavior explicitly selected by `README.md`, `ROADMAP.md`, and an approved implementation plan.

## Application Governance

- Treat every new file or feature as undocumented until its purpose, operation, assumptions, failure modes, and verification path are recorded at the closest useful documentation layer.
- Add or update a local `README.md` for every code, script, tool, test, or build directory that gains source files.
- When adding or changing a script, update `scripts/README.md`, provide useful `--help` output, and document prerequisites, side effects, outputs, and common failures.
- When adding or changing build/runtime behavior, update the root `README.md` if normal users or contributors need to run, configure, observe, or troubleshoot it.
- Use an OpenJDK 21 distribution for Android development, CI, and release tooling. Oracle JDK is prohibited for this repository; Eclipse Temurin is the preferred distribution unless a documented compatibility need requires another OpenJDK distribution.
- Keep executable application code under `cmd/` and `internal/`.
- Keep source-reference checkouts under `third_party/`; application imports must not depend on `third_party/salvagecore`.
- Use `cmd/apparat` for the GUI console and `cmd/apparatd` for the headless worker/service entry point.
- Keep shared runtime orchestration in `internal/app`.
- Keep product rules and durable concepts in `internal/domain`.
- Keep external-system integrations in `internal/adapters`.
- Keep OS and platform lifecycle boundaries in `internal/platform`.
- Keep code files at or below 400 physical lines; `make check-code-size` enforces this for included source files.
- Decompose any over-limit code file into smaller package files and document the split in that directory before expecting `make verify` to pass.
- Run `make check-docs` before considering documentation-governance work complete.
- Use structured logging with stable component, event, command, correlation, and error fields.
- Redact secrets, tokens, private keys, passphrases, raw audio, message bodies, project file contents, and other sensitive payloads before logging.

## Framework Commands

Initialize the framework after cloning:

```bash
git submodule update --init --recursive agentic-pipelines
```

Regenerate or validate host plan indexes:

```bash
python agentic-pipelines/scripts/regenerate_plan_indexes.py --repo-root .
python agentic-pipelines/scripts/regenerate_plan_indexes.py --check --repo-root .
```

When updating the submodule, compare upstream changes with host-managed copies and synthesize them without overwriting host-specific behavior.
