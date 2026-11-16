// Owns an OS-backed exclusive node lock and safe diagnostic metadata.
// The file remains after release; only the OS lock establishes ownership.
package platform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LockInfo struct {
	Root     string `json:"root"`
	PID      int    `json:"pid"`
	Artifact string `json:"artifact"`
	Mode     string `json:"mode"`
	Started  string `json:"started"`
}

type ContentionError struct{ Owner LockInfo }

func (e *ContentionError) Error() string {
	return fmt.Sprintf("node is already owned by pid=%d artifact=%s; stop that owner or choose another --runtime-dir", e.Owner.PID, e.Owner.Artifact)
}

type Lock struct {
	mu   sync.Mutex
	file *os.File
}

func Acquire(root, artifact, mode string) (*Lock, error) {
	err := os.MkdirAll(root, 0700)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(root, ".node.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	err = lockFile(file)
	if err != nil {
		var owner LockInfo
		_ = json.NewDecoder(file).Decode(&owner)
		_ = file.Close()
		if lockContended(err) {
			return nil, &ContentionError{Owner: owner}
		}
		return nil, err
	}
	info := LockInfo{Root: root, PID: os.Getpid(), Artifact: artifact, Mode: mode, Started: time.Now().UTC().Format(time.RFC3339Nano)}
	err = file.Truncate(0)
	if err == nil {
		err = json.NewEncoder(file).Encode(info)
	}
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Lock{file: file}, nil
}

func (lock *Lock) Close() error {
	lock.mu.Lock()
	defer lock.mu.Unlock()
	if lock.file == nil {
		return nil
	}
	err := unlockFile(lock.file)
	closeErr := lock.file.Close()
	lock.file = nil
	if err != nil {
		return err
	}
	return closeErr
}
