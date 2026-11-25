# Owner terminal sessions

Linux hosts run an isolated PTY with a read-only project, private temporary storage and no network. A worker requires explicit per-project terminal grants. One device controls each running session. Private history retains the last screen and up to 256 output chunks of 32 KiB each per session. An ended process can be inspected after restart.

- `pty_linux.go`: Starts the interactive shell inside a private Linux namespace with a read-only checkout.

- `pty_other.go`: Offers remote terminal clients on targets without qualified local PTY isolation.

- `terminal.go`: Owns restricted interactive processes and their independent controller leases.

- `terminal_linux_test.go`: Exercises the isolated PTY through the same session actions used by the GUI.

- `history.go`: Persists terminal bytes in numbered chunks under the session's private directory.

Chat-scoped workspace actions use the `chat-terminal-` prefix and require a canonical chat session ID in `Thread`. The host stores these sessions under the user and project terminal directory in `chats/<thread>/<terminal>`. The original project terminal actions retain their existing directories. Each scope has independent paging, live control and retained screens.

- `chat_policy.go`: Stores the chat's private permission setting and binds a tracked command to its exact bytes and deadline. The job ID names a separate terminal under the chat. An interrupted process with no saved exit status has an unknown outcome.
