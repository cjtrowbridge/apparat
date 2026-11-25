// Delivers bounded command exit receipts even after the user leaves the chat.
// Recovery queries the original execution host and never submits another command.
package app

import (
	"context"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func (n *Runtime) reconcileTerminalCallbacks(ctx context.Context) {
	user := n.PlanningUser()
	if user == "" {
		return
	}
	items, err := n.database.TerminalCallbacks(ctx, user)
	if err != nil {
		return
	}
	for _, q := range items {
		if ctx.Err() != nil || n.PlanningUser() != user {
			return
		}
		n.finishTerminalCallback(ctx, user, q)
	}
}
func (n *Runtime) finishTerminalCallback(parent context.Context, user string, q domain.WorkspaceRequest) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	q.Action = "chat-terminal-result"
	q.Approved = false
	message := "Wait expired. Inspect the terminal for its current process state."
	failed := true
	if time.Now().Before(q.Deadline) {
		v, err := n.ChatTerminal(ctx, q)
		if err != nil || v.Detail == "Running" {
			return
		}
		message = v.Detail + "\n" + v.Report
		if len(message) > 4096 {
			message = strings.ToValidUTF8(message[:4088], "") + "…"
		}
		failed = v.Detail != "Exited 0"
	}
	if n.PlanningUser() != user {
		return
	}
	var err error
	if q.ProjectID == "" {
		base := "ok"
		if failed {
			base = "failed"
		}
		err = n.SavePlanning(ctx, user, "global-terminal-result", domain.GlobalRequest{Session: q.Thread, Turn: q.ID, Input: message, Base: base})
	} else {
		_, err = n.Workspace(ctx, domain.WorkspaceRequest{Action: "agent-terminal-result", ProjectID: q.ProjectID, Thread: q.Thread, ID: q.ID, Path: ".", Approved: !failed, Argument: message})
	}
	if err == nil {
		_ = n.database.ForgetTerminalCallback(ctx, user, q.ID)
	}
}
