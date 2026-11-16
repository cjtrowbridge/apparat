// Exposes transport-neutral node and service inspection plus local commands.
// Returned values contain no endpoint, credential or filesystem authority.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"runtime"
)

type Diagnostic struct {
	Healthy        bool
	Mode           Mode
	RootDir        string
	LastRunPath    string
	DatabasePath   string
	IdentityStatus identity.Status
	Message        string
}

func (node *Runtime) Health() domain.NodeHealth {
	node.mu.Lock()
	defer node.mu.Unlock()
	return domain.NodeHealth{Version: "0.1.0", State: node.state, Ready: node.state == "ready"}
}
func (node *Runtime) Device() domain.Device {
	return domain.Device{OS: runtime.GOOS, Architecture: runtime.GOARCH, IdentityStatus: string(identity.Classify(node.config.IdentityDir))}
}
func (node *Runtime) List(ctx context.Context) ([]domain.Service, error) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return nil, domain.ErrUnavailable
	}
	return node.serviceManager.List(ctx)
}
func (node *Runtime) Get(ctx context.Context, id string) (domain.Service, error) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.Service{}, domain.ErrUnavailable
	}
	return node.serviceManager.Get(ctx, id)
}
func (node *Runtime) Capabilities(ctx context.Context) ([]domain.Capability, error) {
	list, err := node.List(ctx)
	if err != nil {
		return nil, err
	}
	result := []domain.Capability{}
	for _, service := range list {
		if service.Enabled && service.Health == "healthy" {
			result = append(result, service.Capabilities...)
		}
	}
	return result, nil
}
func (node *Runtime) CreateService(ctx context.Context, name, driver string) (domain.Service, error) {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.Service{}, domain.ErrUnavailable
	}
	return node.serviceManager.Create(ctx, name, driver)
}
func (node *Runtime) UpdateService(ctx context.Context, service domain.Service) error {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.ErrUnavailable
	}
	return node.serviceManager.Update(ctx, service)
}
func (node *Runtime) RemoveService(ctx context.Context, id string) error {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.ErrUnavailable
	}
	return node.serviceManager.Remove(ctx, id)
}
func (node *Runtime) Refresh(ctx context.Context) error {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.ErrUnavailable
	}
	return node.serviceManager.Refresh(ctx)
}
func (node *Runtime) Changed() <-chan struct{} {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.serviceManager == nil {
		return nil
	}
	return node.serviceManager.Changed()
}
func (node *Runtime) Address() string {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.server == nil {
		return ""
	}
	return node.server.Address()
}
func (node *Runtime) Doctor(ctx context.Context) Diagnostic {
	diag := Diagnostic{Mode: node.mode, RootDir: node.config.RootDir, LastRunPath: node.config.LastRunPath, DatabasePath: node.config.DatabasePath, IdentityStatus: identity.Classify(node.config.IdentityDir), Message: "ok"}
	err := node.Initialize(ctx)
	if err == nil {
		node.mu.Lock()
		err = node.database.Integrity(ctx)
		node.mu.Unlock()
	}
	diag.Healthy = err == nil
	if err != nil {
		diag.Message = err.Error()
	}
	_ = node.RecordLastRun("info", "doctor", "healthy", "doctor completed", map[string]any{"healthy": diag.Healthy})
	return diag
}
func (node *Runtime) Backup(ctx context.Context, path string) error {
	node.mu.Lock()
	defer node.mu.Unlock()
	if node.state != "ready" {
		return domain.ErrUnavailable
	}
	return node.database.Backup(ctx, path)
}
