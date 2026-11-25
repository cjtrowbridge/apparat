// Exercises project and global terminal tool handoffs on a disposable authenticated worker.
// The fixture models a completed inference step and drives the same callbacks as the HUD.
package app

import (
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func checkClusterTerminalToolContinuation(t *testing.T, nodes []inferenceFixture, roster domain.InferenceRoster, user *identity.Personal, device *identity.Session) {
	t.Helper()
	ctx := t.Context()
	w := nodes[0]
	id := func() string {
		s, e := domain.NewID(time.Now(), nil)
		mustInference(t, e)
		return s
	}
	project := id()
	mustInference(t, w.p.db.RegisterProject(ctx, domain.Project{ID: project, Name: "Terminal fixture", Location: t.TempDir()}))
	model := domain.ModelRef{Provider: "ollama", Name: "test", Digest: "test"}
	var err error
	for _, global := range []bool{false, true} {
		a := domain.ChatAddress{Project: project, Session: id(), Turn: id()}
		q := domain.WorkspaceRequest{Action: "agent-post", Path: ".", ProjectID: project, Thread: a.Session, ID: a.Session, Argument: "Read the terminal screen", Model: model}
		a.Turn = a.Session
		g := domain.GlobalRequest{Session: a.Session, Turn: a.Turn, Input: q.Argument, Title: "Terminal fixture", Model: model, TerminalHost: identity.Fingerprint(device.Public())}
		if global {
			a.Project = "global/" + user.Fingerprint()
			for _, candidate := range nodes {
				if candidate.peer.ID == roster.GlobalWriter(user.Fingerprint(), a.Session) {
					w = candidate
					break
				}
			}
			_, err = w.c.Inference(ctx, domain.InferenceCommand{Operation: "global-post", Global: g})
		} else {
			_, err = w.c.Workspace(ctx, q)
		}
		mustInference(t, err)
		n := w.p.runtime
		s, e := n.DriveChat(ctx, a, "cluster", queuedProvider{})
		mustInference(t, e)
		job := s.Turns[0].Steps[0].Job
		output := domain.ChatOutput{Finish: "tool_calls", Calls: []domain.ToolCall{{ID: "screen", Name: "terminal_screen", Arguments: "{}"}}}
		_, e = n.DeliverChatOutput(ctx, a, job, output)
		mustInference(t, e)
		s, e = n.DriveChat(ctx, a, "cluster", queuedProvider{})
		mustInference(t, e)
		if s.TerminalWait() == nil || len(s.Turns[0].Steps) != 1 {
			t.Fatal("model did not suspend")
		}
		if global {
			g.Turn, g.Input = job, "/project"
			_, err = w.c.Inference(ctx, domain.InferenceCommand{Operation: "global-terminal-result", Global: g})
		} else {
			q.Action, q.ID, q.Argument, q.Approved = "agent-terminal-result", job, "/project", true
			_, err = w.c.Workspace(ctx, q)
		}
		mustInference(t, err)
		s, e = n.DriveChat(ctx, a, "cluster", queuedProvider{})
		mustInference(t, e)
		if s.TerminalWait() != nil || len(s.Turns[0].Steps) != 2 || s.Turns[0].Steps[0].Results[0].Content != "/project" {
			t.Fatal("receipt did not resume the expected turn")
		}
	}
}
