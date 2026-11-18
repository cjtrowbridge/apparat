// Parses the owner-defined versioned Task catalog without evaluating repository code.
// Only implemented action kinds and supported scoped inputs become runnable.
package projectgit

import (
	"encoding/json"
	"errors"
	"strings"
)

type LocalTask struct {
	ID, Label, Action, Path string
	Enabled                 bool
	SecretRefs              []string
	TimeoutSeconds          int
	RetainSeconds           int64
}

func (t LocalTask) Runnable() bool {
	return t.UnavailableReason() == ""
}
func (t LocalTask) UnavailableReason() string {
	switch {
	case !t.Enabled:
		return "disabled in the project configuration"
	case len(t.SecretRefs) > 0:
		return "this executor cannot supply secret references"
	case t.TimeoutSeconds < 0 || t.TimeoutSeconds > 60:
		return "timeout must be between 0 and 60 seconds"
	case t.RetainSeconds < 0 || t.RetainSeconds > 31536000:
		return "retention must be between 0 and 31536000 seconds"
	}
	switch t.Action {
	case "git-status":
		if t.Path != "" && t.Path != "." {
			return "git-status requires the checkout root"
		}
	case "snapshot-text":
		if !WorkspacePath(t.Path) || t.Path == "." {
			return "snapshot-text requires a project-relative file path"
		}
	default:
		return "action is unavailable on the local executor"
	}
	return ""
}
func LocalTasks(raw string) ([]LocalTask, error) {
	var spec struct {
		Version int
		Tasks   []LocalTask
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&spec)
	if err != nil || !json.Valid([]byte(raw)) || spec.Version != 1 {
		return nil, errors.New("unsupported Task configuration; expected version 1")
	}
	ids := map[string]bool{}
	for _, t := range spec.Tasks {
		if t.ID == "" || ids[t.ID] {
			return nil, errors.New("task IDs must be nonempty and unique")
		}
		ids[t.ID] = true
	}
	return spec.Tasks, nil
}
func EnabledTasks(raw string) []string {
	tasks, _ := LocalTasks(raw)
	var enabled []string
	for _, t := range tasks {
		if t.Runnable() {
			enabled = append(enabled, t.ID)
		}
	}
	return enabled
}
