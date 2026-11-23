// Loads immutable service projections outside the render/update goroutine.
// One worker coalesces committed change hints and owns each cancellable query.
package gui

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"reflect"
	"time"
)

type ServiceQueries interface {
	List(context.Context) ([]domain.Service, error)
	Changed() <-chan struct{}
}
type ServiceProjection struct {
	Devices     []domain.ClusterDevice
	DeviceError bool
	Services    []domain.Service
	State       string
}

func WatchServices(ctx context.Context, queries ServiceQueries) <-chan ServiceProjection {
	results := make(chan ServiceProjection, 1)
	go func() {
		defer close(results)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		var previous ServiceProjection
		changed := queries.Changed()
		for {
			queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			list, err := queries.List(queryCtx)
			var devices []domain.ClusterDevice
			var deviceErr error
			q, ok := queries.(interface {
				KnownWorkers(context.Context) ([]domain.ClusterDevice, error)
			})
			if ok {
				devices, deviceErr = q.KnownWorkers(queryCtx)
			}
			cancel()
			if ctx.Err() != nil {
				return
			}
			state := "ready"
			if err != nil {
				state = "query-error"
				list = previous.Services
			} else if len(list) == 0 {
				state = "empty"
			}
			// Retain no timestamp-only redraw dependency; shared queries own freshness.
			list = append([]domain.Service{}, list...)
			for index := range list {
				list[index].ObservedAt = 0
			}

			projection := ServiceProjection{Devices: devices, DeviceError: deviceErr != nil, Services: list, State: state}
			if !reflect.DeepEqual(projection, previous) {
				select {
				case <-results:
				default:
				}
				select {
				case results <- projection:
				case <-ctx.Done():
					return
				}
				previous = projection
			}
			select {
			case <-ctx.Done():
				return
			case <-changed:
			case <-ticker.C:
			}
		}
	}()
	return results
}
