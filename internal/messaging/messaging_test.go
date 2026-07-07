package messaging

import (
	"context"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/database"
)

func TestMessagingPersistenceReplayAndRetry(t *testing.T) {
	db, err := database.Open(context.Background(), t.TempDir()+"/apparat.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	store := New(db.SQL, 2)
	if err := store.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := store.EnqueueOutbox(context.Background(), "m1", "payload"); err != nil {
		t.Fatal(err)
	}
	msg, err := store.Get(context.Background(), "m1")
	if err != nil {
		t.Fatal(err)
	}
	if msg.Direction != "outbox" {
		t.Fatalf("message = %+v", msg)
	}
	seen, err := store.Seen(context.Background(), "m1")
	if err != nil || seen {
		t.Fatalf("first seen = %v %v", seen, err)
	}
	seen, err = store.Seen(context.Background(), "m1")
	if err != nil || !seen {
		t.Fatalf("second seen = %v %v", seen, err)
	}
	msg, err = store.ScheduleRetry(context.Background(), "m1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Attempts != 1 || msg.NextAttemptMs <= 1000 {
		t.Fatalf("retry = %+v", msg)
	}
	if err := store.SetCursor(context.Background(), "events", "42"); err != nil {
		t.Fatal(err)
	}
	cursor, err := store.Cursor(context.Background(), "events")
	if err != nil || cursor != "42" {
		t.Fatalf("cursor = %q %v", cursor, err)
	}
}
