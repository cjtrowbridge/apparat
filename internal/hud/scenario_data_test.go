package hud

import "testing"

func TestDefaultTabsIncludeMultiPageScenarioData(t *testing.T) {
	for _, tab := range DefaultTabs(DefaultConfigManager{}.Config()) {
		if len(tab.Sections) < 8 {
			t.Fatalf("%s sections = %d, want at least 8", tab.ID(), len(tab.Sections))
		}
		rows := 0
		for _, section := range tab.Sections {
			rows += len(section.Rows)
		}
		if rows < 24 {
			t.Fatalf("%s rows = %d, want at least 24", tab.ID(), rows)
		}
	}
}
