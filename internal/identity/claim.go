// Computes the stable URL-safe SHA-256 identifier used by the cluster protocol.
// Inputs are domain-separated by their callers.
package identity

import (
	"crypto/sha256"
	"encoding/base64"
)

func ClaimDigest(invitation string) string {
	hash := sha256.Sum256([]byte(invitation))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
