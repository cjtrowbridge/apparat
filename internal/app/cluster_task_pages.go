// Presents the authenticated user's live pool queues and verified artifacts.
// Reading a task is independent of explicit submission and cancellation.
package app

import (
	"context"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"slices"
)

func (b *clusterPageBuilder) taskPage(ctx context.Context) error {
	c, r := b.client, b.route
	if r.Page == "work" {
		b.Title = "Pool tasks"
		b.runtime.mu.Lock()
		db := b.runtime.database
		b.runtime.mu.Unlock()
		if db == nil {
			return domain.ErrUnavailable
		}
		_, err := c.Sync(ctx, db)
		if err != nil {
			return err
		}
		replica, _, err := db.LocalReplica(ctx, c.Worker, c.Access)
		if err != nil {
			return err
		}
		for _, pool := range replica.View.Pools {
			b.link(pool, "tasks", pool, "")
		}
		if len(b.Options) == 0 {
			b.Detail = "No accessible pool offers hash-text."
		}
		return nil
	}
	b.Title = r.Pool + " · My tasks"
	states := map[string]persistence.QueueChange{}
	var after int64
	for {
		events, err := c.QueueChanges(ctx, r.Pool, after)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.Revision <= after {
				return errors.New("invalid queue revision")
			}
			after = event.Revision
			states[event.ID] = event
		}
		if len(events) < 64 {
			break
		}
	}
	if r.Page == "task" {
		task, ok := states[r.Target]
		if !ok {
			return errors.New("task unavailable")
		}
		b.Title = "Task · " + task.State
		b.Detail = task.ID
		if task.State == "succeeded" {
			result, err := c.Result(ctx, r.Pool, task.ID)
			if err != nil {
				return err
			}
			b.Detail += "\nVerified SHA-256 result\n" + result
		}
		if task.State == "queued" || task.State == "running" {
			b.Options = append(b.Options, domain.ClusterOption{Label: "Cancel task", Confirm: "Cancel this task in " + r.Pool + "?", Run: func(ctx context.Context, _ string) error {
				return c.Cancel(ctx, r.Pool, task.ID)
			}})
		}
		return nil
	}
	id := commandID()
	b.Options = append(b.Options, domain.ClusterOption{Label: "Submit hash task", Input: "Text to hash", Run: func(ctx context.Context, value string) error {
		return c.SubmitHash(ctx, r.Pool, id, value)
	}})
	sorted := make([]persistence.QueueChange, 0, len(states))
	for _, e := range states {
		sorted = append(sorted, e)
	}
	slices.SortFunc(sorted, func(a, b persistence.QueueChange) int {
		if a.Revision > b.Revision {
			return -1
		}
		return 1
	})
	for _, t := range sorted {
		b.link(t.State+" · "+t.ID, "task", r.Pool, t.ID)
	}
	if len(states) == 0 {
		b.Detail = "No tasks submitted yet."
	}
	return nil
}
