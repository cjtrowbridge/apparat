# Local identity dialogs

- `desktop.go`: Selects the native dialog helper, passing requests and responses over private pipes.
- `linux.py`: Uses GTK 3 file and password dialogs and a user Secret Service collection when available.
- `windows.ps1`: Uses WinForms, masked password entry and user-scoped DPAPI.
- `macos.js`: Uses Standard Additions for local macOS choices, password entry and file selection.
- `mac_stub.go`: Leaves Apple credential operations unavailable outside a native macOS build.

Cancellation leaves the node locked. Restore and backup operations use a bounded Apparat private-key file. Test on each native platform by cancelling setup, generating an identity in an empty node, restarting and restoring its encrypted backup to a second node.
- `mac_security.go`: Stores a device-local macOS Keychain credential with user presence through Security.framework.
