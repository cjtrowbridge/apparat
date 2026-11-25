// Pulls compatible jobs in pool order whenever this worker's model slot is free.
// Full-roster grants and applied ownership precede the once-only provider start.
package app

import (
	"context"
	"fmt"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/chatprovider"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (p *inferencePool) run(ctx context.Context) {
	for ctx.Err() == nil {
		progress := p.dispatchNext(ctx)
		if progress {
			continue
		}
		timer := time.NewTimer(p.timing.Snapshot().Recheck)
		select {
		case <-ctx.Done():
		case <-p.wake:
		case <-timer.C:
		}
		timer.Stop()
	}
}
func (p *inferencePool) dispatchNext(ctx context.Context) bool {
	p.mu.RLock()
	table := p.table
	p.mu.RUnlock()
	roster := p.votingRoster()
	if table == nil || len(roster.Peers) == 0 {
		return false
	}
	inventory := p.inventory.Snapshot()
	blocked := map[string]bool{}
	const batch = 16
	for offset := 0; ctx.Err() == nil; offset += batch {
		rows, err := p.db.RunnableInference(ctx, p.self, inventory.HasText(time.Now()), offset, batch)
		if err != nil {
			return false
		}
		for _, r := range rows {
			if ctx.Err() != nil {
				return false
			}
			if p.tryInference(ctx, roster, inventory, r, blocked) {
				return true
			}
		}
		if len(rows) < batch {
			p.observer.prune(blocked)
			break
		}
	}
	return false
}
func (p *inferencePool) tryInference(ctx context.Context, roster domain.InferenceRoster, inventory domain.WorkerInventory, r domain.InferenceRecord, blocked map[string]bool) bool {
	j := r.Job
	if r.State == "finished" {
		return p.publishResult(ctx, roster, r.Claim, *r.Output)
	}
	if !inventory.HasText(time.Now()) {
		return false
	}
	eligible := false
	for _, model := range inventory.Models {
		if model.Text && j.Packet.Model.Provider == "ollama" && model.Name == j.Packet.Model.Name && model.Digest == j.Packet.Model.Digest {
			eligible = true
			break
		}
	}
	if !eligible {
		return false
	}
	allowed, e := p.db.InferenceAccess(ctx, roster.Cluster, j.User, j.Pool, p.self, r.State == "owned")
	if e != nil || !allowed {
		return false
	}
	provider, e := chatprovider.NewCompletions(fmt.Sprintf("http://127.0.0.1:%d/api/chat", inventory.Port), j.Packet.Model.Name, "")
	if e != nil {
		return false
	}
	if p.execute(ctx, roster, r, provider) {
		p.observer.stalled(j.User, false)
		return true
	}
	blocked[j.User] = true
	p.observer.stalled(j.User, true)
	return false
}
func allReplies(roster domain.InferenceRoster, replies map[string]domain.InferenceReply, c domain.InferenceClaim, states ...string) bool {
	if len(replies) != len(roster.Peers) {
		return false
	}
	for _, peer := range roster.Peers {
		a := replies[peer.ID].Answer
		ok := false
		for _, state := range states {
			ok = ok || a.State == state
		}
		if a.Holder != c || !ok {
			return false
		}
	}
	return true
}
func (p *inferencePool) execute(ctx context.Context, roster domain.InferenceRoster, r domain.InferenceRecord, provider domain.ChatProvider) bool {
	c := r.Claim
	if r.State == "queued" {
		p.mu.Lock()
		now := time.Now().Truncate(time.Millisecond)
		if !now.After(p.lastAttempt) {
			now = p.lastAttempt.Add(time.Millisecond)
		}
		p.lastAttempt = now
		p.mu.Unlock()
		attempt, err := domain.NewID(now, nil)
		if err != nil {
			return false
		}
		p.mu.RLock()
		revision := p.roster.Revision()
		p.mu.RUnlock()
		c = domain.InferenceClaim{Job: r.Job.Packet.Job, Revision: revision, Digest: domain.InferenceDigest(r.Job), Executor: p.self, Attempt: attempt}
	}
	p.mu.Lock()
	previous, retained := p.active[c.Job]
	if retained {
		c = previous
	}
	p.active[c.Job] = c
	p.mu.Unlock()
	keep := false
	defer func() {
		if keep {
			return
		}
		p.mu.Lock()
		delete(p.active, c.Job)
		p.mu.Unlock()
	}()
	q := domain.InferenceCommand{Operation: "confirm", Claim: c}
	local, err := p.InferencePeer(ctx, p.self, q)
	if err != nil {
		return false
	}
	if local.Answer.Holder != c {
		p.recheckReservation(ctx, roster, local.Answer)
		return false
	}
	replies := p.fanout(ctx, roster, q)
	if !allReplies(roster, replies, c, "reserved", "owned") {
		for _, peer := range roster.Peers {
			reply, answered := replies[peer.ID]
			if !answered || reply.Answer.State != "missing" || peer.ID == p.self {
				continue
			}
			client, e := p.peerClient(peer)
			if e == nil {
				_, _ = client.Call(ctx, domain.InferenceCommand{Operation: "replicate-queued", Job: r.Job})
			}
		}
		var winner domain.InferenceClaim
		if r.State == "queued" && len(replies) == len(roster.Peers) {
			best := c
			contended := false
			for _, v := range replies {
				h := v.Answer
				if h.Holder == c {
					continue
				}
				if h.State != "reserved" {
					contended = false
					break
				}
				contended = true
				if domain.CompareClaim(h.Holder, best) < 0 {
					best = h.Holder
				}
			}
			if contended {
				keep = best == c
				winner = best
			}
		}
		if keep {
			return false
		}
		if r.State == "queued" {
			if winner.Executor != "" {
				// Yield our local vote to the observed contender before another local scan.
				p.table.Transition(p.self, "withdraw", c)
				p.table.Transition(winner.Executor, "confirm", winner)
			}
			q.Operation = "withdraw"
			p.fanout(ctx, roster, q)
		}
		return false
	}
	q.Operation = "own"
	_, err = p.InferencePeer(ctx, p.self, q)
	if err != nil {
		return false
	}
	replies = p.fanout(ctx, roster, q)
	if !allReplies(roster, replies, c, "owned") {
		return false
	}
	err = p.db.StartInference(ctx, c)
	if err != nil {
		return false
	}
	p.observer.slot(true)
	defer p.observer.slot(false)
	output, err := p.dispatchPreview(ctx, r, provider)
	if err != nil || output == nil {
		output = &domain.ChatOutput{Finish: "indeterminate"}
	}
	err = p.db.FinishInference(ctx, c, *output)
	if err != nil {
		return false
	}
	return p.publishResult(ctx, roster, c, *output)
}
func (p *inferencePool) publishResult(ctx context.Context, roster domain.InferenceRoster, c domain.InferenceClaim, out domain.ChatOutput) bool {
	q := domain.InferenceCommand{Operation: "finish", Claim: c, Output: &out}
	replies := p.fanout(ctx, roster, q)
	if !allReplies(roster, replies, c, "finished") {
		return false
	}
	return p.db.InferenceDelivered(ctx, c) == nil
}
func (p *inferencePool) recheckReservation(ctx context.Context, r domain.InferenceRoster, a domain.ClaimAnswer) {
	if a.State != "reserved" {
		return
	}
	for _, peer := range r.Peers {
		if peer.ID != a.Holder.Executor {
			continue
		}
		c, err := p.peerClient(peer)
		if err != nil {
			return
		}
		v, err := c.Call(ctx, domain.InferenceCommand{Operation: "recheck", Claim: a.Holder})
		if err == nil && v.State == "withdrawn" {
			p.table.Transition(a.Holder.Executor, "withdraw", a.Holder)
		}
	}
}
