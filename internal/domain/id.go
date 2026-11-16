// Creates and validates canonical ULID identifiers without runtime services.
// Time and entropy are supplied explicitly for deterministic domain tests.
package domain

import (
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"strings"
	"time"
)

const idAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func ValidID(id string) bool {
	if len(id) != 26 || id[0] > '7' {
		return false
	}
	for _, ch := range id {
		if !strings.ContainsRune(idAlphabet, ch) {
			return false
		}
	}
	return true
}
func NewID(now time.Time, entropy io.Reader) (string, error) {
	millis := now.UnixMilli()
	if millis < 0 || millis > (1<<48)-1 {
		return "", errors.New("ULID time outside 48-bit range")
	}
	if entropy == nil {
		entropy = rand.Reader
	}
	var data [16]byte
	for index := 5; index >= 0; index-- {
		data[index] = byte(millis)
		millis >>= 8
	}
	_, err := io.ReadFull(entropy, data[6:])
	if err != nil {
		return "", err
	}
	number := new(big.Int).SetBytes(data[:])
	divisor := big.NewInt(32)
	remainder := new(big.Int)
	result := make([]byte, 26)
	for index := 25; index >= 0; index-- {
		number.QuoRem(number, divisor, remainder)
		result[index] = idAlphabet[remainder.Int64()]
	}
	return string(result), nil
}
