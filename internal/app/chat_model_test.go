// Verifies durable provider handoff, replay and scoped tool continuations.
// All model responses come from a controllable in-process async provider.
package app_test

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

type captureProvider func(context.Context, domain.ChatRequest) (*domain.ChatOutput, error)

func (p captureProvider) Dispatch(ctx context.Context, r domain.ChatRequest) (*domain.ChatOutput, error) {
	return p(ctx, r)
}

func TestAsynchronousChatToolRecovery(t *testing.T) {
	root, repo := t.TempDir(), t.TempDir()
	ctx := context.Background()
	gitFixtureCommand(t, repo, "init", "-b", "main")
	node := nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	other := t.TempDir()
	gitFixtureCommand(t, other, "init", "-b", "main")
	requireSuccess(t, node.StoreProject(ctx, "", other, ""))
	projects, err := node.ProjectPage(ctx, "local", "")
	requireSuccess(t, err)
	id := func() string {
		value, err := domain.NewID(time.Now(), nil)
		requireSuccess(t, err)
		return value
	}
	a := domain.ChatAddress{Project: projects[0].ID, Session: id(), Turn: id()}
	_, err = node.CreateChat(ctx, a.Project, a.Session, domain.ProjectChat, "Catalog query")
	requireSuccess(t, err)
	firstModel := domain.ModelRef{Provider: "ollama", Name: "qwen:a", Digest: "a"}
	_, err = node.SetChatSessionModel(ctx, a.Project, a.Session, firstModel)
	requireSuccess(t, err)
	_, err = node.SubmitChatTurn(ctx, a.Project, a.Session, a.Turn, "What project is this?")
	requireSuccess(t, err)
	requests := make(chan domain.ChatRequest, 8)
	var count atomic.Int32
	p := captureProvider(func(_ context.Context, r domain.ChatRequest) (*domain.ChatOutput, error) {
		count.Add(1)
		requests <- r
		return nil, nil
	})
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			_, e := node.DriveChat(ctx, a, "fixture", p)
			if e != nil {
				t.Error(e)
			}
		})
	}
	group.Wait()
	if count.Load() != 1 {
		t.Fatal("duplicate or missing dispatch", count.Load())
	}
	first := <-requests
	_, err = node.SetChatSessionModel(ctx, a.Project, a.Session, domain.ModelRef{Provider: "ollama", Name: "qwen:b", Digest: "b"})
	requireSuccess(t, err)
	if first.Model != firstModel {
		t.Fatal("incorrect initial turn model")
	}
	requireSuccess(t, node.Close())
	node = nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	_, err = node.DriveChat(ctx, a, "fixture", p)
	requireSuccess(t, err)
	recovered, err := node.OutstandingChatRequest(ctx, a)
	requireSuccess(t, err)
	if recovered.Model != firstModel || recovered.Job != first.Job || len(recovered.Messages) != 1 {
		t.Fatal("lost claimed packet")
	}
	if count.Load() != 1 {
		t.Fatal("redispatched after restart")
	}
	out := domain.ChatOutput{Finish: "tool_calls", Calls: []domain.ToolCall{{ID: "c1", Name: "projects", Arguments: "{}"}, {ID: "c2", Name: "unknown", Arguments: "{}"}}}
	for range 2 {
		_, err = node.DeliverChatOutput(ctx, a, first.Job, out)
		requireSuccess(t, err)
	}
	_, err = node.DriveChat(ctx, a, "fixture", p)
	requireSuccess(t, err)
	next := <-requests
	if next.Job == first.Job || len(next.Messages) != 4 || next.Messages[2].CallID != "c1" || !strings.Contains(next.Messages[2].Text, a.Project) || strings.Contains(next.Messages[2].Text, projects[1].ID) || !strings.Contains(next.Messages[3].Text, "Unknown tool") {
		t.Fatal("incorrect correlated tool history", next.Messages)
	}
	answer := domain.ChatOutput{Finish: "stop", Text: "Found your project."}
	_, err = node.DeliverChatOutput(ctx, a, next.Job, answer)
	requireSuccess(t, err)
	_, err = node.DeliverChatOutput(ctx, a, next.Job, domain.ChatOutput{Finish: "stop", Text: "Changed"})
	if err == nil {
		t.Fatal("accepted conflicting replay")
	}
	a.Turn = id()
	_, err = node.SubmitChatTurn(ctx, a.Project, a.Session, a.Turn, "Please continue.")
	requireSuccess(t, err)
	_, err = node.DriveChat(ctx, a, "fixture", p)
	requireSuccess(t, err)
	last := <-requests
	if len(last.Messages) != 6 || last.Messages[4].Text != answer.Text {
		t.Fatal("lost conversation history")
	}
	_, err = node.StopChatTurn(ctx, a)
	requireSuccess(t, err)
	_, err = node.DeliverChatOutput(ctx, a, last.Job, answer)
	if err == nil {
		t.Fatal("late reply revived cancelled turn")
	}
}

func TestHundredSessionsCanWaitWhileOneFinishes(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	gitFixtureCommand(t, repo, "init", "-b", "main")
	node := nodeFor(t, t.TempDir(), false)
	requireSuccess(t, node.Initialize(ctx))
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	rows, err := node.ProjectPage(ctx, "local", "")
	requireSuccess(t, err)
	project := rows[0].ID
	requests := make(chan domain.ChatRequest, 100)
	p := captureProvider(func(_ context.Context, r domain.ChatRequest) (*domain.ChatOutput, error) {
		requests <- r
		return nil, nil
	})
	var group sync.WaitGroup
	for range 100 {
		group.Go(func() {
			// Old timestamps exercise the lack of a session-age admission deadline.
			sid, e := domain.NewID(time.Unix(1, 0), nil)
			if e != nil {
				t.Error(e)
				return
			}
			turn, e := domain.NewID(time.Unix(2, 0), nil)
			if e != nil {
				t.Error(e)
				return
			}
			_, e = node.CreateChat(ctx, project, sid, domain.ProjectChat, "Waiting for capacity")
			if e != nil {
				t.Error(e)
				return
			}
			_, e = node.SubmitChatTurn(ctx, project, sid, turn, "Wait for a reply")
			if e != nil {
				t.Error(e)
				return
			}
			_, e = node.DriveChat(ctx, domain.ChatAddress{Project: project, Session: sid, Turn: turn}, "pool", p)
			if e != nil {
				t.Error(e)
			}
		})
	}
	group.Wait()
	if len(requests) != 100 {
		t.Fatal("lost a session", len(requests))
	}
	r := <-requests
	s, err := node.DeliverChatOutput(ctx, r.ChatAddress, r.Job, domain.ChatOutput{Finish: "length", Calls: []domain.ToolCall{{ID: "partial", Name: "projects", Arguments: "{"}}})
	requireSuccess(t, err)
	if s.Pending() != "" || len(s.Turns[0].Steps[0].Results) != 0 {
		t.Fatal("executed truncated response")
	}
	r = <-requests
	s, err = node.Chat(ctx, project, r.Session)
	requireSuccess(t, err)
	if s.Pending() != r.Turn {
		t.Fatal("unrelated session lost its wait")
	}
	seen := map[string]bool{}
	cursor := ""
	for {
		headers, err := node.Chats(ctx, project, cursor)
		requireSuccess(t, err)
		for _, h := range headers {
			if seen[h.ID] {
				t.Fatal("repeated session across pages")
			}
			seen[h.ID] = true
		}
		if len(headers) < 32 {
			break
		}
		cursor = headers[len(headers)-1].ID
	}
	if len(seen) != 100 {
		t.Fatal("incomplete session directory")
	}
}
