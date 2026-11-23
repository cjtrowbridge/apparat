//go:build gui

// Releases directory watches when the Projects view loses focus or changes owner.
// Burst invalidations share one pending refresh and never replace an active editor.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/hajimehoshi/ebiten/v2"
	"strings"
)

type liveDirectory struct {
	project, dir, session string
	cancel                context.CancelFunc
	dirty                 chan *domain.WorkspaceView
}
type directoryWatcher interface {
	WatchDirectory(context.Context, string, string, string, func(*domain.WorkspaceView)) error
}

func (g *Game) pollWorkspaceLive() {
	f := &g.projects
	w := &f.workspace
	l := &f.live
	watcher, ok := g.runtimeInfo.Queries.(directoryWatcher)
	session := ""
	if w.mode == "terminal-session" {
		session = w.runID
	}
	active := ok && f.active && ebiten.IsFocused() && g.inWorkspace() && (w.mode == "files" || session != "") && strings.HasPrefix(w.id, "worker-")
	if l.cancel != nil && (!active || l.project != w.id || l.dir != w.dir || l.session != session) {
		l.cancel()
		*l = liveDirectory{}
	}
	if !active {
		return
	}
	if l.cancel == nil {
		ctx, cancel := context.WithCancel(f.ctx)
		ch := make(chan *domain.WorkspaceView, 1)
		*l = liveDirectory{w.id, w.dir, session, cancel, ch}
		id, dir := w.id, w.dir
		f.workers.Add(1)
		go func() {
			defer f.workers.Done()
			_ = watcher.WatchDirectory(ctx, id, dir, session, func(view *domain.WorkspaceView) {
				select {
				case ch <- view:
				default:
					select {
					case <-ch:
					default:
					}
					ch <- view
				}
			})
		}()
	}
	if f.pending == nil {
		select {
		case view := <-l.dirty:
			if view != nil {
				w.report = view.Report
				w.selected.State = view.Detail
				g.layoutDirty = true
			} else {
				g.workspaceRequest("list")
			}
		default:
		}
	}
}
