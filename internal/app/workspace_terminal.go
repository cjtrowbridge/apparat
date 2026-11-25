// Maps terminal viewing and control to separate explicit grants on the project owner.
// The worker periodically checks the starting user's admission and terminates revoked sessions.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func terminalScope(action string) string {
	switch action {
	case "chat-terminal-permissions", "chat-terminal-result", "chat-terminal-observe":
		return "terminal-view"
	case "chat-terminal-permission-set", "chat-terminal-approved":
		return "terminal-control"
	}
	switch strings.TrimPrefix(action, "chat-") {
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
	user := "local"
	if strings.HasPrefix(q.Action, "chat-terminal-") {
		user = n.PlanningUser()
		if user == "" {
			return domain.WorkspaceView{}, errors.New("unlock your personal identity to use the chat terminal")
		}
	}
	return n.terminals.Do(ctx, filepath.Join(n.config.RootDir, "terminal"), location, user, "local", q, func() bool {
		return user == "local" || n.PlanningUser() == user
	})
}

// ChatTerminal routes human terminal interaction from the signed-in HUD.
// Project addresses retain the owner's workspace authorization and terminal grants.
func (n *Runtime) ChatTerminal(ctx context.Context, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	if !strings.HasPrefix(q.Action, "chat-terminal-") || terminalScope(q.Action) == "" && q.Action != "chat-terminal-hosts" && q.Action != "chat-terminal-bind" || !domain.ValidID(q.Thread) {
		return domain.WorkspaceView{}, errors.New("invalid chat terminal request")
	}
	if q.Action == "chat-terminal-approved" {
		user := n.PlanningUser()
		if user == "" || !domain.ValidID(q.ID) || time.Until(q.Deadline) <= 0 || time.Until(q.Deadline) > 30*time.Minute {
			return domain.WorkspaceView{}, errors.New("a signed-in user, valid command ID and current deadline are required")
		}
		err := n.database.RememberTerminalCallback(ctx, user, q)
		if err != nil {
			return domain.WorkspaceView{}, err
		}
	}
	q.Path = "."
	if q.ProjectID != "" {
		return n.Workspace(ctx, q)
	}
	user := n.PlanningUser()
	if user == "" {
		return domain.WorkspaceView{}, errors.New("unlock your personal identity to use the global terminal")
	}
	if q.Action == "chat-terminal-hosts" {
		var v domain.WorkspaceView
		for _, p := range n.remoteProjects(ctx) {
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: p.ID, Label: p.RemoteHost + " / " + p.Name})
		}
		if len(v.Choices) == 0 {
			v.Report = "No authorized worker checkout is available."
		}
		return v, nil
	}
	choice := ""
	if q.Action == "chat-terminal-bind" {
		if !strings.HasPrefix(q.Argument, "worker-") {
			return domain.WorkspaceView{}, errors.New("select a worker checkout")
		}
		probe := q
		probe.Action = "chat-terminal-permissions"
		probe.ProjectID = q.Argument
		_, err := n.Workspace(ctx, probe)
		if err != nil {
			return domain.WorkspaceView{}, err
		}
		choice = q.Argument
	}
	host, err := n.database.TerminalProject(ctx, user, q.Thread, choice)
	if err != nil {
		return domain.WorkspaceView{}, err
	}
	if q.Action == "chat-terminal-bind" {
		return domain.WorkspaceView{Detail: host}, nil
	}
	if host != "" {
		q.ProjectID = host
		v, err := n.Workspace(ctx, q)
		v.Base = host
		return v, err
	}
	location := filepath.Join(n.config.RootDir, "data", "repos")
	if q.Action == "chat-terminal-start" || q.Action == "chat-terminal-approved" {
		err := os.MkdirAll(location, 0700)
		if err != nil {
			return domain.WorkspaceView{}, err
		}
	}
	q.ProjectID = "global"
	return n.terminals.Do(ctx, filepath.Join(n.config.RootDir, "terminal"), location, user, "local", q, func() bool {
		return n.PlanningUser() == user
	})
}

func (n *Runtime) ChatTerminalDevice() string {
	if n.Keys == nil || len(n.Keys.Public()) == 0 {
		return ""
	}
	return identity.Fingerprint(n.Keys.Public())
}
