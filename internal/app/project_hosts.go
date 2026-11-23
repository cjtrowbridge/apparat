// Resolves selected worker fingerprints before provider operations and catalog reads.
// A remote project remains a display-only, host-scoped entry on the client.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"time"
)

func (n *Runtime) projectHost(ctx context.Context, host string) (*httpapi.WorkerClient, error) {
	rows, err := n.KnownWorkers(ctx)
	if err != nil {
		return nil, err
	}
	for _, w := range rows {
		if host == "worker-"+w.Fingerprint {
			return n.openSavedWorker(ctx, w)
		}
	}
	return nil, errors.New("unknown project host")
}

func (n *Runtime) GitProviders(ctx context.Context, host string) ([]domain.GitProvider, error) {
	if host == "local" {
		return projectgit.Providers(ctx), nil
	}
	c, err := n.projectHost(ctx, host)
	if err != nil {
		return nil, err
	}
	return c.Providers(ctx)
}

func (n *Runtime) GitOperation(ctx context.Context, host string, req domain.GitRequest) (domain.GitStatus, error) {
	if host == "local" {
		if req.Operation != "install" {
			return domain.GitStatus{}, errors.New("unsupported local operation")
		}
		err := projectgit.InstallProvider(ctx, req.Provider)
		return domain.GitStatus{State: "completed"}, err
	}
	client, err := n.projectHost(ctx, host)
	if err != nil {
		return domain.GitStatus{}, err
	}
	return client.GitOperation(ctx, req)
}

func (n *Runtime) remoteProjects(ctx context.Context) []domain.Project {
	workers, err := n.KnownWorkers(ctx)
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	var out []domain.Project
	for _, w := range workers {
		if ctx.Err() != nil {
			break
		}
		if !w.Admin {
			continue
		}
		c, err := n.openSavedWorker(ctx, w)
		if err != nil {
			continue
		}
		rows, err := c.RemoteProjects(ctx)
		if err != nil {
			continue
		}
		for _, p := range rows {
			out = append(out, domain.Project{ID: "worker-" + w.Fingerprint + "/" + p.ID, RemoteHost: w.DisplayName(), Name: p.Name, Location: p.Location, Branch: p.Branch, Revision: p.Revision, Available: p.Available})
		}
	}
	return out
}

// GitSession reuses the device-bound bearer for one operation's status requests.
func (n *Runtime) GitSession(ctx context.Context, host string) (func(context.Context, domain.GitRequest) (domain.GitStatus, error), error) {
	if host == "local" {
		return func(ctx context.Context, req domain.GitRequest) (domain.GitStatus, error) {
			return n.GitOperation(ctx, host, req)
		}, nil
	}
	client, err := n.projectHost(ctx, host)
	if err != nil {
		return nil, err
	}
	renew := time.Now().Add(4 * time.Minute)
	return func(ctx context.Context, req domain.GitRequest) (domain.GitStatus, error) {
		if time.Now().After(renew) {
			next, err := n.projectHost(ctx, host)
			if err != nil {
				return domain.GitStatus{}, err
			}
			client = next
			renew = time.Now().Add(4 * time.Minute)
		}
		return client.GitOperation(ctx, req)
	}, nil
}
