// Pages registered checkouts by insertion order without inspecting their trees.
// The continuation ID resolves to a registration row only on its owning database.
package persistence

import (
	"context"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (db *DB) RegistryPage(ctx context.Context, before string, worker bool, user string, admin bool) ([]domain.Project, error) {
	query := "SELECT p.id,p.name,p.location,p.account,p.owner FROM project_registry p WHERE 1=1"
	var args []any
	if before != "" {
		query += " AND p.rowid<(SELECT rowid FROM project_registry WHERE id=?)"
		args = append(args, before)
	}
	if worker {
		query += " AND p.id LIKE 'remote-%'"
		if !admin {
			query += " AND EXISTS(SELECT 1 FROM workspace_grants g WHERE g.project=p.id AND g.user=? AND g.scope='read')"
			args = append(args, user)
		}
	}
	rows, err := db.sql.QueryContext(ctx, query+" ORDER BY p.rowid DESC LIMIT 33", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Project{}
	for rows.Next() {
		if len(out) == 32 {
			out[31].NextCursor = out[31].ID
			break
		}
		var p domain.Project
		err = rows.Scan(&p.ID, &p.Name, &p.Location, &p.Account, &p.Owner)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
