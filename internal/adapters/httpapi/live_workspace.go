// Multiplexes project invalidations on one pinned authenticated WebSocket per active owner.
// Only visible directories are watched, and each published signal rechecks live authorization.
package httpapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type Watch struct{ Project, Directory, Session string }
type Notice struct {
	Resource Watch
	Sequence uint64
	Reset    bool
	Screen   *domain.WorkspaceView
}

func (a *workerAuth) live(w http.ResponseWriter, r *http.Request, record credential, worker string) {
	if r.Method != "GET" || !a.db.OwnerConfirmed(r.Context(), worker) {
		respond(w, 403, map[string]string{"error": "access_denied"})
		return
	}
	websocket.Server{Handshake: func(*websocket.Config, *http.Request) error {
		if r.Header.Get("Origin") != "https://apparat.invalid" {
			return errors.New("invalid origin")
		}
		return nil
	}, Handler: func(ws *websocket.Conn) {
		defer ws.Close()
		ws.MaxPayloadBytes = 8192
		_ = ws.SetDeadline(time.Now().Add(4 * time.Second))
		var subs []Watch
		if websocket.JSON.Receive(ws, &subs) != nil || len(subs) == 0 || len(subs) > 16 {
			return
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return
		}
		defer watcher.Close()
		ctx := r.Context()
		watchers := map[string][]Watch{}
		for _, sub := range subs {
			scope := "read"
			if sub.Session != "" {
				scope = "terminal-view"
				if !domain.ValidID(sub.Session) {
					return
				}
			}
			location, err := a.db.WorkspaceLocation(ctx, sub.Project, record.access.User, scope)
			if err != nil || !projectgit.WorkspacePath(sub.Directory) {
				return
			}
			path := filepath.Join(location, filepath.FromSlash(sub.Directory))
			if sub.Session != "" {
				path = filepath.Join(filepath.Dir(filepath.Dir(a.artifacts)), "terminal", sub.Project, record.access.User, sub.Session)
			}
			real, err := filepath.EvalSymlinks(path)
			if err != nil || real != path {
				return
			}
			err = watcher.Add(path)
			if err != nil {
				return
			}
			watchers[path] = append(watchers[path], sub)
		}
		sequence := uint64(0)
		send := func(sub Watch, reset bool) bool {
			if time.Now().After(record.until) {
				return false
			}
			_, err := a.db.CurrentAccess(ctx, record.access.Cluster, record.access.User, record.access.Device)
			if err != nil {
				return false
			}
			scope := "read"
			var screen *domain.WorkspaceView
			if sub.Session != "" {
				scope = "terminal-view"
			}
			_, err = a.db.WorkspaceLocation(ctx, sub.Project, record.access.User, scope)
			if err != nil {
				return false
			}
			if sub.Session != "" {
				view, e := a.workspace(ctx, record.access, domain.WorkspaceRequest{Action: "terminal-screen", ProjectID: sub.Project, Path: ".", ID: sub.Session})
				if e != nil {
					return false
				}
				screen = &view
			}
			sequence++
			_ = ws.SetWriteDeadline(time.Now().Add(2 * time.Second))
			payload, err := json.Marshal(Notice{sub, sequence, reset, screen})
			if err != nil {
				return false
			}
			start := time.Now()
			err = websocket.Message.Send(ws, string(payload))
			if a.metrics != nil {
				a.metrics.Observe(telemetry.Event{Operation: "subscribe", Elapsed: time.Since(start).Microseconds(), Sent: len(payload), Rows: 1, Failed: err != nil})
			}
			return err == nil
		}
		for _, sub := range subs {
			if !send(sub, true) {
				return
			}
		}
		ended := make(chan struct{})
		go func() {
			var unused string
			_ = ws.SetReadDeadline(record.until)
			_ = websocket.Message.Receive(ws, &unused)
			close(ended)
		}()
		tick := time.NewTicker(200 * time.Millisecond)
		defer tick.Stop()
		pending := map[Watch]bool{}
		for {
			select {
			case <-a.gitContext.Done():
				return
			case <-ended:
				return
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
				for _, sub := range subs {
					pending[sub] = true
				}
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}
				for _, path := range []string{e.Name, filepath.Dir(e.Name)} {
					for _, sub := range watchers[path] {
						pending[sub] = true
					}
				}
			case <-tick.C:
				for sub := range pending {
					if !send(sub, false) {
						return
					}
					delete(pending, sub)
				}
			}
		}
	}}.ServeHTTP(w, r)
}
func (c *WorkerClient) Watch(ctx context.Context, subs []Watch, receive func(Notice)) error {
	config, err := websocket.NewConfig("wss"+strings.TrimPrefix(c.base, "https")+"/v1/workspace/live", "https://apparat.invalid")
	if err != nil {
		return err
	}
	config.TlsConfig = c.http.Transport.(*http.Transport).TLSClientConfig.Clone()
	config.TlsConfig.MinVersion = tls.VersionTLS13
	config.Header.Set("Authorization", "Bearer "+c.token)
	ws, err := config.DialContext(ctx)
	if err != nil {
		return err
	}
	defer ws.Close()
	stop := context.AfterFunc(ctx, func() {
		_ = ws.Close()
	})
	defer stop()
	ws.MaxPayloadBytes = 128 << 10
	err = websocket.JSON.Send(ws, subs)
	for err == nil {
		var n Notice
		err = websocket.JSON.Receive(ws, &n)
		if err == nil {
			receive(n)
		}
	}
	return err
}
