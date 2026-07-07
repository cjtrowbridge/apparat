package persistence

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var _ = sql.ErrNoRows
