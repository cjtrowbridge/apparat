//go:build gui

package gui

import (
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
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
	if err := game.shell.SelectTab(6); err != nil {
		t.Fatal(err)
	}
	game.rebuildUI(game.shell.Snapshot())
	if id := game.ActiveTabID(); id != "comrades" {
		t.Fatalf("active tab id before update store = %q, want existing comrades store", id)
	}
	game.activeTabID.Store(string(game.shell.Snapshot().ActiveTab().ID()))
	if id := game.ActiveTabID(); id != "settings" {
		t.Fatalf("active tab id = %q, want settings", id)
	}
	game.ensureActiveTabVisible()
	if game.tabScroll == nil {
		t.Fatal("tab scroll missing after rebuild")
	}
	if got := game.tabScroll.ScrollLeft; got != 1 {
		t.Fatalf("settings tab scroll = %.2f, want 1", got)
	}
}

func TestSettingsContentIncludesAllSectionsAndUpdateButton(t *testing.T) {
	game := NewGame()
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	labels := collectTextLabels(content)

	for _, want := range []string{"UPDATES", "HUD CONFIGURATION", "BINDINGS", "DIAGNOSTICS"} {
		if !containsLabel(labels, want) {
			t.Fatalf("settings labels missing %q in %#v", want, labels)
		}
	}
	if findButtonByLabel(content, "Check for update") == nil {
		t.Fatal("settings content missing Check for update button")
	}
	if findCheckboxByLabel(content, "Open Debug UI overlay") == nil {
		t.Fatal("settings content missing Debug UI overlay checkbox")
	}
}

func TestSettingsDebugOverlayCheckboxLabelClosesWhenOpen(t *testing.T) {
	game := NewGame()
	game.debugOverlayOpen = true
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	if findCheckboxByLabel(content, "Close Debug UI overlay") == nil {
		t.Fatal("settings content missing close Debug UI overlay checkbox label")
	}
}

func TestSettingsDebugOverlayCheckboxTogglesState(t *testing.T) {
	game := NewGame()
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	checkbox := findCheckboxByLabel(content, "Open Debug UI overlay")
	if checkbox == nil {
		t.Fatal("settings content missing Debug UI overlay checkbox")
	}
	checkbox.SetState(widget.WidgetChecked)
	if !game.debugOverlayOpen {
		t.Fatal("debug overlay did not open after checkbox checked")
	}
	checkbox.SetState(widget.WidgetUnchecked)
	if game.debugOverlayOpen {
		t.Fatal("debug overlay did not close after checkbox unchecked")
	}
}

func TestSettingsUpdateButtonInvokesCallback(t *testing.T) {
	game := NewGame()
	called := false
	game.SetOnCheckForUpdate(func() bool {
		called = true
		return true
	})
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	button.Click()
	if !called {
		t.Fatal("update button did not invoke callback")
	}
	if got := button.Text().Label; got != "Checking..." {
		t.Fatalf("button label after callback = %q, want Checking...", got)
	}
	if got := game.UpdateStatus(); got != "Checking..." {
		t.Fatalf("game update status = %q, want Checking...", got)
	}
}

func TestSettingsUpdateButtonAppliesExternalStatus(t *testing.T) {
	game := NewGame()
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	game.SetUpdateStatus("Already current")
	game.applyUpdateStatus()
	if got := button.Text().Label; got != "Already current" {
		t.Fatalf("button label after external status = %q, want Already current", got)
	}
}

func TestSettingsUpdateButtonShowsUnavailableWithoutCallback(t *testing.T) {
	game := NewGame()
	settings := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[6]
	content := game.buildSettingsContent(settings)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	button.Click()
	if got := button.Text().Label; got != "Update unavailable" {
		t.Fatalf("button label without callback = %q, want Update unavailable", got)
	}
	if got := game.UpdateStatus(); got != "Update unavailable" {
		t.Fatalf("game update status = %q, want Update unavailable", got)
	}
}

func TestMasterDetailContentIncludesSectionButtonsAndSummary(t *testing.T) {
	game := NewGame()
	projectTab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[1]
	body := game.buildMasterDetailTab(projectTab)
	container, ok := body.(*widget.Container)
	if !ok {
		t.Fatalf("master-detail body type = %T, want *widget.Container", body)
	}
	if findButtonByLabel(container, "Projects") == nil {
		t.Fatal("master-detail body missing left pane Projects button")
	}
	labels := collectTextLabels(container)
	if !containsLabel(labels, projectTab.Summary) {
		t.Fatalf("master-detail body missing summary %q in %#v", projectTab.Summary, labels)
	}
	if findButtonByLabel(container, "|") == nil {
		t.Fatal("expanded master-detail body missing draggable divider")
	}
}

func TestCollapsedMasterDetailStartsWithListAndCanBuildDetailBack(t *testing.T) {
	game := NewGame()
	game.width = 640
	projectTab := hud.DefaultTabs(hud.DefaultConfigManager{}.Config())[1]
	body := game.buildMasterDetailTab(projectTab)
	container, ok := body.(*widget.ScrollContainer)
	if !ok {
		t.Fatalf("collapsed list body type = %T, want *widget.ScrollContainer", body)
	}
	_ = container
	game.selectSection(projectTab.ID(), 0)
	body = game.buildMasterDetailTab(projectTab)
	scroll, ok := body.(*widget.ScrollContainer)
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
	if width := game.masterListWidth(); width != 0 {
		t.Fatalf("collapsed master list width = %d, want 0", width)
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

func collectTextLabels(root *widget.Container) []string {
	var labels []string
	walkWidget(root, func(node widget.PreferredSizeLocateableWidget) {
		switch v := node.(type) {
		case *widget.Text:
			labels = append(labels, v.Label)
		case *widget.Button:
			if text := v.Text(); text != nil {
				labels = append(labels, text.Label)
			}
		}
	})
	return labels
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
		if text := button.Text(); text != nil && text.Label == label {
			found = button
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
		if text := checkbox.Text(); text != nil && text.Label == label {
			found = checkbox
		}
	})
	return found
}

func walkWidget(node widget.PreferredSizeLocateableWidget, visit func(widget.PreferredSizeLocateableWidget)) {
	visit(node)
	if container, ok := node.(*widget.Container); ok {
		for _, child := range container.Children() {
			walkWidget(child, visit)
		}
	}
}

func containsLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}
