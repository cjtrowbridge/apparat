# Owner terminal sessions

Linux hosts run an isolated PTY with a read-only project, private temporary storage and no network. A worker requires explicit per-project terminal grants. One device controls each running session. Private screen history retains at most 256 checkpoints per session. An ended process can be inspected after restart.

- `pty_linux.go`: Starts the interactive shell inside a private Linux namespace with a read-only checkout.

- `pty_other.go`: Offers remote terminal clients on targets without qualified local PTY isolation.

- `terminal.go`: Owns restricted interactive processes and their independent controller leases.

- `terminal_linux_test.go`: Exercises the isolated PTY through the same session actions used by the GUI.

- `history.go`: Persists terminal bytes in numbered chunks under the session's private directory.
