// Claims inference steps and commits their correlated results to the owner transcript.
// Dispatch happens after durable publication and outside all runtime or workspace locks.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

var chatRouteID = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,96}$`)

const maxChatSteps = 16

// DriveChat returns on asynchronous acceptance, completion, or a provider error.
// Existing outstanding jobs are never dispatched a second time.
func (n *Runtime) DriveChat(ctx context.Context, a domain.ChatAddress, route string, provider domain.ChatProvider) (domain.ChatSession, error) {
	if provider == nil || !chatRouteID.MatchString(route) {
		return domain.ChatSession{}, errors.New("a provider and stable route identifier are required")
	}
	for {
		var request domain.ChatRequest
		s, err := n.withChat(ctx, a.Project, a.Session, nil, func(s *domain.ChatSession) (bool, error) {
			if len(s.Turns) == 0 || s.Turns[len(s.Turns)-1].ID != a.Turn {
				return false, errors.New("select the current session turn")
			}
			if s.Pending() == "" {
				return false, nil
			}
			t := &s.Turns[len(s.Turns)-1]
			if len(t.Steps) > 0 {
				last := t.Steps[len(t.Steps)-1]
				if last.Route != route {
					return false, errors.New("turn is bound to another route")
				}
				if last.Output == nil {
					return false, nil
				}
			}
			if len(t.Steps) >= maxChatSteps {
				t.Stopped = "Tool step budget exhausted"
				return true, nil
			}
			job, err := domain.NewID(time.Now(), nil)
			if err != nil {
				return false, err
			}
			request, err = chatPacket(a, route, job, *s)
			if err != nil {
				return false, err
			}
			data, err := json.Marshal(request)
			if err != nil {
				return false, err
			}
			t.Steps = append(t.Steps, domain.ChatStep{Job: job, Route: route, Digest: projectgit.Digest(string(data))})
			return true, nil
		})
		if err != nil || request.Job == "" {
			return s, err
		}
		output, err := provider.Dispatch(ctx, request)
		if err != nil || output == nil {
			return s, err
		}
		s, err = n.DeliverChatOutput(ctx, a, request.Job, *output)
		if err != nil || s.Pending() == "" {
			return s, err
		}
	}
}

// DeliverChatOutput is the entry point for delayed worker callbacks and reconciliation.
func (n *Runtime) DeliverChatOutput(ctx context.Context, a domain.ChatAddress, job string, out domain.ChatOutput) (domain.ChatSession, error) {
	if !utf8.ValidString(out.Text) || strings.ContainsRune(out.Text, 0) {
		return domain.ChatSession{}, errors.New("invalid assistant text")
	}
	if len(out.Calls) == 0 {
		out.Calls = nil
	}
	return n.withChat(ctx, a.Project, a.Session, nil, func(s *domain.ChatSession) (bool, error) {
		for i := range s.Turns {
			t := &s.Turns[i]
			if t.ID != a.Turn {
				continue
			}
			for j := range t.Steps {
				step := &t.Steps[j]
				if step.Job != job {
					continue
				}
				if step.Output != nil {
					if !reflect.DeepEqual(*step.Output, out) {
						return false, errors.New("conflicting inference result")
					}
					return false, nil
				}
				if s.Pending() != a.Turn || j != len(t.Steps)-1 {
					return false, errors.New("inference result is stale")
				}
				switch out.Finish {
				case "stop":
					if len(out.Calls) > 0 || strings.TrimSpace(out.Text) == "" {
						return false, errors.New("invalid final assistant message")
					}
					t.Reply = out.Text
				case "tool_calls":
					results, err := n.chatToolBatch(ctx, a.Project, s.Scope, out.Calls)
					if err != nil {
						return false, err
					}
					step.Results = results
				case "length", "error", "cancelled":
					t.Stopped = "Provider finished: " + out.Finish
				default:
					return false, errors.New("unsupported provider finish reason")
				}
				step.Output = &out
				return true, nil
			}
		}
		return false, errors.New("unknown inference job")
	})
}

// StopChatTurn fences late replies. Provider-side cancellation is adapter-specific.
func (n *Runtime) StopChatTurn(ctx context.Context, a domain.ChatAddress) (domain.ChatSession, error) {
	return n.withChat(ctx, a.Project, a.Session, nil, func(s *domain.ChatSession) (bool, error) {
		for i := range s.Turns {
			t := &s.Turns[i]
			if t.ID == a.Turn && t.Stopped == "Cancelled by user" {
				return false, nil
			}
		}
		if s.Pending() != a.Turn || a.Turn == "" {
			return false, errors.New("select a pending turn")
		}
		s.Turns[len(s.Turns)-1].Stopped = "Cancelled by user"
		return true, nil
	})
}

// OutstandingChatRequest reconstructs the exact claimed packet for a queue adapter
// reconciling by job ID. This read does not authorize a fresh provider attempt.
func (n *Runtime) OutstandingChatRequest(ctx context.Context, a domain.ChatAddress) (domain.ChatRequest, error) {
	s, err := n.Chat(ctx, a.Project, a.Session)
	if err != nil {
		return domain.ChatRequest{}, err
	}
	if a.Turn == "" || s.Pending() != a.Turn {
		return domain.ChatRequest{}, errors.New("select a pending turn")
	}
	t := s.Turns[len(s.Turns)-1]
	if len(t.Steps) == 0 || t.Steps[len(t.Steps)-1].Output != nil {
		return domain.ChatRequest{}, errors.New("turn has no outstanding provider job")
	}
	step := t.Steps[len(t.Steps)-1]
	request, err := chatPacket(a, step.Route, step.Job, s)
	if err != nil {
		return domain.ChatRequest{}, err
	}
	data, err := json.Marshal(request)
	if err != nil {
		return domain.ChatRequest{}, err
	}
	if projectgit.Digest(string(data)) != step.Digest {
		return domain.ChatRequest{}, errors.New("request changed since dispatch; reconcile using original job ID")
	}
	return request, nil
}
