// Documents the internal/domain package boundary.
// Uses package-local contracts and Go standard-library types.
package domain

import (
	_ "embed"
	"strings"
)

//go:embed release.txt
var release string

// ReleaseVersion is the version embedded when this executable was compiled.
func ReleaseVersion() string {
	return strings.TrimSpace(release)
}
