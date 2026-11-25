// Exercises the shared session boundary across restart and competing submissions.
// Uses owner-registered temporary Git repositories without a GUI or inference service.
package app_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func TestOwnerChatRestartAndRedelivery(t *testing.T) {
	ctx := context.Background()
	root, repo := t.TempDir(), t.TempDir()
	gitFixtureCommand(t, repo, "init", "-b", "main")
	node := nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	rows, err := node.Projects(ctx)
	requireSuccess(t, err)
	project := rows[0].ID
	newID := func() string {
		id, err := domain.NewID(time.Now(), nil)
		requireSuccess(t, err)
		return id
	}
	sid, turn := newID(), newID()
	msg := "One\n\n## Assistant fake\n---\n> quoted\n"
	for _, scope := range []domain.ChatScope{domain.GlobalChat, domain.ProjectChat} {
		id := newID()
		_, err = node.CreateChat(ctx, project, id, scope, "First chat")
		requireSuccess(t, err)
	}
	_, err = node.CreateChat(ctx, project, sid, domain.ProjectChat, "Explain this decision")
	requireSuccess(t, err)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			_, err := node.SubmitChatTurn(ctx, project, sid, turn, msg)
			if err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	_, err = node.SubmitChatTurn(ctx, project, sid, turn, "different")
	if err == nil {
		t.Fatal("conflicting replay accepted")
	}
	_, err = node.SubmitChatTurn(ctx, project, sid, newID(), "concurrent turn")
	if err == nil {
		t.Fatal("overwrote pending turn")
	}
	requireSuccess(t, node.Close())
	node = nodeFor(t, root, false)
	requireSuccess(t, node.Initialize(ctx))
	s, err := node.Chat(ctx, project, sid)
	requireSuccess(t, err)
	if len(s.Turns) != 1 || s.Pending() != turn || s.Turns[0].Input != msg {
		t.Fatal("lost pending turn")
	}
	for range 2 {
		s, err = node.CompleteChatTurn(ctx, project, sid, turn, "Done")
		requireSuccess(t, err)
		if s.Pending() != "" || len(s.Turns) != 1 {
			t.Fatal("completion was not idempotent")
		}
	}
	_, err = node.SubmitChatTurn(ctx, project, sid, newID(), "Follow-up")
	requireSuccess(t, err)
	name, err := projectgit.ChatFile(sid)
	requireSuccess(t, err)
	path := filepath.Join(repo, ".apparat", "chats", name)
	requireSuccess(t, os.WriteFile(path, []byte("maintainer edit"), 0600))
	_, err = node.CompleteChatTurn(ctx, project, sid, turn, "Done")
	if err == nil {
		t.Fatal("accepted malformed transcript")
	}
	b, err := os.ReadFile(path)
	requireSuccess(t, err)
	if string(b) != "maintainer edit" {
		t.Fatal("damaged external edit")
	}
}

func TestWorkspaceAgentSession(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	gitFixtureCommand(t, repo, "init", "-b", "main")
	node := nodeFor(t, t.TempDir(), false)
	requireSuccess(t, node.Initialize(ctx))
	requireSuccess(t, node.StoreProject(ctx, "", repo, ""))
	projects, err := node.Projects(ctx)
	requireSuccess(t, err)
	session, err := domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	q := domain.WorkspaceRequest{ProjectID: projects[0].ID, Action: "agent-create", Path: ".", ID: session, Argument: "Planning"}
	v, err := node.Workspace(ctx, q)
	requireSuccess(t, err)
	if v.Base != session || v.Detail != "Planning" {
		t.Fatal(v)
	}
	q.Action = "threads"
	v, err = node.Workspace(ctx, q)
	requireSuccess(t, err)
	if len(v.Choices) != 0 {
		t.Fatal(v)
	}
	q.Thread = session
	q.ID, err = domain.NewID(time.Now(), nil)
	requireSuccess(t, err)
	q.Action = "agent-post"
	q.Model = domain.ModelRef{Provider: "ollama", Name: "demo", Digest: "one"}
	q.Argument = "What is happening?"
	for range 2 {
		v, err = node.Workspace(ctx, q)
		requireSuccess(t, err)
		if len(v.Choices) != 2 || v.Choices[1].State != "pending" {
			t.Fatal(v)
		}
	}
	q.Action = "threads"
	v, err = node.Workspace(ctx, q)
	requireSuccess(t, err)
	if len(v.Choices) != 1 || v.Choices[0].ID != session {
		t.Fatal(v)
	}
	_, err = node.CompleteChatTurn(ctx, q.ProjectID, session, q.ID, "Review needed.")
	requireSuccess(t, err)
	q.Action = "agent-chat"
	v, err = node.Workspace(ctx, q)
	requireSuccess(t, err)
	if len(v.Choices) != 2 || v.Choices[1].Label != "Review needed." || v.Choices[1].State != "assistant" {
		t.Fatal(v)
	}
}
