// Translates bounded worker denial codes into recovery instructions for workspace operations.
// Unknown or legacy responses never contribute remote text to a displayed command.
package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"github.com/cjtrowbridge/apparat/internal/adapters/persistence"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type WorkspaceGrantNeeded struct{ Scope string }

func (e WorkspaceGrantNeeded) Error() string {
	return "No matching checkout grant for " + e.Scope
}
func (e WorkspaceGrantNeeded) Unwrap() error {
	return ErrDenied
}

type accessExplanation string

func (e accessExplanation) Error() string {
	return string(e)
}
func (e accessExplanation) Unwrap() error {
	return ErrDenied
}

var ErrWorkspaceLocked error = accessExplanation("Unlock your personal and device identities on this device, then reopen the project.")
var commandID = regexp.MustCompile("^[A-Za-z0-9_-]{1,128}$")

func workspaceDenial(status int, body io.Reader, q domain.WorkspaceRequest, access persistence.Access) error {
	var reason struct{ Error, Scope string }
	data, err := io.ReadAll(io.LimitReader(body, 1025))
	if err == nil && len(data) <= 1024 {
		_ = json.Unmarshal(data, &reason)
	}
	action := map[string]string{
		"list":             "list this folder",
		"directory-status": "read this folder’s Git status",
		"read":             "open this file",
		"save":             "save this file",
		"tasks":            "list project tasks",
		"run":              "run this task",
		"runs":             "read task history",
		"artifact":         "read the task result",
		"terminal-list":    "list terminal sessions",
		"terminal-history": "read terminal history",
		"terminal-screen":  "read the terminal screen",
		"terminal-start":   "start a terminal",
		"terminal-input":   "send terminal input",
		"terminal-stop":    "stop the terminal",
		"terminal-take":    "take terminal control",
		"terminal-resize":  "resize the terminal",
	}[q.Action]
	if action == "" {
		action = "access this workspace"
	}
	advice := "Check that the worker approves this user and device. Ask its operator to verify the checkout permissions, then retry."
	switch {
	case access.Admission != "" && access.Admission != "approved":
		advice = "Ask a cluster admin to approve your account, then reopen this project."
	case status == 401:
		advice = "The worker requires a fresh authenticated session. Unlock both identities and reconnect to the worker, then retry."
	case reason.Error == "owner_confirmation_required":
		advice = "The worker operator must confirm the cluster owner with apparatd trust-host on the worker, then retry."
	case reason.Error == "membership_required":
		advice = "Ask a cluster admin to check approval for this user and device on the worker, then retry."
	case reason.Error == "workspace_grant_required":
		switch reason.Scope {
		case "read", "write", "tasks", "terminal-view", "terminal-control":
			project := q.ProjectID
			if !commandID.MatchString(project) {
				project = "PROJECT_ID"
			}
			user := access.User
			if !commandID.MatchString(user) {
				user = "USER_FINGERPRINT"
			}
			advice = fmt.Sprintf("No matching %s permission for this account and checkout. Ask the worker operator to verify the registration and, if access is intended, run this on the worker:\napparatd workspace-access grant %s %s %s\nThen retry.", reason.Scope, project, user, reason.Scope)
		}
	}
	return accessExplanation("Cannot " + action + ". " + advice)
}
