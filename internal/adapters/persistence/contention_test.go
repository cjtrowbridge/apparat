// Exercises real SQLite lock contention and cancelled transaction admission.
// Busy errors remain distinct from replay uniqueness and storage corruption.
package persistence

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBusyTransactionIsBounded(t *testing.T) {
	db := serviceDB(t)
	ctx := context.Background()
	var sequence int
	var name, path string
	{
		err := db.sql.QueryRowContext(ctx, "PRAGMA database_list").Scan(&sequence, &name, &path)
		if err != nil {
			t.Fatal(err)
		}
	}
	other, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = other.Close()
	}()
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	{
		_, err = tx.ExecContext(ctx, "UPDATE services SET name=name")
		if err != nil {
			t.Fatal(err)
		}
	}
	deadline, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err = other.sql.ExecContext(deadline, "UPDATE services SET name=name")
	elapsed := time.Since(started)
	if err == nil || (!errors.Is(err, context.DeadlineExceeded) && !Retryable(err)) {
		t.Fatal("contention misclassified", err)
	}
	if elapsed > 2*time.Second {
		t.Fatal("busy wait exceeded bounded timeout", elapsed)
	}
	t.Log("bounded contention", elapsed, err)
}
