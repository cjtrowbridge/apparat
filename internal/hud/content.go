// Builds the five mock tab content models.
// Uses package-local contracts and Go standard-library types.
package hud

import "runtime"

func DefaultTabs(config HUDConfig) []Tab {
	tabs := []Tab{
		comradesTab(config),
		projectsTab(),
		clusterTab(),
		researchTab(),
		tab(config, "stmpck", "Offline games", nil),
		settingsTab(config),
	}
	ordered := make([]Tab, 0, len(tabs))
	for _, descriptor := range settingsLast(config.TabView.Tabs) {
		for _, entry := range tabs {
			if entry.ID() == descriptor.ID {
				ordered = append(ordered, entry)
			}
		}
	}
	return ordered
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
	return tab(DefaultConfigManager{}.Config(), TabCluster, "Discover local Apparat hosts and browse example cluster configuration.", clusterMockSections())
}

func selectorHeading(title, description, selectorColor string) Section {
	return Section{Title: title, Description: description, SelectorKind: SelectorHeading, SelectorColor: selectorColor}
}

func settingsTab(config HUDConfig) Tab {
	return tab(config, TabSettings, "Display, input, and diagnostics for this device.", []Section{{ActionID: "updates", Title: "Updates", Description: updateDescription(runtime.GOOS)},
		{ActionID: "personal-identity", Title: "Personal identity", Description: "Manage the screen name associated with your user key."},
		{ActionID: "api-configuration", Title: "API configuration", Description: "Hosting requires an explicit network and interface grant."},
		{ActionID: "command-directory", Title: "Voice Commands", Description: "Find the words to open a tab, project, conversation, or game. You can also open a destination directly from the directory."},
		{ActionID: "hud-configuration", Title: "Appearance"},
		{ActionID: "navigation-layout", Title: "Navigation and Window"},
		{ActionID: "speech-recognition", Title: "Speech Recognition", Description: "Choose the offline model that transcribes push-to-talk audio."},
		{ActionID: "bindings", Title: "Bindings", Description: "Controller, keyboard, mouse, and touch bindings dispatch the same HUD actions.", Rows: []Row{{Label: "Previous tab", Detail: "L1, Ctrl+PageUp"}, {Label: "Next tab", Detail: "R1, Ctrl+PageDown"}, {Label: "Direct tabs", Detail: "Alt+1..Alt+6"}, {Label: "Push-to-talk", Detail: "L2, right Ctrl"}}}, {ActionID: "diagnostics", Title: "Diagnostics", Description: "Local checks and logs support troubleshooting without backend services.", Rows: []Row{{Label: "Doctor", Detail: "run --doctor"}, {Label: "Smoke test", Detail: "run --smoke-test"}, {Label: "Logs", Detail: "last_run.log and logs/apparat.jsonl"}, {Label: "Verify", Detail: "make verify && make check-docs"}}}})
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

func updateDescription(platform string) string {
	switch platform {
	case "android":
		return "Check for a new Android release. Android will ask before installing a verified update."
	case "linux":
		return "Save your work first. If an update is available, Apparat installs it and restarts automatically."
	case "windows":
		return "To update on Windows, replace Apparat.exe with the latest Windows build."
	case "darwin":
		return "To update on macOS, replace the Apparat executable with the latest macOS build."
	default:
		return "Update Apparat using the latest build for this platform."
	}
}
