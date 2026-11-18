// Records a file save before applying it and binds retries to the original inputs.
// A pending write can be reconciled with the file digest after interruption.
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type fileTransaction struct{ Path, Base, Result, State, Reason string }

func (n *Runtime) saveWorkspaceFile(ctx context.Context, location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	tx := fileTransaction{Path: q.Path, Base: q.Base, Result: projectgit.Digest(q.Text), State: "pending"}
	if !domain.ValidID(q.ID) {
		return v, errors.New("save requires a transaction ID")
	}
	kind, record, err := n.database.ProjectEntry(ctx, q.ProjectID, q.ID)
	if err == nil {
		var previous fileTransaction
		err = json.Unmarshal([]byte(record), &previous)
		if err != nil || kind != "save" || previous.Path != tx.Path || previous.Base != tx.Base || previous.Result != tx.Result {
			return v, errors.New("transaction ID belongs to different inputs")
		}
		switch previous.State {
		case "applied":
			return domain.WorkspaceView{Report: "Already applied"}, nil
		case "conflicted":
			return v, errors.New(previous.Reason)
		}
	} else if errors.Is(err, sql.ErrNoRows) {
		data, _ := json.Marshal(tx)
		err = n.database.PrepareWorkspaceSave(ctx, q, string(data))
		if err != nil {
			return v, err
		}
	} else {
		return v, err
	}
	v, writeErr := n.workspaceFile(ctx, location, q)
	tx.State = "applied"
	if writeErr != nil {
		tx.State = "conflicted"
		tx.Reason = writeErr.Error()
	}
	data, _ := json.Marshal(tx)
	err = n.database.FinishWorkspaceSave(context.WithoutCancel(ctx), q.ProjectID, q.ID, q.Path, q.Base, q.Text, string(data), writeErr == nil)
	if err != nil {
		return v, err
	}
	if writeErr != nil {
		return v, writeErr
	}
	v.Report = "Saved to checkout"
	return v, nil
}

func (n *Runtime) reconcileFileSaves(ctx context.Context, q domain.WorkspaceRequest, observed domain.WorkspaceView) error {
	items, err := n.database.PendingFileSaves(ctx, q.ProjectID, q.Path)
	if err != nil {
		return err
	}
	for _, r := range items {
		var tx fileTransaction
		err = json.Unmarshal([]byte(r.Label), &tx)
		if err != nil {
			return err
		}
		applied := tx.Result == observed.Base
		switch {
		case applied:
			tx.State = "applied"
		case tx.Base == observed.Base:
			tx.State = "interrupted"
			tx.Reason = "Checkout preserved; draft available to retry"
		default:
			tx.State = "conflicted"
			tx.Reason = "Checkout changed before recovery"
		}
		record, _ := json.Marshal(tx)
		err = n.database.FinishWorkspaceSave(ctx, q.ProjectID, r.ID, q.Path, tx.Base, observed.Text, string(record), applied)
		if err != nil {
			return err
		}
	}
	return nil
}
