// Coordinates service intent, bounded probes and coalesced change hints.
// Each refresh runs at most one probe; revisions fence late results after edits.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sync"
	"time"
)

type ServiceStore interface {
	SaveService(context.Context, domain.Service, int64) error
	SeedServices(context.Context, []domain.Service) error
	RemoveService(context.Context, string) error
	Services(context.Context) ([]domain.Service, error)
	Observe(context.Context, domain.Observation) error
}
type Probe func(context.Context, domain.Service) (domain.Observation, error)
type Services struct {
	store     ServiceStore
	drivers   map[string]Probe
	now       func() time.Time
	id        func() (string, error)
	refreshMu sync.Mutex
	commands  sync.Mutex
	changed   chan struct{}
}

func NewServices(store ServiceStore, drivers map[string]Probe, now func() time.Time, id func() (string, error)) *Services {
	copied := map[string]Probe{}
	for kind, probe := range drivers {
		copied[kind] = probe
	}
	return &Services{store: store, drivers: copied, now: now, id: id, changed: make(chan struct{}, 1)}
}
func (services *Services) Changed() <-chan struct{} {
	return services.changed
}
func (services *Services) notify() {
	select {
	case services.changed <- struct{}{}:
	default:
	}
}
func (services *Services) List(ctx context.Context) ([]domain.Service, error) {
	list, err := services.store.Services(ctx)
	if err != nil {
		return nil, err
	}
	for index := range list {
		if list[index].Enabled && list[index].ObservedAt > 0 && services.now().UnixMilli()-list[index].ObservedAt > 15000 {
			list[index].Health = "stale"
			list[index].Capabilities = []domain.Capability{}
		}
	}
	return list, nil
}
func (services *Services) Get(ctx context.Context, id string) (domain.Service, error) {
	list, err := services.List(ctx)
	if err != nil {
		return domain.Service{}, err
	}
	for _, service := range list {
		if service.ID == id {
			return service, nil
		}
	}
	return domain.Service{}, domain.ErrNotFound
}
func (services *Services) Create(ctx context.Context, name, driver string) (domain.Service, error) {
	services.commands.Lock()
	defer services.commands.Unlock()
	list, err := services.List(ctx)
	if err != nil {
		return domain.Service{}, err
	}
	if len(list) >= 128 {
		return domain.Service{}, errors.New("local service limit reached")
	}
	id, err := services.id()
	if err != nil {
		return domain.Service{}, err
	}
	service := domain.Service{ID: id, Name: name, Driver: driver, Enabled: true, Revision: 1}
	err = domain.ValidateService(service)
	if err != nil {
		return domain.Service{}, err
	}
	if services.drivers[driver] == nil {
		return domain.Service{}, errors.New("driver unavailable")
	}
	err = services.store.SaveService(ctx, service, 0)
	if err == nil {
		services.notify()
	}
	return service, err
}
func (services *Services) Update(ctx context.Context, service domain.Service) error {
	services.commands.Lock()
	defer services.commands.Unlock()
	previous := service.Revision
	service.Revision++
	err := domain.ValidateService(service)
	if err != nil {
		return err
	}
	if services.drivers[service.Driver] == nil {
		return errors.New("driver unavailable")
	}
	err = services.store.SaveService(ctx, service, previous)
	if err == nil {
		services.notify()
	}
	return err
}
func (services *Services) Remove(ctx context.Context, id string) error {
	services.commands.Lock()
	defer services.commands.Unlock()
	err := services.store.RemoveService(ctx, id)
	if err == nil {
		services.notify()
	}
	return err
}
func (services *Services) Refresh(ctx context.Context) error {
	// Skip duplicate concurrent refreshes rather than building an unbounded queue.
	if !services.refreshMu.TryLock() {
		return nil
	}
	defer services.refreshMu.Unlock()
	list, err := services.List(ctx)
	if err != nil {
		return err
	}
	for _, service := range list {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !service.Enabled {
			continue
		}
		probe := services.drivers[service.Driver]
		if probe == nil {
			continue
		}
		probeCtx, cancel := context.WithTimeout(ctx, time.Second)
		observation, probeErr := probe(probeCtx, service)
		cancel()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		observation.ServiceID = service.ID
		observation.Revision = service.Revision
		observation.At = services.now().UnixMilli()
		if probeErr != nil || !observation.Healthy {
			observation.Healthy = false
			observation.Error = "probe_failed"
			observation.Capabilities = nil
		} else {
			observation.Error = ""
		}
		err = services.store.Observe(ctx, observation)
		if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		services.notify()
	}
	return nil
}
func (services *Services) EnsureDemo(ctx context.Context) error {
	list, err := services.List(ctx)
	if err != nil {
		return err
	}
	if len(list) > 0 {
		return nil
	}
	batch := []domain.Service{}
	for _, entry := range [][2]string{{"Text one", "mock-text"}, {"Text two", "mock-text"}, {"Image one", "mock-image"}} {
		id, err := services.id()
		if err != nil {
			return err
		}
		batch = append(batch, domain.Service{ID: id, Name: entry[0], Driver: entry[1], Enabled: true, Revision: 1})
	}
	err = services.store.SeedServices(ctx, batch)
	if err != nil {
		return err
	}
	services.notify()
	return services.Refresh(ctx)
}
