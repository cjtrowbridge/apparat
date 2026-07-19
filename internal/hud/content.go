package hud

func DefaultTabs(config HUDConfig) []Tab {
	return []Tab{
		comradesTab(config),
		projectsTab(),
		researchTab(),
		clusterTab(),
		settingsTab(config),
	}
}

func comradesTab(config HUDConfig) Tab {
	return tab(config, TabComrades, "Future friend chat and owner-controlled shared compute.", []Section{{Title: "Future Capabilities", Description: "Comrades will become a two-pane thread list and conversation surface.", Rows: []Row{{Label: "Real-friend chat", Detail: "planned secure conversations", Future: true}, {Label: "Comrade queues", Detail: "low-priority inference access when owner capacity is idle", Future: true}, {Label: "Sharing posture", Detail: config.Privacy.SharingDefault, Disabled: true}}}, {Title: "Controls", Description: "Backend-dependent sharing actions stay disabled until durable permissions exist.", Rows: []Row{{Label: "Grant access", Detail: "disabled until sharing backend exists", Disabled: true, Future: true}, {Label: "Revoke access", Detail: "planned audited action", Disabled: true, Future: true}}}})
}

func projectsTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabProjects, "Local project workspace preview with mock chats, files, artifacts, Git state, and pipelines.", []Section{{Title: "Projects", Description: "This will become a project list with selected project context beside it.", Rows: []Row{{Label: "apparat", Detail: "selected local project mock"}, {Label: "offline draft", Detail: "transaction concept only", Future: true}}}, {Title: "Workspace", Description: "Selected project context groups chats, files, artifacts, and safe Git status.", Rows: []Row{{Label: "Chat preview", Detail: "architecture sketch"}, {Label: "Files", Detail: "README.md, ROADMAP.md placeholders"}, {Label: "Artifacts", Detail: "mock transcript"}, {Label: "Git", Detail: "clean placeholder"}}}, {Title: "Pipelines", Description: "Mock pipeline-building tasks stay future-facing until durable workflows exist.", DetailSections: pipelineSections()}})
}

func researchTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabResearch, "Future BOINC delegation and validated public-interest compute.", []Section{{Title: "Validated Research", Description: "Research starts as review content and later becomes selectable project catalog work.", Rows: []Row{{Label: "Project catalog", Detail: "placeholder validated BOINC projects", Future: true}, {Label: "Budget", Detail: "opt-in only", Disabled: true}, {Label: "Schedule", Detail: "future quiet-hours compute", Future: true}, {Label: "Gameplay validation", Detail: "planned contribution proof", Future: true}}}, {Title: "Execution", Description: "Execution controls stay grouped with their validation context.", Rows: []Row{{Label: "Start BOINC work", Detail: "disabled until Research phase", Disabled: true, Future: true}}}})
}

func clusterTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabCluster, "Local diagnostics, mock device capability inventory, typed workload routing, and future task automation.", []Section{selectorHeading("Devices", "Local diagnostics and capability inventory."), {Title: "Local Runtime", Description: "Runtime diagnostics summarize this device before real cluster enrollment exists.", Rows: []Row{{Label: "Identity", Detail: "classified by doctor"}, {Label: "Runtime root", Detail: "shown in Settings"}, {Label: "last_run.log", Detail: "reset on each start"}}}, {Title: "Mock Devices", Description: "Cluster will become a device list with selected-device context.", Rows: []Row{{Label: "steamdeck", Detail: "GUI console"}, {Label: "worker", Detail: "text_generation, speech_to_text"}, {Label: "workstation", Detail: "image_generation, video_generation, research_boinc"}}}, selectorHeading("Operations", "Routing and future task automation."), {Title: "Routing", Description: "Typed workload queues, compatibility filters, and fallback routing belong to Cluster.", DetailSections: routingSections()}, tasksSection()})
}

func routingSections() []Section {
	return append([]Section{{Title: "Workload Classes", Description: "Routes are grouped by explicit workload classes rather than one generic compute flag.", Rows: []Row{{Label: "Text generation", Detail: "OpenAI-compatible, Ollama, llama.cpp"}, {Label: "Image generation", Detail: "future adapter contract", Future: true}, {Label: "Video generation", Detail: "future long-running adapter", Future: true}, {Label: "Speech-to-text", Detail: "whisper.cpp reference"}, {Label: "Text-to-speech", Detail: "service-backed first", Future: true}, {Label: "BOINC", Detail: "research compute, not model inference"}}}, {Title: "Queues", Description: "Routing will become a queue list with selected-route context and controls.", Rows: []Row{{Label: "Priority", Detail: "owner work above comrade/research"}, {Label: "Fallback", Detail: "compatible devices only"}}}}, routingScenarioSections()...)
}

func pipelineSections() []Section {
	return []Section{{Title: "Pipeline Builder", Description: "Mock composition stages for a future durable workflow editor.", Rows: []Row{{Label: "Trigger", Detail: "manual, schedule, webhook, or event", Future: true}, {Label: "Inputs", Detail: "typed project, chat, artifact, or form data", Future: true}, {Label: "Steps", Detail: "mock ordered transforms and typed workload submissions", Future: true}}}, {Title: "Safety And Routing", Description: "Future pipeline execution must keep approval and compatible destinations explicit.", Rows: []Row{{Label: "Approval gate", Detail: "mock owner confirmation before side effects", Disabled: true, Future: true}, {Label: "Workload route", Detail: "mock Cluster routing profile selection", Future: true}, {Label: "Fallback", Detail: "mock compatible queue order", Future: true}}}, {Title: "Run History", Description: "Mock durable execution records until Tasks owns real workflows.", Rows: []Row{{Label: "Draft pipeline", Detail: "mock editable, not persisted", Future: true}, {Label: "Validation run", Detail: "mock waiting for owner", Future: true}, {Label: "Published run", Detail: "mock disabled until durable task storage", Disabled: true, Future: true}}}}
}

func selectorHeading(title, description string) Section {
	return Section{Title: title, Description: description, SelectorKind: SelectorHeading}
}

func tasksSection() Section {
	return Section{Title: "Tasks", Description: "Future schedules, webhooks, event rules, approvals, and run history belong to Cluster operations.", DetailSections: append([]Section{{Title: "Task Types", Description: "Tasks will become a run list with selected-task history and approval context.", Rows: []Row{{Label: "Scheduled task", Detail: "placeholder", Future: true}, {Label: "Webhook", Detail: "placeholder", Future: true}, {Label: "Signal-driven", Detail: "future adapter", Future: true}, {Label: "Manual approval", Detail: "planned safety gate", Future: true}}}, {Title: "Controls", Description: "Creation controls stay disabled until durable task storage exists.", Rows: []Row{{Label: "Create task", Detail: "disabled until durable task storage exists", Disabled: true, Future: true}}}}, taskScenarioSections()...)}
}

func settingsTab(config HUDConfig) Tab {
	return tab(config, TabSettings, "Local configuration defaults, runtime paths, diagnostics, and verification hints.", []Section{{Title: "Updates", Description: "Download the latest tracked APK and let Android ask before installing it."}, {Title: "HUD Configuration", Description: "These defaults are hard-coded now and move to durable settings later.", Rows: []Row{{Label: "Tab placement", Detail: string(config.TabView.Placement)}, {Label: "Theme", Detail: string(config.Display.Theme)}, {Label: "Scale", Detail: "1.0"}, {Label: "Font size", Detail: "18"}, {Label: "Settings persistence", Detail: "hard-coded now, future SQLite user settings", Future: true}}}, {Title: "Bindings", Description: "Controller, keyboard, mouse, and touch bindings dispatch the same HUD actions.", Rows: []Row{{Label: "Previous tab", Detail: "L1, Ctrl+PageUp"}, {Label: "Next tab", Detail: "R1, Ctrl+PageDown"}, {Label: "Direct tabs", Detail: "Alt+1..Alt+5"}, {Label: "Push-to-talk", Detail: "R2, right Ctrl"}}}, {Title: "Diagnostics", Description: "Local checks and logs support troubleshooting without backend services.", Rows: []Row{{Label: "Doctor", Detail: "run --doctor"}, {Label: "Smoke test", Detail: "run --smoke-test"}, {Label: "Logs", Detail: "last_run.log and logs/apparat.jsonl"}, {Label: "Verify", Detail: "make verify && make check-docs"}}}})
}

func tab(config HUDConfig, id TabID, summary string, sections []Section) Tab {
	descriptors := config.TabView.Tabs
	for _, descriptor := range descriptors {
		if descriptor.ID == id {
			return Tab{Descriptor: descriptor, Summary: summary, Sections: append(sections, scenarioSections(id)...)}
		}
	}
	return Tab{Descriptor: TabDescriptor{ID: id, Label: string(id), Visible: true, Enabled: true}, Summary: summary, Sections: append(sections, scenarioSections(id)...)}
}
