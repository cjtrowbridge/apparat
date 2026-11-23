// Reads bounded VS Code JSONC catalogs on the checkout owner without running their commands.
// Apparat entrypoints remain independently executable and retain their artifact history.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cjtrowbridge/apparat/internal/domain"
	"github.com/tailscale/hujson"
	"os"
	"strconv"
	"strings"
)

func (n *Runtime) workspaceCatalog(ctx context.Context, location string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	v, taskErr := n.projectTask(ctx, location, q)
	for i := range v.Choices {
		v.Choices[i].Label = "Apparat · " + v.Choices[i].Label
	}
	for _, file := range []string{"tasks", "launch"} {
		raw, err := n.workspaceFile(ctx, location, domain.WorkspaceRequest{Path: ".vscode/" + file + ".json", Action: "read"})
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				v.Report += "\nCould not read .vscode/" + file + ".json"
			}
			continue
		}
		data, err := hujson.Standardize([]byte(raw.Text))
		var spec struct {
			Tasks          []struct{ Label string }
			Configurations []struct{ Name string }
		}
		if err != nil || json.Unmarshal(data, &spec) != nil {
			v.Report += "\nInvalid .vscode/" + file + ".json"
			continue
		}
		for _, t := range spec.Tasks {
			if t.Label != "" {
				v.Choices = append(v.Choices, domain.WorkspaceChoice{Label: "VS Code task · " + t.Label, State: "definition", Detail: "This task requires a VS Code task executor."})
			}
		}
		for _, t := range spec.Configurations {
			if t.Name != "" {
				v.Choices = append(v.Choices, domain.WorkspaceChoice{Label: "Launch · " + t.Name, State: "definition", Detail: "This launch configuration requires its debugger and extensions."})
			}
		}
	}
	if taskErr != nil && len(v.Choices) == 0 {
		v.Report += "\n" + taskErr.Error()
	}
	start, err := strconv.Atoi(q.Argument)
	if q.Argument == "" {
		start = 0
		err = nil
	}
	if err != nil || start < 0 || start > len(v.Choices) {
		return domain.WorkspaceView{}, errors.New("catalog changed; reopen Tasks")
	}
	end := min(start+32, len(v.Choices))
	if end < len(v.Choices) {
		v.Cursor = strconv.Itoa(end)
	}
	v.Choices = v.Choices[start:end]
	v.Report = strings.TrimSpace(v.Report)
	return v, nil
}
