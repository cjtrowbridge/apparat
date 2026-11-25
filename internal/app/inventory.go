// Refreshes authenticated per-user inventory independently of policy replication cadence.
// Manual configuration is applied on the selected owner and publishes its new observation.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/inventory"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/adapters/telemetry"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"sync"
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
			var mu sync.Mutex
			views := map[string]domain.InferenceRoster{}
			clients := map[string]*httpapi.WorkerClient{}
			group.SetLimit(4)
			for _, w := range workers {
				group.Go(func() error {
					started := time.Now()
					call, cancel := context.WithTimeout(ctx, 25*time.Second)
					defer cancel()
					client, err := n.workspaceClient(call, "worker-"+w.Fingerprint)
					if err == nil {
						v, readErr := client.Inventory(call)
						err = readErr
						key := client.Access.User + "/" + w.Fingerprint
						if readErr == nil {
							n.peerInventory.Store(key, v)
							n.notifyInventory()
							n.refreshCapacity(call, client, w, v)
							reply, statusErr := client.Inference(call, domain.InferenceCommand{Operation: "status", Roster: membershipFor(w.Cluster, workers)})
							if statusErr == nil {
								n.peerInference.Store(key, reply.Status)
								if reply.Protocol == 2 {
									mu.Lock()
									views[w.Fingerprint] = reply.Roster
									clients[w.Fingerprint] = client
									mu.Unlock()
								}
							} else {
								n.inferenceOffline(key)
							}
						} else {
							n.peerInventory.Delete(key)
							n.inferenceOffline(key)
						}
					} else {
						key := n.Personal.Fingerprint() + "/" + w.Fingerprint
						n.peerInventory.Delete(key)
						n.inferenceOffline(key)
					}
					if errors.Is(err, httpapi.ErrDenied) {
						n.rosterHardware.Delete(n.Personal.Fingerprint() + "/" + w.Fingerprint)
						n.mu.Lock()
						if n.database != nil {
							_ = n.database.SaveHardware(ctx, w.Fingerprint, persistence.Access{User: n.Personal.Fingerprint(), Cluster: w.Cluster}, nil)
						}
						n.mu.Unlock()
					}
					n.observations.Observe(telemetry.Event{Operation: "worker-inventory", Peer: w.Fingerprint, Failed: err != nil, Elapsed: time.Since(started).Microseconds()})
					return nil
				})
			}
			_ = group.Wait()
			n.convergeMemberships(ctx, views, clients, workers)
			n.mu.Lock()
			if n.serviceManager != nil && len(workers) > 0 {
				n.serviceManager.notify()
			}
			n.mu.Unlock()
		}
		select {
		case <-ctx.Done():
			return
		case <-n.inventoryHint:
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

func (n *Runtime) notifyInventory() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.serviceManager != nil {
		n.serviceManager.notify()
	}
}

func (n *Runtime) refreshCapacity(ctx context.Context, c *httpapi.WorkerClient, w domain.ClusterDevice, v domain.WorkerInventory) {
	reports, err := c.MemberHardware(ctx, n.LocalInventory().Capacity())
	if err != nil && !errors.Is(err, httpapi.ErrDenied) {
		reports = w.Hardware
	}
	if reports == nil {
		reports = map[string]domain.HardwareCapacity{}
	}
	capacity := v.Capacity()
	if capacity.Threads == 0 {
		capacity = w.Hardware[w.Fingerprint]
	}
	reports[w.Fingerprint] = capacity
	n.mu.Lock()
	db := n.database
	n.mu.Unlock()
	if db != nil {
		err = db.SaveHardware(ctx, w.Fingerprint, c.Access, reports)
		if err == nil {
			n.notifyInventory()
		}
	}
}
