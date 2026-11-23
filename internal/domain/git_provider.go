// Worker provider discovery exposes capabilities and account labels to administrators.
// Operation progress contains only display-safe fields and host-owned project metadata.
package domain

type GitProvider struct {
	ID, Name, Tool, Reason           string
	Installed, CanInstall, CanEnroll bool
	Accounts                         []string
}
type GitRequest struct{ Provider, Operation, ID, Repo, Account string }
type GitStatus struct {
	State, Message, Code, Username string
	Project                        *Project
}
