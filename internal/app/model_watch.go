// Follows model invalidations on currently accessible workers and repairs dropped watches.
// Reconnect always begins with a fresh REST read and rechecks the user's current membership.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sync"
	"time"
)

func (n *Runtime) watchModelWorkers(ctx context.Context) {
	defer n.discoveryWorkers.Done()
	active := map[string]context.CancelFunc{}
	var group sync.WaitGroup
	defer group.Wait()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		rows, _ := n.KnownWorkers(ctx)
		wanted := map[string]bool{}
		for _, row := range rows {
			key := n.Personal.Fingerprint() + "/" + row.Fingerprint
			wanted[key] = true
			if active[key] != nil {
				continue
			}
			watch, cancel := context.WithCancel(ctx)
			active[key] = cancel
			group.Add(1)
			go func(w domain.ClusterDevice) {
				defer group.Done()
				for watch.Err() == nil {
					c, err := n.openSavedWorker(watch, w)
					if err == nil && c.Access.Admission == "approved" {
						_ = c.WatchModels(watch, func() {
							_, _ = n.syncGlobalModel(watch, c)
							n.mu.Lock()
							if n.serviceManager != nil {
								n.serviceManager.notify()
							}
							n.mu.Unlock()
						})
					}
					select {
					case <-watch.Done():
						return
					case <-time.After(10 * time.Second):
					}
				}
			}(row)
		}
		for key, cancel := range active {
			if !wanted[key] {
				cancel()
				delete(active, key)
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
