// Executes versioned owner-local text snapshot tasks declared by a project.
// The action reads a scoped file, stores immutable bytes, and records its digest.
package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/adapters/projectgit"
	"github.com/cjtrowbridge/apparat/internal/domain"
)

type artifactRun struct {
	Expires, Started                int64
	Task, Path, Digest, State, MIME string
	Definition, Revision, Reason    string
	Size                            int
}

func (n *Runtime) projectTask(ctx context.Context, location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	if q.Action == "runs" {
		rows, err := n.database.RecentEntries(ctx, q.ProjectID, "run")
		v.Report = "No local Task runs"
		if len(rows) > 0 {
			v.Report = ""
		}
		for _, row := range rows {
			var run artifactRun
			if json.Unmarshal([]byte(row.Label), &run) == nil {
				label := run.Task + " · " + run.State
				if run.Reason != "" {
					label += " · " + run.Reason
				}
				if run.Started > 0 {
					label += " · " + time.Unix(run.Started, 0).Local().Format("Jan 2 15:04")
				}
				v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: row.ID, Label: label, State: run.State, Detail: run.Path + " " + run.Reason})
			}
		}
		return v, err
	}
	if q.Action == "artifact" || q.Action == "run" || q.Action == "delete-artifact" {
		if !domain.ValidID(q.ID) {
			return v, errors.New("choose a valid run ID")
		}
		kind, data, lookup := n.database.ProjectEntry(ctx, q.ProjectID, q.ID)
		if lookup == nil && kind == "run" {
			var run artifactRun
			err = json.Unmarshal([]byte(data), &run)
			if err != nil {
				return v, err
			}
			if q.Action == "run" && run.Task != q.Argument {
				return v, errors.New("run ID belongs to a different task")
			}
			if run.State != "completed" {
				return v, errors.New("run state: " + run.State)
			}
			if len(run.Digest) != 64 || strings.ContainsAny(run.Digest, "/\\") {
				return v, errors.New("invalid artifact metadata")
			}
			if q.Action == "delete-artifact" || run.Expires > 0 && time.Now().Unix() >= run.Expires {
				run.State = "expired"
				if q.Action == "delete-artifact" {
					run.State = "deleted"
				}
				data, _ := json.Marshal(run)
				err = n.database.UpdateRun(ctx, q.ProjectID, q.ID, string(data))
				if err != nil {
					return v, err
				}
				used, err := n.database.ArtifactInUse(ctx, run.Digest)
				if err != nil {
					return v, err
				}
				if !used {
					err = os.Remove(filepath.Join(n.config.ArtifactsDir, run.Digest))
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return v, err
					}
				}
				v.Report = "Artifact " + run.State
				return v, nil
			}
			f, err := os.Open(filepath.Join(n.config.ArtifactsDir, run.Digest))
			if err != nil {
				return v, errors.New("artifact unavailable")
			}
			defer func() {
				_ = f.Close()
			}()
			b, err := io.ReadAll(io.LimitReader(f, domain.WorkspaceTextLimit+1))
			if err != nil || projectgit.Digest(string(b)) != run.Digest {
				return v, errors.New("artifact integrity check failed")
			}
			v.Detail = fmt.Sprintf("%s · %d bytes\nSHA-256 %s\nCheckout %s", run.MIME, run.Size, run.Digest, run.Revision)
			v.Report = "Verified " + run.Path + "\n" + string(b)
			return v, nil
		}
		if lookup != nil && !errors.Is(lookup, sql.ErrNoRows) {
			return v, lookup
		}
		if lookup == nil {
			return v, errors.New("ID already used")
		}
		if q.Action != "run" {
			return v, errors.New("no artifact for that run in this project")
		}
	}
	read := domain.WorkspaceRequest{Action: "read", Path: ".apparat/tasks.json"}
	config, err := n.workspaceFile(ctx, location, read)
	if err != nil {
		return v, errors.New("no readable .apparat/tasks.json; VS Code command labels alone cannot run here")
	}
	tasks, err := projectgit.LocalTasks(config.Text)
	if err != nil {
		return v, err
	}
	var selected projectgit.LocalTask
	var lines []string
	for _, task := range tasks {
		if task.Runnable() {
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: task.ID, Label: task.Label + " · " + task.ID, Detail: task.Action + " " + task.Path + " → local text artifact"})

			if task.ID == q.Argument {
				selected = task
			}
		} else {
			lines = append(lines, task.ID+" · "+task.UnavailableReason())
		}
	}
	if q.Action == "tasks" {
		v.Report = strings.Join(lines, "\n")
		if v.Report == "" && len(v.Choices) == 0 {
			v.Report = "No configured entrypoints"
		}
		return v, nil
	}
	if !q.Approved {
		return v, errors.New("confirm the local Task to run")
	}
	if selected.ID == "" {
		return v, errors.New("choose an enabled local task")
	}
	read.Path = selected.Path

	run := artifactRun{Started: time.Now().Unix(), Task: selected.ID, Path: selected.Path, State: "running", Definition: projectgit.Digest(config.Text)}
	observed, observeErr := n.inspectProject(location)
	if observeErr == nil {
		run.Revision = observed.Revision
	}
	if selected.RetainSeconds > 0 {
		run.Expires = time.Now().Unix() + selected.RetainSeconds
	}
	initial, _ := json.Marshal(run)
	err = n.database.AddProjectEntry(ctx, q.ProjectID, q.ID, "run", string(initial))
	if err != nil {
		return v, err
	}
	timeout := selected.TimeoutSeconds
	if timeout == 0 {
		timeout = 15
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	var file domain.WorkspaceView
	if selected.Action == "git-status" {
		read.Path = "."
		read.Action = "status"
		if strings.HasPrefix(location, "content://") {
			file, err = n.workspaceFile(ctx, location, read)
			file.Text = file.Report
		} else {
			file.Text, err = projectgit.GitAction(location, read)
		}
		run.Path = "Git status"
		if !projectgit.ValidText(file.Text) {
			err = errors.New("status exceeds artifact limits")
		}
	} else {
		file, err = n.workspaceFile(ctx, location, read)
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		run.Digest = projectgit.Digest(file.Text)
		run.Size = len(file.Text)
		run.MIME = "text/plain; charset=utf-8"
		err = publishArtifact(n.config.ArtifactsDir, run.Digest, file.Text)
	}
	if err == nil {
		run.State = "completed"
	} else if ctx.Err() != nil {
		run.State = "cancelled"
	} else {
		run.State = "failed"
	}
	if err != nil {
		run.Reason = err.Error()
	}
	data, _ := json.Marshal(run)
	recordErr := n.database.UpdateRun(context.WithoutCancel(ctx), q.ProjectID, q.ID, string(data))
	if recordErr != nil {
		return v, recordErr
	}
	if err != nil {
		return v, fmt.Errorf("local task %s: %w", run.State, err)
	}
	v.Report = fmt.Sprintf("Completed %s · %d bytes\nRun %s\nSHA-256 %s", selected.ID, run.Size, q.ID, run.Digest)
	return v, nil
}
func publishArtifact(dir, digest, body string) error {
	f, err := os.CreateTemp(dir, ".snapshot-")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	_, err = f.WriteString(body)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	target := filepath.Join(dir, digest)
	existing, err := os.Open(target)
	if err == nil {
		defer func() {
			_ = existing.Close()
		}()
		b, err := io.ReadAll(io.LimitReader(existing, domain.WorkspaceTextLimit+1))
		if err != nil || string(b) != body {
			return errors.New("existing artifact failed integrity check")
		}
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// The runtime owns this private directory and serializes workspace writes.
	return os.Rename(f.Name(), target)
}
