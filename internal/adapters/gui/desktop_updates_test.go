//go:build gui && linux && !android

// Verifies the user-visible Linux update states with a local in-memory HTTP channel.
// Installation uses a disposable executable, never the test runner or real Apparat binary.
package gui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/speech"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"github.com/ebitenui/ebitenui/event"
)

type updateTransport func(*http.Request) (*http.Response, error)

func (f updateTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestLinuxUpdateSingleActivation(t *testing.T) {
	game := NewGame()
	game.initDesktopUpdates(context.Background())
	defer game.closeDesktopUpdates()
	u := game.desktopUpdates
	target := filepath.Join(t.TempDir(), "apparat")
	err := os.WriteFile(target, []byte("old installed binary"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	u.binary = target
	path, _ := os.Executable()
	binary, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(binary)
	digest := hex.EncodeToString(sum[:])
	const revision = "1234567890123456789012345678901234567890"
	u.client.HTTP = &http.Client{Transport: updateTransport(func(r *http.Request) (*http.Response, error) {
		var data []byte
		switch {
		case strings.HasSuffix(r.URL.Path, "/commits/main"):
			data, _ = json.Marshal(map[string]string{"sha": revision})
		case strings.HasSuffix(r.URL.Path, "/releases/verified.json"):
			data, _ = json.Marshal(map[string]any{"accepted": true, "edition": "ordinary", "ref": "refs/heads/main", "delivery_schema": 1,
				"targets": []map[string]any{{"name": "linux/amd64/apparat", "output": "releases/linux/amd64/apparat/latest", "state": "passed", "sha256": digest}}})
		default:
			data = binary
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data)), ContentLength: int64(len(data))}, nil
	})}
	finish := func() {
		deadline := time.Now().Add(5 * time.Second)
		for u.busy && time.Now().Before(deadline) {
			game.pollDesktopUpdates()
			time.Sleep(time.Millisecond)
		}
		if u.busy {
			t.Fatal("update operation did not finish")
		}
	}
	selectTabThroughStrip(t, game, hud.TabSettings)
	game.Layout(320, 720)
	game.rebuildUI(game.shell.Snapshot())
	renderNavigation(game)
	button := game.updateButton
	if button.Text().Label != "Check for update" {
		t.Fatalf("initial button label: %s", button.Text().Label)
	}
	button.Click()
	event.ExecuteDeferred()
	game.requestUpdateAction() // Duplicate input during the same operation.
	game.applyUpdateStatus()
	busyLabel := button.Text().Label
	if (busyLabel != "Checking…" && busyLabel != "Updating…" && !strings.HasPrefix(busyLabel, "Downloading")) || !button.GetWidget().Disabled {
		t.Fatal("busy update button must show progress and reject duplicate input")
	}
	finish()
	if !game.pollDesktopUpdates() {
		t.Fatal("one activation must install and request graceful restart")
	}
	previous, _ := os.ReadFile(target + ".previous")
	if string(previous) != "old installed binary" {
		t.Fatal("previous executable missing")
	}
	u.restart = false
	game.requestUpdateAction()
	finish()
	game.applyUpdateStatus()
	if button.Text().Label != "Up to date" || button.GetWidget().Disabled || u.restart {
		t.Fatalf("current build status: %s", game.UpdateStatusDetail())
	}
	u.client.HTTP.Transport = updateTransport(func(*http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	button.Click()
	event.ExecuteDeferred()
	finish()
	game.applyUpdateStatus()
	if button.Text().Label != "Check for update" || u.restart || button.GetWidget().Disabled || !strings.Contains(game.updateDetail.Label, "retry") {
		t.Fatal("failed update must offer a retry in Settings without restarting")
	}
	renderNavigation(game)
	br := button.GetWidget().Rect
	dr := game.updateDetail.GetWidget().Rect
	bw, _ := button.PreferredSize()
	_, dh := game.updateDetail.PreferredSize()
	if br.Dx() < bw || dr.Min.Y < br.Max.Y || dr.Dy() < dh || dr.Max.X > game.width {
		t.Fatalf("narrow update card overflow: button=%v detail=%v", br, dr)
	}
}

func TestCommandPTTStartsWithoutTextFocus(t *testing.T) {
	game := NewGame()
	game.voice.engine = &speech.Engine{}
	game.beginPTT("controller")
	if game.voice.field != nil || game.voice.canvas == nil {
		t.Fatal("PTT outside an editor must capture a voice command")
	}
	game.cancelPTT()
}

func TestGamescopeDetectionAndNativeAspect(t *testing.T) {
	t.Setenv("GAMESCOPE_WAYLAND_DISPLAY", "")
	if gamescopeSession() {
		t.Fatal("ordinary desktop detected as Gaming Mode")
	}
	t.Setenv("GAMESCOPE_WAYLAND_DISPLAY", "gamescope-0")
	if !gamescopeSession() {
		t.Fatal("Gaming Mode not detected")
	}
	game := NewGame()
	for _, scale := range []int{75, 100, 125, 150} {
		game.prefs.Scale = scale
		w, h := game.Layout(1280, 800)
		difference := w*800 - h*1280
		if difference < -1280 || difference > 1280 {
			t.Fatalf("layout distorts native display at %d%%: %dx%d", scale, w, h)
		}
	}
}
