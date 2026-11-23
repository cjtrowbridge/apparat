// Stores group membership alongside the cluster's other signed policy projections.
// The grant view expands group principals into users for admission, dispatch and sync.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/cjtrowbridge/apparat/internal/identity"
)

const groupSchema = `
CREATE TABLE cluster_groups(cluster TEXT NOT NULL REFERENCES worker_cluster(cluster),name TEXT NOT NULL,PRIMARY KEY(cluster,name));
CREATE TABLE cluster_group_members(cluster TEXT NOT NULL,name TEXT NOT NULL,user_id TEXT NOT NULL,
 PRIMARY KEY(cluster,name,user_id),FOREIGN KEY(cluster,name) REFERENCES cluster_groups(cluster,name),FOREIGN KEY(cluster,user_id) REFERENCES cluster_users(cluster,user_id));
CREATE VIEW effective_pool_grants AS
 SELECT cluster,pool,principal,service FROM pool_grants WHERE principal NOT LIKE 'group:%'
 UNION ALL SELECT g.cluster,g.pool,m.user_id,g.service FROM pool_grants g JOIN cluster_group_members m
 ON g.cluster=m.cluster AND g.principal='group:'||m.name;
`

type GroupMember struct{ Name, User string }

func groupEdit(ctx context.Context, tx *sql.Tx, e identity.PolicyEdit) (sql.Result, error) {
	name := e.Pool
	if name == "" || strings.TrimSpace(name) != name {
		return nil, errors.New("invalid group name")
	}
	if e.Kind == "add_group" {
		return tx.ExecContext(ctx, "INSERT INTO cluster_groups VALUES(?,?)", e.Cluster, name)
	}
	if e.Value == "remove" {
		return tx.ExecContext(ctx, "DELETE FROM cluster_group_members WHERE cluster=? AND name=? AND user_id=?", e.Cluster, name, e.Target)
	}
	if e.Value != "add" {
		return nil, errors.New("invalid membership")
	}
	return tx.ExecContext(ctx, "INSERT INTO cluster_group_members VALUES(?,?,?)", e.Cluster, name, e.Target)
}

func (v *GovernanceView) readGroups(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, "SELECT g.name,coalesce(m.user_id,'') FROM cluster_groups g LEFT JOIN cluster_group_members m USING(cluster,name) WHERE g.cluster=? ORDER BY g.name,m.user_id", v.Cluster)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m GroupMember
		err = rows.Scan(&m.Name, &m.User)
		if err != nil {
			return err
		}
		if len(v.Groups) == 0 || v.Groups[len(v.Groups)-1] != m.Name {
			v.Groups = append(v.Groups, m.Name)
		}
		if m.User != "" {
			v.Memberships = append(v.Memberships, m)
		}
	}
	return rows.Err()
}
