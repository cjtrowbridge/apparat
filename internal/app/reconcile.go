// Reconciles a signed-in personal installation's worker views on activity and policy timers.
// A single goroutine coalesces triggers and owns each worker's retry deadline.
package app

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

func (n *Runtime) RequestSync() {
	select {
	case n.inventoryHint <- struct{}{}:
	default:
	}
	select {
	case n.syncHint <- struct{}{}:
	default:
	}
}

func (n *Runtime) reconcile(ctx context.Context, db *persistence.DB) {
	defer n.discoveryWorkers.Done()
	type schedule struct {
		next, last time.Time
		failures   int
	}
	peers := map[string]schedule{}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		activity := false
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		case <-n.syncHint:
			activity = true
		}
		user := n.Personal.Fingerprint()
		next := time.Now().Add(5 * time.Minute)
		if user != "" && len(n.Keys.Public()) > 0 {
			workers, err := db.KnownWorkers(ctx, user)
			if err == nil {
				for _, worker := range workers {
					key := user + "/" + worker.Fingerprint
					s := peers[key]
					now := time.Now()
					if !now.Before(s.next) || activity && now.Sub(s.last) >= 5*time.Second {
						request, cancel := context.WithTimeout(ctx, 25*time.Second)
						delay, err := n.syncOne(request, db, worker, user)
						cancel()
						s.last = now
						if err != nil {
							s.failures = min(s.failures+1, 6)
							delay = time.Second * time.Duration(5<<s.failures)
						} else {
							s.failures = 0
						}
						jitter := time.Duration(rand.Int64N(max(1, int64(delay/20))))
						s.next = time.Now().Add(delay + jitter)
						peers[key] = s
					}
					if s.next.Before(next) {
						next = s.next
					}
				}
			}
		}
		timer.Reset(max(time.Second, time.Until(next)))
	}
}

func (n *Runtime) syncOne(ctx context.Context, db *persistence.DB, w domain.ClusterDevice, user string) (time.Duration, error) {
	a := persistence.Access{Cluster: w.Cluster, User: user, Device: identity.Fingerprint(n.Keys.Public())}
	c, err := n.openSavedWorker(ctx, w)
	if err == nil && c.Access.Cluster == a.Cluster && c.Access.User == a.User && c.Access.Device == a.Device {
		if c.Access.Admission != "approved" {
			err = httpapi.ErrDenied
		} else {
			_, _ = n.syncGlobalModel(ctx, c)
			for range 8 {
				var seconds int
				seconds, err = c.Sync(ctx, db)
				if err != nil {
					break
				}
				if seconds > 0 {
					return time.Duration(seconds) * time.Second, nil
				}
				// A new replica can add or remove devices while Status is open.
				n.notifyInventory()
				select {
				case n.inventoryHint <- struct{}{}:
				default:
				}
			}
			if err == nil {
				return time.Second, nil
			}
		}
	}
	var change *httpapi.ChangedWorker
	if errors.As(err, &change) {
		n.notifyReplacement(w, change)
	}
	if errors.Is(err, httpapi.ErrDenied) {
		return 0, errors.Join(err, db.ForgetReplica(ctx, w.Fingerprint, a))
	}
	if err == nil {
		err = errors.New("worker identity mismatch")
	}
	return 0, err
}
