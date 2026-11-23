// Exercises the native verification prompt contract without displaying or reading real identities.
// Verifies explicit consent on first contact and safe handling of an unavailable dialog.
package app

import (
	"encoding/json"
	"testing"
)

func TestWorkerFingerprintRequiresExplicitConsent(t *testing.T) {
	if approveTrust(nil, "worker", "compare fingerprints") {
		t.Fatal("accepted unavailable dialog")
	}
	for _, reply := range []string{"null", "invalid", `"Keep disconnected"`, `"Continue"`, `"I verified the fingerprint"`} {
		actual := approveTrust(func(request string) string {
			var r struct {
				Op, Title, Message string
				Options            []string
			}
			err := json.Unmarshal([]byte(request), &r)
			if err != nil || r.Op != "trust" || r.Title != "worker" || len(r.Options) != 2 || r.Options[0] != "Keep disconnected" {
				t.Fatal("unsafe trust prompt")
			}
			return reply
		}, "worker", "compare fingerprints")
		if actual != (reply == `"I verified the fingerprint"`) {
			t.Fatal("unexpected trust decision", reply)
		}
	}
}
