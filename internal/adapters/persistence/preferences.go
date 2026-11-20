// Stores node-local display preferences separately from desktop placement.
// Each upsert atomically replaces one complete snapshot.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const preferencesSchema = `
CREATE TABLE preferences(id INTEGER PRIMARY KEY CHECK(id=1), theme TEXT NOT NULL,font INTEGER NOT NULL,scale INTEGER NOT NULL,landing TEXT NOT NULL,overlay INTEGER NOT NULL);
CREATE TABLE desktop_window(id INTEGER PRIMARY KEY CHECK(id=1),width INTEGER NOT NULL,height INTEGER NOT NULL,x INTEGER NOT NULL,y INTEGER NOT NULL,maximized INTEGER NOT NULL,monitor TEXT NOT NULL);`

func (db *DB) Preferences(ctx context.Context) (domain.Preferences, error) {
	p := domain.DefaultPreferences()
	err := db.sql.QueryRowContext(ctx, "SELECT theme,font,scale,landing,overlay,placement,spoken_choices FROM preferences WHERE id=1").Scan(&p.Theme, &p.Font, &p.Scale, &p.Landing, &p.Overlay, &p.Placement, &p.SpokenChoices)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultPreferences(), nil
	}
	if err == nil && !p.Valid() {
		return domain.DefaultPreferences(), errors.New("invalid stored preferences")
	}
	return p, err
}
func (db *DB) SavePreferences(ctx context.Context, p domain.Preferences) error {
	if !p.Valid() {
		return errors.New("invalid preferences")
	}
	_, err := db.sql.ExecContext(ctx, "INSERT OR REPLACE INTO preferences VALUES(1,?,?,?,?,?,?,?)", p.Theme, p.Font, p.Scale, p.Landing, p.Overlay, p.Placement, p.SpokenChoices)
	return err
}
func (db *DB) WindowState(ctx context.Context) (domain.WindowState, error) {
	var w domain.WindowState
	err := db.sql.QueryRowContext(ctx, "SELECT width,height,x,y,maximized,monitor FROM desktop_window WHERE id=1").Scan(&w.Width, &w.Height, &w.X, &w.Y, &w.Maximized, &w.Monitor)
	if errors.Is(err, sql.ErrNoRows) {
		return w, nil
	}
	return w, err
}
func (db *DB) SaveWindowState(ctx context.Context, w domain.WindowState) error {
	if w.Width <= 0 || w.Height <= 0 {
		return errors.New("invalid window size")
	}
	_, err := db.sql.ExecContext(ctx, "INSERT OR REPLACE INTO desktop_window VALUES(1,?,?,?,?,?,?,?)", w.Width, w.Height, w.X, w.Y, w.Maximized, w.Monitor)
	return err
}

func (db *DB) SpeechModel(ctx context.Context) (string, error) {
	var id string
	err := db.sql.QueryRowContext(ctx, "SELECT value FROM local_runtime WHERE key='speech_model'").Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}
func (db *DB) SetSpeechModel(ctx context.Context, id string) error {
	_, err := db.sql.ExecContext(ctx, "INSERT INTO local_runtime VALUES('speech_model',?,0) ON CONFLICT(key) DO UPDATE SET value=excluded.value", id)
	return err
}
