// Records the package dependency boundary.
// Dependencies: modernc.org/sqlite.
package persistence

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var _ = sql.ErrNoRows
