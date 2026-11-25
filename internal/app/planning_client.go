// Reconciles each GUI's private planning outbox with the pinned session writer.
// Session reads can use any available replica. Retries reuse the original turn ID.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) PlanningUser() string {
	if n.Personal == nil {
		return ""
	}
	return n.Personal.Fingerprint()
}
func (n *Runtime) SavePlanning(ctx context.Context, user, op string, q domain.GlobalRequest) error {
	if user == "" || user != n.PlanningUser() {
		return errors.New("unlock your personal identity")
	}
	return n.database.SavePlanning(ctx, user, op, q)
}

func (n *Runtime) planningClients(ctx context.Context, user string) (domain.InferenceRoster, map[string]*httpapi.WorkerClient, string) {
	workers, lookup := n.KnownWorkers(ctx)
	clients := map[string]*httpapi.WorkerClient{}
	var roster domain.InferenceRoster
	reason := "No saved worker is available. Connect a cluster worker to sync this conversation."
	if lookup != nil || len(workers) > 0 {
		reason = "No worker returned inference status for this identity. Check the worker connection in Cluster Status."
	}
	for _, w := range workers {
		c, e := n.workspaceClient(ctx, "worker-"+w.Fingerprint)
		if e != nil || c.Access.User != user {
			continue
		}
		v, e := c.Inference(ctx, domain.InferenceCommand{Operation: "status"})
		if e != nil {
			continue
		}
		if len(v.Roster.Peers) == 0 {
			reason = "No enrolled inference validators reported. An administrator can enroll them in Cluster Status → Inference validators."
			continue
		}
		if len(roster.Peers) > 0 && v.Roster.Revision() != roster.Revision() {
			return domain.InferenceRoster{}, nil, "Workers reported conflicting validator rosters. Check their enrollment in Cluster Status."
		}
		roster = v.Roster
		clients[w.Fingerprint] = c
	}
	for _, peer := range roster.Peers {
		if clients[peer.ID] != nil {
			continue
		}
		c, e := n.cachedWorker(ctx, peer.ID+"/"+peer.Address, func() (*httpapi.WorkerClient, error) {
			return httpapi.OpenWorker(ctx, peer.Address, peer.ID, n.Personal, n.Keys, &n.observations)
		})
		if e == nil && c.Access.User == user && c.Access.Cluster == roster.Cluster && c.Access.Admission == "approved" {
			clients[peer.ID] = c
		}
	}
	if len(roster.Peers) > 0 {
		reason = ""
	}
	return roster, clients, reason
}
func (n *Runtime) syncPlanning(ctx context.Context, user string, entries []persistence.PlanningEntry, roster domain.InferenceRoster, clients map[string]*httpapi.WorkerClient) {
	blocked := map[string]bool{}
	for _, e := range entries {
		id := e.Request.Session
		c := clients[roster.GlobalWriter(user, id)]
		if c == nil || blocked[id] {
			continue
		}
		v, err := c.Inference(ctx, domain.InferenceCommand{Operation: e.Operation, Global: e.Request})
		if err == nil && v.Artifact != nil && v.Artifact.User == user {
			err = n.database.PutGlobalArtifact(ctx, *v.Artifact, false)
			if err == nil {
				err = n.database.AcknowledgePlanning(ctx, e.Seq, user)
			}
		}
		blocked[id] = err != nil
	}
}

func (n *Runtime) PlanningView(ctx context.Context, user, id string) (domain.PlanningView, error) {
	var v domain.PlanningView
	if n.PlanningUser() != user || user == "" {
		return v, errors.New("unlock your personal identity")
	}
	roster, clients, reason := n.planningClients(ctx, user)

	headers := map[string]domain.ChatHeader{}
	known, _ := n.database.GlobalVersion(ctx, user, id)
	for _, c := range clients {
		cursor := ""
		for id == "" && ctx.Err() == nil {
			reply, e := c.Inference(ctx, domain.InferenceCommand{Operation: "global-list", Global: domain.GlobalRequest{Cursor: cursor}})
			if e != nil {
				break
			}
			for _, h := range reply.Headers {
				headers[h.ID] = h
			}
			if len(reply.Headers) < 32 {
				break
			}
			cursor = reply.Headers[len(reply.Headers)-1].ID
		}
		if id != "" {
			q := domain.InferenceCommand{Operation: "global-version", Global: domain.GlobalRequest{Session: id}}
			head, e := c.Inference(ctx, q)
			if e != nil || head.Chunk == nil || head.Chunk.Revision <= known {
				continue
			}
			q.Operation = "global-get"
			reply, e := c.Inference(ctx, q)
			if e == nil && reply.Artifact != nil && reply.Artifact.User == user {
				e = n.database.PutGlobalArtifact(ctx, *reply.Artifact, false)
				if e == nil {
					known = reply.Artifact.Version
				}
			}
		}
	}
	a, e := n.database.GlobalArtifact(ctx, user, id)
	if e == nil {
		v.Session = &a.Session
		v.Acknowledged = true
		headers[id] = a.Session.ChatHeader
	}
	cursor := ""
	for id == "" {
		page, e := n.database.GlobalHeaders(ctx, user, cursor)
		if e != nil {
			return v, e
		}
		for _, h := range page {
			_, ok := headers[h.ID]
			if !ok {
				headers[h.ID] = h
			}
		}
		if len(page) < 32 {
			break
		}
		cursor = page[len(page)-1].ID
	}
	entries, err := n.database.PlanningEntries(ctx, user)
	for _, entry := range entries {
		g := entry.Request
		at, _ := domain.ChatTime(g.Session)
		h, exists := headers[g.Session]
		if !exists {
			h = domain.ChatHeader{Version: 1, ID: g.Session, Scope: domain.GlobalChat, Started: at, Model: g.Model, Description: g.Title}
		}
		if entry.Operation == "global-rename" {
			h.Description = g.Title
		}
		headers[g.Session] = h
		if g.Session == id {
			if v.Session == nil {
				v.Session = &domain.ChatSession{ChatHeader: h}
			}
			if entry.Operation == "global-post" {
				_, _ = v.Session.Apply(g.Turn, g.Input, false)
			}
			v.Session.Description = h.Description
			v.Acknowledged = false
			v.SyncReason = reason
			if reason == "" && clients[roster.GlobalWriter(user, id)] == nil {
				v.SyncReason = "The worker assigned to this session is unreachable. Apparat will retry automatically."
			}
		}
	}
	for _, h := range headers {
		v.Headers = append(v.Headers, h)
	}
	return v, err
}
