// Supplies deterministic local mock probes for the first durable service slice.
// No network, endpoint credentials, real models or background process is involved.
package mockprovider

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/app"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func Drivers() map[string]app.Probe {
	return map[string]app.Probe{"mock-text": probe("text_generation"), "mock-image": probe("image_generation")}
}
func probe(workload string) app.Probe {
	return func(ctx context.Context, service domain.Service) (domain.Observation, error) {
		if ctx.Err() != nil {
			return domain.Observation{}, ctx.Err()
		}
		if service.Fail {
			return domain.Observation{}, errors.New("mock failure")
		}
		return domain.Observation{Healthy: true, Capabilities: []domain.Capability{{ID: service.ID, ServiceID: service.ID, Workload: workload}}}, nil
	}
}
