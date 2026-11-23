// Maps terminal viewing and control to separate explicit grants on the project owner.
// The worker periodically checks the starting user's admission and terminates revoked sessions.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"path/filepath"
	"strings"
)

func terminalScope(action string) string {
	switch action {
	case "terminal-list", "terminal-history", "terminal-screen":
		return "terminal-view"
	case "terminal-start", "terminal-stop", "terminal-input", "terminal-resize", "terminal-take":
		return "terminal-control"
	}
	return ""
}
func (n *Runtime) ownerTerminal(ctx context.Context, caller persistence.Access, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	scope := terminalScope(q.Action)
	if scope == "" {
		return domain.WorkspaceView{}, httpapi.ErrDenied
	}
	location, err := n.grantedLocation(ctx, q.ProjectID, caller.User, scope)
	if err != nil {
		return domain.WorkspaceView{}, err
	}
	_, err = n.grantedLocation(ctx, q.ProjectID, caller.User, "terminal-view")
	if err != nil {
		return domain.WorkspaceView{}, err
	}
	db := n.database
	return n.terminals.Do(ctx, filepath.Join(n.config.RootDir, "terminal"), location, caller.User, caller.Device, q, func() bool {
		access, err := db.CurrentAccess(context.Background(), caller.Cluster, caller.User, caller.Device)
		if err != nil || access.Admission != "approved" {
			return false
		}
		_, err = db.WorkspaceLocation(context.Background(), q.ProjectID, caller.User, "terminal-control")
		if err != nil {
			return false
		}
		_, err = db.WorkspaceLocation(context.Background(), q.ProjectID, caller.User, "terminal-view")
		return err == nil
	})
}
func (n *Runtime) localTerminal(ctx context.Context, location string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	if strings.HasPrefix(location, "content://") {
		return domain.WorkspaceView{}, errors.New("Select a Linux terminal host")
	}
	return n.terminals.Do(ctx, filepath.Join(n.config.RootDir, "terminal"), location, "local", "local", q, func() bool {
		return true
	})
}
