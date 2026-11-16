//go:build gui

// Verifies checks selector-panel geometry and selection behavior.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func TestClusterSelectorHeadingsAreTextAndTasksOwnsTheContentPanel(t *testing.T) {
	game := NewGame()
	cluster := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[2]
	game.selectSection(cluster.ID(), 0)
	{
		got := game.selectedSectionIndex(cluster)
		if got != -1 {
			t.Fatalf("heading selection resolved to %d, want no selection", got)
		}
	}
	body := game.buildMasterDetailTab(cluster)
	{
		_, ok := unwrapBounded(body).(*widget.Container)
		if !ok {
			t.Fatalf("cluster body type = %T, want *widget.Container", body)
		}
	}
	heading := game.selectorHeading(cluster.Sections[0])
	texts := collectTextNodes(heading)
	if len(texts) != 2 || texts[0].Label != "CLUSTER DEVICES" || texts[1].Label != cluster.Sections[0].Description {
		t.Fatalf("selector heading texts = %#v, want title and description", texts)
	}
	for _, text := range texts {
		if text.MaxWidth <= 0 {
			t.Fatalf("selector heading text %q MaxWidth = %.1f, want positive", text.Label, text.MaxWidth)
		}
	}
	{
		details := game.detailSections(cluster)
		if details != nil {
			t.Fatalf("initial cluster details = %#v, want blank content panel", details)
		}
	}
	game.selectSection(cluster.ID(), 1)
	{
		got := game.selectedSectionIndex(cluster)
		if got != 1 {
			t.Fatalf("explicit selection = %d, want 1", got)
		}
	}
	game.selectSection(cluster.ID(), 10)
	details := game.detailSections(cluster)
	if len(details) != 1 || details[0].Title != "Every Hour" {
		t.Fatalf("selected cluster details = %#v, want Every Hour", details)
	}
}

func TestMasterDetailTabsStartWithBlankContentPanels(t *testing.T) {
	game := NewGame()
	for _, tabData := range hud.DefaultTabs(hud.DefaultConfigManager{}.Config()) {
		if tabData.ID() == hud.TabSettings {
			continue
		}
		{
			got := game.selectedSectionIndex(tabData)
			if got != -1 {
				t.Fatalf("%s initial selection = %d, want none", tabData.ID(), got)
			}
		}
		{
			details := game.detailSections(tabData)
			if details != nil {
				t.Fatalf("%s initial details = %#v, want blank panel", tabData.ID(), details)
			}
		}
	}
}
