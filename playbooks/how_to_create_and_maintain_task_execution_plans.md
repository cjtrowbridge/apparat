# Playbook: Create and Maintain Task Execution Plans

*Status: Draft*

## Objective

Define a repeatable workflow for creating, selecting, executing, revising, and archiving plan files so all implementation remains bound to approved outcomes and material constraints.

## Prerequisites

* Read `README.md`, `ROADMAP.md`, `AGENTS.md`, and `references/interaction_checkpoints_and_automation_boundaries.md`.
* Apply `agentic-pipelines/references/deliverable_driven_simplicity.md` for outcome scope, implementation choices, measurements and stopping conditions.
* User-approved strategy/documentation refinement alone does not require a new implementation plan. Updating a future plan does not begin its execution.
* Plans directories exist:
  * `./plans/future/`
  * `./plans/current/`
  * `./plans/past/`
* Per-folder indexes exist:
  * `./plans/future/index.md`
  * `./plans/current/index.md`
  * `./plans/past/index.md`
* Plan format requirements:
  * YAML front matter with `plan_id`, `title`, `summary`, `status`, `created_at`
  * Key line must be: Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task
  * Coherent outcomes with protected behavior, necessary dependencies, acceptance evidence and a stopping condition
  * Template: `./templates/task_execution_plan.md`

## Step-by-Step Instructions

1. **Find or Confirm Active Plan**
   * Check `./plans/current/index.md` for an existing active plan.
   * If multiple candidates exist, use the plan identified by the user or established scope; ask only when the choice remains materially ambiguous.
   * If no suitable active plan exists, continue to Step 2.

2. **Create or Select the Governing Plan**
    * If work is new:
       * Create a new plan in `./plans/future/` using `YYYY-MM-DD-HH-mm-ss_slug.md`.
       * Keep slug lowercase and hyphenated (letters, numbers, hyphens only).
       * Keep `plan_id` exactly equal to the filename stem.
       * State the observable outcome, roadmap binding, protected behavior, simplest existing path, material uncertainty, acceptance evidence and stopping condition.
       * Decompose only to expose a necessary dependency, resolve uncertainty or distinguish an independently verifiable outcome. Stop when the next implementation/check is clear.
       * Use a short outcome checklist with relevant verification. Checklist items do not imply separate components, services, interfaces, agents or plans.
    * If work is already in-flight:
       * Move/promote the plan to `./plans/current/` immediately before the first non-trivial implementation edit.

3. **Bind the Task Run to Plan Items**
   * Identify exact checklist items that will be worked in this checkpoint.
   * Include the active plan path and targeted outcome/checklist references in the change plan summary.
   * Treat `./templates/change_plan.md` output as proposal-only context; it does not authorize execution by itself.
   * Keep implementation within approved outcomes and constraints; record routine implementation choices without requiring a separate approved checkbox for each code edit.

4. **Establish Execution Authority**
   * Present:
     * active plan path,
     * checklist items to execute,
     * files expected to change,
     * verification approach.
   * Use existing explicit implementation authorization when it covers the outcome. Otherwise request it before implementation. Plan drafting/refinement authorization alone is not execution authority. Branch creation, commit, push and publication retain their separate approval boundaries.

5. **Execute Approved Outcomes**
    * Implement exactly the approved checklist scope.
    * Mark completed items as `[x]`.
    * Mark uncertain outcomes as `[?]` with follow-up validation items.
    * Mark intentionally de-scoped work as `[-]` (closed).

6. **Handle Mid-Run Divergence**
   * Record routine adjustments within the approved outcome and continue.
   * If behavior, authority, irreversible actions, material cost or scope changes, stop the dependent work, propose a concrete revision and request the missing approval. Continue independent authorized work.
   * Review the governing roadmap alongside every plan revision; reconcile scope, status, sequence, dependencies, deliverables and acceptance in the same approved checkpoint. Record unresolved mismatches explicitly.
   * Give material simplification findings a disposition and selected improvements comparable baseline/after evidence, regression checks and a retain/revert decision. Stop when acceptance and relevant budgets are met.

7. **Maintain Plan Lifecycle State**
   * `future -> current` immediately before the first non-trivial implementation edit.
   * `current -> past` when no further work is planned and all remaining items are either complete or intentionally closed.
   * Keep only actively executed plans in `./plans/current/`.
   * It is valid for `./plans/current/` to be empty when no plan is active.

8. **Refresh Plan Indexes**
   * Ensure each affected status index includes current entries in one-line format:
     * `last_modified | path | title | summary`
   * Order:
     * `future/current`: filesystem last-write time descending.
     * `past`: `created_at` descending.
   * Run `python scripts/regenerate_plan_indexes.py` after plan create/update/move/archive.
   * If this framework is mounted as `./agentic-pipelines` in a host repo using host-owned `./plans/`, run:
     * `python agentic-pipelines/scripts/regenerate_plan_indexes.py --repo-root .`

9. **Checkpoint Summary Requirements**
   * Before commit/push flow, summarize:
     * active plan path,
     * checklist items changed,
     * lifecycle moves (`future/current/past`) if any,
     * index files refreshed.
   * Ensure journal checkpoint details reference the same plan path and checklist deltas.

10. **Verification**
   * Confirm all implemented changes map to checklist items in the active plan.
   * Confirm plan front matter and key line remain valid.
   * Confirm plan indexes match filesystem state for affected plan directories.
   * Confirm no work was executed outside approved plan scope.

## Anti-Patterns

* Executing work without an active approved plan.
* Treating playbooks as direct execution authority without plan binding.
* Expanding scope silently instead of proposing plan revisions.
* Marking plan items complete without corresponding implementation evidence.
* Leaving stale plans in `./plans/current/` after work has ended.
* Recursive decomposition, speculative interfaces, repeated permission requests within approved scope, or optimization without a measured need.

## Lifecycle Compliance

Prompt -> Select/Create Plan (using relevant playbook guidance) -> Establish authorization -> Execute approved outcomes -> Plan update -> Docs update -> Verification.

If this occurs inside a git repo:
* Review `git status` and relevant diffs.
* Ensure checkpoint summary includes active plan path and checklist item changes.
* Suggest a commit message that summarizes the completed checkpoint.
* Commit only when explicitly authorized; implementation approval alone is insufficient.
