// Defines local service intent and safe immutable query values.
// Validation is deterministic and independent of persistence and transports.
package domain

import "errors"

var ErrNotFound = errors.New("service not found")
var ErrConflict = errors.New("service revision changed")
var ErrUnavailable = errors.New("node not ready")

type Service struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Driver       string       `json:"driver"`
	Enabled      bool         `json:"enabled"`
	Fail         bool         `json:"-"`
	Revision     int64        `json:"revision"`
	Health       string       `json:"health"`
	ObservedAt   int64        `json:"observed_at_ms"`
	Error        string       `json:"error,omitempty"`
	Capabilities []Capability `json:"capabilities"`
}
type Capability struct {
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	Workload  string `json:"workload_class"`
}
type Observation struct {
	ServiceID    string
	Revision     int64
	At           int64
	Healthy      bool
	Error        string
	Capabilities []Capability
}
type NodeHealth struct {
	Version string `json:"version"`
	State   string `json:"state"`
	Ready   bool   `json:"ready"`
}
type Device struct {
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	IdentityStatus string `json:"identity_status"`
}

func ValidateService(service Service) error {
	if !ValidID(service.ID) {
		return errors.New("invalid service ID")
	}
	if len(service.Name) == 0 || len(service.Name) > 80 {
		return errors.New("service name must contain 1 to 80 bytes")
	}
	if service.Driver != "mock-text" && service.Driver != "mock-image" {
		return errors.New("unsupported driver")
	}
	if service.Revision < 1 {
		return errors.New("invalid service revision")
	}
	return nil
}
