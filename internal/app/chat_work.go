// Reconciles durable owner turns with the authenticated inference replicas.
// The owner releases the project lock before contacting workers or waiting for a reply.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"golang.org/x/sync/errgroup"
	"time"
)

type queuedProvider struct{}

func (queuedProvider) Dispatch(context.Context, domain.ChatRequest) (*domain.ChatOutput, error) {
	return nil, nil
}

func (n *Runtime) queueChatTurn(ctx context.Context, project, user string, s *domain.ChatSession, id, input string) (bool, error) {
	changed, err := n.applyChatInput(ctx, project, s, id, input)
	if err != nil || s.Pending() != id || user == "" {
		return changed, err
	}
	pool := "Project Pool"
	if s.Scope == domain.GlobalChat {
		pool = "Chat Pool"
	}
	w := persistence.ChatWork{ChatAddress: domain.ChatAddress{Project: project, Session: s.ID, Turn: id}, User: user, Pool: pool}
	return changed, n.database.AddChatWork(ctx, w)
}

func (n *Runtime) startChatWork(ctx context.Context) {
	if n.inference != nil || n.Personal != nil {
		n.discoveryWorkers.Go(func() {
			n.runChatWork(ctx)
		})
	}
}
func (n *Runtime) runChatWork(ctx context.Context) {
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for ctx.Err() == nil {
		if n.Personal != nil {
			user := n.PlanningUser()
			entries, e := n.database.PlanningEntries(ctx, user)
			if user != "" && e == nil && len(entries) > 0 {
				r, c := n.planningClients(ctx, user)
				n.syncPlanning(ctx, user, entries, r, c)
			}
		}
		if n.inference != nil {
			n.mirrorGlobalMarkdown(ctx)
			n.inference.publishGlobals(ctx)
		}
		work, err := n.database.ChatWork(ctx)
		if err == nil {
			var group errgroup.Group
			group.SetLimit(8)
			for _, w := range work {
				if ctx.Err() != nil {
					break
				}
				group.Go(func() error {
					n.reconcileChatWork(ctx, w)
					return nil
				})
			}
			_ = group.Wait()
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func (n *Runtime) reconcileChatWork(ctx context.Context, w persistence.ChatWork) {
	s, err := n.Chat(ctx, w.Project, w.Session)
	if err != nil || len(s.Turns) == 0 {
		return
	}
	if s.Pending() != w.Turn {
		for _, t := range s.Turns {
			if t.ID == w.Turn {
				_ = n.database.CompleteChatWork(ctx, w)
				break
			}
		}
		return
	}
	s, err = n.DriveChat(ctx, w.ChatAddress, "cluster", queuedProvider{})
	if err != nil || s.Pending() != w.Turn {
		return
	}
	steps := s.Turns[len(s.Turns)-1].Steps
	if len(steps) == 0 {
		return
	}
	job := steps[len(steps)-1].Job
	self := identity.Fingerprint(n.Keys.Public())
	var roster domain.InferenceRoster
	var clients map[string]*httpapi.WorkerClient
	if n.inference != nil {
		n.inference.mu.RLock()
		roster = n.inference.roster
		n.inference.mu.RUnlock()
	} else {
		if n.Personal == nil || n.Personal.Fingerprint() != w.User {
			return
		}
		roster, clients = n.planningClients(ctx, w.User)
	}
	if len(roster.Peers) == 0 {
		return
	}
	packet, err := n.database.BoundPacket(ctx, job, roster.Revision())
	if err != nil {
		return
	}
	j := domain.InferenceJob{Revision: roster.Revision(), Owner: self, User: w.User, Pool: w.Pool, Packet: packet}
	admitted, err := n.database.PacketAdmitted(ctx, job)
	if err != nil {
		return
	}
	var out *domain.ChatOutput
	if n.inference != nil {
		r, e := n.database.InferenceRecord(ctx, job)
		if e == nil {
			out = r.Output
		}
		if out == nil && !admitted {
			answers := n.inference.fanout(ctx, roster, domain.InferenceCommand{Operation: "replicate", Job: j})
			if len(answers) == len(roster.Peers) {
				_ = n.database.AcknowledgePacket(ctx, job)
			}
		}
	} else {
		accepted := 0
		for _, peer := range roster.Peers {
			c := clients[peer.ID]
			if c == nil {
				continue
			}
			v, e := c.Inference(ctx, domain.InferenceCommand{Operation: "result", ID: job})
			if e != nil && !admitted {
				_, e = c.Inference(ctx, domain.InferenceCommand{Operation: "submit", Job: j})
			} else if v.Output != nil && v.State == "finished" {
				out = v.Output
				break
			}
			if e == nil {
				accepted++
				if admitted {
					break
				}
			}
		}
		if !admitted && accepted == len(roster.Peers) {
			_ = n.database.AcknowledgePacket(ctx, job)
		}
	}
	if out != nil {
		_, _ = n.DeliverChatOutput(ctx, w.ChatAddress, job, *out)
	}
}
