// Hosts one enrolled inference worker's authenticated replica and hot reservation table.
// Inputs and irreversible dispatch receipts are durable. Ordinary confirmations read only RAM.
package app

import (
	"context"
	"database/sql"
	"errors"

	"strings"
	"sync"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/inventory"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type inferencePool struct {
	maintenanceMu sync.Mutex
	installing    bool
	withdrawal    persistence.WorkerWithdrawal
	transcripts   [64]sync.Mutex
	lastAttempt   time.Time
	observer      inferenceObserver
	clientsMu     sync.Mutex
	clients       map[string]peerConnection
	runtime       *Runtime
	mu            sync.RWMutex
	db            *persistence.DB
	keys          *identity.Session
	inventory     *inventory.Monitor
	self          string
	roster        domain.InferenceRoster
	table         *domain.ClaimTable
	timing        domain.InferenceTiming
	wake          chan struct{}
	active        map[string]domain.InferenceClaim
}

func newInferencePool(ctx context.Context, db *persistence.DB, keys *identity.Session, inventory *inventory.Monitor) (*inferencePool, error) {
	p := &inferencePool{db: db, keys: keys, inventory: inventory, self: identity.Fingerprint(keys.Public()), wake: make(chan struct{}, 1), active: map[string]domain.InferenceClaim{}}
	r, err := db.InferenceRoster(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	err = db.RecoverInterrupted(ctx, p.self)
	if err != nil {
		return nil, err
	}
	p.withdrawal, err = db.Withdrawal(ctx)
	if err != nil {
		return nil, err
	}
	return p, p.install(ctx, r)
}
func (p *inferencePool) signal() {
	select {
	case p.wake <- struct{}{}:
	default:
	}
}
func (p *inferencePool) install(ctx context.Context, r domain.InferenceRoster) error {
	err := r.Validate(p.self)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	r, err = p.mergeMembership(ctx, r)
	if err != nil {
		return err
	}
	if p.table != nil && domain.InferenceDigest(p.roster) == domain.InferenceDigest(r) {
		return nil
	}
	err = p.db.InstallInferenceRoster(ctx, r)
	if err != nil {
		return err
	}
	if p.table != nil && p.roster.Revision() == r.Revision() {
		p.roster = r
		return nil
	}
	return p.rebuildLocked(ctx, r)
}

func (p *inferencePool) rebuildClaimTable(ctx context.Context, r domain.InferenceRoster) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.rebuildLocked(ctx, r)
}
func (p *inferencePool) rebuildLocked(ctx context.Context, r domain.InferenceRoster) error {
	table, err := domain.NewClaimTable(r.Revision(), r.IDs())
	if err != nil {
		return err
	}
	table.Recovered()
	err = p.db.InferenceRecords(ctx, func(row domain.InferenceRecord) error {

		e := table.Admit(row.Job.Packet.Job, domain.InferenceDigest(row.Job))
		if e != nil {
			return e
		}
		if row.State != "queued" {
			c := row.Claim
			if row.State == "finished" {
				c.Revision = r.Revision()
			}
			table.Transition(c.Executor, "confirm", c)
			table.Transition(c.Executor, "own", c)
			if row.State == "finished" {
				table.Transition(c.Executor, "finish", c)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	p.roster = r
	p.table = table
	p.signal()
	return nil
}
func (p *inferencePool) InferenceUser(ctx context.Context, a persistence.Access, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	switch q.Operation {
	case "session-home":
		_, err = p.db.GlobalVersion(ctx, a.User, q.Global.Session)
		if err == nil {
			p.mu.RLock()
			v.State = p.sessionWriter(ctx, a.User, q.Global.Session, p.roster)
			p.mu.RUnlock()
		}
	case "global-intake":
		return p.receiveTurn(ctx, a, q.Global)
	case "global-chunk":
		chunk, e := p.db.GlobalChunk(ctx, a.User, q.Global.Session, q.Chunk.Revision, q.Chunk.Offset)
		v.Chunk = &chunk
		return v, e
	case "global-version":
		rev, e := p.db.GlobalVersion(ctx, a.User, q.Global.Session)
		v.Chunk = &domain.GlobalChunk{Revision: rev}
		return v, e
	case "global-get", "global-list", "global-post", "global-rename":
		return p.globalOperation(ctx, a, q)
	case "status":
		if len(q.Roster.Peers) != 0 && a.Admin && q.Roster.Cluster == a.Cluster {
			err = p.install(ctx, q.Roster)
			if err != nil {
				v.State = "membership-pending"
			}
		}
		p.mu.RLock()
		v.Protocol = 2
		v.Roster = p.roster
		v.Status.Validators = len(p.roster.Peers)
		v.Status.Updating = p.withdrawal.Worker
		p.mu.RUnlock()
		p.observer.populate(a.User, &v.Status)
		v.Status.Timing = p.timing.Snapshot()
		err = p.db.InferenceCounts(ctx, a.User, &v.Status)
	case "submit":
		if q.Job.User != a.User || q.Job.Owner != a.Device {
			return v, httpapi.ErrDenied
		}
		err = p.admit(ctx, q.Job)
	case "tasks":
		v.Tasks, err = p.db.InferenceTasks(ctx, a.User, q.Pool, q.ID)
	case "task", "result", "job-state":
		row, e := p.db.InferenceRecord(ctx, q.ID)
		if errors.Is(e, sql.ErrNoRows) && q.Operation == "task" {
			v.Task, err = p.db.ReceiptTask(ctx, a.User, q.ID)
			return v, err
		}
		if e != nil || row.Job.User != a.User {
			return v, httpapi.ErrDenied
		}
		v.State = row.State
		if q.Operation != "job-state" {
			v.Output = row.Output
		}
		if q.Operation == "task" {
			v.Task = &row
		}
	default:
		return v, httpapi.ErrDenied
	}
	return
}
func (p *inferencePool) admit(ctx context.Context, j domain.InferenceJob) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.table == nil || !j.Valid() {
		return errors.New("incompatible inference job")
	}
	ok, err := p.db.InferenceAccess(ctx, p.roster.Cluster, j.User, j.Pool, "", false)
	if err != nil || !ok {
		return httpapi.ErrDenied
	}
	err = p.db.AdmitInference(ctx, j)
	if err != nil {
		return err
	}
	err = p.table.Admit(j.Packet.Job, domain.InferenceDigest(j))
	p.signal()
	return err
}
func (p *inferencePool) InferencePeer(ctx context.Context, caller string, q domain.InferenceCommand) (v domain.InferenceReply, err error) {
	p.mu.RLock()
	table := p.table
	valid := table != nil && table.Member(caller)
	p.mu.RUnlock()
	if !valid {
		return v, httpapi.ErrDenied
	}
	if strings.HasPrefix(q.Operation, "maintenance-") {
		return p.maintenance(ctx, caller, q)
	}
	if q.Operation == "forward-turn" {
		return p.forwardedTurn(ctx, q.ID, q.Global)
	}
	if q.Operation == "catalog" {
		return p.readProjectCatalog(ctx, caller, q)
	}
	if q.Operation == "global-chunk" {
		p.mu.RLock()
		writer := p.sessionWriter(ctx, q.Chunk.User, q.Global.Session, p.roster)
		p.mu.RUnlock()
		if writer != p.self {
			return v, httpapi.ErrDenied
		}
		chunk, e := p.db.GlobalChunk(ctx, q.Chunk.User, q.Global.Session, q.Chunk.Revision, q.Chunk.Offset)
		v.Chunk = &chunk
		return v, e
	}
	if q.Operation == "global-replica" {
		return v, p.receiveGlobal(ctx, caller, q.Artifact)
	}
	if q.Operation == "replicate-queued" {
		return v, p.admit(ctx, q.Job)
	}
	if q.Operation == "replicate" {
		if caller != q.Job.Owner {
			return v, httpapi.ErrDenied
		}
		return v, p.admit(ctx, q.Job)
	}
	c := q.Claim
	localRecheck := q.Operation == "recheck" && c.Executor == p.self
	if caller != c.Executor && !localRecheck {
		return v, httpapi.ErrDenied
	}
	switch q.Operation {
	case "confirm", "withdraw", "inspect":
		p.mu.RLock()
		defer p.mu.RUnlock()
		if table != p.table || !p.mayVote(caller) {
			return v, httpapi.ErrDenied
		}
		v.Answer = table.Transition(caller, q.Operation, c)
	case "recheck":
		p.mu.RLock()
		active := p.active[c.Job] == c
		p.mu.RUnlock()
		if active {
			v.State = "active"
			return v, nil
		}
		row, e := p.db.InferenceRecord(ctx, c.Job)
		if e != nil {
			return v, e
		}
		v.State = row.State
		if row.Claim != c {
			v.State = "withdrawn"
		}
		return v, nil
	case "own", "finish":
		p.mu.RLock()
		defer p.mu.RUnlock()
		if table != p.table || !p.mayVote(caller) {
			return v, httpapi.ErrDenied
		}
		// The RAM grant must still match before recording an irrevocable ownership receipt.
		v.Answer = table.Transition(caller, "inspect", c)
		if v.Answer.Holder != c {
			return v, nil
		}
		if q.Operation == "own" {
			if v.Answer.State != "owned" && v.Answer.State != "reserved" {
				return v, nil
			}
			err = p.db.SealInference(ctx, c)
		} else {
			if q.Output == nil {
				return v, errors.New("missing terminal output")
			}
			err = p.db.FinishInference(ctx, c, *q.Output)
		}
		if err == nil {
			v.Answer = table.Transition(caller, q.Operation, c)
		}
	default:
		return v, httpapi.ErrDenied
	}
	return
}

// Each request is authenticated independently. A missing peer leaves the job pending.
func (p *inferencePool) fanout(ctx context.Context, r domain.InferenceRoster, q domain.InferenceCommand) map[string]domain.InferenceReply {
	start := time.Now()
	var mu sync.Mutex
	var group sync.WaitGroup
	replies := map[string]domain.InferenceReply{}
	if q.Operation == "confirm" {
		p.observer.beginConfirmation(r.Peers)
	}
	for _, peer := range r.Peers {
		group.Go(func() {
			sent := time.Now()
			if q.Operation == "confirm" {
				defer p.observer.endConfirmation(peer.ID)
			}
			var v domain.InferenceReply
			var e error
			if peer.ID == p.self {
				v, e = p.InferencePeer(ctx, p.self, q)
			} else {
				var client *httpapi.InferenceClient
				client, e = p.peerClient(peer)
				if e == nil {
					v, e = client.Call(ctx, q)
				}
			}
			if q.Operation == "confirm" || q.Operation == "own" || q.Operation == "finish" {
				p.observer.round(peer.ID, e == nil, time.Since(sent))
			}
			if e == nil {
				mu.Lock()
				replies[peer.ID] = v
				mu.Unlock()
			}
		})
	}
	group.Wait()
	if q.Operation == "replicate" {
		complete := len(replies) == len(r.Peers)
		p.timing.Observe(true, complete, p.observer.propagationTime("admit/"+q.Job.Packet.Job, start, complete))
	}
	if q.Operation == "confirm" || q.Operation == "own" || q.Operation == "finish" {
		expected := map[string]string{"confirm": "reserved", "own": "owned", "finish": "finished"}[q.Operation]
		complete := len(replies) == len(r.Peers)
		for _, v := range replies {
			agreed := v.Answer.State == expected || q.Operation == "confirm" && v.Answer.State == "owned"
			complete = complete && v.Answer.Holder == q.Claim && agreed
		}
		elapsed := time.Since(start)
		if q.Operation != "confirm" {
			elapsed = p.observer.propagationTime(q.Operation+"/"+q.Claim.Job, start, complete)
		}
		p.timing.Observe(q.Operation != "confirm", complete, elapsed)
		p.observer.pulse()
	}
	return replies
}
