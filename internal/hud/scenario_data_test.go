package hud

import "testing"

func TestDefaultTabsUseMockupSelectorData(t *testing.T) {
	for _, tab := range DefaultTabs(DefaultConfigManager{}.Config()) {
		if tab.ID() == TabSettings {
			continue
		}
		headings, colored := 0, 0
		for _, section := range tab.Sections {
			if section.IsSelectorHeading() {
				headings++
			}
			if section.SelectorColor != "" {
				colored++
			}
		}
		if headings == 0 || colored != len(tab.Sections) {
			t.Fatalf("%s headings=%d colored=%d sections=%d", tab.ID(), headings, colored, len(tab.Sections))
		}
	}
}
