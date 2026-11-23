//go:build gui

// Owns one asynchronous Linux update operation and presents safe outcomes on the UI thread.
// Shutdown cancels and joins I/O before restart can release and reopen the same node.
package gui

import (
	"context"
	"errors"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/hud"

	"github.com/cjtrowbridge/apparat/internal/adapters/selfupdate"
)

type desktopUpdateResult struct {
	err                            error
	installed, available, relaunch bool
	comparison                     string
}

type desktopUpdateState struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	client                *selfupdate.Client
	binary                string
	result                chan desktopUpdateResult
	workers               sync.WaitGroup
	busy, restart, inside bool
	checked               time.Time
}

func (g *Game) initDesktopUpdates(ctx context.Context) {
	if runtime.GOOS != "linux" {
		return
	}
	binary, err := os.Executable()
	if err != nil {
		return
	}
	u := &desktopUpdateState{client: selfupdate.NewClient(), binary: binary, result: make(chan desktopUpdateResult, 1)}
	u.ctx, u.cancel = context.WithCancel(ctx)
	g.desktopUpdates = u
	g.linuxUpdateStatus("Check for update", linuxVersions(""))
	g.SetOnCheckForUpdate(func() bool {
		g.startLinuxUpdate(true)
		return true
	})
	g.layoutDirty = true
}

func linuxVersions(published string) string {
	if published == "" {
		published = "unknown"
	}
	return "Installed " + domain.ReleaseVersion() + " · Published " + published
}

func (g *Game) startLinuxUpdate(install bool) {
	u := g.desktopUpdates
	if u.busy || u.restart {
		return
	}
	u.busy = true
	u.checked = time.Now()
	g.linuxUpdateStatus("Checking…", linuxVersions("")+"\nReading Linux release information…")
	u.workers.Add(1)
	go func() {
		defer u.workers.Done()
		u.result <- u.update(install, g.linuxUpdateStatus)
	}()
}

func (u *desktopUpdateState) update(install bool, status func(string, string)) desktopUpdateResult {
	ctx, cancel := context.WithTimeout(u.ctx, 15*time.Second)
	defer cancel()
	release, err := u.client.Latest(ctx)
	result := desktopUpdateResult{comparison: linuxVersions(release.Version), err: err}
	if err != nil {
		return result
	}
	progress := func(label, detail string) {
		status(label, result.comparison+"\n"+detail)
	}
	progress("Checking…", "Comparing the installed executable with the published digest.")
	current, err := selfupdate.Digest(u.binary)
	if err != nil {
		result.err = err
		return result
	}
	if current == release.SHA256 {
		// An open Linux process can still be executing the old, replaced file.
		result.relaunch = release.Version != "" && release.Version != domain.ReleaseVersion()
		result.installed = result.relaunch && install
		return result
	}
	if !install {
		result.available = true
		return result
	}
	progress("Updating…", "Connecting to the Linux update download…")
	candidate, err := u.client.DownloadWithProgress(u.ctx, release, u.binary, func(received, total int64) {
		status := downloadProgress(received, total)
		detail := status.Detail
		if total > 0 && received == total {
			detail += " · Verifying the executable…"
		}
		progress(status.Label, detail)
	})
	if err != nil {
		result.err = err
		return result
	}
	defer candidate.Discard()
	progress("Updating…", "Installing the update and restarting Apparat…")
	err = candidate.Install(u.ctx)
	result.err = err
	result.installed = err == nil
	return result
}

func (g *Game) linuxUpdateStatus(label, detail string) {
	g.updateStatus.Store(updateStatus{Label: label, Detail: detail, Reason: "linux_update"})
}

func (g *Game) pollDesktopUpdates() bool {
	u := g.desktopUpdates
	if u == nil {
		return false
	}
	inside := g.shell.Snapshot().ActiveTab().ID() == hud.TabSettings
	if inside && !u.inside && time.Since(u.checked) > time.Minute {
		g.startLinuxUpdate(false)
	}
	u.inside = inside
	select {
	case result := <-u.result:
		u.busy = false
		u.restart = result.installed
		switch {
		case result.err != nil:
			detail := "Could not download or save the update. Check your connection and available storage. Your installed app was kept."
			if errors.Is(result.err, selfupdate.ErrVerification) || errors.Is(result.err, selfupdate.ErrMetadata) {
				detail = "The published update failed verification. Your installed app was kept."
			} else if errors.Is(result.err, selfupdate.ErrInstall) {
				detail = "Could not replace the executable. Check that its directory is writable. Your installed app was kept."
			}
			g.linuxUpdateStatus("Check for update", result.comparison+"\n"+detail+" Select again to retry.")
		case result.available:
			g.linuxUpdateStatus("Download update", result.comparison)
		case result.relaunch:
			g.linuxUpdateStatus("Restart Apparat", result.comparison+"\nThe published executable is already on disk.")
		case !result.installed:
			g.linuxUpdateStatus("Up to date", result.comparison)
		}
	default:
	}
	return u.restart
}

func (g *Game) closeDesktopUpdates() {
	u := g.desktopUpdates
	if u == nil {
		return
	}
	u.cancel()
	u.workers.Wait()
}
