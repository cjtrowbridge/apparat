// Emits invalidation hints without project IDs, model names or message contents.
// Clients re-read through the ordinary authenticated API after hints and reconnects.
package httpapi

import (
	"context"
	"errors"
	"golang.org/x/net/websocket"
	"net/http"
	"strings"
	"time"
)

func (a *workerAuth) modelPulse() {
	a.Lock()
	defer a.Unlock()
	if a.modelChanged != nil {
		close(a.modelChanged)
	}
	a.modelChanged = make(chan struct{})
}
func (a *workerAuth) modelSignal() <-chan struct{} {
	a.Lock()
	defer a.Unlock()
	if a.modelChanged == nil {
		a.modelChanged = make(chan struct{})
	}
	return a.modelChanged
}
func (a *workerAuth) modelLive(w http.ResponseWriter, r *http.Request, c credential) {
	if r.Method != "GET" {
		respond(w, 405, "GET required")
		return
	}
	websocket.Server{Handshake: func(_ *websocket.Config, r *http.Request) error {
		if r.Header.Get("Origin") != "https://apparat.invalid" {
			return errors.New("invalid origin")
		}
		return nil
	}, Handler: func(ws *websocket.Conn) {
		defer ws.Close()
		end := time.NewTimer(time.Until(c.until))
		defer end.Stop()
		for {
			signal := a.modelSignal()
			_, err := a.db.CurrentAccess(r.Context(), c.access.Cluster, c.access.User, c.access.Device)
			if err != nil {
				return
			}
			_ = ws.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if websocket.JSON.Send(ws, true) != nil {
				return
			}
			select {
			case <-signal:
			case <-end.C:
				return
			case <-a.gitContext.Done():
				return
			case <-r.Context().Done():
				return
			}
		}
	}}.ServeHTTP(w, r)
}
func (c *WorkerClient) WatchModels(ctx context.Context, hint func()) error {
	config, err := websocket.NewConfig("wss"+strings.TrimPrefix(c.base, "https")+"/v1/chat/models/live", "https://apparat.invalid")
	if err != nil {
		return err
	}
	config.TlsConfig = c.http.Transport.(*http.Transport).TLSClientConfig.Clone()
	config.Header.Set("Authorization", "Bearer "+c.token)
	ws, err := config.DialContext(ctx)
	if err != nil {
		return err
	}
	defer ws.Close()
	stop := context.AfterFunc(ctx, func() {
		ws.Close()
	})
	defer stop()
	ws.MaxPayloadBytes = 32
	for {
		var changed bool
		err = websocket.JSON.Receive(ws, &changed)
		if err != nil {
			return err
		}
		if changed {
			hint()
		}
	}
}
