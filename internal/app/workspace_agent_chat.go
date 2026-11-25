// Projects the owner repository transcript into the authenticated workspace view.
// The workspace caller owns the write lock and has already checked the project's grants.
package app

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) workspaceAgentChat(ctx context.Context, location, user string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	var v domain.WorkspaceView
	id := q.Thread
	var initial *domain.ChatHeader
	if q.Action == "agent-create" {
		id = q.ID
		at, err := domain.ChatTime(id)
		if err != nil {
			return v, err
		}
		initial = &domain.ChatHeader{Version: 1, Scope: domain.ProjectChat, ID: id, Started: at, Description: strings.TrimSpace(q.Argument), Model: n.initialChatModel(ctx, q.ProjectID, domain.ProjectChat)}
	}
	apply := func(s *domain.ChatSession) (bool, error) {
		if s.Scope != domain.ProjectChat {
			return false, errors.New("expected a project session")
		}
		if q.Action == "agent-rename" {
			title := strings.TrimSpace(q.Argument)
			if s.Description == title {
				return false, nil
			}
			if q.Base != s.Description {
				return false, errors.New("session title changed; refresh before renaming")
			}
			s.Description = title
			return true, s.Validate()
		}
		if q.Action != "agent-post" {
			return false, nil
		}
		if !q.Model.Valid() {
			return false, errors.New("select an available model")
		}
		text := q.Argument
		if q.RelatedRun != "" {
			kind, _, err := n.database.ProjectEntry(ctx, q.ProjectID, q.RelatedRun)
			if err != nil || kind != "run" {
				return false, errors.New("run unavailable in this project")
			}
			text += "\nAttached run: " + q.RelatedRun
		}
		s.Model = q.Model
		return n.queueChatTurn(ctx, q.ProjectID, user, s, q.ID, text)
	}
	s, err := projectgit.WithChat(location, id, initial, apply)
	if q.Action == "agent-post" && q.ID == id && errors.Is(err, os.ErrNotExist) {
		at, timeErr := domain.ChatTime(id)
		if timeErr != nil {
			return v, timeErr
		}
		header := domain.ChatHeader{Version: 1, Scope: domain.ProjectChat, ID: id, Started: at, Description: domain.PromptTitle(q.Argument, 512), Model: q.Model}
		s, err = projectgit.WithChat(location, id, &header, apply)
	}
	if err != nil {
		return v, err
	}
	v.Detail = s.Description
	v.Base = s.ID
	// Keep the rendered page bounded independently of the full repository transcript.
	start := max(0, len(s.Turns)-30)
	for _, t := range s.Turns[start:] {
		v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: t.ID, Label: t.Input, State: "user"})
		switch {
		case t.Reply != "":
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: t.ID, Label: t.Reply, State: "assistant"})
		case t.Stopped != "":
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: t.ID, Label: t.Stopped, State: "stopped"})
		default:
			label := "Saved to project · awaiting dispatch"
			if len(t.Steps) > 0 {
				label = "Waiting for reply"
			}
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: t.ID, Label: label, State: "pending"})
		}
	}
	return v, nil
}
