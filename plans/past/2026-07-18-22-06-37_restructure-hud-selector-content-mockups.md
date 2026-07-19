---
plan_id: 2026-07-18-22-06-37_restructure-hud-selector-content-mockups
title: Restructure HUD Selector And Content Mockups
summary: Reorder Research, replace generic placeholder surfaces with faithful mockups, and define color-grouped selector panels for Comrades, Projects, Cluster, and Research.
status: past
created_at: 2026-07-18-22-06-37
---

# Restructure HUD Selector And Content Mockups

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

## Roadmap Binding

- Roadmap section: `ROADMAP.md` Phase 4 Basic HUD Tabs And Content.
- Product contract: preserve controller-first, responsive selector panels and content panels; all example records remain clearly mock/non-durable until their real backend exists.
- Governing playbook: `playbooks/how_to_add_or_modify_hud_tab_contents.md`.
- Source inbox: the pending items in `TODO.md` captured on 2026-07-18.

## Source-Fidelity Requirements

- Preserve the user-specified labels, ordering, capitalization, parenthetical details, color palette, and descriptions below unless a later user-approved plan revision changes a specific item.
- Do not remove Settings content.
- Replace generic placeholder content elsewhere with presentation mockups that resemble the intended product, while clearly avoiding claims that chats, routes, schedules, devices, research contributions, or inference services are real.
- Use `selector panel` and `content panel` terminology in user-facing and contributor documentation.

## Approved Product Decisions

- [x] 0. Resolve the source-label ambiguity: the NPCs/Comrades/chat selector belongs in Comrades, not Cluster.
- [x] 0.1 Define the first mock chat content-panel layout for Librarian and all Comrades: header/identity, static relevant conversation history, status/trust context, an editable EbitenUI reply field anchored at the bottom, and a Send button. Typing and button activation must have no send, persistence, network, or other effect yet.
- [x] 0.2 Define the distinct Projects content-panel layout: each selected project uses in-panel tabs for a mock Git repository view and a mock project-chat view.
- [x] 0.3 Define the Pipeline distinction and content-panel direction: a Project can be any kind of workspace; a Pipeline is a project-oriented workspace with an entry point executed as a Cluster task. Its content panel follows the Project Git/Chat shape plus mock entry-point declarations and a no-op Run button. YAML or another entry-point schema, actual task execution, routing, approval, and run persistence remain later design/implementation work.
- [x] 0.4 Confirm the five-tab order: Comrades, Projects, Cluster, Research, Settings, with Research immediately after Cluster.

## Checklist

- [x] 1. Establish the governed HUD-mockup checkpoint.
  - [x] 1.1 Promote this plan from `future` to `current` immediately before implementation.
  - [x] 1.2 Mark each selected `TODO.md` line `[-]` in progress without rewriting its text.
  - [x] 1.3 Update the roadmap/current HUD contract only after the ordering and ownership decisions are approved.

- [x] 2. Reorder top-level navigation.
  - [x] 2.1 Move Research to immediately after Cluster in default descriptors, tab construction, snapshot tests, runtime tests, and direct keyboard mapping.
  - [x] 2.2 Preserve five top-level tabs and map `Alt+1` through `Alt+5` to their reordered visible positions.
  - [x] 2.3 Update README, ROADMAP, HUD/GUI docs, screenshots, and accessibility labels for the new order.

- [x] 3. Replace generic non-Settings placeholder content with faithful presentation mockups.
  - [x] 3.1 Inventory generic placeholder labels in Comrades, Projects, Research, and Cluster while explicitly excluding Settings from removal.
  - [x] 3.2 Replace only the selected generic content with approved mock records, static status context, and disabled/non-durable controls.
  - [x] 3.3 Retain explicit mock/future/disabled truthfulness where a backend capability does not exist.
  - [x] 3.4 Keep Settings sections, update controls, diagnostics, and current content intact.

- [x] 4. Add semantic selector-panel group colors.
  - [x] 4.1 Define the ordered bisexual-lighting-adjacent palette exactly as supplied: `#0032AB`, `#6028A7`, `#8C159F`, `#AF0093`, `#CB0084`, `#E10072`, `#F10060`.
  - [x] 4.2 Assign each selector subgroup a palette color in order, looping only when a tab has more groups than palette entries; expect no more than three or four groups per tab under the current design.
  - [x] 4.3 Render each selector-heading title and its subgroup buttons with matching identifiable color treatment: heading text and a distinct matching button-outline color.
  - [x] 4.4 Keep button label text legible, selected state distinct, hover state distinct, and headings/descriptions non-clickable and non-focusable.
  - [x] 4.5 Add regression tests for deterministic palette assignment, loop behavior, heading/button color pairing, selection contrast, and preserved touch-target size.

- [x] 5. Build the Comrades NPC/Comrade mock chat selector group.
  - [x] 5.1 Add heading `NPCs` with button `Librarian`.
  - [x] 5.2 Add heading `Comrades` with buttons in this order: `Self`, `Zvyo`, `Iskra`, `Puck`, `Kilo`, `Glitchi`, `Pico`, `Neon`, `Sprout`, `Zephyr`, `Bina`, `Lumen`, `Eco`, `Mira`, `Veda`, `Aura`.
  - [x] 5.3 Make every listed person open a generic static chat conversation in the content panel.
  - [x] 5.4 Add an editable EbitenUI text-entry field and Send button at the bottom of every mock chat content panel; neither may send, persist, or otherwise act on replies yet.
  - [x] 5.5 Validate controller/keyboard/touch navigation skips headings, selects people, reaches the reply field and Send button, and retains narrow-screen Back behavior.

- [x] 6. Build the Projects mock selector groups after the content-panel discussion is approved.
  - [x] 6.1 Add heading `Projects` with description `Open-ended workspaces for any kind of work.` and buttons in this order: `Solving Incompleteness`, `Invent Quantum AI`, `Solve Unification`, `Mechanical Computers`.
  - [x] 6.2 Add heading `Pipelines` with description `Project-oriented workspaces with a Cluster-task entry point.` and buttons in this order: `What's in the news?`, `What's happening tonight?`, `Oppo Analysis`.
  - [x] 6.3 Implement the approved project content-panel mockup for every project record with in-panel `Git` and `Chat` tabs, without filesystem, Git, network, or durable-state effects.
  - [x] 6.4 Implement the approved Pipeline content-panel mockup for every pipeline record: the same mock Git/Chat workspace shape as a Project, plus static entry-point declarations and a visible no-op Run button. Do not add YAML parsing, scheduling, routing, workflow execution, or durable-state effects.
  - [x] 6.5 Preserve selector state and narrow-screen Back behavior independently for Projects and Pipelines.

- [x] 7. Build the Cluster device, routing, task, and inference mock selector groups.
  - [x] 7.1 Add heading `Cluster Devices` with buttons in this order: `Pixel 2 AI Server (4gb ram)`, `Trash Can Mac Pro (256gb ram)`, `(Add New Device)`.
  - [x] 7.2 Add heading `Routing` with description `How your cluster prioritizes different kinds of tasks.` and buttons in this order: `Chat Pool (High priority)`, `Project Pool (Medium Priority)`, `Comrade Pool (Low Priority)`, `Research Pool (Last Priority)`.
  - [x] 7.3 Add heading `Tasks` with description `Things your cluster needs to do at certain times or for certain events.` and buttons in this order: `Every Hour`, `Every Day`, `Every Week`, `Every Month`, `Webhooks`.
  - [x] 7.4 Add heading `Inference Types` with description `Which devices in your cluster are set up for each kind of inference.` and buttons in this order: `Text Generation`, `Image Generation`, `Text-to-Speech`, `Speech-to-Text`, `Video Generation`.
  - [x] 7.5 Define truthful, static content-panel mockups for each device, pool, schedule/webhook, and inference-type selector without adding enrollment, route execution, scheduler, webhook, or inference behavior.
  - [x] 7.6 Keep `(Add New Device)` visibly disabled or a no-op mock until a later approved enrollment flow exists.

- [x] 8. Build the Research mock selector groups.
  - [x] 8.1 Add heading `Your Research` with description `The total compute power you've donated to each type of research.` and buttons `Curing Cancer (2.4 pflops)` and `Finding Aliens (6.2 gflops)`.
  - [x] 8.2 Add heading `Other Research Opportunities` with button `Drug Research`, button `Einstein@Home` with description `Analyze gravity waves to help find new neutron stars.`, and button `NFS@Home` with description `Find new factorizations of large integers.`.
  - [x] 8.3 Define truthful static research content panels showing mock contribution/context data without BOINC enrollment, scheduling, or execution.

- [x] 9. Verify, document, and publish.
  - [x] 9.1 Add focused HUD tests for the specified selector data, exact ordering, descriptions, disabled mock controls, top-tab order, and palette groups.
  - [x] 9.2 Add focused GUI tests for selector/content-panel rendering, non-focusable headings, color pairing, chat reply field placement, text wrapping, touch targets, narrow Back navigation, and selected/hover contrast.
  - [x] 9.3 Capture desktop and Android screenshots for Comrades/NPCs, Projects/Projects, Projects/Pipelines, Cluster devices/routing/tasks/inference, and Research selector panels.
  - [x] 9.4 Run focused tests, `make test`, GUI compile, documentation/code-size checks, `git diff --check`, and plan-index validation.
  - [x] 9.5 Rebuild and install the Android APK when the connected target is explicitly authorized; verify the updater-retrieved artifact matches the tested build.
  - [x] 9.6 Mark corresponding TODO entries complete only after approved mockups and visual evidence pass.
  - [x] 9.7 Update the plan, index, and journal; confirm no Salvagecore files or unapproved user-authored edits are staged; review pending downtime reports.
  - [x] 9.8 Commit and push directly to `main` after the user-approved checkpoint summary; do not open a pull request.

## Scope Boundaries

- This is HUD presentation work, not implementation of real chats, messaging, projects, Git operations, device enrollment, task scheduling, webhooks, routing, inference execution, BOINC, or persistence.
- Do not modify Settings content while replacing generic placeholder content elsewhere.
- Keep the approved Comrades ownership and Project/Pipeline presentation boundaries; defer only YAML/schema and real execution details.
- Preserve the existing tab-strip drag, selection, and responsive selector/content-panel behavior.

## Approved Plan Revision: 2026-07-18 Blank Initial Content Panels

- [x] 10. Require explicit selector-panel choice before displaying a content panel.
  - [x] 10.1 Represent the absence of a selected selector item without falling back to the first selectable item.
  - [x] 10.2 Render a blank content panel on initial tab entry at desktop and narrow widths; do not check any selector button by default.
  - [x] 10.3 Preserve explicit selection, Back behavior, and selected content after a user chooses an item.
  - [x] 10.4 Add focused HUD/GUI regression coverage and capture Android evidence of blank initial Comrades, Projects, Cluster, and Research content panels.

## Completion Record

- 2026-07-19: The user completed on-device review and confirmed that all requested mockup, palette, selector/content-panel, tab-order, and blank-initial-panel behaviors look correct. Automated HUD/GUI tests, repository tests, GUI compilation, documentation, code-size checks, Android installation, and rendered tablet evidence were completed during the checkpoint. This acceptance closes the remaining implementation and validation checklist items.
