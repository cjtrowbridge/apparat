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
	ProjectID, Action, Path, Base, Text, Argument, ID, Thread, RelatedRun string
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
	Changes                                  []WorkspaceChange
	Choices                                  []WorkspaceChoice
	Entries                                  []WorkspaceEntry
	Text, Base, Report, ComparedBase, Detail string
	Editing                                  bool
}

const WorkspaceTextLimit = 64 * 1024
