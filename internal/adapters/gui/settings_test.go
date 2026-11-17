//go:build gui

// Verifies Settings diagnostics, update callbacks and status rendering.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/event, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"path/filepath"
	"testing"
	"time"

	"github.com/ebitenui/ebitenui/event"
	"github.com/ebitenui/ebitenui/widget"
)

func TestSettingsContentIncludesAllSections(t *testing.T) {
	game := NewGame()
	content := buildSettingsContentForTest(t, game)
	labels := collectTextLabels(content)

	for _, want := range []string{"UPDATES", "HUD CONFIGURATION", "BINDINGS", "DIAGNOSTICS"} {
		if !containsLabel(labels, want) {
			t.Fatalf("settings labels missing %q in %#v", want, labels)
		}
	}
	if findButtonByLabel(content, "Check for update") != nil {
		t.Fatal("desktop Settings unexpectedly offers an update check")
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
	game.SetOnCheckForUpdate(func() bool {
		return true
	})
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

type settingsTestRuntime struct {
	*persistence.DB
}

func (s settingsTestRuntime) Changed() <-chan struct{} {
	return nil
}
func (s settingsTestRuntime) List(ctx context.Context) ([]domain.Service, error) {
	return s.Services(ctx)
}

func TestSettingsControlSavesAndAppliesAtRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "apparat.db")
	open := func() settingsTestRuntime {
		db, err := persistence.Open(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		err = db.Migrate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = db.Close()
		})
		return settingsTestRuntime{db}
	}
	store := open()
	game := NewGameWithRuntimeInfo(RuntimeInfo{Queries: store})
	click := func(label string) {
		button := findButtonByLabel(buildSettingsContentForTest(t, game), label)
		if button == nil {
			t.Fatalf("missing control %s", label)
		}
		button.Click()
		event.ExecuteDeferred()
		deadline := time.Now().Add(2 * time.Second)
		for game.prefs.pending != nil && time.Now().Before(deadline) {
			game.receivePreferences()
			time.Sleep(time.Millisecond)
		}
		if game.prefs.pending != nil {
			t.Fatal("preference save timed out")
		}
	}
	click("UI scale: 100%")
	click("Startup tab: comrades")
	click("Tab placement: Top")
	if game.prefs.Placement != "bottom" || game.sideRail() != "" {
		t.Fatal("confirmed placement did not reach layout")
	}
	width, _ := game.Layout(1000, 800)
	if width != 800 {
		t.Fatalf("live width %d", width)
	}
	game.Close()
	reopened := NewGameWithRuntimeInfo(RuntimeInfo{Queries: open()})
	defer reopened.Close()
	width, _ = reopened.Layout(1000, 800)
	if width != 800 || reopened.ActiveTabID() != "projects" || reopened.prefs.Placement != "bottom" {
		t.Fatalf("restored width %d and tab %s", width, reopened.ActiveTabID())
	}
}
