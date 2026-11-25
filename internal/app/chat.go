// Exposes trusted owner-local session operations to every runtime mode.
// Each call resolves the registered repository and releases the owner lock before waiting.
package app

import (
	"context"
	"errors"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

// CreateChat uses caller-generated session IDs so lost acknowledgements are retryable.
// Global scope takes an explicitly selected repository for its transcript.
func (n *Runtime) CreateChat(ctx context.Context, project, id string, scope domain.ChatScope, description string) (domain.ChatSession, error) {
	started, err := domain.ChatTime(id)
	if err != nil {
		return domain.ChatSession{}, err
	}
	h := domain.ChatHeader{Model: n.initialChatModel(ctx, project, scope), Version: 1, ID: id, Started: started, Scope: scope, Description: description}
	err = h.Validate()
	if err != nil {
		return domain.ChatSession{}, err
	}
	return n.withChat(ctx, project, id, &h, nil)
}

func (n *Runtime) Chat(ctx context.Context, project, id string) (domain.ChatSession, error) {
	return n.withChat(ctx, project, id, nil, nil)
}

func (n *Runtime) SubmitChatTurn(ctx context.Context, project, id, turn, text string) (domain.ChatSession, error) {
	return n.withChat(ctx, project, id, nil, func(s *domain.ChatSession) (bool, error) {
		if !s.Model.Valid() {
			s.Model = n.initialChatModel(ctx, project, s.Scope)
		}
		return n.applyChatInput(ctx, project, s, turn, text)
	})
}

func (n *Runtime) applyChatInput(ctx context.Context, project string, s *domain.ChatSession, turn, text string) (bool, error) {
	changed, err := s.Apply(turn, text, false)
	if err != nil || !changed || !strings.EqualFold(strings.TrimSpace(text), "list projects") {
		return changed, err
	}
	result, err := n.chatProjectList(ctx, project, s.ID, s.Scope, "", "")
	if err != nil {
		return false, err
	}
	return s.Apply(turn, result, true)
}

func (n *Runtime) CompleteChatTurn(ctx context.Context, project, id, turn, reply string) (domain.ChatSession, error) {
	return n.withChat(ctx, project, id, nil, func(s *domain.ChatSession) (bool, error) {
		return s.Apply(turn, reply, true)
	})
}

func (n *Runtime) withChat(ctx context.Context, project, id string, h *domain.ChatHeader, apply func(*domain.ChatSession) (bool, error)) (domain.ChatSession, error) {
	user, ok := globalUser(project)
	if ok {
		if n.inference == nil || !domain.ValidID(id) {
			return domain.ChatSession{}, domain.ErrUnavailable
		}
		lock := &n.inference.transcripts[id[len(id)-1]%64]
		lock.Lock()
		defer lock.Unlock()
		return n.withGlobal(ctx, user, id, h, apply)
	}
	n.workspaceMu.Lock()
	defer n.workspaceMu.Unlock()
	location, err := n.chatLocation(ctx, project)
	if err != nil {
		return domain.ChatSession{}, err
	}
	return projectgit.WithChat(location, id, h, apply)
}

// Chats returns up to 32 repository-owned sessions, newest first.
func (n *Runtime) Chats(ctx context.Context, project, before string) ([]domain.ChatHeader, error) {
	user, ok := globalUser(project)
	if ok {
		return n.database.GlobalHeaders(ctx, user, before)
	}
	location, err := n.chatLocation(ctx, project)
	if err != nil {
		return nil, err
	}
	return projectgit.ChatHeaders(location, before)
}

func (n *Runtime) chatLocation(ctx context.Context, project string) (string, error) {
	if !domain.ValidID(project) {
		return "", errors.New("expected an owner-local project registration")
	}

	n.mu.Lock()
	if n.state != "ready" {
		n.mu.Unlock()
		return "", domain.ErrUnavailable
	}
	location, err := n.database.ProjectLocation(ctx, project)
	n.mu.Unlock()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(location, "content://") {
		return "", errors.New("session transcripts require a filesystem checkout")
	}
	err = ctx.Err()
	if err != nil {
		return "", err
	}
	return location, nil
}
