// Defines the owner-local checkout registration and current checkout observation.
// Credentials and working tree contents remain outside the registration.
package domain

type Project struct {
	ID, Owner, Name, Location, Account string
	RemoteHost                         string
	Branch, Revision                   string
	Threads                            []WorkspaceChoice
	Entrypoints                        []string
	Actions                            []string
	Available                          bool
}

// LocalProjectSetup describes one checkout operation on the executing node.
type LocalProjectSetup struct {
	Mode, Parent, Name, Account string
}

// WorkspaceRequest addresses one registration and repository-relative path.
type WorkspaceRequest struct {
	Known                                                                 string
	ProjectID, Action, Path, Base, Text, Argument, ID, Thread, RelatedRun string
	Correlation                                                           string
	Approved                                                              bool
}
type WorkspaceEntry struct {
	Name      string
	Directory bool
}
type WorkspaceChoice struct{ ID, Label, Detail, State string }
type WorkspaceChange struct {
	Marker                     string
	Path                       string
	Staged, Unstaged, Conflict bool
	Directory                  bool `json:",omitempty"`
}
type WorkspaceView struct {
	Unchanged                                bool
	Changes                                  []WorkspaceChange
	Choices                                  []WorkspaceChoice
	Entries                                  []WorkspaceEntry
	Text, Base, Report, ComparedBase, Detail string
	Cursor                                   string
	Editing                                  bool
}

const WorkspaceTextLimit = 64 * 1024
