//go:build gui

// Discovers Android versions from the small manifest published beside the matching APK.
// The GUI thread owns entry tracking and results; package transfer requires a separate click.
package gui

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/hud"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const releaseManifest = "https://raw.githubusercontent.com/cjtrowbridge/apparat/main/android/apparat/AndroidManifest.xml"

type updateDiscovery struct {
	installed func() string
	inside    bool
	checked   time.Time
	pending   chan updateStatus
	newer     bool
}

func (g *Game) SetInstalledRelease(read func() string) {
	g.discovery.installed = read
}
func (g *Game) pollUpdateDiscovery() {
	d := &g.discovery
	if d.installed == nil {
		return
	}
	select {
	case status := <-d.pending:
		d.pending = nil
		d.newer = status.Reason == UpdateStatusAvailable
		g.updateStatus.Store(status)
	default:
	}
	inside := g.shell.Snapshot().ActiveTab().ID() == hud.TabSettings
	if inside && !d.inside && time.Since(d.checked) > time.Minute {
		if !g.discoverRelease() {
			return
		}
	}
	d.inside = inside
}
func (g *Game) discoverRelease() bool {
	d := &g.discovery
	if d.pending != nil {
		return true
	}
	installed := d.installed()
	if installed == "" {
		return false
	}
	result := make(chan updateStatus, 1)
	d.pending = result
	d.checked = time.Now()
	d.newer = false
	g.SetUpdateStatus(UpdateStatusChecking)
	ctx := g.projects.ctx
	go func() {
		result <- readRelease(ctx, releaseManifest, installed)
	}()
	return true
}
func readRelease(ctx context.Context, url, installed string) updateStatus {
	unavailable := updateStatus{Label: "Retry update check", Detail: "Could not read the release version. Check your connection and retry."}
	code, name, ok := strings.Cut(installed, " ")
	local, err := strconv.Atoi(code)
	if !ok || err != nil {
		return unavailable
	}
	unavailable.Detail = "Installed " + name + ". " + unavailable.Detail
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return unavailable
	}
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return unavailable
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		unavailable.Detail = "Installed " + name + ". Release information is unavailable to this device. Private Git delivery requires a host update."
		return unavailable
	}
	if resp.StatusCode != http.StatusOK {
		return unavailable
	}
	var manifest struct {
		XMLName xml.Name `xml:"manifest"`
		Package string   `xml:"package,attr"`
		Code    int      `xml:"http://schemas.android.com/apk/res/android versionCode,attr"`
		Name    string   `xml:"http://schemas.android.com/apk/res/android versionName,attr"`
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32769))
	if err != nil || len(data) > 32768 || xml.Unmarshal(data, &manifest) != nil || manifest.Package != "com.cjtrowbridge.apparat" || manifest.Code <= 0 || manifest.Name == "" {
		return unavailable
	}
	detail := fmt.Sprintf("Installed %s · Published %s", name, manifest.Name)
	if manifest.Code > local {
		return updateStatus{Label: "Download update", Detail: detail, Reason: UpdateStatusAvailable}
	}
	return updateStatus{Label: "Up To Date!", Detail: detail, Reason: UpdateStatusUpToDate}
}
func (g *Game) requestUpdateAction() {
	if g.discovery.installed != nil && !g.discovery.newer {
		g.discoverRelease()
		return
	}
	if g.desktopUpdates == nil {
		g.SetUpdateStatus(UpdateStatusChecking)
	}
	if g.onCheckForUpdate == nil || !g.onCheckForUpdate() {
		g.SetUpdateStatus(UpdateStatusUnableBridge)
	}
}

func (game *Game) SetOnCheckForUpdate(f func() bool) {
	game.onCheckForUpdate = f
}

func (game *Game) SetUpdateStatus(code UpdateStatusCode) {
	game.updateStatus.Store(updateStatusFor(code))
}

func (game *Game) UpdateStatus() string {
	{
		status, ok := game.updateStatus.Load().(updateStatus)
		if ok {
			return status.Label
		}
	}
	return ""
}

func (game *Game) UpdateStatusDetail() string {
	{
		status, ok := game.updateStatus.Load().(updateStatus)
		if ok {
			return status.Detail
		}
	}
	return ""
}
