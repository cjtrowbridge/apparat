// Exercises strict identity parsing, key consistency and private-file permissions.
// Malformed inputs must fail before expensive derivation or identity readiness.
package identity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIdentityRejectsMalformedPrivateFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "key.json")
	for _, data := range []string{`{}`, `{"salt":"x","nonce":"x","ciphertext":"x"}`, `{"unknown":true}`, `{} {}`, `null`} {
		{
			err := os.WriteFile(path, []byte(data), 0600)
			if err != nil {
				t.Fatal(err)
			}
		}
		{
			_, err := LoadEncrypted(path, []byte("pass"))
			if err == nil {
				t.Fatal("accepted", data)
			}
		}
	}
	pair, _ := Generate()
	{
		err := SaveEncrypted(path, pair.Private, []byte("pass"))
		if err != nil {
			t.Fatal(err)
		}
	}
	data, _ := os.ReadFile(path)
	var payload map[string]string
	{
		err := json.Unmarshal(data, &payload)
		if err != nil {
			t.Fatal(err)
		}
	}
	payload["algorithm"] = "unsupported"
	bad, _ := json.Marshal(payload)
	{
		err := os.WriteFile(path, bad, 0600)
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		_, err := LoadEncrypted(path, []byte("pass"))
		if err == nil {
			t.Fatal("unsupported algorithm")
		}
	}
	if runtime.GOOS != "windows" {
		{
			err := os.WriteFile(path, data, 0600)
			if err != nil {
				t.Fatal(err)
			}
		}
		{
			err := os.Chmod(path, 0644)
			if err != nil {
				t.Fatal(err)
			}
		}
		{
			_, err := LoadEncrypted(path, []byte("pass"))
			if err == nil {
				t.Fatal("public private key accepted")
			}
		}
	}
	pair.Private[63] ^= 1
	if SaveEncrypted(filepath.Join(t.TempDir(), "key"), pair.Private, []byte("pass")) == nil {
		t.Fatal("inconsistent key accepted")
	}
}
func TestAtomicIdentityFailureCleansTemporary(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "occupied")
	{
		err := os.Mkdir(destination, 0700)
		if err != nil {
			t.Fatal(err)
		}
	}
	if atomicWrite(destination, []byte("private"), 0600) == nil {
		t.Fatal("directory overwritten")
	}
	files, err := filepath.Glob(filepath.Join(dir, ".identity-*"))
	if err != nil || len(files) != 0 {
		t.Fatal("temporary files remain", files, err)
	}
}
