// Samples node capacity and serializes background and administrator discovery runs.
// Readers get immutable snapshots without waiting for loopback requests.
package inventory

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sync"
	"time"
)

const DefaultPort = 11434

type Monitor struct {
	run   sync.Mutex
	mu    sync.RWMutex
	view  domain.WorkerInventory
	cache map[string]bool
	port  int
	save  func(context.Context, int) error
}

func New(port int, save func(context.Context, int) error) *Monitor {
	configured := port >= 1 && port <= 65535
	if !configured {
		port = DefaultPort
	}
	return &Monitor{port: port, save: save, view: domain.WorkerInventory{Port: port, Configured: configured}, cache: map[string]bool{}}
}
func (m *Monitor) Snapshot() domain.WorkerInventory {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.view
}
func (m *Monitor) Refresh(ctx context.Context, trial int) (domain.WorkerInventory, error) {
	if !m.run.TryLock() {
		return m.Snapshot(), errors.New("a discovery check is already running; retry after it finishes")
	}
	defer m.run.Unlock()
	port := m.port
	if trial != 0 {
		port = trial
	}
	if port < 1 || port > 65535 {
		return m.Snapshot(), errors.New("port must be 1 through 65535")
	}
	configured := m.Snapshot().Configured
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	v, err := probe(ctx, port, m.cache)
	if err != nil {
		if trial != 0 && trial != m.port {
			return m.Snapshot(), err
		}
		v = domain.WorkerInventory{Port: port, Problem: err.Error()}
	} else if !configured || port != m.port {
		err = m.save(ctx, port)
		if err != nil {
			return m.Snapshot(), errors.New("Ollama responded but its loopback port could not be saved")
		}
		m.port = port
	}
	v.Configured = err == nil || configured
	v.Resources = resources(ctx)
	v.Observed = time.Now().Unix()
	m.mu.Lock()
	m.view = v
	m.mu.Unlock()
	return v, err
}
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()
	for {
		_, _ = m.Refresh(ctx, 0)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
