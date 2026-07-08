package hud

func DefaultTabs(config HUDConfig) []Tab {
	return []Tab{
		comradesTab(config),
		projectsTab(),
		researchTab(),
		clusterTab(),
		routingTab(),
		tasksTab(),
		settingsTab(config),
	}
}

func comradesTab(config HUDConfig) Tab {
	return tab(config, TabComrades, "Future friend chat and owner-controlled shared compute.", []Section{{Title: "Future Capabilities", Rows: []Row{{Label: "Real-friend chat", Detail: "planned secure conversations", Future: true}, {Label: "Comrade queues", Detail: "low-priority inference access when owner capacity is idle", Future: true}, {Label: "Sharing posture", Detail: config.Privacy.SharingDefault, Disabled: true}}}, {Title: "Controls", Rows: []Row{{Label: "Grant access", Detail: "disabled until sharing backend exists", Disabled: true, Future: true}, {Label: "Revoke access", Detail: "planned audited action", Disabled: true, Future: true}}}})
}

func projectsTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabProjects, "Local project workspace preview with mock chats, files, artifacts, and Git state.", []Section{{Title: "Projects", Rows: []Row{{Label: "apparat", Detail: "selected local project mock"}, {Label: "offline draft", Detail: "transaction concept only", Future: true}}}, {Title: "Workspace", Rows: []Row{{Label: "Chat preview", Detail: "architecture sketch"}, {Label: "Files", Detail: "README.md, ROADMAP.md placeholders"}, {Label: "Artifacts", Detail: "mock transcript"}, {Label: "Git", Detail: "clean placeholder"}}}})
}

func researchTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabResearch, "Future BOINC delegation and validated public-interest compute.", []Section{{Title: "Validated Research", Rows: []Row{{Label: "Project catalog", Detail: "placeholder validated BOINC projects", Future: true}, {Label: "Budget", Detail: "opt-in only", Disabled: true}, {Label: "Schedule", Detail: "future quiet-hours compute", Future: true}, {Label: "Gameplay validation", Detail: "planned contribution proof", Future: true}}}, {Title: "Execution", Rows: []Row{{Label: "Start BOINC work", Detail: "disabled until Research phase", Disabled: true, Future: true}}}})
}

func clusterTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabCluster, "Local diagnostics and mock device capability inventory.", []Section{{Title: "Local Runtime", Rows: []Row{{Label: "Identity", Detail: "classified by doctor"}, {Label: "Runtime root", Detail: "shown in Settings"}, {Label: "last_run.log", Detail: "reset on each start"}}}, {Title: "Mock Devices", Rows: []Row{{Label: "steamdeck", Detail: "GUI console"}, {Label: "worker", Detail: "text_generation, speech_to_text"}, {Label: "workstation", Detail: "image_generation, video_generation, research_boinc"}}}})
}

func routingTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabRouting, "Typed workload queues, compatibility filters, and fallback routing.", []Section{{Title: "Workload Classes", Rows: []Row{{Label: "Text generation", Detail: "OpenAI-compatible, Ollama, llama.cpp"}, {Label: "Image generation", Detail: "future adapter contract"}, {Label: "Video generation", Detail: "future long-running adapter"}, {Label: "Speech-to-text", Detail: "whisper.cpp reference"}, {Label: "Text-to-speech", Detail: "service-backed first"}, {Label: "BOINC", Detail: "research compute, not model inference"}}}, {Title: "Queues", Rows: []Row{{Label: "Priority", Detail: "owner work above comrade/research"}, {Label: "Fallback", Detail: "compatible devices only"}}}})
}

func tasksTab() Tab {
	return tab(DefaultConfigManager{}.Config(), TabTasks, "Future schedules, webhooks, event rules, approvals, and run history.", []Section{{Title: "Task Types", Rows: []Row{{Label: "Scheduled task", Detail: "placeholder", Future: true}, {Label: "Webhook", Detail: "placeholder", Future: true}, {Label: "Signal-driven", Detail: "future adapter", Future: true}, {Label: "Manual approval", Detail: "planned safety gate", Future: true}}}, {Title: "Controls", Rows: []Row{{Label: "Create task", Detail: "disabled until durable task storage exists", Disabled: true, Future: true}}}})
}

func settingsTab(config HUDConfig) Tab {
	return tab(config, TabSettings, "Local configuration defaults, runtime paths, diagnostics, and verification hints.", []Section{{Title: "HUD Configuration", Rows: []Row{{Label: "Tab placement", Detail: string(config.TabView.Placement)}, {Label: "Theme", Detail: string(config.Display.Theme)}, {Label: "Scale", Detail: "1.0"}, {Label: "Font size", Detail: "18"}, {Label: "Settings persistence", Detail: "hard-coded now, future SQLite user settings", Future: true}}}, {Title: "Bindings", Rows: []Row{{Label: "Previous tab", Detail: "L1, Ctrl+PageUp"}, {Label: "Next tab", Detail: "R1, Ctrl+PageDown"}, {Label: "Direct tabs", Detail: "Alt+1..Alt+7"}, {Label: "Push-to-talk", Detail: "R2, right Ctrl"}}}, {Title: "Diagnostics", Rows: []Row{{Label: "Doctor", Detail: "run --doctor"}, {Label: "Smoke test", Detail: "run --smoke-test"}, {Label: "Logs", Detail: "last_run.log and logs/apparat.jsonl"}, {Label: "Verify", Detail: "make verify && make check-docs"}}}})
}

func tab(config HUDConfig, id TabID, summary string, sections []Section) Tab {
	descriptors := config.TabView.Tabs
	for _, descriptor := range descriptors {
		if descriptor.ID == id {
			return Tab{Descriptor: descriptor, Summary: summary, Sections: sections}
		}
	}
	return Tab{Descriptor: TabDescriptor{ID: id, Label: string(id), Visible: true, Enabled: true}, Summary: summary, Sections: sections}
}
