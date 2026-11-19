// Checks that cancelled or corrupt model downloads preserve the installed bytes.
// The cases also check temporary file cleanup after each failure.
package speech

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallIsAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "model")
	m := Model{Bytes: 4, SHA256: fmt.Sprintf("%x", sha256.Sum256([]byte("good")))}
	err := os.WriteFile(path, []byte("good"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"bad!", "extra", "cut", "cancel"} {
		ctx, cancel := context.WithCancel(context.Background())
		if input == "cancel" {
			cancel()
		}
		_, err := installStream(ctx, path, m, strings.NewReader(input), nil)
		cancel()
		if err == nil {
			t.Fatalf("accepted %q", input)
		}
		kept, err := os.ReadFile(path)
		files, listErr := os.ReadDir(dir)
		if err != nil || string(kept) != "good" || listErr != nil || len(files) != 1 {
			t.Fatal("previous model or temp cleanup failed")
		}
	}
}
