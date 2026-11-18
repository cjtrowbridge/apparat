// Runs project-scoped workspace commands against the registered checkout.
// Durable drafts retain their original base until the user explicitly rebases them.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type documentWorkspace interface {
	WorkspaceFile(string, domain.WorkspaceRequest) (domain.WorkspaceView, error)
}

func (n *Runtime) workspaceFile(location string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	if strings.HasPrefix(location, "content://") {
		docs, ok := n.documents.(documentWorkspace)
		if !ok {
			return domain.WorkspaceView{}, errors.New("this folder provider has no workspace access")
		}
		return docs.WorkspaceFile(location, q)
	}
	return projectgit.Files(location, q)
}
func (n *Runtime) Workspace(ctx context.Context, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	n.workspaceMu.Lock()
	defer n.workspaceMu.Unlock()
	err = ctx.Err()
	if err != nil {
		return v, err
	}
	if !domain.ValidID(q.ProjectID) || !projectgit.WorkspacePath(q.Path) || !projectgit.ValidText(q.Text) || len(q.Argument) > 4096 {
		return v, errors.New("invalid workspace request")
	}
	projects, err := n.database.Projects(ctx)
	if err != nil {
		return v, err
	}
	var location string
	for _, p := range projects {
		if p.ID == q.ProjectID {
			location = p.Location
		}
	}
	if location == "" {
		return v, errors.New("project registration unavailable")
	}
	switch q.Action {
	case "read", "create", "list", "mkdir":
		v, err = n.workspaceFile(location, q)
		if err != nil || !v.Editing {
			return v, err
		}
		err = n.reconcileFileSaves(ctx, q, v)
		if err != nil {
			return v, err
		}
		base, body, found, err := n.database.Draft(ctx, q.ProjectID, q.Path)
		if err != nil {
			return v, err
		}
		if found {
			v.Report = "Recovered draft"
			if v.Base != base {
				v.Report = "Draft conflicts with the current file. Review both versions before rebasing."
			}
			v.Base = base
			v.Text = body
		}
		return v, nil
	case "draft", "save", "rebase", "compare":
		if q.Base == "" {
			return v, errors.New("open the file before editing")
		}
		if q.Action == "compare" || q.Action == "rebase" {
			read := q
			read.Action = "read"
			v, err = n.workspaceFile(location, read)
			if err != nil {
				return v, err
			}
			if q.Action == "compare" {
				v.ComparedBase = v.Base
				v.Report = "Current checkout contents\n" + v.Text
				v.Text = q.Text
				v.Base = q.Base
				return v, nil
			}
			if q.Argument != "" && q.Argument != v.Base {
				return v, errors.New("checkout changed again; review the latest version")
			}
			q.Base = v.Base
		}
		if q.Action == "save" {
			return n.saveWorkspaceFile(ctx, location, q)
		}
		err = n.database.SaveDraft(ctx, q.ProjectID, q.Path, q.Base, q.Text)
		if err != nil {
			return v, err
		}
		v = domain.WorkspaceView{Editing: true, Text: q.Text, Base: q.Base, Report: "Draft saved"}
		return v, nil
	case "discard":
		err = n.database.ClearDraft(ctx, q.ProjectID, q.Path)
		if err != nil {
			return v, err
		}
		q.Action = "read"
		return n.workspaceFile(location, q)
	case "move", "delete":
		v, err = n.workspaceFile(location, q)
		if err == nil {
			err = n.database.ClearDraft(ctx, q.ProjectID, q.Path)
		}
		return v, err
	case "review", "changes":
		var raw string
		if strings.HasPrefix(location, "content://") {
			file, e := n.workspaceFile(location, q)
			raw, err = file.Report, e
		} else {
			raw, err = projectgit.GitAction(location, q)
		}
		if err == nil {
			err = json.Unmarshal([]byte(raw), &v)
		}
		return v, err
	case "transactions":
		v.Report, err = n.database.ProjectEntries(ctx, q.ProjectID, "save")
		return v, err
	case "threads":
		v.Choices, err = n.database.ChatThreads(ctx, q.ProjectID)
		return v, err
	case "chat", "post":
		thread := strings.TrimSpace(q.Thread)
		if thread == "" {
			thread = "General"
		}
		if len(thread) > 100 {
			return v, errors.New("thread title exceeds 100 bytes")
		}
		kind := "chat:" + thread
		if q.Action == "post" {
			if strings.TrimSpace(q.Argument) == "" || !domain.ValidID(q.ID) {
				return v, errors.New("enter a message")
			}
			body := "Pending on this device · " + q.Argument
			if q.RelatedRun != "" {
				runKind, record, lookup := n.database.ProjectEntry(ctx, q.ProjectID, q.RelatedRun)
				if lookup != nil || runKind != "run" {
					return v, errors.New("run unavailable in this project")
				}
				var run artifactRun
				err = json.Unmarshal([]byte(record), &run)
				if err != nil {
					return v, err
				}
				body += "\nAttached run: " + run.Task + " · " + run.State + " (" + q.RelatedRun + ")"
			}
			err = n.database.AddProjectEntry(ctx, q.ProjectID, q.ID, kind, body)
			if err != nil {
				return v, err
			}
		}
		messages, readErr := n.database.RecentEntries(ctx, q.ProjectID, kind)
		err = readErr
		for i := len(messages) - 1; i >= 0; i-- {
			v.Report += messages[i].Label + "\n"
		}
		if v.Report == "" {
			v.Report = "No pending project messages"
		}
		return v, err
	case "tasks", "run", "runs", "artifact", "delete-artifact":
		return n.projectTask(ctx, location, q)
	default:
		if strings.HasPrefix(location, "content://") {
			return n.workspaceFile(location, q)
		}
		v.Report, err = projectgit.GitAction(location, q)
		if len(v.Report) > domain.WorkspaceTextLimit {
			v.Report = v.Report[:domain.WorkspaceTextLimit] + "\nDisplay limit reached"
		}
		if err == nil && q.Action == "commit" {
			v.Report = fmt.Sprintf("%s · %s", v.Report, time.Now().Format("15:04"))
		}
		return v, err
	}
}
