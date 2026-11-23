// Exposes the harmless hash-text queue through current worker authorization.
// Only the submitting user can retrieve the resulting content-addressed artifact.
package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/identity"
)

type TaskRequest struct {
	Pool   string
	ID     string
	Prompt string
	After  int64
}

func (a *workerAuth) tasks(w http.ResponseWriter, r *http.Request, access persistence.Access, keys *identity.Session) {
	worker := identity.Fingerprint(keys.Public())
	var q TaskRequest
	if r.Method != http.MethodPost || !boundedJSON(w, r, &q) || q.Pool == "" || len(q.Pool) > 128 || a.artifacts == "" {
		respond(w, 400, map[string]string{"error": "invalid_task_request"})
		return
	}
	ctx := r.Context()
	var result any = map[string]string{"state": "accepted"}
	var err error
	switch r.URL.Path {
	case "/v1/tasks/submit":
		err = a.db.SubmitHash(ctx, access, q.Pool, q.ID, q.Prompt)
		if err == nil {
			// The first capability is bounded and local. Resume queued work when the
			// submitter retries or checks its pool after a worker restart.
			_ = a.runOne(ctx, access.Cluster, q.Pool, keys)
		}
	case "/v1/tasks/changes":
		if a.db.MayDispatch(ctx, access, q.Pool, worker) {
			_ = a.runOne(ctx, access.Cluster, q.Pool, keys)
		}
		result, err = a.db.QueueAfter(ctx, access, q.Pool, q.After)
	case "/v1/tasks/cancel":
		err = a.db.CancelTask(ctx, access, q.Pool, q.ID)
	case "/v1/tasks/result":
		var digest string
		digest, err = a.db.ResultReference(ctx, access, q.Pool, q.ID)
		if err == nil {
			var body []byte
			body, err = a.readArtifact(digest)
			if err == nil {
				receipt, receiptErr := a.db.ResultReceipt(ctx, access, q.Pool, q.ID, digest)
				result = map[string]any{"sha256": digest, "text": string(body)}
				if receiptErr == nil {
					result.(map[string]any)["receipt"] = receipt
				}
			}
		}
	}
	if err != nil {
		respond(w, 403, map[string]string{"error": "task_unavailable"})
		return
	}
	respond(w, 200, result)
}

func (a *workerAuth) runOne(ctx context.Context, cluster, pool string, keys *identity.Session) error {
	worker := identity.Fingerprint(keys.Public())
	task, err := a.db.LeaseHash(ctx, cluster, pool, worker, time.Now())
	if err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(task.Prompt))
	body := []byte(hex.EncodeToString(sum[:]) + "\n")
	hash := sha256.Sum256(body)
	digest := hex.EncodeToString(hash[:])
	err = os.MkdirAll(a.artifacts, 0700)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(a.artifacts, ".pool-result-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write(body)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(f.Name(), filepath.Join(a.artifacts, digest))
	}
	if err != nil {
		return err
	}
	stored, err := a.readArtifact(digest)
	if err != nil || !bytes.Equal(stored, body) {
		return errors.New("worker result failed integrity check")
	}
	return a.db.CommitResult(ctx, task, keys, digest, time.Now())
}

func (a *workerAuth) readArtifact(digest string) ([]byte, error) {
	hash, err := hex.DecodeString(digest)
	if err != nil || len(hash) != 32 {
		return nil, errors.New("invalid artifact reference")
	}
	file, err := os.Open(filepath.Join(a.artifacts, digest))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, 66))
	actual := sha256.Sum256(body)
	if err != nil || len(body) != 65 || !bytes.Equal(actual[:], hash) {
		return nil, errors.New("invalid artifact")
	}
	return body, nil
}
