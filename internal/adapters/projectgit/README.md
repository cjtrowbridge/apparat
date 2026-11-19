# Local Git adapter

- `checkout.go`: validates and inspects working trees and shallow-clones an explicit GitHub HTTPS repo. A selected Linux account is resolved through the system gh process. Tokens stay in memory. Public clones need no account. Only explicit labeled tasks from .vscode/tasks.json are displayed. Existing destinations are preserved. Validate with the project registry tests and a real public checkout on the Pixel 4a.

- `install_linux.go`: Atomically publishes a completed temporary clone without replacing an occupied destination on Linux and Android.
- `install_other.go`: Reports an unsupported clone publication operation on unqualified platforms.
- `checkout_test.go`: Covers cancellation, preservation of user files, concurrent clone publication and isolated account lookup.

workspace_files.go bounds UTF-8 files to 64 KiB and compares draft base digests before replacement. workspace_git.go uses go-git for status, diff, staging, commit, branches and history. Commits refuse unrelated staged files.

- `workspace_git.go`: Implements local Git actions without launching repository commands or hooks.

- `workspace_files.go`: Exposes bounded working-tree text and directory operations.

- `task_config.go`: Validates version 1 project entrypoints and identifies enabled supported Tasks.

- `workspace_paths.go`: Pins workspace directories and checks their identity before opening children.

- `workspace_index.go`: Publishes a locked index after storing the selected blob, scopes unstaging to one path and exposes Git merge stages for explicit resolution.

- `workspace_refs.go`: Locks and verifies a branch reference before replacing it, including on Android document providers without flock support.

- `workspace_git_test.go`: Simulates failure to publish a Git lockfile and verifies that the previous branch tip and user data survive.

- `workspace_changes.go`: Returns changed paths with stage state and the working-file digest reviewed for a scoped Git operation.

- `workspace_directory.go`: Compares immediate children against the index and the corresponding HEAD tree, applies ancestor ignore files, and cancels file hashing when browsing moves elsewhere. Changes continues to use repository-wide Git status.
- `create.go`: Initializes a named, empty local working tree on main and publishes it without replacing an existing folder.
