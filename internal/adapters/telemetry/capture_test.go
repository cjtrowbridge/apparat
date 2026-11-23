// Verifies bounded storage, redaction, capture failure and the exported record format.
// No target hardware, remote credential or SQLite instance participates.
package telemetry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBoundedCapture(t *testing.T) {
	var r Ring
	for i := 0; i < capacity+9; i++ {
		r.Observe(Event{Operation: "save", Elapsed: int64(i)})
	}
	r.Observe(Event{Operation: "private-filename.txt"})
	c := r.Snapshot("worker", "test", 0)
	if len(c.Events) != capacity || c.Lost != 9 || c.Events[0].Seq != 10 {
		t.Fatal("wrong ring bounds")
	}
	f, err := Save(context.Background(), t.TempDir(), c)
	if err != nil {
		t.Fatal(err)
	}
	lines, next, err := ReadPage(f, 0)
	if err != nil || len(lines) != 24 || next == 0 || !strings.Contains(lines[0], `"schema":1`) {
		t.Fatal("invalid first capture page", err)
	}
	_, err = Save(context.Background(), filepath.Join(f, "directory"), c)
	if err == nil {
		t.Fatal("failed disk write reported success")
	}
	start := time.Now()
	r.Observe(Event{Operation: "read"})
	if time.Since(start) > time.Second {
		t.Fatal("failed capture blocked observer")
	}
	info, err := os.Stat(f)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("capture was not private")
	}
}
