//go:build gui

// Verifies Settings diagnostics, update callbacks and status rendering.
// Dependencies: github.com/cjtrowbridge/apparat/internal/hud, github.com/ebitenui/ebitenui/event, github.com/ebitenui/ebitenui/widget.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
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

	for _, want := range []string{"UPDATES", "VOICE COMMANDS", "APPEARANCE", "NAVIGATION AND WINDOW", "SPEECH RECOGNITION", "BINDINGS", "DIAGNOSTICS"} {
		if !containsLabel(labels, want) {
			t.Fatalf("settings labels missing %q in %#v", want, labels)
		}
	}
	if findButtonByLabel(content, "Browse available commands") == nil {
		t.Fatal("missing command directory button")
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
	click("Startup tab: projects")
	click("Tab placement: Responsive")
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
	if width != 800 || reopened.ActiveTabID() != "cluster" || reopened.prefs.Placement != "bottom" {
		t.Fatalf("restored width %d and tab %s", width, reopened.ActiveTabID())
	}
}

func TestResponsiveLandingAndScrollDistance(t *testing.T) {
	game := NewGame()
	defer game.Close()
	game.Layout(390, 800)
	if game.ActiveTabID() != "projects" || !game.bottomTabs() {
		t.Fatal("fresh portrait must show Projects with bottom tabs")
	}
	game.Layout(1280, 800)
	if game.bottomTabs() || game.sideRail() != "left_rail" {
		t.Fatal("wide landscape must use the responsive rail")
	}
	game.prefs.Placement = "top"
	game.Layout(390, 800)
	if game.bottomTabs() || game.sideRail() != "" {
		t.Fatal("explicit top placement was lost on rotation")
	}
	for _, overflow := range []int{500, 1000, 2400} {
		if math.Abs((scrollFraction(.5, -100, overflow)-.5)*float64(overflow)-100) > 0.001 {
			t.Fatal("drag distance changed with content height")
		}
		if scrollFraction(0, 100, overflow) != 0 || scrollFraction(1, -100, overflow) != 1 {
			t.Fatal("scrolled past content edge")
		}
	}
	if scrollFraction(.5, -100, -300) != 0 || scrollFraction(.5, 100, 0) != 0 {
		t.Fatal("short viewport scrolled")
	}
}

func TestReleaseDiscovery(t *testing.T) {
	for _, tc := range []struct {
		code                   int
		body, installed, label string
	}{
		{200, "valid", "10 1.0", "Up To Date!"},
		{200, "valid", "9 0.9", "Download update"},
		{404, "", "10 1.0", "Retry update check"},
		{200, "invalid", "10 1.0", "Retry update check"},
	} {
		func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
				if tc.body == "valid" {
					_, _ = io.WriteString(w, `<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="com.cjtrowbridge.apparat" android:versionCode="10" android:versionName="1.0"/>`)
				}
			}))
			defer srv.Close()
			status := readRelease(context.Background(), srv.URL, tc.installed)
			if status.Label != tc.label {
				t.Fatal(tc, status)
			}
			srv.Close()
			if readRelease(context.Background(), srv.URL, tc.installed).Label != "Retry update check" {
				t.Fatal("offline lookup reported success")
			}
		}()
	}
}

func TestPersonalIdentityShowsNameAndPencil(t *testing.T) {
	g := NewGame()
	g.runtimeInfo.EditName = func() {}
	g.serviceProjection.ScreenName = "Example owner"
	card := widget.NewContainer()
	g.settingsCardControls(card, "personal-identity")
	labels := collectTextLabels(card)
	if !containsLabel(labels, "Example owner") || containsLabel(labels, "Edit screen name") {
		t.Fatal("expected screen name beside icon")
	}
	var targets []focusTarget
	collectTargets(card, &targets)
	if len(targets) != 1 {
		t.Fatal("expected one focusable pencil")
	}
}

func TestUpdateLabelFitsAfterChange(t *testing.T) {
	for _, width := range []int{320, 1100} {
		g := NewGame()
		g.SetOnCheckForUpdate(func() bool {
			return true
		})
		selectTabThroughStrip(t, g, hud.TabSettings)
		g.Layout(width, 720)
		g.rebuildUI(g.shell.Snapshot())
		renderNavigation(g)
		for _, label := range []string{"Checking…", "Download update", "Unable to check for updates"} {
			g.updateStatus.Store(updateStatus{Label: label})
			g.applyUpdateStatus()
			renderNavigation(g)
			w, h := g.updateButton.Text().PreferredSize()
			box := g.updateButton.GetWidget().Rect
			pad := g.theme.ButtonTheme.TextPadding
			if w+pad.Left+pad.Right > box.Dx() || h+pad.Top+pad.Bottom > box.Dy() {
				t.Fatalf("width %d, label %q measured %dx%d inside %v", width, label, w, h, box)
			}
		}
	}
}
