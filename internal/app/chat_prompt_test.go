// Checks byte-stable priming, legacy reconstruction and repository instruction selection.
// These fixtures exercise request construction without a model or persistent KV backend.
package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cjtrowbridge/apparat/internal/domain"
)

func TestChatPrimingPrefix(t *testing.T) {
	dir := t.TempDir()
	read := func() string {
		t.Helper()
		policy, err := readChatInstructions(dir)
		if err != nil {
			t.Fatal(err)
		}
		return policy
	}
	if read() != "" {
		t.Fatal("unexpected project instructions")
	}
	for _, name := range []string{"agents.md", "AGENTS.md"} {
		err := os.WriteFile(filepath.Join(dir, name), []byte("Follow "+name), 0600)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(read(), "Follow "+name) {
			t.Fatal("missing policy", name)
		}
	}
	session := domain.ChatSession{}
	build := func(route, job, policy string, legacy bool) domain.ChatRequest {
		t.Helper()
		p, err := chatPacket(domain.ChatAddress{}, route, job, session, policy, legacy)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	first := build("cluster", "first", read(), false)
	second := build("cluster", "second", read(), false)
	if first.System != second.System || !strings.HasSuffix(first.System, primeInstruction) ||
		!strings.Contains(first.System, "Follow AGENTS.md") || strings.Contains(first.System, "Follow agents.md") {
		t.Fatal("unstable or incomplete project prefix")
	}
	if first.Messages[0].Text != domain.ChatReadyReply || len(first.Tools) != len(chatTools) {
		t.Fatal("incomplete primed request")
	}
	legacy := build("one", "old", "", true)
	if len(legacy.Messages) != 0 || strings.Contains(legacy.System, primeInstruction) {
		t.Fatal("legacy request changed")
	}
}

// A changed digest requires an intentional cache-prefix migration.
func TestSharedChatPrefixFingerprint(t *testing.T) {
	p := domain.ChatRequest{Route: "cluster"}
	p.System = chatSystem(&p, "", false)
	prefix, err := json.Marshal([]any{p.System, p.Tools, domain.ChatReadyReply})
	if err != nil {
		t.Fatal(err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(prefix))
	if got != "fb75aab1540028f36d0d4b125cfe0df2ae0c87373f5f803bd4cf702f76d16b78" {
		t.Fatalf("shared prefix changed: %s", got)
	}
}
