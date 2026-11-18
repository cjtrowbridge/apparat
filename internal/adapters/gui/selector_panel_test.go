//go:build gui

// Verifies checks selector-panel geometry and selection behavior.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
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

func TestSelectorButtonsFitNarrowPaneAndReserveActionIcon(t *testing.T) {
	game := NewGame()
	game.width = 960
	game.splitWidth = minSplitWidth
	tabs := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())
	cases := []struct {
		tab     hud.Tab
		section hud.Section
	}{
		{tabs[2], hud.Section{Title: "Project Pool (Medium priority)"}},
		{tabs[1], hud.Section{Title: "unbroken-repository-name-with-many-characters", ContentKind: hud.ContentPipeline}},
	}
	for _, tc := range cases {
		button := game.sectionButton(tc.tab, tc.section, 0)
		button.GetWidget().SetTheme(game.theme)
		button.Validate()
		width, height := button.PreferredSize()
		available := game.splitWidth - 2*selectorListInset
		if width > available || height <= tabHeight {
			t.Errorf("%q preferred size = %dx%d, want width <= %d and height > %d", tc.section.Title, width, height, available, tabHeight)
		}
	}
}

func TestLocalProjectTabIncludesOnlyRegisteredCheckouts(t *testing.T) {
	game := NewGameWithRuntimeInfo(RuntimeInfo{})
	defer game.Close()
	tab := game.projectTab(hud.Tab{})
	if len(tab.Sections) != 2 || tab.Sections[1].Title != "Add or clone repo" {
		t.Fatal("unexpected rows in empty registry")
	}
	game.projects.rows = []domain.Project{
		{ID: "a", Name: "vibe-modeling", Available: true, Branch: "main", Actions: []string{"Rebuild stale CAD designs"}},
		{ID: "b", Name: "plain", Available: true},
	}
	tab = game.projectTab(hud.Tab{})
	if tab.Sections[2].ContentKind != hud.ContentPipeline || tab.Sections[3].ContentKind != hud.ContentProject || tab.Sections[2].SelectorColor == tab.Sections[3].SelectorColor {
		t.Fatal("action decoration does not follow real checkout state")
	}
}
