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
	WorkspaceFile(context.Context, string, domain.WorkspaceRequest) (domain.WorkspaceView, error)
}

func (n *Runtime) workspaceFile(ctx context.Context, location string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	if strings.HasPrefix(location, "content://") {
		docs, ok := n.documents.(documentWorkspace)
		if !ok {
			return domain.WorkspaceView{}, errors.New("this folder provider has no workspace access")
		}
		return docs.WorkspaceFile(ctx, location, q)
	}
	return projectgit.Files(location, q)
}
func (n *Runtime) workspaceThread(ctx context.Context, location, user string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	switch q.Action {
	case "agent-create", "agent-chat", "agent-rename", "agent-terminal-result", "agent-post":
		return n.workspaceAgentChat(ctx, location, user, q)
	case "threads":
		v.Choices, err = n.database.ChatThreads(ctx, q.ProjectID)
		if err != nil || strings.HasPrefix(location, "content://") {
			return v, err
		}
		headers, err := projectgit.ChatHeaders(location, q.Known)
		for _, h := range headers {
			if h.Scope == domain.ProjectChat {
				v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: h.ID, Label: h.Description, State: "session"})
			}
		}
		if len(headers) == 32 {
			v.Cursor = headers[len(headers)-1].ID
		}
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
			q.Thread = thread
			err = n.database.AddChatEntry(ctx, q, kind, body)
			if err != nil {
				return v, err
			}
		}
		messages, _, readErr := n.database.EntryPage(ctx, q.ProjectID, kind, "")
		err = readErr
		for i := len(messages) - 1; i >= 0; i-- {
			message := messages[i]
			v.Report += message.Label + "\n"
			message.Label = strings.TrimPrefix(message.Label, "Pending on this device · ")
			v.Choices = append(v.Choices, message)
		}
		if v.Report == "" {
			v.Report = "No pending project messages"
		}
		return v, err
	}
	return v, errors.New("unknown conversation operation")
}

func (n *Runtime) Workspace(ctx context.Context, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	start := time.Now()
	if q.Correlation == "" {
		q.Correlation, _ = domain.NewID(start, nil)
	}
	defer func() {
		n.observeWorkspace(start, q, v, err)
	}()
	if strings.HasPrefix(q.ProjectID, "worker-") {
		return n.remoteWorkspace(ctx, q)
	}
	// Directory decoration is a read-only observation and must not hold up navigation or edits.
	if q.Action != "directory-status" {
		n.workspaceMu.Lock()
		defer n.workspaceMu.Unlock()
	}
	err = ctx.Err()
	if err != nil {
		return v, err
	}
	if !domain.ValidID(q.ProjectID) || !projectgit.WorkspacePath(q.Path) || !projectgit.ValidText(q.Text) || len(q.Argument) > 4096 {
		return v, errors.New("invalid workspace request")
	}
	location, err := n.database.ProjectLocation(ctx, q.ProjectID)
	if err != nil {
		return v, errors.New("project registration unavailable")
	}
	if terminalScope(q.Action) != "" {
		return n.localTerminal(ctx, location, q)
	}
	switch q.Action {
	case "read", "create", "list", "mkdir":
		if q.Action == "list" && !strings.HasPrefix(location, "content://") {
			return n.pages.Page("local/"+q.ProjectID, location, q.Path, q.Argument)
		}
		v, err = n.workspaceFile(ctx, location, q)
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
			v, err = n.workspaceFile(ctx, location, read)
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
		if q.Action == "draft" && projectgit.Digest(q.Text) == q.Base {
			err = n.database.ClearDraft(ctx, q.ProjectID, q.Path)
		} else {
			err = n.database.SaveDraft(ctx, q.ProjectID, q.Path, q.Base, q.Text)
		}
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
		return n.workspaceFile(ctx, location, q)
	case "move", "delete":
		v, err = n.workspaceFile(ctx, location, q)
		if err == nil {
			err = n.database.ClearDraft(ctx, q.ProjectID, q.Path)
		}
		return v, err
	case "review", "changes", "directory-status":
		return n.directoryView(ctx, location, q)
	case "transactions":
		v.Report, err = n.database.ProjectEntries(ctx, q.ProjectID, "save")
		return v, err
	case "threads", "chat", "post", "agent-create", "agent-chat", "agent-rename", "agent-terminal-result", "agent-post":
		var user string
		if n.Personal != nil {
			user = n.Personal.Fingerprint()
		}
		return n.workspaceThread(ctx, location, user, q)
	case "tasks":
		return n.workspaceCatalog(ctx, location, q)
	case "run", "runs", "artifact", "delete-artifact":
		return n.projectTask(ctx, location, q)
	default:
		if strings.HasPrefix(location, "content://") {
			return n.workspaceFile(ctx, location, q)
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

func (n *Runtime) directoryView(ctx context.Context, location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	var raw string
	if strings.HasPrefix(location, "content://") {
		file, e := n.workspaceFile(ctx, location, q)
		raw, err = file.Report, e
	} else {
		raw, err = projectgit.GitActionContext(ctx, location, q)
	}
	if err == nil {
		err = json.Unmarshal([]byte(raw), &v)
	}
	return v, err
}
