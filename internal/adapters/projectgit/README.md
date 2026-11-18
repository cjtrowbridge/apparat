# Local Git adapter

- `checkout.go`: validates and inspects working trees and shallow-clones an explicit GitHub HTTPS repo. A selected Linux account is resolved through the system gh process. Tokens stay in memory. Public clones need no account. Only explicit labeled tasks from .vscode/tasks.json are displayed. Existing destinations are preserved. Validate with the project registry tests and a real public checkout on the Pixel 4a.

- `install_linux.go`: Atomically publishes a completed temporary clone without replacing an occupied destination on Linux and Android.
- `install_other.go`: Reports an unsupported clone publication operation on unqualified platforms.
- `checkout_test.go`: Covers cancellation, preservation of user files, concurrent clone publication and isolated account lookup.
