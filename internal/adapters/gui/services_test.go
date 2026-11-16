// Verifies direct query projection and cancellation without a display or HTTP.
// A bounded fake supplies commit hints and immutable service values.
package gui

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"sync"
	"testing"
	"time"
)

type projectionSource struct {
	mu       sync.Mutex
	services []domain.Service
	err      error
	changed  chan struct{}
	called   chan struct{}
}

func (source *projectionSource) Changed() <-chan struct{} {
	return source.changed
}
func (source *projectionSource) List(context.Context) ([]domain.Service, error) {
	source.mu.Lock()
	defer source.mu.Unlock()
	if source.called != nil {
		source.called <- struct{}{}
	}
	return append([]domain.Service{}, source.services...), source.err
}
func receiveProjection(t *testing.T, ch <-chan ServiceProjection) ServiceProjection {
	t.Helper()
	select {
	case p := <-ch:
		return p
	case <-time.After(time.Second):
		t.Fatal("query did not converge")
		return ServiceProjection{}
	}
}
func TestDirectProjectionStatesAndShutdown(t *testing.T) {
	source := &projectionSource{changed: make(chan struct{}, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	updates := WatchServices(ctx, source)
	if receiveProjection(t, updates).State != "empty" {
		t.Fatal("missing empty state")
	}
	source.mu.Lock()
	source.services = []domain.Service{{ID: "00000000000000000000000000", Name: "one", Health: "stale", Enabled: true, ObservedAt: 1}}
	source.mu.Unlock()
	source.changed <- struct{}{}
	projection := receiveProjection(t, updates)
	if projection.Services[0].Health != "stale" {
		t.Fatal(projection)
	}
	source.mu.Lock()
	source.err = errors.New("private query failure")
	source.mu.Unlock()
	source.changed <- struct{}{}
	projection = receiveProjection(t, updates)
	if projection.State != "query-error" || len(projection.Services) != 1 {
		t.Fatal(projection)
	}
	cancel()
	select {
	case _, open := <-updates:
		if open {
			t.Fatal("worker did not close")
		}
	case <-time.After(time.Second):
		t.Fatal("worker leaked")
	}
}

func TestTimestampHintsDoNotRequestWidgetRebuilds(t *testing.T) {
	source := &projectionSource{changed: make(chan struct{}, 1), called: make(chan struct{}, 1), services: []domain.Service{{ID: "00000000000000000000000000", Health: "healthy", ObservedAt: 1}}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	updates := WatchServices(ctx, source)
	_ = receiveProjection(t, updates)
	<-source.called
	for index := int64(2); index <= 102; index++ {
		source.mu.Lock()
		source.services[0].ObservedAt = index
		source.mu.Unlock()
		source.changed <- struct{}{}
		select {
		case <-source.called:
		case <-ctx.Done():
			t.Fatal("query worker did not progress")
		}
	}
	select {
	case update := <-updates:
		t.Fatal("timestamp-only widget rebuild request", update)
	default:
	}
}
