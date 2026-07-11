//go:build gui

package gui

import (
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/widget"
)

func TestNewGameBuildsStretchedTabBookRoot(t *testing.T) {
	game := NewGame()
	root, ok := game.ui.Container.(*widget.Container)
	if !ok {
		t.Fatalf("root type = %T, want *widget.Container", game.ui.Container)
	}
	children := root.Children()
	if len(children) != 1 {
		t.Fatalf("root children = %d, want 1", len(children))
	}
	tabBook, ok := children[0].(*widget.TabBook)
	if !ok {
		t.Fatalf("root child type = %T, want *widget.TabBook", children[0])
	}
	layoutData, ok := tabBook.GetWidget().LayoutData.(widget.AnchorLayoutData)
	if !ok {
		t.Fatalf("tabbook layout data = %T, want widget.AnchorLayoutData", tabBook.GetWidget().LayoutData)
	}
	if !layoutData.StretchHorizontal || !layoutData.StretchVertical {
		t.Fatalf("tabbook stretch = horizontal %t vertical %t, want both true", layoutData.StretchHorizontal, layoutData.StretchVertical)
	}
}

func TestTabBookInitialTabFollowsHUDSnapshot(t *testing.T) {
	game := NewGame()
	if err := game.shell.SelectTab(6); err != nil {
		t.Fatal(err)
	}
	game.rebuildUI(game.shell.Snapshot())
	root := game.ui.Container.(*widget.Container)
	tabBook := root.Children()[0].(*widget.TabBook)
	game.ui.Container.GetWidget().SetTheme(game.theme)
	game.ui.Container.Validate()
	tabButton := tabBook.GetTabButton(tabBook.Tab())
	if tabButton == nil || tabButton.Text() == nil || !strings.Contains(tabButton.Text().Label, "Settings") {
		label := "<nil>"
		if tabButton != nil && tabButton.Text() != nil {
			label = tabButton.Text().Label
		}
		t.Fatalf("active tab label = %q, want Settings", label)
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
