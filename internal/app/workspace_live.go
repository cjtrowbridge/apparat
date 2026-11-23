// Subscribes only while a foreground client is using an owner-backed directory.
// Reconnection requests fresh snapshots after watch loss and backs off on transport errors.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"time"
)

func (n *Runtime) WatchDirectory(ctx context.Context, id, dir, session string, update func(*domain.WorkspaceView)) error {
	host, project, _ := strings.Cut(id, "/")
	delay := time.Second
	for ctx.Err() == nil {
		client, err := n.workspaceClient(ctx, host)
		if err == nil {
			err = client.Watch(ctx, []httpapi.Watch{{Project: project, Directory: dir, Session: session}}, func(note httpapi.Notice) {
				delay = time.Second
				update(note.Screen)
			})
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = min(30*time.Second, delay*2)
	}
	return ctx.Err()
}
