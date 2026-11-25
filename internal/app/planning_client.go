// Reconciles each GUI's private planning outbox with the pinned session writer.
// Session reads can use any available replica. Retries reuse the original turn ID.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"time"
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
	_, workers, lookup := n.modelCluster(ctx)
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
		v, e := c.Inference(ctx, domain.InferenceCommand{Operation: "status", Roster: membershipFor(w.Cluster, workers)})
		if e != nil {
			continue
		}
		if v.Protocol != 2 {
			reason = "Worker update pending. Apparat will retry automatically."
			continue
		}
		clients[w.Fingerprint] = c
		if len(v.Roster.Peers) == 0 {
			reason = "Waiting for authenticated worker membership to synchronize."
			continue
		}
		if len(roster.Peers) > 0 && v.Roster.Revision() != roster.Revision() {
			continue
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
		home, _ := n.database.SessionHome(ctx, user, id, "")
		if home == "" {
			home = roster.GlobalWriter(user, id)
		}
		c := clients[home]
		if blocked[id] {
			continue
		}
		if e.Operation == "global-post" {
			g := e.Request
			home := roster.GlobalWriter(user, id)
			artifact, readErr := n.database.GlobalArtifact(ctx, user, id)
			if readErr == nil && artifact.Session.Home == "" {
				for _, worker := range clients {
					probe, probeErr := worker.Inference(ctx, domain.InferenceCommand{Operation: "session-home", Global: domain.GlobalRequest{Session: id}})
					if probeErr == nil && probe.State != "" {
						home = probe.State
						break
					}
				}
			}
			if readErr == nil && artifact.Session.Home != "" {
				home = artifact.Session.Home
			}
			if home == "" || g.Parent != nil && *g.Parent == "" && clients[home] == nil {
				for _, worker := range clients {
					home = worker.Worker
					break
				}
			}
			var err error
			g.Home, err = n.database.SessionHome(ctx, user, id, home)
			if err != nil {
				blocked[id] = true
				continue
			}
			accepted := e.Received
			for _, worker := range clients {
				if accepted {
					break
				}
				v, sendErr := worker.Inference(ctx, domain.InferenceCommand{Operation: "global-intake", Global: g})
				if sendErr == nil && v.State == "received" {
					accepted = true
					break
				}
			}
			// Retain the local projection until the replicated transcript includes this exact turn.
			if accepted {
				err = n.database.MarkPlanningReceived(ctx, user, e.Seq)
				for _, worker := range clients {
					v, readErr := worker.Inference(ctx, domain.InferenceCommand{Operation: "global-get", Global: g})
					if readErr != nil || v.Artifact == nil {
						continue
					}
					for _, turn := range v.Artifact.Session.Turns {
						if turn.ID == g.Turn && turn.Input == g.Input {
							err = n.database.PutGlobalArtifact(ctx, *v.Artifact, false)
							if err == nil {
								err = n.database.AcknowledgePlanning(ctx, e.Seq, user)
							}
							break
						}
					}
				}
			}
			blocked[id] = !accepted || err != nil
			continue
		}
		if c == nil {
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
			known = n.refreshGlobalRevision(ctx, c, user, id, known)
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
			v.Acknowledged = entry.Received
			v.SyncReason = reason
			if entry.Received {
				v.SyncReason = "Queued · Chat Pool"
			}
			if !entry.Received && reason == "" && clients[roster.GlobalWriter(user, id)] == nil {
				v.SyncReason = "The worker assigned to this session is unreachable. Apparat will retry automatically."
			}
		}
	}
	if v.Session != nil && v.Session.Pending() != "" {
		t := v.Session.Turns[len(v.Session.Turns)-1]
		if !v.Acknowledged {
			start, _ := domain.ChatTime(t.ID)
			v.SyncReason = reason + " Saved " + start.Local().Format("15:04:05") + "."
		} else if len(t.Steps) > 0 {
			for _, c := range clients {
				state, e := c.Inference(ctx, domain.InferenceCommand{Operation: "job-state", ID: t.Steps[len(t.Steps)-1].Job})
				if e != nil {
					continue
				}
				if state.State == "dispatching" {
					v.SyncReason = "Worker processing request"
					break
				}
				if state.State == "queued" {
					v.SyncReason = "Queued · Chat Pool"
				}
			}
		}
		if v.SyncReason == "Queued · Chat Pool" && len(roster.Peers) > 0 {
			checked, available := 0, false
			for _, peer := range roster.Peers {
				cached, ok := n.peerInventory.Load(user + "/" + peer.ID)
				if !ok {
					continue
				}
				inv := cached.(domain.WorkerInventory)
				if inv.Retrying || !inv.Fresh(time.Now()) {
					continue
				}
				checked++
				for _, m := range inv.Models {
					if m.Text && m.Name == t.Model.Name && m.Digest == t.Model.Digest {
						available = true
					}
				}
			}
			if checked == len(roster.Peers) && !available {
				v.SyncReason = "Queued · waiting for " + t.Model.Name
			}
		}
	}
	for _, h := range headers {
		v.Headers = append(v.Headers, h)
	}
	return v, err
}

// Fetch a transcript only when a worker has a newer revision for this user.
func (n *Runtime) refreshGlobalRevision(ctx context.Context, c *httpapi.WorkerClient, user, id string, known int64) int64 {
	q := domain.InferenceCommand{Operation: "global-version", Global: domain.GlobalRequest{Session: id}}
	head, e := c.Inference(ctx, q)
	if e != nil || head.Chunk == nil || head.Chunk.Revision <= known {
		return known
	}
	q.Operation = "global-get"
	reply, e := c.Inference(ctx, q)
	if e != nil || reply.Artifact == nil || reply.Artifact.User != user {
		return known
	}
	e = n.database.PutGlobalArtifact(ctx, *reply.Artifact, false)
	if e != nil {
		return known
	}
	return reply.Artifact.Version
}

func (n *Runtime) syncPersonalPlanning(ctx context.Context) {
	user := n.PlanningUser()
	if user == "" {
		return
	}
	entries, e := n.database.PlanningEntries(ctx, user)
	if e != nil {
		return
	}
	pending, e := n.database.PendingPlanningRevisions(ctx, user)
	if e != nil || len(entries) == 0 && len(pending) == 0 {
		return
	}
	membership, clients, _ := n.planningClients(ctx, user)
	if len(entries) > 0 {
		n.syncPlanning(ctx, user, entries, membership, clients)
	}
	for id, known := range pending {
		for _, c := range clients {
			known = n.refreshGlobalRevision(ctx, c, user, id, known)
		}
	}
}
