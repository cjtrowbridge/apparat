//go:build gui

// Verifies checks selector-panel geometry and selection behavior.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"runtime"
	"strings"
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
	for i, s := range cluster.Sections {
		if s.Title == "Every Hour" {
			game.selectSection(cluster.ID(), i)
			break
		}
	}
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
		button, _ := game.sectionButton(tc.tab, tc.section, 0)
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
	if len(tab.Sections) != 4 || tab.Sections[1].Title != "Clone remote repo" || tab.Sections[2].Title != "Init new repo" || tab.Sections[3].Title != "Import local repo" {
		t.Fatal("unexpected rows in empty registry")
	}
	game.projects.rows = []domain.Project{
		{ID: "a", Name: "vibe-modeling", Available: true, Branch: "main", Actions: []string{"Rebuild stale CAD designs"}},
		{ID: "b", Name: "plain", Available: true},
	}
	tab = game.projectTab(hud.Tab{})
	if !tab.Sections[len(projectEntrySections)].HasConfiguredActions || tab.Sections[len(projectEntrySections)].ContentKind != hud.ContentProject || tab.Sections[len(projectEntrySections)+1].ContentKind != hud.ContentProject || tab.Sections[len(projectEntrySections)].SelectorColor == tab.Sections[len(projectEntrySections)+1].SelectorColor {
		t.Fatal("action decoration does not follow real checkout state")
	}
}

func TestProjectSelectorShowsFullCheckoutLocation(t *testing.T) {
	game := NewGame()
	game.width = 960
	game.splitWidth = minSplitWidth
	tab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[1]
	for _, tc := range []struct{ location, want string }{
		{"/src/two  words/repo", "local:///src/two  words/repo"},
		{"content://com.android.externalstorage.documents/tree/primary%3ADocuments/document/primary%3ADocuments%2Frepo", "local://~/Documents/repo"},
		{"content://com.android.externalstorage.documents/tree/1234-5678/document/1234-5678%3ADocuments%2Frepo", "local://1234-5678/Documents/repo"},
		{"content://example/tree/1/document/42", "local://example/42/repo"},
	} {
		game.projects.rows = []domain.Project{{ID: "one", Name: "repo", Location: tc.location, Available: true, Branch: "main", Revision: "0123456789ab", Actions: []string{"Build"}}}
		section := game.projectTab(tab).Sections[len(projectEntrySections)]
		button, subtitle := game.sectionButton(tab, section, len(projectEntrySections))
		if subtitle == nil || strings.ReplaceAll(subtitle.Label, "\n", "") != tc.want+"main @ 0123456789ab" {
			t.Fatal("missing checkout path", tc.want)
		}
		button.GetWidget().SetTheme(game.theme)
		button.Validate()
		_, height := button.PreferredSize()
		width, subHeight := subtitle.PreferredSize()
		if width > int(subtitle.MaxWidth) || height < subHeight+24 {
			t.Fatalf("path exceeds its row: %dx%d, row height %d", width, subHeight, height)
		}
	}
}

func TestProjectDataAliasMatchesDirectoryBoundary(t *testing.T) {
	root := "/arbitrary/install/node/data"
	for _, tc := range []struct{ path, want string }{
		{root + "/repos/owner/repo", "local://apparat/repos/owner/repo"},
		{root + "-other/repo", "local:///arbitrary/install/node/data-other/repo"},
		{root, "local://apparat"},
	} {
		got := projectListLocation(domain.Project{Location: tc.path}, root)
		if got != tc.want {
			t.Errorf("projectListLocation(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestProjectEntryFormsOnlyOfferTheirOwnActions(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.projects.store = layoutRegistry{}
	actions := []string{"Preview clone", "Clone", "Initialize repo", "Import repo", "Save registration"}
	selectTabThroughStrip(t, g, hud.TabProjects)
	g.Layout(360, 720)
	for i, section := range projectEntrySections[1:] {
		g.selectSection(hud.TabProjects, i+1)
		renderNavigation(g)
		if !g.detailOpen[hud.TabProjects] || g.inWorkspace() {
			t.Fatalf("%s did not open its detail pane", section.Title)
		}
		form := g.buildSectionContainer(section)
		form.GetWidget().SetTheme(g.theme)
		form.Validate()
		for _, label := range actions {
			want := section.ContentKind == cloneRepoForm && (label == "Preview clone" || label == "Clone") ||
				section.ContentKind == initRepoForm && label == "Initialize repo" ||
				section.ContentKind == importRepoForm && label == "Import repo"
			found := findButtonByLabel(form, label) != nil
			if found != want {
				t.Errorf("%s shows %q = %t, want %t", section.Title, label, found, want)
			}
		}
		fields := 0
		for _, target := range form.GetFocusers() {
			_, ok := target.(*widget.TextInput)
			if ok {
				fields++
			}
		}
		expected := map[hud.ContentKind]int{cloneRepoForm: 3, initRepoForm: 2, importRepoForm: 1}
		if runtime.GOOS == "android" {
			expected[cloneRepoForm] = 2
		}
		if fields != expected[section.ContentKind] {
			t.Errorf("%s has %d fields, want %d", section.Title, fields, expected[section.ContentKind])
		}
	}
}

func TestClusterDirectoryUsesSavedWorkers(t *testing.T) {
	g := NewGame()
	defer g.Close()
	g.runtimeInfo.Queries = &projectionSource{}
	g.serviceProjection.Devices = []domain.ClusterDevice{{Host: "saved-worker", Fingerprint: "public-key", Address: "192.0.2.1:4000", Cluster: "home"}}
	base := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[2]
	tab := g.presentedTab(base)
	for i, s := range tab.Sections {
		if s.ActionID == "add-new-device" {
			if tab.Sections[i+1].ActionID != "worker-public-key" || tab.Sections[i+1].Title != "saved-worker" || tab.Sections[i+2].Title != "Routing" {
				t.Fatal("incorrect device rows")
			}
			continue
		}
		if s.ActionID == "pixel-2-ai-server" || s.ActionID == "trash-can-mac-pro" {
			t.Fatal("mock device remained")
		}
	}
	actions := g.queryActions("open saved-worker", true)
	if !actions.Exact || len(actions.Choices) != 1 {
		t.Fatal("missing device action")
	}
	g.serviceProjection.Devices = nil
	g.invokeAction(actions.Choices[0].Address)
	if len(g.queryActions("open saved-worker", true).Choices) != 0 {
		t.Fatal("stale device action")
	}
}
