// Delivers private, transient generation deltas over a pinned worker WebSocket.
// Only a live job's submitting identity can subscribe. Slow readers cannot block dispatch.
package httpapi

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"golang.org/x/net/websocket"
	"net/http"
	"time"
)

func (a *workerAuth) previewSocket(w http.ResponseWriter, r *http.Request, c credential) {
	job := r.URL.Query().Get("job")
	provider, ok := a.inference.(interface {
		OpenPreview(string, string) (func(int) domain.PreviewFrame, error)
	})
	if !ok || !domain.ValidID(job) || r.Method != "GET" {
		respond(w, 404, "preview unavailable")
		return
	}
	read, err := provider.OpenPreview(c.access.User, job)
	if err != nil {
		respond(w, 404, "preview unavailable")
		return
	}
	websocket.Server{Handshake: func(_ *websocket.Config, r *http.Request) error {
		if r.Header.Get("Origin") != "https://apparat.invalid" {
			return errors.New("invalid origin")
		}
		return nil
	}, Handler: func(ws *websocket.Conn) {
		defer ws.Close()
		disconnected := make(chan struct{})
		go func() {
			var b [1]byte
			_, _ = ws.Read(b[:])
			close(disconnected)
		}()
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()
		offset := 0
		last := ""
		for time.Now().Before(c.until) {
			frame := read(offset)
			if last != frame.Stage || frame.Text != "" || frame.Done {
				_, err := a.db.CurrentAccess(r.Context(), c.access.Cluster, c.access.User, c.access.Device)
				if err != nil {
					return
				}
				_ = ws.SetWriteDeadline(time.Now().Add(2 * time.Second))
				if websocket.JSON.Send(ws, frame) != nil || frame.Done {
					return
				}
				offset += len(frame.Text)
				last = frame.Stage
			}
			select {
			case <-disconnected:
				return
			case <-r.Context().Done():
				return
			case <-tick.C:
			}
		}
	}}.ServeHTTP(w, r)
}
func (c *WorkerClient) WatchPreview(ctx context.Context, job string, update func(domain.PreviewFrame)) error {
	if !domain.ValidID(job) {
		return ErrDenied
	}
	ws, err := c.openSocket(ctx, "/v1/inference/preview?job="+job)
	if err != nil {
		return err
	}
	defer ws.Close()
	stop := context.AfterFunc(ctx, func() {
		ws.Close()
	})
	defer stop()
	ws.MaxPayloadBytes = 32 << 20
	for {
		var f domain.PreviewFrame
		err = websocket.JSON.Receive(ws, &f)
		if err != nil {
			return err
		}
		update(f)
		if f.Done {
			return nil
		}
	}
}
