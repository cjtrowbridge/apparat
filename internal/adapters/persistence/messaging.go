// Persists message envelopes and replay state.
// Uses package-local contracts and Go standard-library types.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"modernc.org/sqlite"
	"time"
)

type Message struct {
	ID            string
	Direction     string
	Payload       string
	Attempts      int
	NextAttemptMs int64
}
type Store struct {
	db          *sql.DB
	maxAttempts int
}

func NewMessages(db *sql.DB, maxAttempts int) Store {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	return Store{db: db, maxAttempts: maxAttempts}
}

func (store Store) Init(ctx context.Context) error {
	return (&DB{sql: store.db}).Migrate(ctx)
}

func (store Store) EnqueueOutbox(ctx context.Context, id, payload string) error {
	return store.insert(ctx, Message{ID: id, Direction: "outbox", Payload: payload})
}
func (store Store) RecordInbox(ctx context.Context, id, payload string) error {
	return store.insert(ctx, Message{ID: id, Direction: "inbox", Payload: payload})
}

func (store Store) insert(ctx context.Context, message Message) error {
	_, err := store.db.ExecContext(ctx, `INSERT INTO messages(id, direction, payload, attempts, next_attempt_ms) VALUES (?, ?, ?, ?, ?)`, message.ID, message.Direction, message.Payload, message.Attempts, message.NextAttemptMs)
	return err
}

func (store Store) Seen(ctx context.Context, id string) (bool, error) {
	_, err := store.db.ExecContext(ctx, `INSERT INTO replay_seen(id, seen_at_ms) VALUES (?, ?)`, id, time.Now().UTC().UnixMilli())
	if err == nil {
		return false, nil
	}
	var sqliteError *sqlite.Error
	if errors.As(err, &sqliteError) && sqliteError.Code() == 1555 {
		return true, nil
	}
	return false, err
}

func (store Store) ScheduleRetry(ctx context.Context, id string, nowMs int64) (Message, error) {
	msg, err := store.Get(ctx, id)
	if err != nil {
		return msg, err
	}
	if msg.Attempts >= store.maxAttempts {
		return msg, errors.New("retry attempts exhausted")
	}
	msg.Attempts++
	msg.NextAttemptMs = nowMs + int64(msg.Attempts*msg.Attempts*1000)
	_, err = store.db.ExecContext(ctx, `UPDATE messages SET attempts = ?, next_attempt_ms = ? WHERE id = ?`, msg.Attempts, msg.NextAttemptMs, id)
	return msg, err
}

func (store Store) Get(ctx context.Context, id string) (Message, error) {
	var msg Message
	err := store.db.QueryRowContext(ctx, `SELECT id, direction, payload, attempts, next_attempt_ms FROM messages WHERE id = ?`, id).Scan(&msg.ID, &msg.Direction, &msg.Payload, &msg.Attempts, &msg.NextAttemptMs)
	return msg, err
}

func (store Store) SetCursor(ctx context.Context, name, cursor string) error {
	_, err := store.db.ExecContext(ctx, `INSERT INTO event_cursors(name, cursor) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET cursor=excluded.cursor`, name, cursor)
	return err
}
func (store Store) Cursor(ctx context.Context, name string) (string, error) {
	var cursor string
	err := store.db.QueryRowContext(ctx, `SELECT cursor FROM event_cursors WHERE name = ?`, name).Scan(&cursor)
	return cursor, err
}
