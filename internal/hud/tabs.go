package hud

func DefaultTabDescriptors() []TabDescriptor {
	return []TabDescriptor{
		{ID: TabComrades, Label: "Comrades", Glyph: "☭", AccessibilityLabel: "Comrades, future friend chat and shared compute", Visible: true, Enabled: true},
		{ID: TabProjects, Label: "Projects", Glyph: "⌘", AccessibilityLabel: "Projects, chats files artifacts and Git", Visible: true, Enabled: true},
		{ID: TabResearch, Label: "Research", Glyph: "⚗", AccessibilityLabel: "Research, future BOINC validated compute", Visible: true, Enabled: true},
		{ID: TabCluster, Label: "Cluster", Glyph: "◌", AccessibilityLabel: "Cluster, devices health and capabilities", Visible: true, Enabled: true},
		{ID: TabRouting, Label: "Routing", Glyph: "⇄", AccessibilityLabel: "Routing, typed workload queues", Visible: true, Enabled: true},
		{ID: TabTasks, Label: "Tasks", Glyph: "✓", AccessibilityLabel: "Tasks, schedules webhooks and automations", Visible: true, Enabled: true},
		{ID: TabSettings, Label: "Settings", Glyph: "⚙", AccessibilityLabel: "Settings, local configuration and diagnostics", Visible: true, Enabled: true},
	}
}

func tabIndexByID(tabs []Tab, id TabID) int {
	for index, tab := range tabs {
		if tab.ID() == id {
			return index
		}
	}
	return 0
}
