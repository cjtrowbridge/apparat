// Routes remote project operations to their registered owner and keeps drafts on the editing device.
// Cluster membership permits browsing and the current admin role permits file writes.
package app

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"strings"
	"time"
)

func (n *Runtime) observeWorkspace(start time.Time, q domain.WorkspaceRequest, v domain.WorkspaceView, err error) {
	correlation := q.Correlation
	if !domain.ValidID(correlation) {
		correlation = ""
	}
	n.observations.Observe(telemetry.Event{Operation: q.Action, Elapsed: time.Since(start).Microseconds(), Received: len(q.Text), Sent: len(v.Text) + len(v.Report), Rows: len(v.Entries) + len(v.Choices), Failed: err != nil, Correlation: correlation})
}
func (n *Runtime) grantedLocation(ctx context.Context, project, user, scope string) (string, error) {
	location, err := n.database.WorkspaceLocation(ctx, project, user, scope)
	if errors.Is(err, sql.ErrNoRows) {
		return "", httpapi.WorkspaceGrantNeeded{Scope: scope}
	}
	return location, err
}
func (n *Runtime) serveWorkspace(ctx context.Context, caller persistence.Access, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	start := time.Now()
	defer func() {
		n.observeWorkspace(start, q, v, err)
	}()
	if !projectgit.WorkspacePath(q.Path) || !projectgit.ValidText(q.Text) || len(q.Path) > 4096 || len(q.Argument) > 4096 {
		return v, httpapi.ErrDenied
	}
	if q.Action == "performance-capture" && caller.Admin {
		var data strings.Builder
		err = n.capture(ctx, 0).Write(&data)
		v.Detail = data.String()
		return v, err
	}
	if terminalScope(q.Action) != "" {
		return n.ownerTerminal(ctx, caller, q)
	}
	scope := "read"
	switch q.Action {
	case "save", "post", "agent-create", "agent-rename", "agent-terminal-result", "agent-post":
		scope = "write"
	case "run":
		scope = "tasks"
	case "chat", "agent-chat", "threads", "list", "read", "tasks", "runs", "artifact", "directory-status":
	default:
		return v, httpapi.ErrDenied
	}
	location, err := n.grantedLocation(ctx, q.ProjectID, caller.User, scope)
	if err != nil {
		return v, err
	}
	if scope != "read" {
		_, err = n.grantedLocation(ctx, q.ProjectID, caller.User, "read")
		if err != nil {
			return v, err
		}
	}
	switch q.Action {
	case "chat", "threads", "post", "agent-chat", "agent-create", "agent-rename", "agent-terminal-result", "agent-post":
		n.workspaceMu.Lock()
		defer n.workspaceMu.Unlock()
		return n.workspaceThread(ctx, location, caller.User, q)
	case "list":
		v, err = n.pages.Page(caller.Device+"/"+q.ProjectID, location, q.Path, q.Argument)
		return conditional(q, v, err)
	case "read":
		v, err = n.workspaceFile(ctx, location, q)
		v, err = conditional(q, v, err)
		v.ReadOnly = !caller.Admin
		return v, err
	case "tasks":
		return n.workspaceCatalog(ctx, location, q)
	case "directory-status":
		return n.directoryView(ctx, location, q)
	case "save":
		n.workspaceMu.Lock()
		defer n.workspaceMu.Unlock()
		return n.remoteSave(ctx, caller.User, location, q)
	default:
		return n.projectTask(ctx, location, q)
	}
}
func (n *Runtime) remoteSave(ctx context.Context, user, location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	if !domain.ValidID(q.ID) || len(q.Base) != 64 {
		return v, errors.New("invalid save")
	}
	digest := projectgit.Digest(q.Text)
	old, done, err := n.database.PrepareRemoteSave(ctx, q.ProjectID, user, q.ID, q.Path, q.Base, digest)
	if err != nil {
		return v, err
	}
	if done {
		return domain.WorkspaceView{Base: digest}, nil
	}
	read := domain.WorkspaceRequest{Action: "read", Path: q.Path}
	current, err := n.workspaceFile(ctx, location, read)
	if err != nil {
		return v, err
	}
	if current.Base != q.Base && !(old && current.Base == digest) {
		return v, httpapi.ErrWorkspaceConflict
	}
	if current.Base != digest {
		_, err = n.workspaceFile(ctx, location, q)
		if err != nil {
			return v, err
		}
	}
	err = n.database.FinishRemoteSave(ctx, q.ProjectID, user, q.ID)
	return domain.WorkspaceView{Base: digest}, err
}
func (n *Runtime) remoteWorkspace(ctx context.Context, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	host, id, ok := strings.Cut(q.ProjectID, "/")
	if !ok || !projectgit.WorkspacePath(q.Path) || !projectgit.ValidText(q.Text) {
		return v, httpapi.ErrDenied
	}
	if n.Personal == nil {
		return v, httpapi.ErrWorkspaceLocked
	}
	user := n.Personal.Fingerprint()
	if user == "" {
		return v, httpapi.ErrWorkspaceLocked
	}
	full := q.ProjectID
	base, body, op, found, err := n.database.RemoteDraft(ctx, user, full, q.Path)
	if err != nil {
		return v, err
	}
	if q.Action == "draft" {
		if projectgit.Digest(q.Text) == q.Base {
			err = n.database.ClearRemoteDraft(ctx, user, full, q.Path, base, body)
		} else {
			err = n.database.PutRemoteDraft(ctx, user, full, q.Path, q.Base, q.Text, "")
		}
		return domain.WorkspaceView{Editing: true, Base: q.Base, Text: q.Text}, err
	}
	draftText := q.Text
	original := q.Action
	if q.Action == "save" {
		if found && base == q.Base && body == q.Text && op != "" {
			q.ID = op
		}
		err = n.database.PutRemoteDraft(ctx, user, full, q.Path, q.Base, q.Text, q.ID)
		if err != nil {
			return v, err
		}
	}
	c, err := n.workspaceClient(ctx, host)
	if err != nil {
		return v, err
	}
	if c.Access.User != user {
		return v, httpapi.ErrDenied
	}
	if original == "compare" || original == "rebase" {
		q.Action = "read"
	}
	q.ProjectID = id
	if q.Action != "save" {
		q.Text = ""
	}
	key := cacheKey{user, full, q.Path, q.Action}
	cacheable := (q.Action == "read" || q.Action == "list" && q.Argument == "")
	var cached domain.WorkspaceView
	if cacheable {
		cached, _ = n.workspaceCache.get(key)
		q.Known = cached.Base
	}
	v, err = c.Workspace(ctx, q)
	if errors.Is(err, httpapi.ErrDenied) {
		connection := &n.workspaceConnection
		connection.mu.Lock()
		for k, v := range connection.entries {
			if v.client == c {
				delete(connection.entries, k)
			}
		}
		connection.mu.Unlock()
	}
	if err != nil {
		return v, err
	}
	if cacheable {
		if v.Unchanged {
			cached.ReadOnly = v.ReadOnly
			v = cached
		} else if v.Cursor == "" {
			n.workspaceCache.put(key, v)
		}
	}
	switch original {
	case "save":
		err = n.database.ClearRemoteDraft(ctx, user, full, q.Path, q.Base, draftText)
	case "read":
		if found && !v.ReadOnly {
			v.Report = "Recovered local draft"
			if base != v.Base {
				v.Report = "Owner file changed. Compare before saving."
			}
			v.Text = body
			v.Base = base
		}
	case "compare":
		v.ComparedBase = v.Base
		v.Report = "Current checkout contents\n" + v.Text
		v.Text = body
		v.Base = base
	case "rebase":
		if q.Argument == "" || q.Argument != v.Base {
			return v, httpapi.ErrWorkspaceConflict
		}
		v.Text = draftText
		err = n.database.PutRemoteDraft(ctx, user, full, q.Path, v.Base, draftText, "")
	}
	return v, err
}
