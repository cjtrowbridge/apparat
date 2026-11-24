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
	h := domain.ChatHeader{Version: 1, ID: id, Started: started, Scope: scope, Description: description}
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
		return s.Apply(turn, text, false)
	})
}

func (n *Runtime) CompleteChatTurn(ctx context.Context, project, id, turn, reply string) (domain.ChatSession, error) {
	return n.withChat(ctx, project, id, nil, func(s *domain.ChatSession) (bool, error) {
		return s.Apply(turn, reply, true)
	})
}

func (n *Runtime) withChat(ctx context.Context, project, id string, h *domain.ChatHeader, apply func(*domain.ChatSession) (bool, error)) (domain.ChatSession, error) {
	if !domain.ValidID(project) {
		return domain.ChatSession{}, errors.New("expected an owner-local project registration")
	}
	n.workspaceMu.Lock()
	defer n.workspaceMu.Unlock()
	n.mu.Lock()
	if n.state != "ready" {
		n.mu.Unlock()
		return domain.ChatSession{}, domain.ErrUnavailable
	}
	location, err := n.database.ProjectLocation(ctx, project)
	n.mu.Unlock()
	if err != nil {
		return domain.ChatSession{}, err
	}
	if strings.HasPrefix(location, "content://") {
		return domain.ChatSession{}, errors.New("session transcripts require a filesystem checkout")
	}
	err = ctx.Err()
	if err != nil {
		return domain.ChatSession{}, err
	}
	return projectgit.WithChat(location, id, h, apply)
}
