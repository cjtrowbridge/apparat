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

	"github.com/cjtrowbridge/apparat/internal/adapters/selfupdate"
)

type desktopUpdateResult struct {
	err       error
	installed bool
}

type desktopUpdateState struct {
	ctx           context.Context
	cancel        context.CancelFunc
	client        *selfupdate.Client
	binary        string
	result        chan desktopUpdateResult
	workers       sync.WaitGroup
	busy, restart bool
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
	g.linuxUpdateStatus("Check for update", "")
	g.SetOnCheckForUpdate(func() bool {
		if u.busy || u.restart {
			return true
		}
		u.busy = true
		g.linuxUpdateStatus("Checking…", "Checking for a Linux update…")
		u.workers.Add(1)
		go func() {
			defer u.workers.Done()
			u.result <- u.update(g.linuxUpdateStatus)
		}()
		return true
	})
	g.layoutDirty = true
}

func (u *desktopUpdateState) update(progress func(string, string)) desktopUpdateResult {
	release, err := u.client.Latest(u.ctx)
	if err != nil {
		return desktopUpdateResult{err: err}
	}
	current, err := selfupdate.Digest(u.binary)
	if err != nil || current == release.SHA256 {
		return desktopUpdateResult{err: err}
	}
	progress("Updating…", "Downloading and verifying the update…")
	candidate, err := u.client.Download(u.ctx, release, u.binary)
	if err != nil {
		return desktopUpdateResult{err: err}
	}
	defer candidate.Discard()
	progress("Updating…", "Installing the update and restarting Apparat…")
	err = candidate.Install(u.ctx)
	return desktopUpdateResult{err: err, installed: err == nil}
}

func (g *Game) linuxUpdateStatus(label, detail string) {
	g.updateStatus.Store(updateStatus{Label: label, Detail: detail, Reason: "linux_update"})
}

func (g *Game) pollDesktopUpdates() bool {
	u := g.desktopUpdates
	if u == nil {
		return false
	}
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
			g.linuxUpdateStatus("Check for update", detail + " Select again to retry.")
		case !result.installed:
			g.linuxUpdateStatus("Up to date", "Select again to check for a newer build.")
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
