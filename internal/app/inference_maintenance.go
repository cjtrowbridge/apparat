// Coordinates a planned worker update using authenticated, durable acknowledgements.
// Queue admission stays open. A partition cannot initiate a withdrawal.
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"slices"
	"time"
)

func (p *inferencePool) votingRoster() domain.InferenceRoster {
	p.mu.RLock()
	defer p.mu.RUnlock()
	r := p.roster
	w := p.withdrawal
	if w.Phase == "" {
		return r
	}
	if w.Phase != "away" || w.Worker == p.self {
		r.Peers = nil
		return r
	}
	r.Peers = slices.DeleteFunc(slices.Clone(r.Peers), func(peer domain.InferencePeer) bool {
		return peer.ID == w.Worker
	})
	return r
}

func (p *inferencePool) maintenance(ctx context.Context, caller string, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w := p.withdrawal
	if q.Roster.Revision() != p.roster.Revision() || !domain.ValidID(q.ID) {
		return v, errors.New("stale maintenance request")
	}
	if w.Worker != "" && w.Worker != caller && q.Operation == "maintenance-release" {
		return v, p.db.SetWithdrawal(ctx, persistence.WorkerWithdrawal{Worker: caller, Generation: q.ID, Phase: "released"})
	}
	if w.Worker != "" && (w.Worker != caller || w.Generation != q.ID) {
		return v, errors.New("another update is active")
	}
	if q.Operation == "maintenance-ledger" {
		if w.Phase != "closing" {
			return v, errors.New("worker is still accepting claims")
		}
		r, e := p.db.NextInferenceRecord(ctx, q.Global.Cursor)
		if errors.Is(e, sql.ErrNoRows) {
			return v, nil
		}
		if e != nil {
			return v, e
		}
		raw, e := json.Marshal(r)
		if e != nil {
			return v, e
		}
		offset := q.Chunk.Offset
		if offset < 0 || offset >= len(raw) {
			return v, errors.New("invalid record offset")
		}
		v.State = r.Job.Packet.Job
		if q.Global.Turn != "" && q.Global.Turn != v.State {
			return v, errors.New("record page changed")
		}
		v.Chunk = &domain.GlobalChunk{Offset: offset, Total: len(raw), Bytes: raw[offset:min(offset+domain.GlobalChunkBytes, len(raw))]}
		return v, nil
	}
	phase := map[string]string{"maintenance-prepare": "prepared", "maintenance-away": "away", "maintenance-close": "closing", "maintenance-release": "released"}[q.Operation]
	if phase == "" {
		return v, errors.New("unknown maintenance phase")
	}
	if (phase == "prepared" || phase == "closing") && phase != w.Phase {
		if len(p.active) > 0 || p.table.Reserved() {
			return v, errors.New("claim in progress")
		}
		busy, e := p.db.OpenOwnership(ctx)
		if busy || e != nil {
			return v, errors.New("waiting for inference ownership to settle")
		}
	}
	if phase == "away" && w.Phase != "prepared" && w.Phase != "away" {
		return v, errors.New("withdrawal is not prepared")
	}
	w = persistence.WorkerWithdrawal{Worker: caller, Generation: q.ID, Phase: phase}
	err = p.db.SetWithdrawal(ctx, w)
	if err != nil {
		return v, err
	}
	if phase == "released" {
		w = persistence.WorkerWithdrawal{}
	}
	p.withdrawal = w
	p.signal()
	return v, nil
}

// PrepareInferenceUpdate is called immediately before installation after the updater drains.
func (n *Runtime) PrepareInferenceUpdate(ctx context.Context) bool {
	p := n.inference
	if p == nil {
		return true
	}
	p.maintenanceMu.Lock()
	defer p.maintenanceMu.Unlock()
	p.mu.RLock()
	r := p.roster
	p.mu.RUnlock()
	if len(r.Peers) < 2 {
		return true
	}
	id, err := domain.NewID(time.Now(), nil)
	if err != nil {
		return false
	}
	q := domain.InferenceCommand{ID: id, Roster: r, Operation: "maintenance-prepare"}
	_, err = p.maintenance(ctx, p.self, q)
	if err != nil {
		return false
	}
	for _, op := range []string{"maintenance-prepare", "maintenance-away"} {
		q.Operation = op
		if op == "maintenance-away" {
			_, err = p.maintenance(ctx, p.self, q)
			if err != nil {
				p.resumeMaintenance(ctx)
				return false
			}
		}
		replies := p.fanout(ctx, r, q)
		if len(replies) != len(r.Peers) {
			p.resumeMaintenance(ctx)
			return false
		}
	}
	p.installing = true
	return true
}

// An unsuccessful installation resumes the current binary. A successful one resumes on restart.
func (n *Runtime) CancelInferenceUpdate(ctx context.Context) {
	if n.inference == nil {
		return
	}
	p := n.inference
	p.maintenanceMu.Lock()
	defer p.maintenanceMu.Unlock()
	p.installing = false
	p.resumeMaintenance(ctx)
}

func (p *inferencePool) recoverMaintenance(ctx context.Context) {
	if !p.maintenanceMu.TryLock() {
		return
	}
	defer p.maintenanceMu.Unlock()
	if !p.installing {
		p.resumeMaintenance(ctx)
	}
}

func (p *inferencePool) resumeMaintenance(ctx context.Context) {
	p.mu.RLock()
	r, w := p.roster, p.withdrawal
	p.mu.RUnlock()
	if w.Worker != p.self {
		return
	}
	q := domain.InferenceCommand{Operation: "maintenance-close", ID: w.Generation, Roster: r}
	// Never close another worker's concurrent proposal. Release only a partially prepared
	// transition when every local vote was still fenced and no reduced round could run.
	if w.Phase == "prepared" || w.Phase == "releasing" {
		q.Operation = "maintenance-release"
		p.releaseWithdrawal(ctx, r, q)
		return
	}
	if len(p.fanout(ctx, r, q)) != len(r.Peers) {
		return
	}
	for _, peer := range r.Peers {
		if peer.ID == p.self {
			continue
		}
		client, err := p.peerClient(peer)
		if err != nil {
			return
		}
		q.Operation = "maintenance-ledger"
		q.Global.Cursor = ""
		for ctx.Err() == nil {
			record, e := maintenanceRecord(ctx, client, q)
			if e != nil {
				return
			}
			if record == nil {
				break
			}
			row := *record
			if row.Job.Packet.Job <= q.Global.Cursor {
				return
			}
			err = p.db.RestoreInferenceRecord(ctx, row)
			if err != nil {
				return
			}
			q.Global.Cursor = row.Job.Packet.Job
		}
	}
	err := p.rebuildClaimTable(ctx, r)
	if err != nil {
		return
	}
	q.Operation = "maintenance-release"
	p.mu.Lock()
	w.Phase = "releasing"
	err = p.db.SetWithdrawal(ctx, w)
	if err == nil {
		p.withdrawal = w
	}
	p.mu.Unlock()
	if err == nil {
		p.releaseWithdrawal(ctx, r, q)
	}
}

func (p *inferencePool) releaseWithdrawal(ctx context.Context, r domain.InferenceRoster, q domain.InferenceCommand) {
	// Keep our own recovery intent until all peers have persisted their acknowledgements.
	for _, peer := range r.Peers {
		if peer.ID == p.self {
			continue
		}
		c, e := p.peerClient(peer)
		if e != nil {
			return
		}
		_, e = c.Call(ctx, q)
		if e != nil {
			return
		}
	}
	_, _ = p.maintenance(ctx, p.self, q)
}

// Called under p.mu. The withdrawing worker cannot grant or propose an execution.
func (p *inferencePool) mayVote(caller string) bool {
	w := p.withdrawal
	return w.Phase == "" || w.Phase == "away" && w.Worker != p.self && w.Worker != caller
}

func maintenanceRecord(ctx context.Context, c *httpapi.InferenceClient, q domain.InferenceCommand) (*domain.InferenceRecord, error) {
	var raw []byte
	for {
		v, e := c.Call(ctx, q)
		if e != nil {
			return nil, e
		}
		if v.Chunk == nil && len(raw) == 0 {
			return nil, nil
		}
		chunk := v.Chunk
		if chunk == nil || chunk.Offset != len(raw) || len(chunk.Bytes) == 0 || len(chunk.Bytes) > domain.GlobalChunkBytes || chunk.Total < len(raw)+len(chunk.Bytes) {
			return nil, errors.New("incomplete worker ledger")
		}
		raw = append(raw, chunk.Bytes...)
		if len(raw) == chunk.Total {
			var r domain.InferenceRecord
			e = json.Unmarshal(raw, &r)
			if e != nil || r.Job.Packet.Job != v.State {
				return nil, errors.New("invalid worker ledger")
			}
			return &r, nil
		}
		q.Chunk.Offset = len(raw)
		q.Global.Turn = v.State
	}
}
