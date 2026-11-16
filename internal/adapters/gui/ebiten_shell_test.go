//go:build gui

// Verifies updates and draws the Ebitengine HUD shell.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/event, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestNewGameBuildsScrollableTabStripRoot(t *testing.T) {
	game := NewGame()
	root, ok := game.ui.Container.(*widget.Container)
	if !ok {
		t.Fatalf("root type = %T, want *widget.Container", game.ui.Container)
	}
	children := root.Children()
	if len(children) != 1 {
		t.Fatalf("root children = %d, want 1", len(children))
	}
	shell, ok := children[0].(*widget.Container)
	if !ok {
		t.Fatalf("root child type = %T, want *widget.Container", children[0])
	}
	layoutData, ok := shell.GetWidget().LayoutData.(widget.AnchorLayoutData)
	if !ok {
		t.Fatalf("shell layout data = %T, want widget.AnchorLayoutData", shell.GetWidget().LayoutData)
	}
	if !layoutData.StretchHorizontal || !layoutData.StretchVertical {
		t.Fatalf("shell stretch = horizontal %t vertical %t, want both true", layoutData.StretchHorizontal, layoutData.StretchVertical)
	}
	if game.tabScroll == nil {
		t.Fatal("root did not retain tab scroll container")
	}
	if game.tabButtonCount != len(game.shell.Snapshot().Tabs) {
		t.Fatalf("tab button count = %d, want %d", game.tabButtonCount, len(game.shell.Snapshot().Tabs))
	}
}

func TestActiveTabStripFollowsHUDSnapshot(t *testing.T) {
	game := NewGame()
	selectTabThroughStrip(t, game, hud.TabSettings)
	{
		id := game.shell.Snapshot().ActiveTab().ID()
		if id != hud.TabSettings {
			t.Fatalf("active tab id = %q, want settings", id)
		}
	}
	game.ensureActiveTabVisible()
	if game.tabScroll == nil {
		t.Fatal("tab scroll missing after rebuild")
	}
	{
		got := game.tabScroll.ScrollLeft
		if got != 0 {
			t.Fatalf("settings tab scroll = %.2f, want left-aligned 0 when tabs fit", got)
		}
	}
}

func selectTabThroughStrip(t *testing.T, game *Game, id hud.TabID) {
	t.Helper()
	for index, tab := range game.shell.Snapshot().Tabs {
		if tab.ID() != id {
			continue
		}
		game.tabRadioGroup.SetActive(game.tabButtons[index])
		event.ExecuteDeferred()
		return
	}
	t.Fatalf("tab %q was not present", id)
}

func buildSettingsContentForTest(t *testing.T, game *Game) *widget.Container {
	t.Helper()
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[4]
	content := game.buildSettingsContent(settings)
	content.Validate()
	return content
}

func TestTabStripManualScrollIsNotOverwrittenWithoutRequest(t *testing.T) {
	game := NewGame()
	game.tabScroll.ScrollLeft = 0.42
	game.ensureActiveTabVisible()
	{
		got := game.tabScroll.ScrollLeft
		if got != 0.42 {
			t.Fatalf("manual tab scroll = %.2f, want preserved 0.42", got)
		}
	}
	game.requestActiveTabVisible()
	game.ensureActiveTabVisible()
	{
		got := game.tabScroll.ScrollLeft
		if got != 0 {
			t.Fatalf("requested active tab scroll = %.2f, want 0 for first tab", got)
		}
	}
}

func TestMasterDetailShowsOnlyTheSelectedGroupedDetail(t *testing.T) {
	game := NewGame()
	tabs := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())
	cluster := tabs[2]
	game.selectSection(cluster.ID(), 5)
	details := game.detailSections(cluster)
	if len(details) != 1 || details[0].Title != "Chat Pool (High priority)" {
		t.Fatalf("selected cluster details = %#v, want Chat Pool", details)
	}
	labels := collectTextLabels(game.buildSectionContainer(details[0]))
	for _, want := range []string{"CHAT POOL (HIGH PRIORITY)", "Priority: high"} {
		if !containsLabel(labels, want) {
			t.Fatalf("routing detail missing %q in %#v", want, labels)
		}
	}
	projects := tabs[1]
	game.selectSection(projects.ID(), 6)
	details = game.detailSections(projects)
	if len(details) != 1 || details[0].Title != "What's in the news?" {
		t.Fatalf("selected project details = %#v, want What's in the news?", details)
	}
}

func TestCollapsedMasterDetailStartsWithListAndCanBuildDetailBack(t *testing.T) {
	game := NewGame()
	game.width = 640
	projectTab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[1]
	body := game.buildMasterDetailTab(projectTab)
	container, ok := unwrapBounded(body).(*widget.ScrollContainer)
	if !ok {
		t.Fatalf("collapsed list body type = %T, want *widget.ScrollContainer", body)
	}
	_ = container
	game.selectSection(projectTab.ID(), 0)
	body = game.buildMasterDetailTab(projectTab)
	scroll, ok := unwrapBounded(body).(*widget.ScrollContainer)
	if !ok {
		t.Fatalf("collapsed detail body type = %T, want *widget.ScrollContainer", body)
	}
	if scroll.GetWidget().LayoutData == nil {
		t.Fatal("collapsed detail missing layout data")
	}
}

func TestCollapsedMasterListDoesNotKeepExpandedSplitWidth(t *testing.T) {
	game := NewGame()
	game.width = 640
	{
		width := game.masterListWidth()
		if width != 0 {
			t.Fatalf("collapsed master list width = %d, want 0", width)
		}
	}
}

func TestTabStripHasSingleCheckedButtonAfterSelection(t *testing.T) {
	game := NewGame()
	if len(game.tabButtons) < 2 {
		t.Fatal("tab strip did not retain buttons")
	}
	game.selectTab(1, game.tabButtons[1])
	checked := 0
	for _, button := range game.tabButtons {
		if button.State() == widget.WidgetChecked {
			checked++
		}
	}
	if checked != 1 {
		t.Fatalf("checked tab buttons = %d, want 1", checked)
	}
}

func TestNarrowBodyPreferredWidthsAreBounded(t *testing.T) {
	game := NewGame()
	game.width = 360
	tabs := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())
	for index, tabData := range tabs {
		body := game.buildActiveTabBody(hud.Snapshot{Tabs: tabs, ActiveIndex: index})
		width, _ := body.PreferredSize()
		{
			maxWidth := game.hudPreferredWidth()
			if width > maxWidth {
				t.Fatalf("%s preferred width = %d, want <= %d", tabData.ID(), width, maxWidth)
			}
		}
	}
}

func TestHUDBodyTextUsesWrappingWidth(t *testing.T) {
	game := NewGame()
	game.width = 360
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[4]
	for _, text := range collectTextNodes(game.buildSettingsContent(settings)) {
		if text.MaxWidth <= 0 {
			t.Fatalf("settings text %q MaxWidth = %.1f, want positive", text.Label, text.MaxWidth)
		}
	}
	section := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[1].Sections[0]
	for _, text := range collectTextNodes(game.buildSectionContainer(section)) {
		if text.MaxWidth <= 0 {
			t.Fatalf("detail text %q MaxWidth = %.1f, want positive", text.Label, text.MaxWidth)
		}
	}
}

func collectTextLabels(root *widget.Container) []string {
	var labels []string
	walkWidget(root, func(node widget.PreferredSizeLocateableWidget) {
		switch v := node.(type) {
		case *widget.Text:
			labels = append(labels, v.Label)
		case *widget.Button:
			{
				text := v.Text()
				if text != nil {
					labels = append(labels, text.Label)
				}
			}
		}
	})
	return labels
}

func collectTextNodes(root widget.PreferredSizeLocateableWidget) []*widget.Text {
	var texts []*widget.Text
	walkWidget(root, func(node widget.PreferredSizeLocateableWidget) {
		{
			text, ok := node.(*widget.Text)
			if ok {
				texts = append(texts, text)
			}
		}
	})
	return texts
}

func findButtonByLabel(root *widget.Container, label string) *widget.Button {
	var found *widget.Button
	walkWidget(root, func(node widget.PreferredSizeLocateableWidget) {
		if found != nil {
			return
		}
		button, ok := node.(*widget.Button)
		if !ok {
			return
		}
		{
			text := button.Text()
			if text != nil && text.Label == label {
				found = button
			}
		}
	})
	return found
}

func findCheckboxByLabel(root *widget.Container, label string) *widget.Checkbox {
	var found *widget.Checkbox
	walkWidget(root, func(node widget.PreferredSizeLocateableWidget) {
		if found != nil {
			return
		}
		checkbox, ok := node.(*widget.Checkbox)
		if !ok {
			return
		}
		{
			text := checkbox.Text()
			if text != nil && text.Label == label {
				found = checkbox
			}
		}
	})
	return found
}

func walkWidget(node widget.PreferredSizeLocateableWidget, visit func(widget.PreferredSizeLocateableWidget)) {
	visit(node)
	{
		bounded, ok := node.(*boundedPreferredWidget)
		if ok {
			walkWidget(bounded.child, visit)
			return
		}
	}
	{
		container, ok := node.(*widget.Container)
		if ok {
			for _, child := range container.Children() {
				walkWidget(child, visit)
			}
		}
	}
}

func unwrapBounded(node widget.PreferredSizeLocateableWidget) widget.PreferredSizeLocateableWidget {
	{
		bounded, ok := node.(*boundedPreferredWidget)
		if ok {
			return bounded.child
		}
	}
	return node
}

func containsLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}
