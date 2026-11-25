// Owns private, host-side chat permission choices and command execution receipts.
// A tracked command gets its own terminal and cannot replay after a host restart.
package terminal

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func chatPermission(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "permission"))
	if os.IsNotExist(err) {
		return "read-only", nil
	}
	if err != nil {
		return "", err
	}
	if string(data) == "always-ask" || string(data) == "yolo" {
		return string(data), nil
	}
	return "read-only", nil
}

func setChatPermission(dir, mode string) error {
	if mode != "read-only" && mode != "always-ask" && mode != "yolo" {
		return errors.New("unsupported agent permission")
	}
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "permission"), []byte(mode), 0600)
}

type commandReceipt struct {
	Terminal string
	Command  string
	Deadline time.Time
}

func commandStatus(dir string, q domain.WorkspaceRequest) (domain.WorkspaceView, error) {
	var v domain.WorkspaceView
	data, err := os.ReadFile(filepath.Join(dir, "command.json"))
	if err != nil {
		return v, err
	}
	var receipt commandReceipt
	err = json.Unmarshal(data, &receipt)
	if err != nil || receipt.Terminal != q.RelatedRun || receipt.Command != q.Argument || !receipt.Deadline.Equal(q.Deadline) {
		return v, errors.New("command request changed")
	}
	screen, err := os.ReadFile(filepath.Join(dir, "screen"))
	if err == nil {
		v.Report = string(screen)
	}
	status, err := os.ReadFile(filepath.Join(dir, "exit"))
	if err == nil {
		v.Detail = string(status)
	} else if os.IsNotExist(err) {
		v.Detail = "Outcome unknown after interruption"
	} else {
		return v, err
	}
	return v, nil
}

func saveCommand(dir string, q domain.WorkspaceRequest) error {
	data, err := json.Marshal(commandReceipt{q.RelatedRun, q.Argument, q.Deadline})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "command.json"), data, 0600)
}

func selectionFile(chat, device string) string {
	return filepath.Join(chat, fmt.Sprintf("selected-%x", sha256.Sum256([]byte(device))))
}
