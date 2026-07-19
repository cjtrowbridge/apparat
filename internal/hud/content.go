package hud

func DefaultTabs(config HUDConfig) []Tab {
	return []Tab{
		comradesTab(config),
		projectsTab(),
		clusterTab(),
		researchTab(),
		settingsTab(config),
	}
}

func comradesTab(config HUDConfig) Tab {
	return tab(config, TabComrades, "Mock conversations with NPCs and trusted comrades; replies stay local to the input field.", comradesMockSections())
}

func projectsTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabProjects, "Mock workspaces with Git and Chat panels; pipelines add mock task entry points.", projectsMockSections())
}

func researchTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabResearch, "Mock research contribution and opportunity records; no research work is enrolled or running.", researchMockSections())
}

func clusterTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabCluster, "Mock devices, pool policy, schedules, and inference capabilities; no Cluster operation is live.", clusterMockSections())
}

func routingSections() []Section {
	return append([]Section{{Title: "Workload Classes", Description: "Routes are grouped by explicit workload classes rather than one generic compute flag.", Rows: []Row{{Label: "Text generation", Detail: "OpenAI-compatible, Ollama, llama.cpp"}, {Label: "Image generation", Detail: "future adapter contract", Future: true}, {Label: "Video generation", Detail: "future long-running adapter", Future: true}, {Label: "Speech-to-text", Detail: "whisper.cpp reference"}, {Label: "Text-to-speech", Detail: "service-backed first", Future: true}, {Label: "BOINC", Detail: "research compute, not model inference"}}}, {Title: "Queues", Description: "Routing will become a queue list with selected-route context and controls.", Rows: []Row{{Label: "Priority", Detail: "owner work above comrade/research"}, {Label: "Fallback", Detail: "compatible devices only"}}}}, routingScenarioSections()...)
}

func pipelineSections() []Section {
	return []Section{{Title: "Pipeline Builder", Description: "Mock composition stages for a future durable workflow editor.", Rows: []Row{{Label: "Trigger", Detail: "manual, schedule, webhook, or event", Future: true}, {Label: "Inputs", Detail: "typed project, chat, artifact, or form data", Future: true}, {Label: "Steps", Detail: "mock ordered transforms and typed workload submissions", Future: true}}}, {Title: "Safety And Routing", Description: "Future pipeline execution must keep approval and compatible destinations explicit.", Rows: []Row{{Label: "Approval gate", Detail: "mock owner confirmation before side effects", Disabled: true, Future: true}, {Label: "Workload route", Detail: "mock Cluster routing profile selection", Future: true}, {Label: "Fallback", Detail: "mock compatible queue order", Future: true}}}, {Title: "Run History", Description: "Mock durable execution records until Tasks owns real workflows.", Rows: []Row{{Label: "Draft pipeline", Detail: "mock editable, not persisted", Future: true}, {Label: "Validation run", Detail: "mock waiting for owner", Future: true}, {Label: "Published run", Detail: "mock disabled until durable task storage", Disabled: true, Future: true}}}}
}

func selectorHeading(title, description, selectorColor string) Section {
	return Section{Title: title, Description: description, SelectorKind: SelectorHeading, SelectorColor: selectorColor}
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
			return Tab{Descriptor: descriptor, Summary: summary, Sections: appendScenarioSections(id, sections)}
		}
	}
	return Tab{Descriptor: TabDescriptor{ID: id, Label: string(id), Visible: true, Enabled: true}, Summary: summary, Sections: appendScenarioSections(id, sections)}
}

func appendScenarioSections(id TabID, sections []Section) []Section {
	if id == TabSettings {
		return append(sections, scenarioSections(id)...)
	}
	return sections
}
