// Publishes payload-free change notifications and a current view of the worker slot.
// Confirmation and propagation durations use only the observing worker's clock.
package app

import (
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"slices"
	"sync"
	"time"
)

type peerConnection struct {
	peer    domain.InferencePeer
	expires time.Time
	client  *httpapi.InferenceClient
}
type inferenceObserver struct {
	awaiting    map[string]int
	mu          sync.Mutex
	signal      chan struct{}
	propagation map[string]time.Time
	busy        bool
	blocked     map[string]time.Time
	peer        map[string]domain.PeerObservation
}

func (p *inferencePool) InferenceSignal() <-chan struct{} {
	o := &p.observer
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.signal == nil {
		o.signal = make(chan struct{})
	}
	return o.signal
}
func (o *inferenceObserver) pulse() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.signal != nil {
		close(o.signal)
		o.signal = nil
	}
}
func (o *inferenceObserver) slot(busy bool) {
	o.mu.Lock()
	o.busy = busy
	o.mu.Unlock()
	o.pulse()
}
func (o *inferenceObserver) stalled(user string, blocked bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.blocked == nil {
		o.blocked = map[string]time.Time{}
	}
	if !blocked {
		delete(o.blocked, user)
	} else if o.blocked[user].IsZero() {
		o.blocked[user] = time.Now()
	}
}
func (o *inferenceObserver) round(peer string, ok bool, elapsed time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.peer == nil {
		o.peer = map[string]domain.PeerObservation{}
	}
	o.peer[peer] = domain.PeerObservation{Received: ok, Elapsed: elapsed, Observed: time.Now().UnixMilli()}
}
func (o *inferenceObserver) populate(user string, v *domain.InferenceStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	v.Busy = o.busy
	for id := range o.awaiting {
		v.Awaiting = append(v.Awaiting, id)
	}
	slices.Sort(v.Awaiting)
	t := o.blocked[user]
	if !t.IsZero() && !o.busy {
		v.BlockedSince = t.UnixMilli()
	}
	v.Peers = make(map[string]domain.PeerObservation, len(o.peer))
	for k, val := range o.peer {
		v.Peers[k] = val
	}
}
func (p *inferencePool) peerClient(peer domain.InferencePeer) (*httpapi.InferenceClient, error) {
	p.clientsMu.Lock()
	defer p.clientsMu.Unlock()
	c := p.clients[peer.ID]
	if c.peer == peer && time.Now().Before(c.expires) {
		return c.client, nil
	}
	client, err := httpapi.NewInferenceClient(peer, p.keys)
	if err != nil {
		return nil, err
	}
	if p.clients == nil {
		p.clients = map[string]peerConnection{}
	}
	p.clients[peer.ID] = peerConnection{peer, time.Now().Add(12 * time.Hour), client}
	return client, nil
}

// Include time between retries until the last validator has applied the same receipt.
func (o *inferenceObserver) propagationTime(key string, start time.Time, complete bool) time.Duration {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.propagation == nil {
		o.propagation = map[string]time.Time{}
	}
	first, ok := o.propagation[key]
	if !ok {
		first = start
	}
	if complete {
		delete(o.propagation, key)
	} else {
		o.propagation[key] = first
	}
	return time.Since(first)
}

// Preserve the beginning of a loss of contact across repeated unsuccessful polls.
func (n *Runtime) inferenceOffline(key string) {
	now := time.Now().UnixMilli()
	old, ok := n.peerInference.Load(key)
	if ok && old.(domain.InferenceStatus).OfflineSince > 0 {
		return
	}
	n.peerInference.Store(key, domain.InferenceStatus{OfflineSince: now})
}

func (o *inferenceObserver) prune(pending map[string]bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for user := range o.blocked {
		if !pending[user] {
			delete(o.blocked, user)
		}
	}
}

func (o *inferenceObserver) beginConfirmation(peers []domain.InferencePeer) {
	o.mu.Lock()
	if o.awaiting == nil {
		o.awaiting = map[string]int{}
	}
	for _, p := range peers {
		o.awaiting[p.ID]++
	}
	o.mu.Unlock()
	o.pulse()
}
func (o *inferenceObserver) endConfirmation(id string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.awaiting[id]--
	if o.awaiting[id] <= 0 {
		delete(o.awaiting, id)
	}
}
