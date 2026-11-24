// Refreshes authenticated per-user inventory independently of policy replication cadence.
// Manual configuration is applied on the selected owner and publishes its new observation.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/inventory"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"time"
)

func (n *Runtime) LocalInventory() domain.WorkerInventory {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.inventory == nil {
		return domain.WorkerInventory{}
	}
	return n.inventory.Snapshot()
}
func (n *Runtime) refreshInventoryPeers(ctx context.Context) {
	defer n.discoveryWorkers.Done()
	tick := time.NewTicker(40 * time.Second)
	defer tick.Stop()
	for {
		workers, err := n.KnownWorkers(ctx)
		if err == nil {
			var group errgroup.Group
			group.SetLimit(4)
			for _, w := range workers {
				group.Go(func() error {
					call, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()
					client, err := n.workspaceClient(call, "worker-"+w.Fingerprint)
					if err == nil {
						v, readErr := client.Inventory(call)
						key := client.Access.User + "/" + w.Fingerprint
						if readErr == nil {
							n.peerInventory.Store(key, v)
						} else {
							n.peerInventory.Delete(key)
						}
					} else {
						n.peerInventory.Delete(n.Personal.Fingerprint() + "/" + w.Fingerprint)
					}
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
func (n *Runtime) inventoryPage(ctx context.Context, r domain.ClusterRoute) (domain.ClusterPage, error) {
	var v domain.WorkerInventory
	var client *httpapi.WorkerClient
	if r.Worker == "local" {
		v = n.LocalInventory()
	} else {
		var err error
		client, err = n.workspaceClient(ctx, "worker-"+r.Worker)
		if err != nil {
			return domain.ClusterPage{}, err
		}
		if !client.Access.Admin {
			return domain.ClusterPage{}, errors.New("worker administrator access required")
		}
		v, err = client.Inventory(ctx)
		if err != nil {
			return domain.ClusterPage{}, err
		}
	}
	port := v.Port
	if port == 0 {
		port = inventory.DefaultPort
	}
	p := domain.ClusterPage{Title: "Configure Ollama", Detail: "Try the Ollama port on this device. A successful metadata check publishes its model inventory."}
	p.Options = []domain.ClusterOption{{Label: "Test and save port", Input: "Loopback port", Run: func(ctx context.Context, input string) error {
		value, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || value < 1 || value > 65535 {
			return errors.New("enter a TCP port from 1 through 65535")
		}
		if client == nil {
			_, err = n.inventory.Refresh(ctx, value)
		} else {
			var result domain.WorkerInventory
			active, openErr := n.workspaceClient(ctx, "worker-"+r.Worker)
			if openErr != nil {
				return openErr
			}
			if active.Access.User != client.Access.User || !active.Access.Admin {
				return errors.New("worker administrator access required")
			}
			result, err = active.TryOllama(ctx, value)
			if result.Observed > 0 {
				n.peerInventory.Store(client.Access.User+"/"+r.Worker, result)
			}
		}
		n.mu.Lock()
		if n.serviceManager != nil {
			n.serviceManager.notify()
		}
		n.mu.Unlock()
		return err
	}}}
	p.InitialValue = strconv.Itoa(port)
	return p, nil
}
