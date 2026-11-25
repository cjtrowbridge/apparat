// Validates single-call terminal requests before the chat owner releases its model slot.
// An owner deadline limits the wait even if the initiating GUI never reconnects.
package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func terminalWait(job string, calls []domain.ToolCall, now time.Time) *domain.ChatTerminalWait {
	if len(calls) != 1 {
		return nil
	}
	c := calls[0]
	if c.ID == "" || len(c.ID) > 256 {
		return nil
	}
	for _, t := range chatTools {
		if t.name != c.Name || t.query != nil {
			continue
		}
		args, ok := t.arguments(c.Arguments)
		if !ok {
			return nil
		}
		wait := &domain.ChatTerminalWait{Job: job, CallID: c.ID, Deadline: now.Add(30 * time.Second)}
		if c.Name == "terminal_screen" {
			wait.Observe = true
			return wait
		}
		seconds, err := strconv.Atoi(args["timeout_seconds"])
		command := args["command"]
		if err != nil || seconds < 1 || seconds > 1800 || strings.TrimSpace(command) == "" || len(command) > 4096 || strings.ContainsRune(command, 0) {
			return nil
		}
		wait.Deadline = now.Add(time.Duration(seconds) * time.Second)
		wait.Command = command
		return wait
	}
	return nil
}
