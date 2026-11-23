// Checks that a large directory streams in bounded pages tied to the caller.
// No subtree traversal or Git checkout is required to enumerate entries.
package projectgit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryPaging(t *testing.T) {
	p := DirectoryPages{}
	defer p.Close()
	dir := t.TempDir()
	for i := 0; i < 2100; i++ {
		err := os.WriteFile(filepath.Join(dir, fmt.Sprint(i)), nil, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	seen := map[string]bool{}
	token := ""
	for {
		v, err := p.Page("device/project", dir, ".", token)
		if err != nil || len(v.Entries) > 64 {
			t.Fatal("page", err)
		}
		for _, entry := range v.Entries {
			if seen[entry.Name] {
				t.Fatal("duplicate entry")
			}
			seen[entry.Name] = true
		}
		token = v.Cursor
		if token == "" {
			break
		}
		_, err = p.Page("other", dir, ".", token)
		if err == nil {
			t.Fatal("cursor crossed caller scope")
		}
	}
	if len(seen) != 2100 {
		t.Fatal(len(seen))
	}
}
