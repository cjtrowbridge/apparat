// Embeds the Linux platform integration in each shipped executable.
// Helpers use system Python and the current user's network services.
package discovery

import (
	"context"
	_ "embed"
	"os/exec"
)

//go:embed linux.py
var source string

func Command(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, "/usr/bin/python3", append([]string{"-c", "import sys;SOURCE=sys.argv.pop(1);exec(SOURCE)", source}, args...)...)
}
