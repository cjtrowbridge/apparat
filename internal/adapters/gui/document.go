//go:build gui

// Opens a platform multiline editor over the same project-scoped draft and save commands.
// Only the selected document crosses the bridge. The editor never receives filesystem access.
package gui

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/ebitenui/ebitenui/widget"
	"io"
	"os/exec"
	"strings"
	"time"
)

//go:embed document_gtk.py
var desktopEditor string

type documentSession struct {
	compared string
	ctx      context.Context
	store    workspaceStore
	request  domain.WorkspaceRequest
	view     domain.WorkspaceView
}

func (s *documentSession) Apply(action, text string) string {
	q := s.request
	switch action {
	case "draft", "save", "compare", "rebase":
	default:
		return "Unknown editor action"
	}
	if action == "rebase" {
		if s.compared == "" {
			return "Review the current checkout first"
		}
		q.Argument = s.compared
	}
	q.Action = action
	q.Text = text
	if action == "save" {
		q.ID, _ = domain.NewID(time.Now(), nil)
	}
	v, err := s.store.Workspace(s.ctx, q)
	if err != nil {
		return err.Error()
	}
	if action == "compare" {
		s.compared = v.ComparedBase
		return v.Report
	}
	s.request.Base = v.Base
	s.view = v
	return ""
}
func openDesktopDocument(ctx context.Context, title, body string, bookmark *[3]int, session DocumentSession) error {
	cmd := exec.CommandContext(ctx, "python3", "-I", "-c", desktopEditor)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	writer := json.NewEncoder(in)
	reader := json.NewDecoder(out)
	err = writer.Encode([]any{title, body, bookmark})
	for err == nil {
		var request [2]string
		err = reader.Decode(&request)
		if err != nil {
			break
		}
		if request[0] == "position" {
			err = json.Unmarshal([]byte(request[1]), bookmark)
			if err != nil {
				break
			}
			err = writer.Encode("")
		} else {
			err = writer.Encode(session.Apply(request[0], request[1]))
		}
	}
	in.Close()
	wait := cmd.Wait()
	if wait != nil {
		return errors.New("document editor unavailable; install Python 3, PyGObject and GTK 4")
	}
	if err != io.EOF {
		return err
	}
	return nil
}
func (g *Game) editDocument() {
	w := &g.projects.workspace
	store, ok := g.runtimeInfo.Queries.(workspaceStore)
	if !ok {
		return
	}
	q := domain.WorkspaceRequest{ProjectID: w.id, Path: w.path, Base: w.base, Action: "document"}
	v := domain.WorkspaceView{Editing: true, Base: w.base, Text: strings.Join(w.lines, "\n"), Report: w.report}
	key := [2]string{w.id, w.path}
	bookmark := g.projects.bookmarks[key]
	host := g.runtimeInfo.DocumentEditor
	g.projectWork(func(ctx context.Context) projectResult {
		session := &documentSession{ctx: ctx, store: store, request: q, view: v}
		var err error
		if host != nil {
			err = host.EditDocument(q.ProjectID, q.Path, v.Text, session)
		} else {
			err = openDesktopDocument(ctx, q.Path, v.Text, &bookmark, session)
		}
		return projectResult{workspace: &workspaceReply{q, session.view, err}, applyField: func() {
			if w.id == q.ProjectID {
				if g.projects.bookmarks == nil {
					g.projects.bookmarks = map[[2]string][3]int{}
				}
				g.projects.bookmarks[key] = bookmark
			}
		}}
	})
}

func (g *Game) projectField(c *widget.Container, label string, value *string) {
	prompt := g.runtimeInfo.PromptValue
	if prompt == nil {
		g.projectInput(c, label, *value, func(text string) {
			*value = text
		})
		return
	}
	caption := label
	if *value != "" {
		caption += " · " + *value
	}
	g.projectButton(c, caption, func() {
		initial := *value
		g.projectWork(func(ctx context.Context) projectResult {
			text, ok := prompt(label, initial)
			return projectResult{applyField: func() {
				if ok {
					*value = text
				}
			}}
		})
	})
}
