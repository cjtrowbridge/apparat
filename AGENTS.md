# AGENTS Instructions

Read `./agents/RULES.md` in its entirety before doing anything in this repository. Follow all instructions in `./agents/RULES.md` as though they are written directly in this file. Do not proceed if you have not read and understood `./agents/RULES.md`.

## Framework Resolution

- Treat `./agents/RULES.md` as canonical policy.
- Use host-managed `./playbooks/`, `./references/`, `./templates/`, and `./scripts/` when present.
- Fall back to `./agents/playbooks/`, `./agents/references/`, `./agents/templates/`, and `./agents/scripts/` when host copies are missing.
- Treat `./plans/`, `./journal/`, `./kanban/`, and `./downtime/reports/` as host-owned operational state.
- Keep agent-facing operating instructions in `AGENTS.md`, not the human-facing `README.md`.

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
