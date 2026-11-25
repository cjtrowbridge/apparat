//go:build gui

// Releases directory watches when the Projects view loses focus or changes owner.
// Burst invalidations share one pending refresh and never replace an active editor.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/hajimehoshi/ebiten/v2"
	"slices"
	"strings"
	"time"
)

type liveDirectory struct {
	project, dir, session string
	cancel                context.CancelFunc
	dirty                 chan *domain.WorkspaceView
	listing               chan workspaceReply
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
		*l = liveDirectory{project: w.id, dir: w.dir, session: session, cancel: cancel, dirty: ch}
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
	if f.pending != nil {
		l.listing = nil
		return
	}
	if l.listing != nil {
		select {
		case r := <-l.listing:
			l.listing = nil
			if r.err != nil || !slices.Equal(w.entries, r.view.Entries) {
				g.applyWorkspace(&r)
				g.layoutDirty = true
			} else {
				g.startDirectoryStatus(r.request)
			}
		default:
		}
		return
	}
	select {
	case view := <-l.dirty:
		if view != nil {
			w.report = view.Report
			w.selected.State = view.Detail
			g.layoutDirty = true
		} else {
			store, ok := g.runtimeInfo.Queries.(workspaceStore)
			if !ok {
				return
			}
			q := domain.WorkspaceRequest{ProjectID: w.id, Path: w.dir, Action: "list"}
			reply := make(chan workspaceReply, 1)
			l.listing = reply
			f.workers.Add(1)
			go func() {
				defer f.workers.Done()
				v, err := store.Workspace(f.ctx, q)
				reply <- workspaceReply{q, v, err}
			}()
		}
	default:
	}
}

// Refreshes active agent turns without stealing focus or touching the compose buffer.
func (g *Game) pollProjectConversation() {
	f := &g.projects
	w := &f.workspace
	if !g.inWorkspace() || w.mode != "conversation" || !w.agentThread || w.chatDraft || f.ctx == nil || f.pending != nil {
		return
	}
	if f.chatRead != nil {
		select {
		case r := <-f.chatRead:
			f.chatRead = nil
			if r.err == nil && r.request.ProjectID == w.id && r.request.Thread == w.thread && (!slices.Equal(r.view.Choices, w.choices) || r.view.Detail != w.threadTitle) {
				w.choices = r.view.Choices
				w.threadTitle = r.view.Detail
				g.layoutDirty = true
			}
		default:
		}
		return
	}
	if time.Now().Before(f.nextChatRead) {
		return
	}
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	f.nextChatRead = time.Now().Add(2 * time.Second)
	q := domain.WorkspaceRequest{Path: ".", ProjectID: w.id, Thread: w.thread, Action: "agent-chat"}
	read := make(chan workspaceReply, 1)
	f.chatRead = read
	f.workers.Go(func() {
		v, e := store.Workspace(f.ctx, q)
		read <- workspaceReply{q, v, e}
	})
}
