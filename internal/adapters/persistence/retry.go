// Classifies SQLite contention without treating cancellation or corruption as retries.
// Callers decide whether their bounded operation may be attempted again.
package persistence

import (
	"errors"
	"modernc.org/sqlite"
)

func Retryable(err error) bool {
	var sqliteError *sqlite.Error
	if !errors.As(err, &sqliteError) {
		return false
	}
	code := sqliteError.Code() & 255
	return code == 5 || code == 6
}
