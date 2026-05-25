package security

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashString generates a SHA256 hash of the input string.
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
