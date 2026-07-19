# HUD Package

This package owns Apparat's controller-first HUD view model.

The HUD model is intentionally independent from Ebitengine, EbitenUI, SQLite, network transports, inference runtimes, and task execution. GUI adapters render snapshots from this package and dispatch named actions back into the shell.

## Phase 4 Scope

Phase 4 turns the mock HUD into a usable local five-tab shell before backend networking work.

The canonical tab order is:

1. Comrades
2. Projects
3. Research
4. Cluster
5. Settings

Each tab is represented by data rather than a hard-coded visual strip. `TabDescriptor` records stable IDs, labels, glyph slots, accessibility labels, visibility, enabled state, and future badge/status metadata. This lets the GUI render tabs across the top today and later realign them into side rails or responsive layouts without rewriting tab content.

## Configuration Manager

`DefaultConfigManager` is the temporary Phase 4 configuration boundary.

It returns hard-coded defaults for:

- Tab placement, density, label mode, and default selected tab.
- Named key/controller/pointer bindings.
- Dark theme, accent color, UI scale, font size, spacing, contrast, and focus ring defaults.
- Interaction behavior such as push-to-talk mode, repeat delay, scroll speed, confirmations, landing tab, and remembered selections.
- Notification categories, toast defaults, mute/volume, and quiet-hours placeholder.
- Diagnostic visibility and log detail defaults.
- Default views for each top-level tab.
- Privacy and safety posture, including no sharing by default.

The field names and value shapes should remain compatible with a future SQLite-backed user configuration table. Do not scatter literals for user-editable behavior through rendering or input code; add them to the configuration manager first.

## Actions And Bindings

Input is represented as named actions such as `previous_tab`, `next_tab`, `push_to_talk`, `cancel_recording`, focus movement, activation, back, context menu, command palette, scrolling, and collection navigation.

Default bindings preserve the accepted controls:

- `L1` / `R1` and `Ctrl+PageUp` / `Ctrl+PageDown` switch tabs.
- `Alt+1` through `Alt+5` select canonical tabs directly.
- `R2` and right `Ctrl` are push-to-talk.
- `Escape` cancels held recording.
- Pointer wheel, pointer drag, touch drag, right-stick scrolling, `PageUp`, and `PageDown` are tracked as scroll defaults.

Bindings are hard-coded for Phase 4 but must be treated as future user-editable settings.

## Tab Content Boundaries

Phase 4 tab content is local, mocked, or placeholder-only:

- Comrades explains future friend chat and comrade queues.
- Projects shows mock chats, files, artifacts, Git status, drafts, transaction concepts, and Pipeline-building detail.
- Research explains future BOINC delegation and validation gameplay.
- Cluster shows local diagnostics, mock device capabilities, Routing, and Tasks. Tasks remains a selector-panel item whose content panel shows placeholder schedules, webhooks, events, approvals, and run history.
- Settings shows runtime/config diagnostics, hard-coded HUD settings, command hints, and verification hints.

Backend-dependent controls must be disabled or clearly marked future until their storage, transport, queue, or execution systems exist.

The Phase 4 scenario data is intentionally fictional and visibly marked `mock`, `planned`, or disabled. It exists to exercise multi-page body/list/detail scrolling and does not represent persisted projects, enrolled devices, trusted comrades, live services, or scheduled tasks.

## Body Layout Patterns

Tab bodies use structured layout patterns rather than free-form text placement:

- Settings is a vertical list of fieldsets. Each fieldset owns its title, explanation, and rows or controls. The temporary `Updates` fieldset stays first while update-in-place is a high-priority Phase 5 test surface; ordinary new Settings groups are appended as fieldsets.
- Comrades, Projects, and Cluster use a selector/content-panel structure. The selector panel lists relevant objects or semantic headings; the content panel shows only the selected-object context. Headings are non-selectable and non-focusable, and may include a short presentation-only description beneath their title. Cluster's Routing and Tasks selectors and Projects' Pipelines selector own grouped nested detail sections without becoming top-level tabs.
- Research can use fieldsets while it remains review/placeholder content, then move to master-detail when selectable research projects exist.
- Native platform controls must correspond to a reserved HUD element and stay hidden outside the owning tab.
- List rows, fieldset rows, buttons, and form controls must keep touch-first target sizing comparable to tab buttons.
- Text and input-like controls are block-level elements. Text wraps, truncates, clips, or scrolls inside the owning fieldset or panel instead of drawing over neighboring content.
- Every body/list/detail viewport starts below the fixed top tab row and ends above the bottom diagnostics/status area. Wheel and vertical drag/touch input scroll the innermost owning viewport; tab-strip horizontal swipes remain separate.
