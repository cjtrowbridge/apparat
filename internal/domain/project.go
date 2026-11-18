// Defines the owner-local checkout registration and current checkout observation.
// Credentials and working tree contents remain outside the registration.
package domain

type Project struct {
	ID, Owner, Name, Location, Account string
	Branch, Revision                   string
	Entrypoints                        []string
	Actions                            []string
	Available                          bool
}

// WorkspaceRequest addresses one registration and repository-relative path.
type WorkspaceRequest struct {
	ProjectID, Action, Path, Base, Text, Argument, ID string
}
type WorkspaceEntry struct {
	Name      string
	Directory bool
}
type WorkspaceView struct {
	Entries            []WorkspaceEntry
	Text, Base, Report string
	Editing            bool
}

const WorkspaceTextLimit = 64 * 1024
