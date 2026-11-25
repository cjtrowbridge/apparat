// Resolves a visible job through one authenticated replica and dials its pinned executor.
// The caller cancels this read when the conversation closes or advances to a new step.
package app

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) WatchChatPreview(ctx context.Context, user, job string, update func(domain.PreviewFrame)) error {
	if user == "" || user != n.PlanningUser() || !domain.ValidID(job) {
		return httpapi.ErrDenied
	}
	workers, err := n.KnownWorkers(ctx)
	if err != nil {
		return err
	}
	for _, w := range workers {
		c, e := n.workspaceClient(ctx, "worker-"+w.Fingerprint)
		if e != nil || c.Access.User != user {
			continue
		}
		v, e := c.Inference(ctx, domain.InferenceCommand{Operation: "job-state", ID: job})
		if e != nil {
			continue
		}
		peer := v.Executor
		if peer.ID == "" {
			return domain.ErrUnavailable
		}
		if peer.ID != w.Fingerprint {
			c, e = n.cachedWorker(ctx, peer.ID+"/"+peer.Address, func() (*httpapi.WorkerClient, error) {
				return httpapi.OpenWorker(ctx, peer.Address, peer.ID, n.Personal, n.Keys, &n.observations)
			})
		}
		if e != nil || c.Access.User != user || c.Access.Cluster != w.Cluster {
			return httpapi.ErrDenied
		}
		return c.WatchPreview(ctx, job, update)
	}
	return domain.ErrUnavailable
}
