// Binds the hash result to its creator, pool, task and fenced worker attempt.
// The signed receipt is committed alongside the result slot and terminal state.
package persistence

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/identity"
	"time"
)

type ResultReceipt struct {
	Domain, Cluster, Pool, Task, Creator, Digest string
	Fence                                        int64
	Worker                                       ed25519.PublicKey
	Signature                                    []byte
}

func (r ResultReceipt) message() []byte {
	r.Signature = nil
	body, _ := json.Marshal(r)
	return body
}
func (r ResultReceipt) Valid(worker string) bool {
	return r.Domain == "apparat/hash-result/v1" && r.Fence > 0 && len(r.Worker) == ed25519.PublicKeySize && identity.Fingerprint(r.Worker) == worker && ed25519.Verify(r.Worker, r.message(), r.Signature)
}
func (db *DB) CommitResult(ctx context.Context, t Task, keys *identity.Session, digest string, now time.Time) error {
	r := ResultReceipt{Domain: "apparat/hash-result/v1", Cluster: t.Cluster, Pool: t.Pool, Task: t.ID, Creator: t.Creator, Fence: t.Fence, Digest: digest, Worker: keys.Public()}
	sig, err := keys.Sign(r.message())
	if err != nil {
		return err
	}
	r.Signature = sig
	worker := identity.Fingerprint(keys.Public())
	if !r.Valid(worker) {
		return errors.New("invalid worker result signature")
	}
	return db.completeHash(ctx, t, worker, digest, now, &r)
}
func (db *DB) ResultReceipt(ctx context.Context, a Access, pool, id, digest string) (ResultReceipt, error) {
	var raw []byte
	var r ResultReceipt
	err := db.sql.QueryRowContext(ctx, `SELECT r.receipt FROM pool_task_artifacts r JOIN pool_tasks t ON r.cluster=t.cluster AND r.pool=t.pool AND r.task=t.id WHERE t.cluster=? AND t.pool=? AND t.id=? AND t.creator=? AND r.slot='result' AND r.digest=?`, a.Cluster, pool, id, a.User, digest).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return r, errors.New("signed result unavailable")
	}
	err = json.Unmarshal(raw, &r)
	return r, err
}
