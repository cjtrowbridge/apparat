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
)

type updateTransport func(*http.Request) (*http.Response, error)

func (f updateTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestLinuxUpdateDownloadInstallAndRestartControls(t *testing.T) {
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
	game.onCheckForUpdate()
	finish()
	if game.UpdateStatus() != "Install update" {
		t.Fatalf("download status: %s", game.UpdateStatus())
	}
	game.onCheckForUpdate()
	finish()
	if game.UpdateStatus() != "Restart Apparat" || !u.installed {
		t.Fatalf("install status: %s", game.UpdateStatus())
	}
	previous, _ := os.ReadFile(target + ".previous")
	if string(previous) != "old installed binary" {
		t.Fatal("previous executable missing")
	}
	game.onCheckForUpdate()
	if !game.pollDesktopUpdates() {
		t.Fatal("restart control did not request graceful termination")
	}
	// A new checker observing those installed bytes reports a no-op without staging.
	u.installed = false
	u.restart = false
	game.onCheckForUpdate()
	finish()
	if game.UpdateStatus() != "Up To Date!" || u.candidate != nil {
		t.Fatalf("current build status: %s", game.UpdateStatus())
	}
}

func TestUntargetedPTTIsSilent(t *testing.T) {
	game := NewGame()
	game.voice.engine = &speech.Engine{}
	game.beginPTT("controller")
	if game.voice.preview != "" || game.voice.canvas != nil {
		t.Fatal("untargeted PTT must not create a session or floating hint")
	}
	game.endPTT()
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
