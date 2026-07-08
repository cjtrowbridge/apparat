# AGENTS Instructions

Read `./agents/RULES.md` in its entirety before doing anything in this repository. Follow all instructions in `./agents/RULES.md` as though they are written directly in this file. Do not proceed if you have not read and understood `./agents/RULES.md`.

## Framework Resolution

- Treat `./agents/RULES.md` as the canonical shared-policy baseline; explicit host-specific overrides in this file take precedence for this repository.
- Use host-managed `./playbooks/`, `./references/`, `./templates/`, and `./scripts/` when present.
- Fall back to `./agents/playbooks/`, `./agents/references/`, `./agents/templates/`, and `./agents/scripts/` when host copies are missing.
- Treat `./plans/`, `./journal/`, `./kanban/`, and `./downtime/reports/` as host-owned operational state.
- Keep agent-facing operating instructions in `AGENTS.md`, not the human-facing `README.md`.

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
git submodule update --init --recursive agents
```

Regenerate or validate host plan indexes:

```bash
python agents/scripts/regenerate_plan_indexes.py --repo-root .
python agents/scripts/regenerate_plan_indexes.py --check --repo-root .
```

When updating the submodule, compare upstream changes with host-managed copies and synthesize them without overwriting host-specific behavior.
