// Defines stable tab identities and order.
// Uses package-local contracts and Go standard-library types.
package hud

func DefaultTabDescriptors() []TabDescriptor {
	return settingsLast([]TabDescriptor{
		{ID: TabComrades, Label: "Comrades", Glyph: "☭", AccessibilityLabel: "Comrades, future friend chat and shared compute", Visible: true, Enabled: true},
		{ID: TabProjects, Label: "Projects", Glyph: "🔨", AccessibilityLabel: "Projects, chats files artifacts and Git", Visible: true, Enabled: true},
		{ID: TabCluster, Label: "Cluster", Glyph: "🖧", AccessibilityLabel: "Cluster, devices health and capabilities", Visible: true, Enabled: true},
		{ID: TabResearch, Label: "Research", Glyph: "🧪", AccessibilityLabel: "Research, future BOINC validated compute", Visible: true, Enabled: true},
		{ID: "stmpck", Label: "STMPCK", Glyph: "💉", AccessibilityLabel: "STMPCK games", Visible: true, Enabled: true},
		{ID: TabGlobalSession, Label: "Apparat Global Planning Agent", Glyph: "🧠", AccessibilityLabel: "Apparat Global Planning Agent, prompt draft", Visible: true, Enabled: true},
		{ID: TabSettings, Label: "Settings", Glyph: "⚙️", AccessibilityLabel: "Settings, local configuration and diagnostics", Visible: true, Enabled: true},
	})
}

func tabIndexByID(tabs []Tab, id TabID) int {
	for index, tab := range tabs {
		if tab.ID() == id {
			return index
		}
	}
	return 0
}

// settingsLast preserves the relative order of every other destination.
func settingsLast(tabs []TabDescriptor) []TabDescriptor {
	result := make([]TabDescriptor, 0, len(tabs))
	for _, descriptor := range tabs {
		if descriptor.ID != TabSettings {
			result = append(result, descriptor)
		}
	}
	for _, descriptor := range tabs {
		if descriptor.ID == TabSettings {
			result = append(result, descriptor)
		}
	}
	return result
}
