// Builds the five mock tab content models.
// Uses package-local contracts and Go standard-library types.
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

func selectorHeading(title, description, selectorColor string) Section {
	return Section{Title: title, Description: description, SelectorKind: SelectorHeading, SelectorColor: selectorColor}
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
