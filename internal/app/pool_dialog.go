// Lets an approved user submit the harmless hash service and inspect their own tasks.
// Each action uses the pinned worker; the server validates current access again.
package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cjtrowbridge/apparat/internal/adapters/httpapi"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"slices"
)

func (n *Runtime) PoolDialog(ctx context.Context, id string, dialog identity.Dialog) {
	if !n.claimMu.TryLock() {
		return
	}
	defer n.claimMu.Unlock()
	workers, err := n.KnownWorkers(ctx)
	if err != nil {
		return
	}
	for _, w := range workers {
		if w.Fingerprint != id {
			continue
		}
		c, err := n.openSavedWorker(ctx, w)
		if err == nil {
			n.mu.Lock()
			db := n.database
			n.mu.Unlock()
			if db == nil {
				return
			}
			_, err = c.Sync(ctx, db)
			if err == nil {
				var r persistence.SignedReplica
				r, _, err = db.LocalReplica(ctx, c.Worker, c.Access)
				if err == nil {
					err = taskDialog(ctx, c, r.View, dialog)
				}
			}
		}
		if err != nil {
			nativeQuestion(dialog, "notice", w.Host, err.Error())
		}
		n.RequestSync()
		return
	}
}

func taskDialog(ctx context.Context, c *httpapi.WorkerClient, v persistence.Replica, dialog identity.Dialog) error {
	if len(v.Pools) == 0 {
		return errors.New("No accessible pool offers the hash-text service.")
	}
	pool, ok := nativeQuestion(dialog, "choose", "Pool tasks", "Choose a pool.", v.Pools...)
	if !ok || !slices.Contains(v.Pools, pool) {
		return nil
	}
	action, ok := nativeQuestion(dialog, "choose", pool, "Choose an operation.", "Submit hash task", "My tasks")
	if !ok {
		return nil
	}
	if action == "Submit hash task" {
		prompt, ok := nativeQuestion(dialog, "text", "Submit hash task", "Enter the text to hash.")
		if !ok {
			return nil
		}
		var id [16]byte
		rand.Read(id[:])
		err := c.SubmitHash(ctx, pool, hex.EncodeToString(id[:]), prompt)
		if err == nil {
			nativeQuestion(dialog, "notice", pool, "The worker accepted your task.")
		}
		return err
	}
	states := map[string]string{}
	var after int64
	for {
		events, err := c.QueueChanges(ctx, pool, after)
		if err != nil {
			return err
		}
		for _, event := range events {
			if event.Revision <= after {
				return errors.New("invalid queue page")
			}
			states[event.ID] = event.State
			after = event.Revision
		}
		if len(events) < 64 {
			break
		}
	}
	if len(states) == 0 {
		return errors.New("You have no tasks in this pool.")
	}
	labels := []string{}
	ids := map[string]string{}
	for id, state := range states {
		label := fmt.Sprintf("%s · %s", state, id)
		labels = append(labels, label)
		ids[label] = id
	}
	slices.Sort(labels)
	label, ok := nativeQuestion(dialog, "choose", "My tasks", "Choose a task.", labels...)
	id := ids[label]
	if !ok || id == "" {
		return nil
	}
	switch states[id] {
	case "succeeded":
		value, err := c.Result(ctx, pool, id)
		if err == nil {
			nativeQuestion(dialog, "notice", "Verified SHA-256 result", value)
		}
		return err
	case "running", "queued":
		_, ok = nativeQuestion(dialog, "choose", "Cancel task", "Stop this task?", "Cancel task")
		if ok {
			return c.Cancel(ctx, pool, id)
		}
	}
	return nil
}
