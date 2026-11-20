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
	candidate *selfupdate.Candidate
	err       error
	installed bool
}

type desktopUpdateState struct {
	ctx                      context.Context
	cancel                   context.CancelFunc
	client                   *selfupdate.Client
	binary                   string
	result                   chan desktopUpdateResult
	workers                  sync.WaitGroup
	candidate                *selfupdate.Candidate
	busy, installed, restart bool
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
	g.SetOnCheckForUpdate(func() bool {
		if u.busy {
			return true
		}
		if u.installed {
			u.restart = true
			return true
		}
		u.busy = true
		candidate := u.candidate
		if candidate != nil {
			g.linuxUpdateStatus("Installing...", "Replacing Apparat. Your projects and settings are preserved.")
		} else {
			g.linuxUpdateStatus("Checking and downloading...", "Checking the published Linux build and verifying its download.")
		}
		u.workers.Add(1)
		go func() {
			defer u.workers.Done()
			if candidate != nil {
				err := candidate.Install(u.ctx)
				candidate.Discard()
				u.result <- desktopUpdateResult{err: err, installed: err == nil}
				return
			}
			r, err := u.client.Latest(u.ctx)
			if err == nil {
				var current string
				current, err = selfupdate.Digest(u.binary)
				if err == nil && current != r.SHA256 {
					candidate, err = u.client.Download(u.ctx, r, u.binary)
				}
			}
			u.result <- desktopUpdateResult{candidate: candidate, err: err}
		}()
		return true
	})
	g.layoutDirty = true
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
		u.candidate = result.candidate
		u.installed = result.installed
		switch {
		case result.err != nil:
			label := "Update failed — retry"
			detail := "Could not download or save the update. Check your connection and available storage. Your installed app was kept."
			if errors.Is(result.err, selfupdate.ErrVerification) || errors.Is(result.err, selfupdate.ErrMetadata) {
				detail = "The published update failed verification. Your installed app was kept."
			} else if errors.Is(result.err, selfupdate.ErrInstall) {
				detail = "Could not replace the executable. Check that its directory is writable. Your installed app was kept."
			}
			g.linuxUpdateStatus(label, detail)
		case result.installed:
			g.linuxUpdateStatus("Restart Apparat", "Update installed. Restart to use it; Steam's launcher stays the same.")
		case result.candidate != nil:
			g.linuxUpdateStatus("Install update", "The Linux download is verified and ready. Install when your work is saved.")
		default:
			g.linuxUpdateStatus("Up To Date!", "This executable matches the latest qualified Linux build.")
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
	if u.candidate != nil {
		u.candidate.Discard()
	}
	select {
	case r := <-u.result:
		if r.candidate != nil {
			r.candidate.Discard()
		}
	default:
	}
}
