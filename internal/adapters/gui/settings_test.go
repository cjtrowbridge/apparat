//go:build gui

// Verifies Settings diagnostics, update callbacks and status rendering.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/event, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"testing"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestSettingsContentIncludesAllSectionsAndUpdateButton(t *testing.T) {
	game := NewGame()
	content := buildSettingsContentForTest(t, game)
	labels := collectTextLabels(content)

	for _, want := range []string{"UPDATES", "HUD CONFIGURATION", "BINDINGS", "DIAGNOSTICS"} {
		if !containsLabel(labels, want) {
			t.Fatalf("settings labels missing %q in %#v", want, labels)
		}
	}
	if findButtonByLabel(content, "Check for update") == nil {
		t.Fatalf("settings content missing Check for update button; labels: %#v", labels)
	}
	if findCheckboxByLabel(content, "Open Debug UI overlay") == nil {
		t.Fatal("settings content missing Debug UI overlay checkbox")
	}
}

func TestSettingsDebugOverlayCheckboxLabelClosesWhenOpen(t *testing.T) {
	game := NewGame()
	game.debugOverlayOpen = true
	content := buildSettingsContentForTest(t, game)
	if findCheckboxByLabel(content, "Close Debug UI overlay") == nil {
		t.Fatal("settings content missing close Debug UI overlay checkbox label")
	}
}

func TestSettingsDebugOverlayCheckboxTogglesState(t *testing.T) {
	game := NewGame()
	content := buildSettingsContentForTest(t, game)
	checkbox := findCheckboxByLabel(content, "Open Debug UI overlay")
	if checkbox == nil {
		t.Fatal("settings content missing Debug UI overlay checkbox")
	}
	checkbox.SetState(widget.WidgetChecked)
	event.ExecuteDeferred()
	if !game.debugOverlayOpen {
		t.Fatal("debug overlay did not open after checkbox checked")
	}
	checkbox.SetState(widget.WidgetUnchecked)
	event.ExecuteDeferred()
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
	content := buildSettingsContentForTest(t, game)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	button.Click()
	event.ExecuteDeferred()
	if !called {
		t.Fatal("update button did not invoke callback")
	}
	{
		got := button.Text().Label
		if got != "Checking..." {
			t.Fatalf("button label after callback = %q, want Checking...", got)
		}
	}
	{
		got := game.UpdateStatus()
		if got != "Checking..." {
			t.Fatalf("game update status = %q, want Checking...", got)
		}
	}
}

func TestSettingsUpdateButtonAppliesExternalStatus(t *testing.T) {
	game := NewGame()
	content := buildSettingsContentForTest(t, game)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	game.SetUpdateStatus(UpdateStatusUpToDate)
	game.applyUpdateStatus()
	{
		got := button.Text().Label
		if got != "Up To Date!" {
			t.Fatalf("button label after external status = %q, want Up To Date!", got)
		}
	}
}

func TestSettingsUpdateButtonShowsUnavailableWithoutCallback(t *testing.T) {
	game := NewGame()
	content := buildSettingsContentForTest(t, game)
	button := findButtonByLabel(content, "Check for update")
	if button == nil {
		t.Fatal("settings content missing Check for update button")
	}
	button.Click()
	event.ExecuteDeferred()
	{
		got := button.Text().Label
		if got != "Unable to check for updates" {
			t.Fatalf("button label without callback = %q, want Unable to check for updates", got)
		}
	}
	{
		got := game.UpdateStatus()
		if got != "Unable to check for updates" {
			t.Fatalf("game update status = %q, want Unable to check for updates", got)
		}
	}
	{
		got := game.UpdateStatusDetail()
		if got != "No update checker is configured in this build." {
			t.Fatalf("update detail = %q, want desktop no-checker explanation", got)
		}
	}
}
