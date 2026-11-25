//go:build linux && !android

// Starts the interactive shell inside a private Linux namespace with a read-only checkout.
// Host home directories, service credentials and network interfaces are outside the namespace.
package terminal

import (
	"github.com/creack/pty"
	"os"
	"os/exec"
)

func prerequisite() string {
	for _, p := range []string{"/usr/bin/bwrap", "/usr/bin/prlimit", "/usr/bin/systemd-run"} {
		_, err := os.Stat(p)
		if err != nil {
			return "Linux terminal hosting requires Bubblewrap, a user systemd session and prlimit."
		}
	}
	return ""
}
func start(location, command string) (*os.File, *exec.Cmd, error) {
	args := []string{"--user", "--scope", "--quiet", "--property=TasksMax=64", "--property=MemoryMax=512M", "--", "/usr/bin/prlimit", "--as=536870912", "--nofile=128", "--fsize=16777216", "--", "/usr/bin/bwrap", "--die-with-parent", "--unshare-all", "--unshare-user", "--disable-userns", "--ro-bind", "/usr", "/usr", "--symlink", "usr/bin", "/bin", "--symlink", "usr/lib", "/lib", "--symlink", "usr/lib64", "/lib64", "--proc", "/proc", "--dev", "/dev", "--size", "16777216", "--tmpfs", "/tmp", "--ro-bind", location, "/project", "--chdir", "/project", "--clearenv", "--setenv", "HOME", "/tmp", "--setenv", "TERM", "xterm", "--setenv", "PATH", "/usr/bin:/bin", "--", "/bin/sh"}
	if command == "" {
		args = append(args, "-i")
	} else {
		args = append(args, "-c", command)
	}
	cmd := exec.Command("/usr/bin/systemd-run", args...)
	cmd.Env = []string{"PATH=/usr/bin:/bin", "XDG_RUNTIME_DIR=" + os.Getenv("XDG_RUNTIME_DIR"), "DBUS_SESSION_BUS_ADDRESS=" + os.Getenv("DBUS_SESSION_BUS_ADDRESS")}
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 24, Cols: 80})
	return f, cmd, err
}
func setSize(f *os.File, cols, rows int) error {
	return pty.Setsize(f, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}
