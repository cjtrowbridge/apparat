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
	Expires                         int64
	Task, Path, Digest, State, MIME string
	Definition, Revision            string
	Size                            int
}

func (n *Runtime) projectTask(ctx context.Context, location string, q domain.WorkspaceRequest) (v domain.WorkspaceView, err error) {
	if q.Action == "runs" {
		rows, err := n.database.RecentEntries(ctx, q.ProjectID, "run")
		v.Report = "No local Task runs"
		if len(rows) > 0 {
			v.Report = "Recent runs"
		}
		for _, row := range rows {
			var run artifactRun
			if json.Unmarshal([]byte(row.Label), &run) == nil {
				label := run.Task + " · " + run.State
				if run.State == "completed" {
					v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: row.ID, Label: label})
				} else {
					v.Report += "\n" + label
				}
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
	config, err := n.workspaceFile(location, read)
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
			v.Choices = append(v.Choices, domain.WorkspaceChoice{ID: task.ID, Label: task.Label + " · " + task.ID})
			lines = append(lines, task.ID+" · "+task.Label+" · snapshot "+task.Path)
			if task.ID == q.Argument {
				selected = task
			}
		} else {
			lines = append(lines, task.ID+" · disabled or unsupported entrypoint")
		}
	}
	if q.Action == "tasks" {
		v.Report = strings.Join(lines, "\n")
		if v.Report == "" {
			v.Report = "No configured entrypoints"
		}
		return v, nil
	}
	if selected.ID == "" {
		return v, errors.New("choose an enabled snapshot-text Task by ID")
	}
	read.Path = selected.Path

	run := artifactRun{Task: selected.ID, Path: selected.Path, State: "running", Definition: projectgit.Digest(config.Text)}
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
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	file, err := n.workspaceFile(location, read)
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
	data, _ := json.Marshal(run)
	recordErr := n.database.UpdateRun(context.WithoutCancel(ctx), q.ProjectID, q.ID, string(data))
	if recordErr != nil {
		return v, recordErr
	}
	if err != nil {
		return v, errors.New("snapshot failed; see run history")
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
	err = os.Link(f.Name(), target)
	if errors.Is(err, os.ErrExist) {
		existing, err := os.ReadFile(target)
		if err != nil || string(existing) != body {
			return errors.New("existing artifact failed integrity check")
		}
		return nil
	}
	return err
}
