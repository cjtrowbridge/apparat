// Submits the bounded hash service and reads only the logged-in user's pool history.
// Result bytes are verified against the worker's content digest before display.
package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
)

func (c *WorkerClient) QueueChanges(ctx context.Context, pool string, after int64) ([]persistence.QueueChange, error) {
	var changes []persistence.QueueChange
	err := c.call(ctx, "POST", "/v1/tasks/changes", TaskRequest{Pool: pool, After: after}, &changes)
	return changes, err
}
func (c *WorkerClient) SubmitHash(ctx context.Context, pool, id, prompt string) error {
	var response map[string]string
	return c.call(ctx, "POST", "/v1/tasks/submit", TaskRequest{Pool: pool, ID: id, Prompt: prompt}, &response)
}
func (c *WorkerClient) Cancel(ctx context.Context, pool, id string) error {
	var response map[string]string
	return c.call(ctx, "POST", "/v1/tasks/cancel", TaskRequest{Pool: pool, ID: id}, &response)
}
func (c *WorkerClient) Result(ctx context.Context, pool, id string) (string, error) {
	var artifact struct {
		Text, SHA256 string
		Receipt      persistence.ResultReceipt
	}
	err := c.call(ctx, "POST", "/v1/tasks/result", TaskRequest{Pool: pool, ID: id}, &artifact)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(artifact.Text))
	if hex.EncodeToString(sum[:]) != artifact.SHA256 {
		return "", errors.New("artifact verification failed")
	}
	r := artifact.Receipt
	if !r.Valid(c.Worker) || r.Cluster != c.Access.Cluster || r.Creator != c.Access.User || r.Pool != pool || r.Task != id || r.Digest != artifact.SHA256 {
		return "", errors.New("signed result could not be verified")
	}
	return artifact.Text, nil
}
