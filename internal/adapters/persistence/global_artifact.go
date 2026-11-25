// Keeps private global session replicas ordered by the session writer's revision.
// Publication is retried across restarts until every pinned worker has acknowledged.
package persistence

import (
	"context"
	"encoding/json"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

const globalSchema = `CREATE TABLE global_chats(user TEXT,session TEXT,revision INTEGER NOT NULL,digest TEXT NOT NULL,body BLOB NOT NULL,pending INTEGER NOT NULL,PRIMARY KEY(user,session));`

func (db *DB) GlobalArtifact(ctx context.Context, user, id string) (a domain.GlobalArtifact, err error) {
	var b []byte
	a.User = user
	err = db.sql.QueryRowContext(ctx, "SELECT revision,body FROM global_chats WHERE user=? AND session=?", user, id).Scan(&a.Version, &b)
	if err == nil {
		err = json.Unmarshal(b, &a.Session)
	}
	return
}

func (db *DB) PutGlobalArtifact(ctx context.Context, a domain.GlobalArtifact, pending bool) error {
	b, err := json.Marshal(a.Session)
	if err != nil {
		return err
	}
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	r, err := tx.ExecContext(ctx, `INSERT INTO global_chats(user,session,revision,digest,body,pending) VALUES(?,?,?,?,?,?) ON CONFLICT(user,session) DO UPDATE SET
	revision=excluded.revision,digest=excluded.digest,body=excluded.body,pending=excluded.pending
	WHERE excluded.revision>revision OR (excluded.revision=revision AND excluded.digest=digest)`, a.User, a.Session.ID, a.Version, domain.InferenceDigest(a.Session), b, pending)
	err = inferenceChange(r, err)
	if err != nil {
		return err
	}
	if pending && a.Session.Pending() != "" {
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO chat_work VALUES(?,?,?,?,?)", "global/"+a.User, a.Session.ID, a.Session.Pending(), a.User, "Chat Pool")
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) GlobalHeaders(ctx context.Context, user, cursor string) ([]domain.ChatHeader, error) {
	rows, err := db.sql.QueryContext(ctx, "SELECT json_remove(CAST(body AS TEXT),'$.Turns') FROM global_chats WHERE user=? AND (?='' OR session<?) ORDER BY session DESC LIMIT 32", user, cursor, cursor)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var out []domain.ChatHeader
	for rows.Next() {
		var b []byte
		var h domain.ChatHeader
		err = rows.Scan(&b)
		if err == nil {
			err = json.Unmarshal(b, &h)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (db *DB) PendingGlobals(ctx context.Context) ([]domain.GlobalArtifact, error) {
	return db.globalArtifacts(ctx, "SELECT user,revision,body FROM global_chats WHERE pending=1 ORDER BY session")
}
func (db *DB) globalArtifacts(ctx context.Context, query string) ([]domain.GlobalArtifact, error) {
	rows, err := db.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var out []domain.GlobalArtifact
	for rows.Next() {
		var a domain.GlobalArtifact
		var b []byte
		err = rows.Scan(&a.User, &a.Version, &b)
		if err == nil {
			err = json.Unmarshal(b, &a.Session)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (db *DB) GlobalDelivered(ctx context.Context, a domain.GlobalArtifact) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE global_chats SET pending=0 WHERE user=? AND session=? AND revision=?", a.User, a.Session.ID, a.Version)
	return err
}

// MarkdownGlobals returns artifact revisions still awaiting local filesystem publication.
func (db *DB) MarkdownGlobals(ctx context.Context) ([]domain.GlobalArtifact, error) {
	return db.globalArtifacts(ctx, "SELECT user,revision,body FROM global_chats WHERE markdown_revision<revision")
}
func (db *DB) MarkGlobalMarkdown(ctx context.Context, a domain.GlobalArtifact) error {
	_, err := db.sql.ExecContext(ctx, "UPDATE global_chats SET markdown_revision=? WHERE user=? AND session=? AND revision=?", a.Version, a.User, a.Session.ID, a.Version)
	return err
}

func (db *DB) GlobalChunk(ctx context.Context, user, id string, revision int64, offset int) (p domain.GlobalChunk, err error) {
	p.User = user
	p.Offset = offset
	err = db.sql.QueryRowContext(ctx, "SELECT revision,length(body),substr(body,?,?) FROM global_chats WHERE user=? AND session=? AND (?=0 OR revision=?) AND ?>=0 AND ?<length(body)", offset+1, domain.GlobalChunkBytes, user, id, revision, revision, offset, offset).Scan(&p.Revision, &p.Total, &p.Bytes)
	return
}

func (db *DB) GlobalVersion(ctx context.Context, user, id string) (revision int64, err error) {
	err = db.sql.QueryRowContext(ctx, "SELECT revision FROM global_chats WHERE user=? AND session=?", user, id).Scan(&revision)
	return
}

// Use the locally cached pending turn to follow a session after the GUI outbox was acknowledged.
func (db *DB) PendingPlanningRevisions(ctx context.Context, user string) (map[string]int64, error) {
	rows, err := db.sql.QueryContext(ctx, `SELECT session,revision FROM global_chats WHERE user=? AND pending=0 AND
 json_extract(body,'$.Turns[#-1].Reply')='' AND json_extract(body,'$.Turns[#-1].Stopped')=''`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var id string
		var revision int64
		err = rows.Scan(&id, &revision)
		if err != nil {
			return nil, err
		}
		out[id] = revision
	}
	return out, rows.Err()
}
