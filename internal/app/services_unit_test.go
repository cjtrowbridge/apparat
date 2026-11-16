// Tests service orchestration without SQLite, networking, GUI or real clocks.
// The small store fake exercises revision fencing and post-commit hints.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sync"
	"testing"
	"time"
)

type memoryServices struct {
	mu           sync.Mutex
	service      domain.Service
	observations int
	failure      error
}

func (store *memoryServices) Services(context.Context) ([]domain.Service, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return []domain.Service{store.service}, nil
}
func (store *memoryServices) SaveService(_ context.Context, s domain.Service, revision int64) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.failure != nil {
		return store.failure
	}
	if store.service.Revision != revision {
		return domain.ErrConflict
	}
	store.service = s
	return nil
}
func (store *memoryServices) RemoveService(context.Context, string) error {
	return nil
}
func (store *memoryServices) Observe(_ context.Context, o domain.Observation) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	if o.Revision != store.service.Revision {
		return domain.ErrConflict
	}
	store.observations++
	return nil
}
func TestLateProbeCannotOverwriteEditedService(t *testing.T) {
	original := domain.Service{ID: "00000000000000000000000000", Name: "one", Driver: "mock-text", Revision: 1, Enabled: true}
	store := &memoryServices{service: original}
	entered := make(chan struct{})
	release := make(chan struct{})
	manager := NewServices(store, map[string]Probe{"mock-text": func(context.Context, domain.Service) (domain.Observation, error) {
		close(entered)
		<-release
		return domain.Observation{Healthy: true}, nil
	}}, func() time.Time {
		return time.UnixMilli(10)
	}, nil)
	done := make(chan error, 1)
	go func() {
		done <- manager.Refresh(context.Background())
	}()
	<-entered
	{
		err := manager.Refresh(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	original.Enabled = false
	{
		err := manager.Update(context.Background(), original)
		if err != nil {
			t.Fatal(err)
		}
	}
	close(release)
	{
		err := <-done
		if err != nil {
			t.Fatal(err)
		}
	}
	if store.observations != 0 {
		t.Fatal("stale probe committed")
	}
	select {
	case <-manager.Changed():
	default:
		t.Fatal("missing committed hint")
	}
	select {
	case <-manager.Changed():
		t.Fatal("unexpected extra hint")
	default:
	}
}
func TestCancellationAndFailedWritesDoNotNotify(t *testing.T) {
	store := &memoryServices{service: domain.Service{ID: "00000000000000000000000000", Name: "one", Driver: "mock-text", Revision: 1, Enabled: true}, failure: errors.New("storage failure")}
	manager := NewServices(store, map[string]Probe{"mock-text": func(ctx context.Context, _ domain.Service) (domain.Observation, error) {
		return domain.Observation{}, ctx.Err()
	}}, func() time.Time {
		return time.UnixMilli(10)
	}, nil)
	if manager.Update(context.Background(), store.service) == nil {
		t.Fatal("missing write failure")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(manager.Refresh(ctx), context.Canceled) {
		t.Fatal("lost cancellation")
	}
	select {
	case <-manager.Changed():
		t.Fatal("uncommitted hint")
	default:
	}
}

func (store *memoryServices) SeedServices(context.Context, []domain.Service) error {
	return nil
}
